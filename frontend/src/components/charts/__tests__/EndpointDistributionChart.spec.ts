import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import EndpointDistributionChart from '../EndpointDistributionChart.vue'

const messages: Record<string, string> = {
  'usage.endpointDistribution': 'Endpoint Distribution',
  'usage.endpoint': 'Endpoint',
  'usage.inbound': 'Inbound',
  'usage.upstream': 'Upstream',
  'usage.path': 'Path',
  'admin.dashboard.requests': 'Requests',
  'admin.dashboard.tokens': 'Tokens',
  'admin.dashboard.actual': 'Actual',
  'admin.dashboard.standard': 'Standard',
  'admin.dashboard.metricTokens': 'By Tokens',
  'admin.dashboard.metricActualCost': 'By Actual Cost',
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
  Doughnut: {
    props: ['data'],
    template: '<div class="chart-data">{{ JSON.stringify(data) }}</div>',
  },
}))

describe('EndpointDistributionChart', () => {
  it('renders the amber Orbital route matrix with its aggregate', () => {
    const wrapper = mount(EndpointDistributionChart, {
      props: {
        telemetry: true,
        endpointStats: [
          { endpoint: '/v1/responses', requests: 8, total_tokens: 1800, cost: 0.8, actual_cost: 0.7 },
          { endpoint: '/v1/messages', requests: 4, total_tokens: 700, cost: 0.4, actual_cost: 0.3 },
        ],
      },
      global: {
        stubs: {
          LoadingSpinner: true,
        },
      },
    })

    expect(wrapper.get('[data-testid="telemetry-distribution-card"]').classes())
      .toContain('telemetry-distribution-card--amber')
    expect(wrapper.get('[data-testid="telemetry-distribution-meta"]').text())
      .toContain('矩阵 // 03')
    expect(wrapper.get('[data-testid="telemetry-donut-total"]').text()).toBe('2.50K')
    expect(wrapper.get('[data-testid="telemetry-donut-active"]').text()).toContain('2')
    expect(wrapper.get('[data-testid="telemetry-distribution-accent"]')).toBeTruthy()
    expect(wrapper.get('[data-testid="telemetry-distribution-footer"]')).toBeTruthy()
    expect(wrapper.findAll('.telemetry-series-swatch').length).toBeGreaterThan(0)
  })
})
