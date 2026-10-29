package service

import (
	"context"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// DiagnoseModelAvailabilityForPlatform reports whether the requested model
// is configured to be served by any persistently eligible OpenAI-compatible
// account in the group for the given platform (e.g. PlatformOpenAI,
// PlatformGrok). The platform scopes the candidate pool so distinct
// OpenAI-compatible platforms do not cross-contaminate diagnosis results.
// The query bypasses scheduler snapshots and ignores transient runtime state.
//
// Safe to call on the error path: returns {true,true} on any internal
// failure or when the inputs preclude meaningful diagnosis (empty model,
// nil service), so callers stay on the 503 fallback branch.
func (s *OpenAIGatewayService) DiagnoseModelAvailabilityForPlatform(
	ctx context.Context,
	groupID *int64,
	requestedModel string,
	platform string,
) ModelAvailabilityDiagnosis {
	if s == nil {
		return ModelAvailabilityDiagnosis{HasAccountsInPool: true, HasModelSupport: true}
	}
	requestedModel = strings.TrimSpace(requestedModel)
	if requestedModel == "" {
		return ModelAvailabilityDiagnosis{HasAccountsInPool: true, HasModelSupport: true}
	}
	if s.accountRepo == nil {
		return ModelAvailabilityDiagnosis{HasAccountsInPool: true, HasModelSupport: true}
	}

	platform = NormalizeOpenAICompatiblePlatform(platform)
	queryGroupID := groupID
	includeGrouped := false
	if s.cfg != nil && s.cfg.RunMode == config.RunModeSimple {
		queryGroupID = nil
		includeGrouped = true
	}
	accounts, err := s.accountRepo.ListModelAvailabilityCandidates(
		ctx,
		queryGroupID,
		[]string{platform},
		includeGrouped,
	)
	if err != nil {
		// Conservative fallback so the caller keeps returning 503; we do not
		// want a transient lookup failure to flip into 404 model_not_found.
		return ModelAvailabilityDiagnosis{HasAccountsInPool: true, HasModelSupport: true}
	}

	diag := ModelAvailabilityDiagnosis{}
	now := time.Now()
	supporting := 0
	rateLimited := 0
	var earliestReset time.Time
	for i := range accounts {
		diag.HasAccountsInPool = true
		if isRetiredOpenAISubscriptionModel(&accounts[i], resolveOpenAIAccountUpstreamModelForRequest(&accounts[i], requestedModel, false)) {
			diag.HasRetiredSubscriptionModel = true
		}
		// Mirrors the per-candidate filter used during account selection
		// (openai_account_scheduler.isAccountRequestCompatible): empty
		// model_mapping accepts everything; otherwise the explicit / wildcard
		// mapping must match.
		if !accounts[i].IsModelSupported(requestedModel) {
			continue
		}
		diag.HasModelSupport = true
		supporting++
		reset, limited := rateLimitWindowForDiagnosis(ctx, &accounts[i], requestedModel, now)
		if !limited {
			continue
		}
		rateLimited++
		if earliestReset.IsZero() || reset.Before(earliestReset) {
			earliestReset = reset
		}
	}
	// Claim "every account serving this model is rate-limited" only when the
	// whole supporting set sits inside a recorded window: that is the case where
	// retrying now cannot succeed and the client should be told to wait.
	if supporting > 0 && rateLimited == supporting {
		diag.AllModelAccountsRateLimited = true
		diag.NextRateLimitResetAt = earliestReset
	}
	return diag
}

// rateLimitWindowForDiagnosis reports whether the account currently sits inside a
// recorded rate-limit window that applies to requestedModel, and when that window
// reopens. Two kinds of window exist and they do not have the same reach:
//   - RateLimitResetAt is account-wide, so it blocks every model the account can serve.
//   - The model_rate_limits snapshot (written by SetModelRateLimit) is scoped to the
//     models its key names, so a window recorded for a sibling model must not be
//     reported as an exhausted pool for this one.
//
// Only windows that actually block this request are counted; anything else stays a
// plain "no available accounts" 503 so the client keeps retrying.
func rateLimitWindowForDiagnosis(
	ctx context.Context,
	account *Account,
	requestedModel string,
	now time.Time,
) (time.Time, bool) {
	if account == nil {
		return time.Time{}, false
	}
	if account.RateLimitResetAt != nil && now.Before(*account.RateLimitResetAt) {
		return *account.RateLimitResetAt, true
	}
	if remaining := account.GetModelRateLimitRemainingTimeWithContext(ctx, requestedModel); remaining > 0 {
		return now.Add(remaining), true
	}
	return time.Time{}, false
}
