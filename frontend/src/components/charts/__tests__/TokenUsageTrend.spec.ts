import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import TokenUsageTrend from '../TokenUsageTrend.vue'

const messages: Record<string, string> = {
  'admin.dashboard.tokenUsageTrend': 'Token Usage Trend',
  'admin.dashboard.noDataAvailable': 'No data available',
}

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => messages[key] ?? key,
    }),
  }
})

vi.mock('vue-chartjs', () => ({
  Line: {
    props: ['data', 'options'],
    template: '<div class="chart-data" :data-options="JSON.stringify(options)">{{ JSON.stringify(data) }}</div>',
  },
}))

const point = {
  date: '2026-09-18 08:00',
  requests: 3,
  input_tokens: 800,
  output_tokens: 200,
  cache_creation_tokens: 100,
  cache_read_tokens: 900,
  cost: 0.02,
  actual_cost: 0.01,
}

function mountChart(trendData = [point], telemetry = false) {
  return mount(TokenUsageTrend, {
    props: { trendData, telemetry },
    global: {
      stubs: { LoadingSpinner: true },
    },
  })
}

describe('TokenUsageTrend', () => {
  it('calculates cache hit rate against all prompt tokens', () => {
    const wrapper = mountChart([
      { ...point, input_tokens: 500, cache_creation_tokens: 0, cache_read_tokens: 1500 },
    ])
    const chartData = JSON.parse(wrapper.get('.chart-data').text())
    const hitRateDataset = chartData.datasets.find((dataset: any) => dataset.label === 'Cache Hit Rate')

    expect(hitRateDataset.data[0]).toBe(75)
  })

  it('returns zero hit rate when all prompt tokens are zero', () => {
    const wrapper = mountChart([
      {
        ...point,
        requests: 0,
        input_tokens: 0,
        output_tokens: 0,
        cache_creation_tokens: 0,
        cache_read_tokens: 0,
        cost: 0,
        actual_cost: 0,
      },
    ])
    const chartData = JSON.parse(wrapper.get('.chart-data').text())
    const hitRateDataset = chartData.datasets.find((dataset: any) => dataset.label === 'Cache Hit Rate')

    expect(hitRateDataset.data[0]).toBe(0)
  })

  it('includes cache creation tokens in the hit-rate denominator without plotting a separate series', () => {
    const wrapper = mountChart([
      { ...point, input_tokens: 200, output_tokens: 50, cache_creation_tokens: 300, cache_read_tokens: 500 },
    ])
    const chartData = JSON.parse(wrapper.get('.chart-data').text())
    const hitRateDataset = chartData.datasets.find((dataset: any) => dataset.label === 'Cache Hit Rate')

    expect(hitRateDataset.data[0]).toBe(50)
    expect(chartData.datasets.map((dataset: any) => dataset.label)).toEqual([
      'Input',
      'Output',
      'Cache Read',
      'Cache Hit Rate',
    ])
  })

  it('renders the design-file telemetry hierarchy and four compact metric chips', () => {
    const wrapper = mountChart([point], true)

    expect(wrapper.get('[data-testid="telemetry-trend-card"]').classes()).toContain('telemetry-trend-card')
    expect(wrapper.get('[data-testid="telemetry-trend-accent"]').exists()).toBe(true)
    expect(wrapper.get('[data-testid="telemetry-trend-meta"]').text()).toContain('矩阵 // 04 · 实时流量')
    expect(wrapper.findAll('[data-testid="telemetry-trend-metric"]')).toHaveLength(4)
    expect(wrapper.findAll('[data-testid="telemetry-trend-metric"]').map((metric) => metric.text())).toEqual([
      '输入800',
      '输出200',
      '缓存读取900',
      '命中率50.0%',
    ])
    expect(wrapper.get('[data-testid="telemetry-trend-footer"]').text()).toContain('流式指标：24小时粒度')
    expect(wrapper.get('[data-testid="telemetry-cache-efficiency"]').text()).toContain('缓存效率：50.0%')

    const options = JSON.parse(wrapper.get('.chart-data').attributes('data-options') || '{}')
    expect(options.plugins.legend.display).toBe(false)
  })
})
