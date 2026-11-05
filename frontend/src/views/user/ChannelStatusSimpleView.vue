<template>
  <AppLayout>
    <div class="space-y-6 pb-12">
      <section class="card sticky top-0 z-20 !rounded-3xl !border-0 p-0 shadow-sm ring-1 ring-gray-900/5 backdrop-blur-sm dark:!bg-dark-800 dark:ring-dark-700 supports-[backdrop-filter]:bg-white/95 dark:supports-[backdrop-filter]:bg-dark-800/95">
        <header class="flex flex-wrap items-center justify-between gap-4 px-5 py-4 sm:px-6">
          <div class="min-w-0">
            <h1 class="flex items-center gap-2 text-xl font-black text-gray-900 dark:text-white">
              <span class="inline-flex h-8 w-8 items-center justify-center rounded-xl bg-blue-50 text-blue-500 dark:bg-blue-900/30 dark:text-blue-400">
                <Icon name="chart" size="sm" />
              </span>
              渠道状态
            </h1>
            <div class="mt-1.5 flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400">
              <span class="relative flex h-2 w-2 shrink-0">
                <span class="relative inline-flex h-2 w-2 rounded-full" :class="loading ? 'bg-gray-400' : 'bg-green-500'"></span>
              </span>
              <span v-if="updatedAt">{{ updatedAt }}</span>
              <span v-else>{{ t('common.loading') }}</span>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <div class="tabs inline-flex" role="group">
              <button
                v-for="option in ranges"
                :key="option.value"
                type="button"
                class="tab !px-2 !py-1 text-xs"
                :class="range === option.value ? 'tab-active' : ''"
                @click="setRange(option.value)"
              >
                {{ option.label }}
              </button>
            </div>
            <button
              class="flex h-8 w-8 items-center justify-center rounded-lg bg-gray-100 text-gray-500 hover:bg-gray-200 dark:bg-dark-700 dark:text-gray-400 dark:hover:bg-dark-600"
              type="button"
              :disabled="loading"
              :title="t('common.refresh')"
              @click="load()"
            >
              <Icon name="refresh" size="sm" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </header>
      </section>

      <div class="space-y-6">
        <section v-for="section in sections" :key="section.key" class="space-y-3">
          <div class="flex items-center gap-2">
            <h2 class="text-sm font-semibold text-gray-900 dark:text-white">{{ section.title }}</h2>
            <span class="rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-500 dark:bg-dark-700 dark:text-dark-300">{{ section.rows.length }}</span>
          </div>
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="row in section.rows"
          :key="rowKey(row)"
          class="cursor-pointer rounded-3xl border border-gray-100 bg-white p-5 shadow-sm ring-1 ring-gray-900/5 transition hover:shadow-md dark:border-dark-700 dark:bg-dark-800 dark:ring-dark-700"
          @click="openDetail(row)"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="flex min-w-0 items-center gap-2.5">
              <span class="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gray-50 text-gray-700 dark:bg-dark-700 dark:text-dark-200">
                <ProviderIcon :provider="row.platform" :size="22" />
              </span>
              <div class="min-w-0">
                <p class="truncate text-sm font-semibold text-gray-900 dark:text-white">{{ row.group_name || row.platform }}</p>
                <p class="mt-0.5 flex items-center gap-1.5 text-xs">
                  <span class="inline-flex items-center rounded-md bg-gray-100 px-1.5 py-0.5 text-gray-500 dark:bg-dark-700 dark:text-dark-300">{{ row.platform }}</span>
                  <span class="truncate text-gray-500 dark:text-dark-400">{{ row.model === '__other__' ? '其他模型' : row.model }}</span>
                  <span v-if="groupRate(row) !== null" class="shrink-0 font-medium text-rose-500">倍率 {{ groupRate(row) }}x</span>
                </p>
              </div>
            </div>
            <span :class="['shrink-0 rounded-full px-2 py-0.5 text-xs font-medium', statusClass(row.health.overall)]">
              {{ statusLabel(row.health.overall) }}
            </span>
          </div>

          <dl class="mt-4 grid grid-cols-3 gap-2">
            <div>
              <dt class="text-xs text-gray-500 dark:text-dark-400">缓存率</dt>
              <dd class="mt-0.5 text-lg font-semibold text-gray-900 dark:text-white">{{ formatPercent(row.metrics.cache_rate) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-gray-500 dark:text-dark-400">可用率</dt>
              <dd class="mt-0.5 text-lg font-semibold text-emerald-600 dark:text-emerald-400">{{ formatPercent(1 - row.metrics.error_rate) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-gray-500 dark:text-dark-400">首 TOKEN</dt>
              <dd class="mt-0.5 text-lg font-semibold text-gray-900 dark:text-white">{{ formatMs(row.metrics.ttft.p50_ms) }}</dd>
            </div>
          </dl>

          <div class="mt-4 flex items-end gap-[2px]">
            <i
              v-for="(slot, index) in alignedBuckets(row)"
              :key="index"
              class="h-3.5 flex-1 rounded-[2px]"
              :class="slot.bucket ? bucketClass(slot.bucket.health) : 'bg-gray-200 dark:bg-dark-600'"
              :title="slot.bucket ? bucketTooltip(slot.bucket) : `${formatBucketTime(slot.start)} · 无流量`"
            ></i>
          </div>
          <div class="mt-1.5 flex justify-between text-[10px] text-gray-400 dark:text-dark-500">
            <span>较早</span>
            <span>现在</span>
          </div>

          <div v-if="iqRounds(row).length" class="mt-3 border-t border-gray-100 pt-3 dark:border-dark-700">
            <div class="flex items-center justify-between text-xs">
              <span class="text-gray-500 dark:text-dark-400">降智检测 · 近 {{ iqRounds(row).length }} 次</span>
              <span :class="['rounded-full px-2 py-0.5 font-medium', iqStatusClass(iqLatest(row)?.status)]">
                {{ iqStatusText(iqLatest(row)?.status) || '暂无数据' }}
              </span>
            </div>
            <div class="mt-2 flex items-end gap-[2px]">
              <i
                v-for="(round, index) in iqRounds(row).slice(0, 24).reverse()"
                :key="index"
                class="h-2.5 flex-1 rounded-[2px]"
                :class="iqDotClass(round.status)"
                :title="iqRoundTooltip(round)"
              ></i>
            </div>
            <div class="mt-2 max-h-20 overflow-auto pr-1">
              <div class="grid grid-cols-4 gap-1">
                <span
                  v-for="round in iqRounds(row).slice(0, 30)"
                  :key="round.round_id"
                  class="inline-flex items-center gap-1 rounded-full border border-gray-100 px-1.5 py-0.5 text-[10px] text-gray-500 dark:border-dark-700 dark:text-dark-400"
                  :title="iqRoundTooltip(round)"
                >
                  <i class="h-1.5 w-1.5 shrink-0 rounded-full" :class="iqDotClass(round.status)"></i>
                  {{ formatBucketTime(round.created_at).slice(6) }}
                </span>
              </div>
            </div>
          </div>
        </article>

          </div>
        </section>
        <div v-if="!loading && !rows.length" class="rounded-3xl border border-dashed border-gray-200 p-10 text-center text-sm text-gray-400 dark:border-dark-700">
          暂无渠道监控数据
        </div>
      </div>
    </div>

    <div v-if="detail" class="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" @click.self="closeDetail()">
      <section class="w-full max-w-3xl rounded-3xl bg-white p-6 shadow-xl dark:bg-dark-800">
        <header class="flex items-start justify-between gap-4">
          <div>
            <h2 class="text-lg font-bold text-gray-900 dark:text-white">{{ detail.group_name || detail.platform }}</h2>
            <p class="mt-1 text-xs text-gray-500 dark:text-dark-400">
              {{ detail.platform }} · {{ detail.model === '__other__' ? '其他模型' : detail.model }}
            </p>
          </div>
          <button class="text-2xl leading-none text-gray-400 hover:text-gray-600 dark:hover:text-gray-200" type="button" @click="closeDetail()">×</button>
        </header>

        <div class="mt-5 overflow-x-auto">
          <table class="w-full text-sm">
            <thead>
              <tr class="border-b border-gray-100 text-left text-xs text-gray-500 dark:border-dark-700 dark:text-dark-400">
                <th class="pb-2 pr-3 font-medium">模型</th>
                <th class="pb-2 pr-3 font-medium">最新状态</th>
                <th class="pb-2 pr-3 font-medium">最新延迟 (ms)</th>
                <th class="pb-2 pr-3 font-medium">7 天可用率</th>
                <th class="pb-2 pr-3 font-medium">30 天可用率</th>
                <th class="pb-2 font-medium">7 天平均延迟 (ms)</th>
              </tr>
            </thead>
            <tbody>
              <tr class="text-gray-900 dark:text-white">
                <td class="py-3 pr-3 font-semibold">{{ detail.model === '__other__' ? '其他模型' : detail.model }}</td>
                <td class="py-3 pr-3">
                  <span :class="['rounded-full px-2 py-0.5 text-xs font-medium', statusClass(detail.health.overall)]">
                    {{ statusLabel(detail.health.overall) }}
                  </span>
                </td>
                <td class="py-3 pr-3">{{ detailLoading ? '…' : detailStats.latency || '-' }}</td>
                <td class="py-3 pr-3">{{ detailLoading ? '…' : detailStats.avail7 || '-' }}</td>
                <td class="py-3 pr-3">{{ detailLoading ? '…' : detailStats.avail30 || '-' }}</td>
                <td class="py-3">{{ detailLoading ? '…' : detailStats.avgLat7 || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <footer class="mt-6 flex justify-end">
          <button class="rounded-xl border border-gray-200 px-4 py-2 text-sm text-gray-700 hover:bg-gray-50 dark:border-dark-600 dark:text-dark-200 dark:hover:bg-dark-700" type="button" @click="closeDetail()">关闭</button>
        </footer>
      </section>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import Icon from '@/components/icons/Icon.vue'
import ProviderIcon from '@/components/user/monitor/ProviderIcon.vue'
import { useAuthStore } from '@/stores/auth'
import { userGroupsAPI } from '@/api/groups'
import * as api from '@/api/channelMonitorV2'
import type { MonitorCoverage, MonitorMatrixBucket, MonitorMatrixRow, MonitorRange } from '@/api/channelMonitorV2'
import { formatMonitorMs, formatMonitorPercent } from '@/features/channel-monitor-v2/monitorFormat'

const { t } = useI18n()
const authStore = useAuthStore()
const isAdmin = computed(() => authStore.isAdmin)

const ranges: Array<{ value: MonitorRange; label: string }> = [
  { value: '90m', label: '90m' },
  { value: '24h', label: '24h' },
  { value: '7d', label: '7d' },
  { value: '30d', label: '30d' },
]
const range = ref<MonitorRange>('90m')
const rows = ref<MonitorMatrixRow[]>([])
const coverage = ref<MonitorCoverage | null>(null)
const loading = ref(false)
const updatedAt = ref('')

const sections = computed(() => {
  const workbuddy = rows.value.filter((row) => (row.group_name || '').startsWith('wb2ds'))
  const bus = rows.value.filter((row) => (row.group_name || '').startsWith('巴士'))
  const other = rows.value.filter((row) => !workbuddy.includes(row) && !bus.includes(row))
  return [
    { key: 'workbuddy', title: 'WorkBuddy', rows: workbuddy },
    { key: 'bus', title: '巴士 Codex', rows: bus },
    { key: 'other', title: '其他渠道', rows: other },
  ].filter((section) => section.rows.length > 0)
})

const rateByGroup = ref<Record<number, number>>({})

async function loadGroupRates() {
  try {
    const [groups, overrides] = await Promise.all([
      userGroupsAPI.getAvailable(),
      userGroupsAPI.getUserGroupRates(),
    ])
    const rates: Record<number, number> = {}
    for (const group of groups) {
      if (typeof group.rate_multiplier === 'number') {
        rates[group.id] = overrides[group.id] ?? group.rate_multiplier
      }
    }
    rateByGroup.value = rates
  } catch {
    // 倍率取不到时隐藏，不影响监控数据。
  }
}

function groupRate(row: MonitorMatrixRow): number | null {
  if (!row.group_id) return null
  const rate = rateByGroup.value[row.group_id]
  return typeof rate === 'number' ? parseFloat(rate.toFixed(2)) : null
}

interface IqRound {
  round_id: string
  status: string
  created_at: string
}

interface IqTarget {
  model: string
  group_id: number
  rounds: IqRound[]
}

const IQ_API = window.location.hostname === '127.0.0.1' || window.location.hostname === 'localhost'
  ? 'http://127.0.0.1:8090'
  : 'https://iq.bbs2dpinyin.com'
const IQ_STATUS: Record<string, string> = { pass: '通过', warn: '需关注', fail: '请求失败' }
const iqTargets = ref<Record<string, IqTarget>>({})

async function loadIq() {
  try {
    const res = await fetch(`${IQ_API}/api/summary`)
    if (!res.ok) return
    const data = (await res.json()) as { targets?: IqTarget[] }
    const map: Record<string, IqTarget> = {}
    for (const target of data.targets || []) {
      map[`${target.model}:${target.group_id}`] = target
    }
    iqTargets.value = map
  } catch {
    // 检测服务不可用时不影响渠道状态展示
  }
}

function iqRounds(row: MonitorMatrixRow): IqRound[] {
  return iqTargets.value[`${row.model}:${row.group_id}`]?.rounds || []
}

function iqLatest(row: MonitorMatrixRow): IqRound | null {
  return iqRounds(row)[0] || null
}

function iqStatusText(status: string | undefined): string {
  return status ? IQ_STATUS[status] || status : ''
}

function iqStatusClass(status: string | undefined): string {
  switch (status) {
    case 'pass':
      return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'warn':
      return 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
    case 'fail':
      return 'bg-rose-50 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400'
    default:
      return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-dark-300'
  }
}

function iqDotClass(status: string | undefined): string {
  switch (status) {
    case 'pass':
      return 'bg-emerald-500'
    case 'warn':
      return 'bg-amber-400'
    case 'fail':
      return 'bg-rose-500'
    default:
      return 'bg-gray-300 dark:bg-dark-500'
  }
}

function iqRoundTooltip(round: IqRound): string {
  return `${formatBucketTime(round.created_at)} · ${iqStatusText(round.status)}`
}

const detail = ref<MonitorMatrixRow | null>(null)
const detailLoading = ref(false)
const detailStats = ref<{ latency: string | null; avail7: string | null; avail30: string | null; avgLat7: string | null }>({
  latency: null,
  avail7: null,
  avail30: null,
  avgLat7: null,
})

let controller: AbortController | null = null

function emptyFilter(r: MonitorRange) {
  return { range: r, platforms: [], groupIds: [], models: [] }
}

const STATUS_TEXT: Record<string, string> = {
  healthy: '正常',
  warning: '降级',
  critical: '异常',
  unknown: '未知',
}

function statusLabel(state: string | undefined | null): string {
  return STATUS_TEXT[state || 'unknown'] || '未知'
}

function statusClass(state: string | undefined | null): string {
  switch (state) {
    case 'healthy':
      return 'bg-emerald-50 text-emerald-600 dark:bg-emerald-900/30 dark:text-emerald-400'
    case 'warning':
      return 'bg-amber-50 text-amber-600 dark:bg-amber-900/30 dark:text-amber-400'
    case 'critical':
      return 'bg-rose-50 text-rose-600 dark:bg-rose-900/30 dark:text-rose-400'
    default:
      return 'bg-gray-100 text-gray-500 dark:bg-dark-700 dark:text-dark-300'
  }
}

function formatPercent(value: number | null | undefined): string {
  return formatMonitorPercent(value ?? 0)
}

function formatMs(value: number | null | undefined): string {
  return formatMonitorMs(value ?? null)
}

function formatBucketTime(value: string | undefined): string {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return ''
  return `${String(d.getMonth() + 1).padStart(2, '0')}/${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
}

interface BucketSlot {
  start: string
  bucket?: MonitorMatrixBucket
}

/** 按覆盖时间生成固定时间轴，缺数据的桶留空（灰色）。 */
function alignedBuckets(row: MonitorMatrixRow): BucketSlot[] {
  const cov = coverage.value
  if (!cov) return row.buckets.map((bucket) => ({ start: bucket.bucket_start, bucket }))
  const step = Math.max(60, cov.bucket_seconds) * 1000
  const start = new Date(cov.requested_start).getTime()
  const endRaw = cov.requested_end ? new Date(cov.requested_end).getTime() : NaN
  const end = Number.isFinite(endRaw) && endRaw > start ? endRaw : new Date(cov.data_through).getTime()
  if (!Number.isFinite(start) || !Number.isFinite(end) || start >= end || step <= 0) {
    return row.buckets.map((bucket) => ({ start: bucket.bucket_start, bucket }))
  }
  const byStart = new Map<number, MonitorMatrixBucket>()
  for (const bucket of row.buckets) {
    byStart.set(new Date(bucket.bucket_start).getTime(), bucket)
  }
  const slots: BucketSlot[] = []
  for (let cursor = Math.floor(start / step) * step; cursor < end; cursor += step) {
    slots.push({
      start: new Date(cursor).toISOString(),
      bucket: byStart.get(cursor),
    })
  }
  return slots.slice(-24)
}

function bucketTooltip(bucket: MonitorMatrixBucket): string {
  const time = formatBucketTime(bucket.bucket_start)
  const avail = formatPercent(1 - bucket.metrics.error_rate)
  const cache = formatPercent(bucket.metrics.cache_rate)
  const ttft = formatMs(bucket.metrics.ttft.p50_ms)
  return `${time} · 可用率 ${avail} · 缓存率 ${cache} · 首 Token ${ttft}`
}

function bucketClass(health: { overall?: string }): string {
  switch (health?.overall) {
    case 'healthy':
      return 'bg-emerald-500'
    case 'warning':
      return 'bg-amber-400'
    case 'critical':
      return 'bg-rose-500'
    default:
      return 'bg-gray-300 dark:bg-dark-500'
  }
}

function rowKey(row: MonitorMatrixRow): string {
  return [row.platform, row.group_id || 0, row.model || ''].join(':')
}

async function load() {
  controller?.abort()
  const request = new AbortController()
  controller = request
  loading.value = true
  try {
    const res = await api.getMatrix(emptyFilter(range.value), 'platform_group_model', isAdmin.value, request.signal)
    coverage.value = res.coverage || null
    // 非管理员接口会把请求数等流量指标脱敏为 0，用 buckets 是否有数据判断有效行。
    rows.value = (res.items || []).filter(
      (row) =>
        row.model &&
        row.model !== 'unknown' &&
        row.model !== '__other__' &&
        (row.buckets.length > 0 || row.metrics.request_count > 0),
    )
    updatedAt.value = res.coverage?.data_through
      ? `更新至 ${formatBucketTime(res.coverage.data_through)}`
      : ''
  } catch (error) {
    if ((error as { name?: string }).name !== 'CanceledError') {
      // 保持上次数据
    }
  } finally {
    if (!request.signal.aborted) loading.value = false
  }
}

function setRange(value: MonitorRange) {
  if (range.value === value) return
  range.value = value
  void load()
}

async function openDetail(row: MonitorMatrixRow) {
  detail.value = row
  detailLoading.value = true
  detailStats.value = { latency: null, avail7: null, avail30: null, avgLat7: null }
  try {
    const [res7, res30] = await Promise.all([
      api.getMatrix(emptyFilter('7d'), 'platform_group_model', isAdmin.value),
      api.getMatrix(emptyFilter('30d'), 'platform_group_model', isAdmin.value),
    ])
    const pick = (res: typeof res7) =>
      (res.items || []).find(
        (item) =>
          item.platform === row.platform &&
          (item.group_id || 0) === (row.group_id || 0) &&
          item.model === row.model,
      )
    const row7 = pick(res7)
    const row30 = pick(res30)
    detailStats.value = {
      latency: formatMs(row.metrics.duration?.avg_ms ?? row.metrics.ttft.p50_ms),
      avail7: row7 ? formatPercent(1 - row7.metrics.error_rate) : null,
      avail30: row30 ? formatPercent(1 - row30.metrics.error_rate) : null,
      avgLat7: row7 ? formatMs(row7.metrics.duration?.avg_ms) : null,
    }
  } finally {
    detailLoading.value = false
  }
}

function closeDetail() {
  detail.value = null
}

let timer: number | null = null

onMounted(() => {
  void load()
  void loadGroupRates()
  void loadIq()
  timer = window.setInterval(() => {
    void load()
    void loadIq()
  }, 60_000)
})

onBeforeUnmount(() => {
  controller?.abort()
  if (timer !== null) window.clearInterval(timer)
})
</script>
