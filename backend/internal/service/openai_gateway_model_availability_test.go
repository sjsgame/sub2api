//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func ptrInt64ForAvailabilityTest(v int64) *int64 { return &v }

// A pool whose every model-capable account is inside a recorded rate-limit
// window must report that explicitly, so handlers can answer 429 with a reset
// time instead of a generic 503 that clients retry blindly.
func TestDiagnoseModelAvailability_RateLimitedPool(t *testing.T) {
	resetAt := time.Now().Add(30 * time.Minute)
	repo := codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		7: {{
			ID:               3,
			Platform:         PlatformOpenAI,
			Type:             AccountTypeOAuth,
			Status:           StatusActive,
			Schedulable:      true,
			RateLimitResetAt: &resetAt,
		}},
	}}
	svc := &OpenAIGatewayService{accountRepo: repo, cfg: &config.Config{}}

	diag := svc.DiagnoseModelAvailabilityForPlatform(
		context.Background(), ptrInt64ForAvailabilityTest(7), "gpt-6.1-terra", PlatformOpenAI)

	require.True(t, diag.HasAccountsInPool)
	require.True(t, diag.HasModelSupport)
	require.True(t, diag.AllModelAccountsRateLimited)
	require.WithinDuration(t, resetAt, diag.NextRateLimitResetAt, time.Second)
}

// One account outside its window means a retry can still succeed, so the
// diagnosis must not claim the whole pool is rate-limited.
func TestDiagnoseModelAvailability_PartialRateLimitIsNotExhausted(t *testing.T) {
	resetAt := time.Now().Add(30 * time.Minute)
	repo := codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		7: {
			{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true, RateLimitResetAt: &resetAt},
			{ID: 4, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true},
		},
	}}
	svc := &OpenAIGatewayService{accountRepo: repo, cfg: &config.Config{}}

	diag := svc.DiagnoseModelAvailabilityForPlatform(
		context.Background(), ptrInt64ForAvailabilityTest(7), "gpt-6.1-terra", PlatformOpenAI)

	require.True(t, diag.HasModelSupport)
	require.False(t, diag.AllModelAccountsRateLimited)
	require.True(t, diag.NextRateLimitResetAt.IsZero())
}

// An expired window is not a limit: the account is schedulable again.
func TestDiagnoseModelAvailability_ExpiredRateLimitWindowIsIgnored(t *testing.T) {
	resetAt := time.Now().Add(-time.Minute)
	repo := codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{
		7: {{
			ID:               3,
			Platform:         PlatformOpenAI,
			Type:             AccountTypeOAuth,
			Status:           StatusActive,
			Schedulable:      true,
			RateLimitResetAt: &resetAt,
		}},
	}}
	svc := &OpenAIGatewayService{accountRepo: repo, cfg: &config.Config{}}

	diag := svc.DiagnoseModelAvailabilityForPlatform(
		context.Background(), ptrInt64ForAvailabilityTest(7), "gpt-6.1-terra", PlatformOpenAI)

	require.False(t, diag.AllModelAccountsRateLimited)
}
