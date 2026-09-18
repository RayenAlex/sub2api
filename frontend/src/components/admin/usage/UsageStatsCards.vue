<template>
  <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
    <article class="card telemetry-kpi telemetry-kpi--blue">
      <div>
        <p class="telemetry-kpi__label">{{ t('usage.totalRequests') }}</p>
        <p class="telemetry-kpi__value">{{ stats?.total_requests?.toLocaleString() || '0' }}</p>
      </div>
      <div class="telemetry-kpi__icon" aria-hidden="true">
        <Icon name="document" size="md" />
      </div>
      <div class="telemetry-kpi__footer">
        <span>{{ t('usage.inSelectedRange') }}</span>
        <span class="font-mono font-semibold tabular-nums">{{ stats?.total_requests?.toLocaleString() || '0' }}</span>
      </div>
    </article>

    <article class="card telemetry-kpi telemetry-kpi--emerald">
      <div>
        <p class="telemetry-kpi__label">{{ t('usage.totalTokens') }}</p>
        <p class="telemetry-kpi__value">{{ formatTokens(stats?.total_tokens || 0) }}</p>
      </div>
      <div class="telemetry-kpi__icon" aria-hidden="true">
        <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9"
          />
        </svg>
      </div>
      <div class="telemetry-kpi__footer flex-wrap justify-start gap-x-2 gap-y-1">
        <span>{{ t('usage.in') }}: <strong class="font-mono text-gray-700 dark:text-dark-200">{{ formatTokens(stats?.total_input_tokens || 0) }}</strong></span>
        <span aria-hidden="true">/</span>
        <span>{{ t('usage.out') }}: <strong class="font-mono text-gray-700 dark:text-dark-200">{{ formatTokens(stats?.total_output_tokens || 0) }}</strong></span>
        <span aria-hidden="true">/</span>
        <span class="group relative inline-flex cursor-help items-center gap-0.5" tabindex="0">
          <span>{{ cacheLabel() }}: <strong class="font-mono text-gray-700 dark:text-dark-200">{{ formatTokens(stats?.total_cache_tokens || 0) }}</strong></span>
          <svg class="h-3.5 w-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
            <path stroke-linecap="round" stroke-linejoin="round" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <span class="pointer-events-none absolute bottom-full left-1/2 z-30 mb-2 hidden w-56 -translate-x-1/2 rounded-xl border border-surface-container bg-white p-3 text-left text-xs text-gray-700 shadow-lg group-hover:block group-focus:block dark:border-dark-600 dark:bg-dark-800 dark:text-dark-200">
            <span class="mb-2 block font-semibold text-gray-900 dark:text-white">{{ cacheDetailLabel() }}</span>
            <span class="flex items-center justify-between gap-3">
              <span>{{ t('usage.cacheCreationTokensLabel') }}</span>
              <span class="font-mono tabular-nums">{{ formatTokens(stats?.total_cache_creation_tokens || 0) }}</span>
            </span>
            <span class="mt-1 flex items-center justify-between gap-3">
              <span>{{ t('usage.cacheReadTokensLabel') }}</span>
              <span class="font-mono tabular-nums">{{ formatTokens(stats?.total_cache_read_tokens || 0) }}</span>
            </span>
          </span>
        </span>
      </div>
    </article>

    <article class="card telemetry-kpi telemetry-kpi--red">
      <div class="min-w-0">
        <p class="telemetry-kpi__label">{{ t('usage.totalCost') }}</p>
        <p class="telemetry-kpi__value">${{ (stats?.total_actual_cost || 0).toFixed(4) }}</p>
      </div>
      <div class="telemetry-kpi__icon" aria-hidden="true">
        <Icon name="dollar" size="md" />
      </div>
      <div class="telemetry-kpi__footer flex-wrap justify-start gap-x-3 gap-y-1">
        <span v-if="showAccountCost && totalAccountCost !== null">
          {{ t('usage.accountCost') }}
          <strong class="font-mono text-gray-700 dark:text-dark-200">${{ totalAccountCost.toFixed(4) }}</strong>
        </span>
        <span>
          {{ t('usage.standardCost') }}
          <strong class="font-mono text-gray-700 dark:text-dark-200" :class="{ 'line-through': strikeStandardCost }">${{ (stats?.total_cost || 0).toFixed(4) }}</strong>
        </span>
      </div>
    </article>

    <article class="card telemetry-kpi telemetry-kpi--violet">
      <div>
        <p class="telemetry-kpi__label">{{ t('usage.avgDuration') }}</p>
        <p class="telemetry-kpi__value">{{ formatDuration(stats?.average_duration_ms || 0) }}</p>
      </div>
      <div class="telemetry-kpi__icon" aria-hidden="true">
        <Icon name="clock" size="md" />
      </div>
      <div class="telemetry-kpi__footer">
        <span>{{ t('usage.inSelectedRange') }}</span>
        <span class="font-mono font-semibold">LATENCY</span>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminUsageStatsResponse } from '@/api/admin/usage'
import type { UsageStatsResponse } from '@/types'
import Icon from '@/components/icons/Icon.vue'

const props = withDefaults(defineProps<{
  stats: (AdminUsageStatsResponse | UsageStatsResponse) | null
  showAccountCost?: boolean
  strikeStandardCost?: boolean
}>(), {
  showAccountCost: true,
  strikeStandardCost: false,
})

const { t } = useI18n()

const totalAccountCost = computed(() => {
  const stats = props.stats as (AdminUsageStatsResponse & { total_account_cost?: number }) | null
  return stats?.total_account_cost ?? null
})
const showAccountCost = computed(() => props.showAccountCost)
const strikeStandardCost = computed(() => props.strikeStandardCost)

const formatDuration = (ms: number) =>
  ms < 1000 ? `${ms.toFixed(0)}ms` : `${(ms / 1000).toFixed(2)}s`

const formatTokens = (value: number) => {
  if (value >= 1e9) return (value / 1e9).toFixed(2) + 'B'
  if (value >= 1e6) return (value / 1e6).toFixed(2) + 'M'
  if (value >= 1e3) return (value / 1e3).toFixed(2) + 'K'
  return value.toLocaleString()
}

const cacheLabel = () => t('usage.cacheTotal')
const cacheDetailLabel = () => t('usage.cacheBreakdown')
</script>
