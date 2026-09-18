<template>
  <div
    v-if="telemetry"
    data-testid="telemetry-trend-card"
    class="telemetry-trend-card"
  >
    <div data-testid="telemetry-trend-accent" class="telemetry-trend-accent" aria-hidden="true"></div>

    <header class="telemetry-trend-header">
      <div class="telemetry-trend-title-row">
        <span class="telemetry-trend-title-dot" aria-hidden="true"></span>
        <h3>{{ t('admin.dashboard.tokenUsageTrend') }}</h3>
        <span data-testid="telemetry-trend-meta" class="telemetry-trend-meta">
          [矩阵 // 04 · 实时流量]
        </span>
      </div>

      <div data-testid="telemetry-trend-metrics" class="telemetry-trend-metrics">
        <div
          v-for="metric in telemetryMetrics"
          :key="metric.label"
          data-testid="telemetry-trend-metric"
          :class="['telemetry-trend-metric', { 'telemetry-trend-metric--rate': metric.rate }]"
        >
          <span
            :class="['telemetry-trend-metric-dot', { 'telemetry-trend-metric-dot--pulse': metric.rate }]"
            :style="{ backgroundColor: metric.color }"
            aria-hidden="true"
          ></span>
          <span class="telemetry-trend-metric-label">{{ metric.label }}</span>
          <strong>{{ metric.value }}</strong>
        </div>
      </div>
    </header>

    <div class="telemetry-trend-chart-body">
      <div v-if="loading" class="telemetry-trend-state">
        <LoadingSpinner />
      </div>
      <div v-else-if="trendData.length > 0 && chartData" class="telemetry-trend-canvas">
        <Line :data="chartData" :options="lineOptions" />
      </div>
      <div v-else class="telemetry-trend-state telemetry-trend-empty">
        {{ t('admin.dashboard.noDataAvailable') }}
      </div>
    </div>

    <footer data-testid="telemetry-trend-footer" class="telemetry-trend-footer">
      <span>流式指标：24小时粒度</span>
      <strong data-testid="telemetry-cache-efficiency">
        缓存效率：{{ telemetryCacheEfficiency }}
      </strong>
    </footer>
  </div>

  <div v-else data-testid="telemetry-trend-card" class="card p-4">
    <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
      {{ t('admin.dashboard.tokenUsageTrend') }}
    </h3>
    <div v-if="loading" class="flex h-48 items-center justify-center">
      <LoadingSpinner />
    </div>
    <div v-else-if="trendData.length > 0 && chartData" class="h-48">
      <Line :data="chartData" :options="lineOptions" />
    </div>
    <div v-else class="flex h-48 items-center justify-center text-sm text-gray-500 dark:text-gray-400">
      {{ t('admin.dashboard.noDataAvailable') }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler,
} from 'chart.js'
import { Line } from 'vue-chartjs'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import type { TrendDataPoint } from '@/types'

ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const { t } = useI18n()

const props = defineProps<{
  trendData: TrendDataPoint[]
  loading?: boolean
  telemetry?: boolean
}>()

const isDarkMode = computed(() => document.documentElement.classList.contains('dark'))

const chartColors = computed(() => ({
  text: isDarkMode.value ? '#c9c7d0' : '#76737c',
  grid: isDarkMode.value ? '#383741' : '#eeedf3',
  baseline: isDarkMode.value ? '#4b4956' : '#e3e2e7',
  input: '#2855e8',
  output: '#4db987',
  cacheRead: '#efa934',
  cacheHitRate: '#8b5cf6',
}))

const totals = computed(() => props.trendData.reduce(
  (result, item) => {
    result.input += item.input_tokens || 0
    result.output += item.output_tokens || 0
    result.cacheRead += item.cache_read_tokens || 0
    result.prompt +=
      (item.input_tokens || 0) + (item.cache_read_tokens || 0) + (item.cache_creation_tokens || 0)
    return result
  },
  { input: 0, output: 0, cacheRead: 0, prompt: 0 }
))

const telemetryCacheEfficiency = computed(() => {
  const efficiency = totals.value.prompt > 0
    ? (totals.value.cacheRead / totals.value.prompt) * 100
    : 0
  return `${efficiency.toFixed(1)}%`
})

const telemetryMetrics = computed(() => [
  { label: '输入', value: formatTokens(totals.value.input), color: chartColors.value.input },
  { label: '输出', value: formatTokens(totals.value.output), color: chartColors.value.output },
  { label: '缓存读取', value: formatTokens(totals.value.cacheRead), color: chartColors.value.cacheRead },
  {
    label: '命中率',
    value: telemetryCacheEfficiency.value,
    color: chartColors.value.cacheHitRate,
    rate: true,
  },
])

const chartData = computed(() => {
  if (!props.trendData?.length) return null

  const sharedLineStyle = {
    borderWidth: 2.25,
    pointRadius: 0,
    pointHoverRadius: 4,
    pointHoverBorderWidth: 2,
    tension: 0.42,
  }

  return {
    labels: props.trendData.map((item) => formatTimeLabel(item.date)),
    datasets: [
      {
        ...sharedLineStyle,
        label: 'Input',
        data: props.trendData.map((item) => item.input_tokens),
        borderColor: chartColors.value.input,
        backgroundColor: `${chartColors.value.input}16`,
        pointHoverBackgroundColor: chartColors.value.input,
        fill: true,
      },
      {
        ...sharedLineStyle,
        label: 'Output',
        data: props.trendData.map((item) => item.output_tokens),
        borderColor: chartColors.value.output,
        backgroundColor: 'transparent',
        pointHoverBackgroundColor: chartColors.value.output,
        fill: false,
      },
      {
        ...sharedLineStyle,
        label: 'Cache Read',
        data: props.trendData.map((item) => item.cache_read_tokens),
        borderColor: chartColors.value.cacheRead,
        backgroundColor: 'transparent',
        pointHoverBackgroundColor: chartColors.value.cacheRead,
        fill: false,
      },
      {
        ...sharedLineStyle,
        label: 'Cache Hit Rate',
        data: props.trendData.map((item) => {
          const promptTokens =
            (item.input_tokens || 0) + (item.cache_read_tokens || 0) + (item.cache_creation_tokens || 0)
          return promptTokens > 0 ? ((item.cache_read_tokens || 0) / promptTokens) * 100 : 0
        }),
        borderColor: chartColors.value.cacheHitRate,
        backgroundColor: 'transparent',
        pointHoverBackgroundColor: chartColors.value.cacheHitRate,
        borderDash: [4, 3],
        fill: false,
        yAxisID: 'yPercent',
      },
    ],
  }
})

const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  normalized: true,
  interaction: {
    intersect: false,
    mode: 'index' as const,
  },
  layout: {
    padding: { top: 2, right: 2, bottom: 0, left: 0 },
  },
  plugins: {
    legend: {
      display: false,
    },
    tooltip: {
      displayColors: true,
      usePointStyle: true,
      backgroundColor: isDarkMode.value ? '#1f1f24' : '#ffffff',
      titleColor: isDarkMode.value ? '#ffffff' : '#111116',
      bodyColor: isDarkMode.value ? '#d6d4dc' : '#76737c',
      borderColor: isDarkMode.value ? '#494751' : '#eeedf3',
      borderWidth: 1,
      cornerRadius: 14,
      padding: 14,
      caretSize: 0,
      titleFont: { family: 'ui-monospace, SFMono-Regular, Menlo, monospace', size: 12, weight: 'bold' as const },
      bodyFont: { family: 'ui-monospace, SFMono-Regular, Menlo, monospace', size: 11 },
      callbacks: {
        title: (items: any[]) => props.trendData[items[0]?.dataIndex]?.date || '',
        label: (context: any) => context.dataset.yAxisID === 'yPercent'
          ? ` 命中率：${Number(context.raw).toFixed(1)}%`
          : ` ${translateDatasetLabel(context.dataset.label)}：${formatTokens(Number(context.raw))}`,
        footer: () => '',
      },
    },
  },
  scales: {
    x: {
      border: { color: chartColors.value.baseline },
      grid: { display: false },
      ticks: {
        color: chartColors.value.text,
        maxRotation: 0,
        minRotation: 0,
        autoSkip: true,
        maxTicksLimit: 7,
        padding: 10,
        font: { family: 'ui-monospace, SFMono-Regular, Menlo, monospace', size: 10 },
      },
    },
    y: {
      beginAtZero: true,
      border: { display: false },
      grid: {
        color: chartColors.value.grid,
        borderDash: [3, 3],
        drawTicks: false,
      },
      ticks: {
        color: chartColors.value.text,
        maxTicksLimit: 5,
        padding: 8,
        callback: (value: string | number) => formatTokens(Number(value)),
        font: { family: 'ui-monospace, SFMono-Regular, Menlo, monospace', size: 9 },
      },
    },
    yPercent: {
      position: 'right' as const,
      min: 0,
      max: 100,
      border: { display: false },
      grid: { drawOnChartArea: false, drawTicks: false },
      ticks: {
        color: chartColors.value.cacheHitRate,
        stepSize: 25,
        padding: 8,
        callback: (value: string | number) => `${value}%`,
        font: { family: 'ui-monospace, SFMono-Regular, Menlo, monospace', size: 9 },
      },
    },
  },
}))

function formatTokens(value: number): string {
  const absolute = Math.abs(value)
  if (absolute >= 1_000_000_000) return compactNumber(value / 1_000_000_000, 'B')
  if (absolute >= 1_000_000) return compactNumber(value / 1_000_000, 'M')
  if (absolute >= 1_000) return compactNumber(value / 1_000, 'K')
  return value.toLocaleString('en-US')
}

function compactNumber(value: number, suffix: string): string {
  const rounded = value.toFixed(1).replace(/\.0$/, '')
  return `${rounded}${suffix}`
}

function formatTimeLabel(value: string): string {
  const timeMatch = value.match(/(?:T|\s)(\d{2}:\d{2})/)
  return timeMatch?.[1] || value
}

function translateDatasetLabel(label: string): string {
  return ({ Input: '输入', Output: '输出', 'Cache Read': '缓存读取' } as Record<string, string>)[label] || label
}
</script>
