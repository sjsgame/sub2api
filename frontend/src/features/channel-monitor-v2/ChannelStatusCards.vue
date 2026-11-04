<template>
  <section class="card !rounded-3xl !border-0 p-4 shadow-sm ring-1 ring-gray-900/5 dark:!bg-dark-800 dark:ring-dark-700 sm:p-5">
    <div v-if="rows.length" class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
      <article
        v-for="row in rows"
        :key="rowKey(row)"
        class="rounded-2xl border border-gray-100 bg-white p-4 dark:border-dark-700 dark:bg-dark-900/40"
      >
        <header class="flex items-start justify-between gap-3">
          <div class="flex min-w-0 items-center gap-2.5">
            <span class="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-gray-50 text-gray-700 dark:bg-dark-800 dark:text-dark-200">
              <ProviderIcon :provider="row.platform" :size="20" />
            </span>
            <div class="min-w-0">
              <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ rowLabel(row) }}</p>
              <p class="mt-0.5 flex items-center gap-1.5 text-xs">
                <span class="inline-flex items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-gray-500 dark:bg-dark-700 dark:text-dark-300">
                  {{ row.platform }}
                </span>
                <span v-if="rateMultiplier(row) !== null" class="font-medium text-rose-500">
                  {{ t('channelMonitorV2.cards.userRate') }} {{ rateMultiplier(row) }}x
                </span>
              </p>
            </div>
          </div>
          <span
            :class="['status-dot shrink-0', cellClass(row.health, row.metrics.request_count)]"
            :title="t('channelMonitorV2.cards.healthDot')"
            aria-hidden="true"
          ></span>
        </header>

        <dl class="mt-4 grid grid-cols-2 gap-3">
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('channelMonitorV2.metrics.successRate') }}</dt>
            <dd class="mt-0.5 text-lg font-semibold text-gray-900 dark:text-white">
              {{ formatMonitorSuccessRateFromError(row.metrics.error_rate) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('channelMonitorV2.metrics.ttftP50') }}</dt>
            <dd class="mt-0.5 text-lg font-semibold text-gray-900 dark:text-white">
              {{ formatMonitorMs(row.metrics.ttft.p50_ms) }}
            </dd>
          </div>
          <div v-if="showThroughput">
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('channelMonitorV2.metrics.tps') }}</dt>
            <dd class="mt-0.5 text-lg font-semibold text-gray-900 dark:text-white">
              {{ formatMonitorTokensPerSecond(row.metrics.tpm) }}
            </dd>
          </div>
          <div>
            <dt class="text-xs text-gray-500 dark:text-dark-400">{{ t('channelMonitorV2.metrics.cacheRate') }}</dt>
            <dd class="mt-0.5 text-lg font-semibold text-gray-900 dark:text-white">
              {{ formatMonitorPercent(row.metrics.cache_rate) }}
            </dd>
          </div>
        </dl>

        <div class="mt-4">
          <div class="flex items-end gap-[2px]" role="img" :aria-label="t('channelMonitorV2.cards.timeline')">
            <i
              v-for="(slot, index) in tailBuckets(row)"
              :key="index"
              class="h-4 flex-1 rounded-[2px]"
              :class="slot ? cellClass(slot.health, slot.metrics.request_count) : 'health-unknown'"
              :title="slot ? bucketTooltip(slot) : t('channelMonitorV2.cards.noTraffic')"
            ></i>
          </div>
          <div class="mt-1.5 flex justify-between text-[10px] text-gray-400 dark:text-dark-500">
            <span>{{ t('channelMonitorV2.cards.past') }}</span>
            <span>{{ t('channelMonitorV2.cards.now') }}</span>
          </div>
        </div>
      </article>
    </div>
    <EmptyState v-else :title="t('channelMonitorV2.cards.empty')" />
  </section>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import type {
  MonitorHealth,
  MonitorMatrixBucket,
  MonitorMatrixRow,
} from '@/api/channelMonitorV2'
import EmptyState from '@/components/common/EmptyState.vue'
import ProviderIcon from '@/components/user/monitor/ProviderIcon.vue'
import {
  formatMonitorMs,
  formatMonitorPercent,
  formatMonitorSuccessRateFromError,
  formatMonitorTokensPerSecond,
  healthScoreClass,
} from '@/features/channel-monitor-v2/monitorFormat'

type HealthMode = 'overall' | 'success' | 'ttft' | 'cache'

const props = withDefaults(
  defineProps<{
    rows: MonitorMatrixRow[]
    healthMode: HealthMode
    showThroughput?: boolean
    /** group_id -> 分组倍率（用户实际看到的倍率）。 */
    rateByGroup?: Record<number, number>
  }>(),
  { showThroughput: true, rateByGroup: () => ({}) },
)

const { t } = useI18n()

/** 卡片时间线只展示最近的桶，和 Kedaya 的「近 N 次」观感一致。 */
const TIMELINE_BUCKETS = 24

function rowKey(row: MonitorMatrixRow): string {
  return [row.platform, row.group_id || 0, row.model || ''].join(':')
}

function rowLabel(row: MonitorMatrixRow): string {
  return row.group_name || row.model || row.platform
}

function rateMultiplier(row: MonitorMatrixRow): string | null {
  if (!row.group_id) return null
  const rate = props.rateByGroup[row.group_id]
  if (typeof rate !== 'number' || !Number.isFinite(rate)) return null
  return parseFloat(rate.toFixed(2)).toString()
}

function tailBuckets(row: MonitorMatrixRow): Array<MonitorMatrixBucket | null> {
  return row.buckets.slice(-TIMELINE_BUCKETS)
}

function cellClass(health: MonitorHealth, requestCount: number): string {
  return healthScoreClass(health, props.healthMode, requestCount)
}

function bucketTooltip(bucket: MonitorMatrixBucket): string {
  const lines = [
    formatMonitorSuccessRateFromError(bucket.metrics.error_rate),
    formatMonitorMs(bucket.metrics.ttft.p50_ms),
    formatMonitorPercent(bucket.metrics.cache_rate),
  ]
  return lines.join(' · ')
}
</script>

<style scoped>
.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 9999px;
  display: inline-block;
}
.health-score10 { background: #16a34a; }
.health-score9 { background: #22c55e; }
.health-score8 { background: #4ade80; }
.health-score7 { background: #a3e635; }
.health-score6 { background: #facc15; }
.health-score5 { background: #fbbf24; }
.health-score4 { background: #f59e0b; }
.health-score3 { background: #f97316; }
.health-score2 { background: #fb7185; }
.health-score1 { background: #f87171; }
.health-score0 { background: rgb(239, 67, 67); }
.health-healthy { background: #22c55e; }
.health-warning { background: #facc15; }
.health-critical { background: #fb7185; }
.health-unknown { background: #d1d5db; }
</style>
