const ipGeoMocks = vi.hoisted(() => ({
  getEntry: vi.fn(() => ({ status: 'idle' as const })),
  fetchOne: vi.fn(),
  fetchBatch: vi.fn(),
}))

const appStoreMocks = vi.hoisted(() => ({
  showSuccess: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/utils/ipGeoLookup', () => ipGeoMocks)
vi.mock('@/stores/app', () => ({ useAppStore: () => appStoreMocks }))

import { describe, expect, it, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'

import UsageTable from '../UsageTable.vue'

const messages: Record<string, string> = {
  'admin.usage.userDeletedBadge': 'Deleted',
  'usage.costDetails': 'Cost Breakdown',
  'admin.usage.inputCost': 'Input Cost',
  'admin.usage.outputCost': 'Output Cost',
  'admin.usage.cacheCreationCost': 'Cache Creation Cost',
  'admin.usage.cacheReadCost': 'Cache Read Cost',
  'usage.inputTokenPrice': 'Input price',
  'usage.outputTokenPrice': 'Output price',
  'usage.perMillionTokens': '/ 1M tokens',
  'usage.serviceTier': 'Service tier',
  'usage.serviceTierPriority': 'Fast',
  'usage.serviceTierUltrafast': 'Ultrafast',
  'usage.serviceTierFlex': 'Flex',
  'usage.serviceTierStandard': 'Standard',
  'usage.rate': 'Rate',
  'usage.accountMultiplier': 'Account rate',
  'usage.original': 'Original',
  'usage.userBilled': 'User billed',
  'usage.accountBilled': 'Account billed',
  'usage.imageUnit': ' images',
  'usage.imageCount': 'Image count',
  'usage.imageBillingSize': 'Billing size',
  'usage.imageInputSize': 'Input size',
  'usage.imageOutputSize': 'Output size',
  'usage.imageSizeSource': 'Size source',
  'usage.imageSizeBreakdown': 'Size breakdown',
  'usage.imageSizeSourceOutput': 'Upstream output',
  'usage.imageSizeSourceInput': 'Request input',
  'usage.imageSizeSourceDefault': 'Default billing tier',
  'usage.imageSizeSourceLegacy': 'Legacy record',
  'usage.imageSizeSourceMissing': 'Not recorded',
  'usage.imageSizeNotRecorded': 'not recorded',
  'usage.imageSizeLegacyUnstandardized': 'legacy unstandardized',
  'usage.imageSizeUnknown': 'unknown',
  'usage.imageUnitPrice': 'Per-image price',
  'usage.imageTotalPrice': 'Image total price',
  'usage.stream': 'Stream',
  'usage.sync': 'Sync',
  'usage.nativeCompactionV2': 'Compaction',
  'admin.usage.billingModeToken': 'Token',
  'admin.usage.billingModePerRequest': 'Per request',
  'admin.usage.billingModeImage': 'Image',
	'admin.usage.requestIdCopied': 'Request ID copied',
	'admin.usage.upstreamRequestIdCopied': 'Upstream ID copied',
	'keys.copied': 'Copied',
	'keys.copyToClipboard': 'Copy to clipboard',
	'common.copyFailed': 'Copy failed',
	'usage.requestedModel': 'Requested',
	'usage.sentUpstreamModel': 'Sent upstream',
	'usage.upstreamResponseModel': 'Upstream response',
	'usage.modelVariant': 'Possible version variant',
	'usage.modelMismatch': 'Different model',
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

const DataTableStub = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.request_id">
        <slot name="cell-model" :row="row" :value="row.model" />
        <slot name="cell-reasoning_effort" :row="row" :value="row.reasoning_effort" />
        <slot name="cell-billing_mode" :row="row" />
        <slot name="cell-tokens" :row="row" />
        <slot name="cell-cache_hit_rate" :row="row" />
        <slot name="cell-cost" :row="row" />
        <slot name="cell-request_id" :row="row" />
        <slot name="cell-upstream_request_id" :row="row" />
      </div>
    </div>
  `,
}

const baseImageRow = {
  request_id: 'req-admin-image',
  model: 'gpt-image-2',
  actual_cost: 0.4,
  total_cost: 0.4,
  account_rate_multiplier: 1,
  rate_multiplier: 1,
  service_tier: null,
  input_cost: 0,
  output_cost: 0,
  cache_creation_cost: 0,
  cache_read_cost: 0,
  input_tokens: 0,
  output_tokens: 0,
  cache_creation_tokens: 0,
  cache_read_tokens: 0,
  cache_creation_5m_tokens: 0,
  cache_creation_1h_tokens: 0,
  cache_ttl_overridden: false,
  billing_mode: 'image',
  image_count: 2,
  image_size: '2K',
  image_input_size: null,
  image_output_size: null,
  image_size_source: null,
  image_size_breakdown: null,
}

describe('admin UsageTable tooltip', () => {
  beforeEach(() => {
    vi.spyOn(HTMLElement.prototype, 'getBoundingClientRect').mockReturnValue({
      x: 0,
      y: 0,
      top: 20,
      left: 20,
      right: 120,
      bottom: 40,
      width: 100,
      height: 20,
      toJSON: () => ({}),
    } as DOMRect)
  })

  it('shows each token request cache-hit rate using the billing token buckets', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-cache-hit',
            billing_mode: 'token',
            input_tokens: 100,
            cache_creation_tokens: 100,
            cache_read_tokens: 800,
          },
          {
            ...baseImageRow,
            request_id: 'req-no-prompt-tokens',
            billing_mode: 'token',
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const rates = wrapper.findAll('[data-testid="cache-hit-rate"]')
    expect(rates).toHaveLength(2)
    expect(rates[0].text()).toContain('80.0%')
    expect(rates[0].classes()).toContain('items-center')
    expect(rates[1].text()).toBe('-')
  })

  it('renders long-context multiplier in context usage without a cost hover trigger', () => {
    const DataTableStubWithUsageCells = {
      props: ['data'],
      template: `
        <div>
          <div v-for="row in data" :key="row.request_id">
            <div data-testid="tokens-cell"><slot name="cell-tokens" :row="row" /></div>
            <div data-testid="cost-cell"><slot name="cell-cost" :row="row" /></div>
          </div>
        </div>
      `,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-long-context-enabled',
            billing_mode: 'token',
            input_tokens: 1_969,
            output_tokens: 622,
            cache_read_tokens: 321_900,
            long_context_billing_applied: true,
          },
          {
            ...baseImageRow,
            request_id: 'req-long-context-disabled',
            billing_mode: 'token',
            input_tokens: 1_969,
            output_tokens: 622,
            cache_read_tokens: 321_900,
            long_context_billing_applied: false,
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStubWithUsageCells,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const tokenCells = wrapper.findAll('[data-testid="tokens-cell"]')
    const costCells = wrapper.findAll('[data-testid="cost-cell"]')

    expect(tokenCells[0].get('[data-testid="long-context-billing-marker"]').text()).toBe('x2')
    expect(tokenCells[1].find('[data-testid="long-context-billing-marker"]').exists()).toBe(false)
    expect(costCells.every((cell) => !cell.find('[data-testid="long-context-billing-marker"]').exists())).toBe(true)
    expect(costCells.every((cell) => !cell.find('.group.relative').exists())).toBe(true)
  })

  it('marks only token rows charged with a peak multiplier', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-peak-token',
            billing_mode: 'token',
            input_tokens: 855,
            output_tokens: 104,
            applied_peak_multiplier: 3,
          },
          {
            ...baseImageRow,
            request_id: 'req-normal-token',
            billing_mode: 'token',
            input_tokens: 855,
            output_tokens: 104,
            applied_peak_multiplier: 1,
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const markers = wrapper.findAll('[data-testid="peak-token-quota-marker"]')
    expect(markers).toHaveLength(1)
    expect(markers[0].text()).toBe('×3')
  })

it('keeps the request type badge and adds a separate badge only for native compaction rows', () => {
    const DataTableStreamStub = {
      props: ['data'],
      template: `
        <div>
          <div v-for="row in data" :key="row.request_id">
            <slot name="cell-stream" :row="row" />
          </div>
        </div>
      `,
    }
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-compaction-stream',
            request_type: 'stream',
            stream: true,
            native_compaction_v2: true,
          },
          {
            ...baseImageRow,
            request_id: 'req-historical-sync',
            request_type: 'sync',
            stream: false,
            native_compaction_v2: false,
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStreamStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const requestBadges = wrapper.findAll('[data-testid="request-type-badge"]')
    expect(requestBadges).toHaveLength(2)
    expect(requestBadges[0].text()).toBe('Stream')
    expect(requestBadges[1].text()).toBe('Sync')
    expect(wrapper.findAll('[data-testid="native-compaction-badge"]')).toHaveLength(1)
    expect(wrapper.get('[data-testid="native-compaction-badge"]').text()).toBe('Compaction')
  })

  it('shows requested and upstream models separately for admin rows', () => {
    const row = {
      request_id: 'req-admin-model-1',
      model: 'claude-sonnet-4',
      upstream_model: 'claude-sonnet-4-20250514',
      actual_cost: 0,
      total_cost: 0,
      account_rate_multiplier: 1,
      rate_multiplier: 1,
      input_cost: 0,
      output_cost: 0,
      cache_creation_cost: 0,
      cache_read_cost: 0,
      input_tokens: 0,
      output_tokens: 0,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('claude-sonnet-4')
    expect(text).toContain('claude-sonnet-4-20250514')
  })

  it('shows requested and forwarded reasoning effort separately when they differ', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-admin-effort-1',
          model: 'gpt-5.4',
          reasoning_effort: 'max',
          upstream_reasoning_effort: 'xhigh',
        }],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('Max')
    expect(text).toContain('XHigh')
    expect(text).toContain('↳')
  })

  it('shows a single reasoning effort when requested matches forwarded', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-admin-effort-2',
          model: 'gpt-5.6-sol',
          reasoning_effort: 'max',
        }],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('Max')
    expect(text).not.toContain('↳')
  })

  it('hides mapped reasoning effort for user rows that only have the requested value', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{
          request_id: 'req-user-effort-1',
          model: 'gpt-5.4',
          reasoning_effort: 'max',
        }],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).toContain('Max')
    expect(wrapper.text()).not.toContain('XHigh')
    expect(wrapper.text()).not.toContain('↳')
  })

	it.each([
		{
			name: 'possible version variant',
			responseModel: 'gpt-5.5-2026-08-01',
			expectedBadge: 'Possible version variant',
		},
		{
			name: 'different upstream model',
			responseModel: 'gpt-5.4',
			expectedBadge: 'Different model',
		},
	])('shows a compact upstream response audit marker for $name', ({ responseModel, expectedBadge }) => {
		const wrapper = mount(UsageTable, {
			props: {
				data: [{
					request_id: `req-${responseModel}`,
					model: 'gpt-5.6-sol',
					upstream_model: 'gpt-5.5',
					model_mapping_chain: 'gpt-5.6-sol→gpt-5.5',
					upstream_response_model: responseModel,
					upstream_model_mismatch: true,
				}],
				loading: false,
				columns: [],
			},
			global: {
				stubs: {
					DataTable: DataTableStub,
					EmptyState: true,
					Icon: true,
					Teleport: true,
				},
			},
		})

		const text = wrapper.text()
		expect(text).toContain('gpt-5.6-sol')
		expect(text).toContain('gpt-5.5')
		expect(text).toContain(responseModel)
		expect(text).toContain(expectedBadge)
	})

  it('displays historical image rows with missing billing_mode as image usage without a 2K fallback', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          {
            ...baseImageRow,
            request_id: 'req-admin-legacy-missing-image',
            billing_mode: null,
            image_size: null,
            image_input_size: null,
            image_output_size: null,
            image_size_source: null,
            image_size_breakdown: null,
          },
        ],
        loading: false,
        columns: [],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const text = wrapper.text()
    expect(text).toContain('2 images')
    expect(text).toContain('not recorded')
    expect(text).not.toContain('(2K)')
  })
})

describe('admin UsageTable request ID column', () => {
  beforeEach(() => {
    appStoreMocks.showSuccess.mockReset()
    appStoreMocks.showError.mockReset()
  })

  it('renders and copies the request ID', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('navigator', { clipboard: { writeText } })

    const wrapper = mount(UsageTable, {
      props: {
        data: [{ ...baseImageRow, request_id: 'req-admin-visible-id' }],
        loading: false,
        columns: [{ key: 'request_id', label: 'Request ID' }],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).toContain('req-admin-visible-id')
    await wrapper.get('button[title="Copy to clipboard"]').trigger('click')

    expect(writeText).toHaveBeenCalledWith('req-admin-visible-id')
    expect(appStoreMocks.showSuccess).toHaveBeenCalledWith('Request ID copied')
  })

  it('renders and copies the upstream ID', async () => {
    const writeText = vi.fn().mockResolvedValue(undefined)
    vi.stubGlobal('navigator', { clipboard: { writeText } })

    const wrapper = mount(UsageTable, {
      props: {
        data: [{ ...baseImageRow, request_id: '', upstream_request_id: '20260903082826779695' }],
        loading: false,
        columns: [{ key: 'upstream_request_id', label: 'Upstream ID' }],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).toContain('20260903082826779695')
    const copyButtons = wrapper.findAll('button[title="Copy to clipboard"]')
    expect(copyButtons).toHaveLength(1)
    await copyButtons[0].trigger('click')

    expect(writeText).toHaveBeenCalledWith('20260903082826779695')
    expect(appStoreMocks.showSuccess).toHaveBeenCalledWith('Upstream ID copied')
  })
})

describe('admin UsageTable IP geolocation batch toolbar', () => {
  const DataTableStubWithIp = {
    props: ['data'],
    template: `
      <div>
        <div v-for="row in data" :key="row.request_id">
          <slot name="cell-ip_address" :row="row" />
        </div>
      </div>
    `,
  }

  beforeEach(() => {
    ipGeoMocks.getEntry.mockReset()
    ipGeoMocks.fetchOne.mockReset()
    ipGeoMocks.fetchBatch.mockReset()
    ipGeoMocks.getEntry.mockReturnValue({ status: 'idle' })
  })

  it('does not render the batch toolbar when the ip_address column is not visible', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [{ request_id: 'r1', ip_address: '8.8.8.8' }],
        loading: false,
        columns: [],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    expect(wrapper.text()).not.toContain('usage.ipGeo.batchFetch')
  })

  it('renders the batch toolbar with a pending count when the ip_address column is visible', () => {
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          { request_id: 'r1', ip_address: '8.8.8.8' },
          { request_id: 'r2', ip_address: '8.8.8.8' },
          { request_id: 'r3', ip_address: '1.1.1.1' },
        ],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    expect(wrapper.text()).toContain('usage.ipGeo.pending')
    const button = wrapper.find('button')
    expect(button.exists()).toBe(true)
    expect((button.element as HTMLButtonElement).disabled).toBe(false)
  })

  it('fetches deduplicated IPs from the current page when the batch button is clicked', async () => {
    ipGeoMocks.fetchBatch.mockResolvedValue(true)
    const wrapper = mount(UsageTable, {
      props: {
        data: [
          { request_id: 'r1', ip_address: '8.8.8.8' },
          { request_id: 'r2', ip_address: '8.8.8.8' },
          { request_id: 'r3', ip_address: '1.1.1.1' },
        ],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    await wrapper.find('button').trigger('click')
    expect(ipGeoMocks.fetchBatch).toHaveBeenCalledWith(['8.8.8.8', '1.1.1.1'])
    expect(wrapper.emitted('ipGeoBatchFailed')).toBeUndefined()
  })

  it('emits ipGeoBatchFailed when the batch request reports a network-level failure', async () => {
    ipGeoMocks.fetchBatch.mockResolvedValue(false)
    const wrapper = mount(UsageTable, {
      props: {
        data: [{ request_id: 'r1', ip_address: '8.8.8.8' }],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    await wrapper.find('button').trigger('click')
    expect(wrapper.emitted('ipGeoBatchFailed')).toHaveLength(1)
  })

  it('renders IpGeoCell content for ip_address cells', () => {
    ipGeoMocks.getEntry.mockReturnValue({ status: 'success', label: 'CN · Guangdong · Shenzhen', detail: {} })
    const wrapper = mount(UsageTable, {
      props: {
        data: [{ request_id: 'r1', ip_address: '121.35.47.43' }],
        loading: false,
        columns: [{ key: 'ip_address', label: 'IP' }],
      },
      global: { stubs: { DataTable: DataTableStubWithIp, EmptyState: true, Teleport: true } },
    })
    expect(wrapper.text()).toContain('121.35.47.43')
    expect(wrapper.text()).toContain('CN · Guangdong · Shenzhen')
  })
})

// A DataTable stub that also renders cell-user, so the deleted badge can be asserted.
const DataTableStubWithUser = {
  props: ['data'],
  template: `
    <div>
      <div v-for="row in data" :key="row.request_id">
        <slot name="cell-user" :row="row" />
        <slot name="cell-model" :row="row" :value="row.model" />
        <slot name="cell-reasoning_effort" :row="row" :value="row.reasoning_effort" />
        <slot name="cell-billing_mode" :row="row" />
        <slot name="cell-tokens" :row="row" />
        <slot name="cell-cost" :row="row" />
      </div>
    </div>
  `,
}

describe('admin UsageTable deleted-user badge', () => {
  it('renders deleted badge for a soft-deleted user row', () => {
    const row = {
      request_id: 'req-deleted-user-1',
      model: 'claude-3',
      user_id: 2,
      user: { id: 2, email: 'd@test.com', deleted_at: '2026-05-28T00:00:00Z' },
      actual_cost: 0,
      total_cost: 0,
      input_cost: 0,
      output_cost: 0,
      rate_multiplier: 1,
      input_tokens: 1,
      output_tokens: 1,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'user', label: 'User' }],
      },
      global: {
        stubs: {
          DataTable: DataTableStubWithUser,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).toContain('Deleted')
    expect(wrapper.text()).toContain('d@test.com')
  })

  it('does NOT render deleted badge for an active user row', () => {
    const row = {
      request_id: 'req-active-user-1',
      model: 'claude-3',
      user_id: 3,
      user: { id: 3, email: 'active@test.com', deleted_at: null },
      actual_cost: 0,
      total_cost: 0,
      input_cost: 0,
      output_cost: 0,
      rate_multiplier: 1,
      input_tokens: 1,
      output_tokens: 1,
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [row],
        loading: false,
        columns: [{ key: 'user', label: 'User' }],
      },
      global: {
        stubs: {
          DataTable: DataTableStubWithUser,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.text()).not.toContain('Deleted')
    expect(wrapper.text()).toContain('active@test.com')
  })

  it('forwards the telemetry ledger variant to the shared data table', () => {
    const DataTableVariantStub = {
      props: ['variant'],
      template: '<div data-testid="data-table-variant" :data-variant="variant" />',
    }

    const wrapper = mount(UsageTable, {
      props: {
        data: [],
        loading: false,
        columns: [],
        telemetry: true,
      },
      global: {
        stubs: {
          DataTable: DataTableVariantStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.get('[data-testid="data-table-variant"]').attributes('data-variant'))
      .toBe('telemetry')
  })

})

describe('admin UsageTable telemetry ledger cells', () => {
  it('renders the reference protocol, token throughput, latency, and time/source composites', () => {
    const TelemetryCellsStub = {
      props: ['data', 'variant'],
      template: `
        <div :data-variant="variant">
          <div v-for="row in data" :key="row.request_id">
            <slot name="cell-user" :row="row" />
            <slot name="cell-model" :row="row" />
            <slot name="cell-reasoning_effort" :row="row" />
            <slot name="cell-endpoint" :row="row" />
            <slot name="cell-stream" :row="row" />
            <slot name="cell-billing_mode" :row="row" />
            <slot name="cell-tokens" :row="row" />
            <slot name="cell-cache_hit_rate" :row="row" />
            <slot name="cell-cost" :row="row" />
            <slot name="cell-latency" :row="row" />
            <slot name="cell-created_at" :row="row" :value="row.created_at" />
          </div>
        </div>
      `,
    }

    const wrapper = mount(UsageTable, {
      props: {
        telemetry: true,
        loading: false,
        columns: [],
        data: [{
          request_id: 'req-ledger-1',
          user_id: 1,
          user: { id: 1, email: 'admin@sub2api.local' },
          model: 'gpt-5.6-terra', upstream_model: 'gpt-5.6-sol', upstream_response_model: 'gpt-5.5',
          reasoning_effort: 'high',
          inbound_endpoint: '/v1/responses',
          upstream_endpoint: '/v1/responses',
          request_type: 'ws_v2',
          billing_mode: 'token',
          input_tokens: 1427,
          output_tokens: 105,
          cache_read_tokens: 80100,
          cache_creation_tokens: 0,
          cache_creation_5m_tokens: 0,
          cache_creation_1h_tokens: 0,
          cache_ttl_overridden: false,
          actual_cost: 0.02014,
          total_cost: 0.02014,
          duration_ms: 6990,
          first_token_ms: 4710,
          created_at: '2026-09-18T14:32:08.410Z',
          ip_address: '104.28.192.12',
        }],
      },
      global: {
        stubs: {
          DataTable: TelemetryCellsStub,
          EmptyState: true,
          Icon: true,
          IpGeoCell: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.get('.telemetry-model-audit').text()).toContain('gpt-5.6-terra')
    expect(wrapper.get('.telemetry-model-audit').text()).toContain('gpt-5.6-sol')
    expect(wrapper.get('.telemetry-model-audit').text()).toContain('gpt-5.5')
    expect(wrapper.get('[data-testid="telemetry-protocol"]').text()).toBe('WS')
    expect(wrapper.get('[data-testid="telemetry-token-throughput"]').text()).toContain('IN1,427')
    expect(wrapper.get('[data-testid="telemetry-token-throughput"]').text()).toContain('OUT105')
    expect(wrapper.get('[data-testid="telemetry-token-throughput"]').text()).toContain('80.1KKV CACHE')
    expect(wrapper.get('[data-testid="telemetry-latency"]').text()).toContain('4.71s / 6.99s')
    expect(wrapper.get('[data-testid="telemetry-time-source"]').text()).toContain(':32:08.410')
    expect(wrapper.get('[data-testid="telemetry-time-source"]').text()).toContain('104.28.192.12')
  })
})

describe('admin UsageTable telemetry expanded layout and token breakdown', () => {
  it('switches to stable wide-table mode when optional columns are visible', () => {
    const wrapper = mount(UsageTable, {
      props: {
        telemetry: true,
        loading: false,
        data: [],
        columns: [
          { key: 'user', label: 'User' },
          { key: 'tokens', label: 'Tokens' },
          { key: 'request_id', label: 'Request ID' },
        ],
      },
      global: {
        stubs: {
          DataTable: true,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    expect(wrapper.get('.telemetry-usage-table').classes()).toContain('telemetry-usage-table--expanded')
  })

  it('keeps reduced column selections dense instead of distributing empty gaps', () => {
    const wrapper = mount(UsageTable, {
      props: {
        telemetry: true,
        loading: false,
        data: [],
        columns: [
          { key: 'user', label: 'User' },
          { key: 'api_key', label: 'API Key' },
          { key: 'account', label: 'Account' },
          { key: 'model', label: 'Model' },
          { key: 'reasoning_effort', label: 'Reasoning' },
          { key: 'endpoint', label: 'Endpoint' },
          { key: 'tokens', label: 'Tokens' },
        ],
      },
      global: {
        stubs: {
          DataTable: true,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    const table = wrapper.get('.telemetry-usage-table')
    expect(table.classes()).toContain('telemetry-usage-table--compact')
    expect(table.attributes('style')).toContain('--telemetry-layout-width: 1175px')
  })

  it('renders the design-system token breakdown popover', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        telemetry: true,
        loading: false,
        columns: [{ key: 'tokens', label: 'Tokens' }],
        data: [{
          ...baseImageRow,
          request_id: 'req-token-breakdown',
          billing_mode: 'token',
          input_tokens: 900,
          output_tokens: 220,
          cache_creation_tokens: 320,
          cache_read_tokens: 540,
        }],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await wrapper.get('[data-testid="telemetry-token-throughput"]').trigger('mouseenter')

    const popover = wrapper.get('[data-testid="telemetry-token-breakdown"]')
    expect(popover.text()).toContain('TOKEN BREAKDOWN')
    expect(popover.text()).toContain('REALTIME')
    expect(popover.text()).toContain('Input (Prompt)900')
    expect(popover.text()).toContain('Output (Completion)220')
    expect(popover.text()).toContain('Cache Read540')
    expect(popover.text()).toContain('TOTAL TOKENS1,980')
    expect(popover.get('[role="progressbar"]').attributes('aria-valuenow')).toBe('27.3')
  })

  it('keeps image-token splits in the redesigned breakdown popover', async () => {
    const wrapper = mount(UsageTable, {
      props: {
        telemetry: true,
        loading: false,
        columns: [{ key: 'tokens', label: 'Tokens' }],
        data: [{
          ...baseImageRow,
          request_id: 'req-image-token-breakdown',
          billing_mode: 'token',
          input_tokens: 100,
          image_input_tokens: 80,
          output_tokens: 50,
          image_output_tokens: 40,
          cache_creation_tokens: 0,
          cache_read_tokens: 0,
        }],
      },
      global: {
        stubs: {
          DataTable: DataTableStub,
          EmptyState: true,
          Icon: true,
          Teleport: true,
        },
      },
    })

    await wrapper.get('[data-testid="telemetry-token-throughput"]').trigger('mouseenter')

    const popover = wrapper.get('[data-testid="telemetry-token-breakdown"]')
    expect(popover.text()).toContain('Text Input (Prompt)20')
    expect(popover.text()).toContain('Image Input80')
    expect(popover.text()).toContain('Text Output (Completion)10')
    expect(popover.text()).toContain('Image Output40')
  })

})
