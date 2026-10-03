import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import UsageTable from '../UsageTable.vue'
import DataTable from '@/components/common/DataTable.vue'
import dashboard from '@/i18n/locales/en/dashboard'
import type { AdminUsageLog, AdminWebSearchEvent } from '@/types'

vi.mock('@/stores/app', () => ({ useAppStore: () => ({ showSuccess: vi.fn(), showError: vi.fn() }) }))
vi.mock('@/utils/ipGeoLookup', () => ({ getEntry: vi.fn(), fetchBatch: vi.fn() }))
vi.mock('vue-i18n', async () => ({
  ...await vi.importActual<typeof import('vue-i18n')>('vue-i18n'),
  useI18n: () => ({
    t: (key: string, params: Record<string, string | number> = {}) => {
      const name = key.replace('usage.webSearch.', '') as keyof typeof dashboard.usage.webSearch
      const message = key.startsWith('usage.webSearch.') ? dashboard.usage.webSearch[name] : key
      return message.replace(/\{(\w+)\}/g, (_: string, token: string) => String(params[token] ?? token))
    },
  }),
}))

const event = (id: number): AdminWebSearchEvent => ({
  id, sequence: id, call_id: `call-${id}`, query: `query ${id}`, status: 'completed',
  source_count: 1, sources: [{ url: `https://example.com/${id}`, title: `Source ${id}` }],
  created_at: '2026-01-02T00:00:00Z',
})
const row = (id = 1, events?: AdminWebSearchEvent[]): AdminUsageLog => ({
  id, user_id: 1, user: { email: 'user@example.com' }, model: 'test-model',
  input_tokens: 1234, output_tokens: 567, actual_cost: 0.123456,
  duration_ms: 1000, first_token_ms: 100, created_at: '2026-01-02T00:00:00Z',
  ...(events === undefined ? {} : { web_search_events: events }),
} as AdminUsageLog)
const columns = ['user', 'tokens', 'cost', 'latency', 'created_at'].map(key => ({ key, label: key, sortable: true }))
const wrappers: ReturnType<typeof mount>[] = []
function render(data: AdminUsageLog[], optIn = true) {
  const wrapper = mount(UsageTable, {
    props: { data, columns, telemetry: true, showWebSearchEvents: optIn },
  })
  wrappers.push(wrapper)
  return wrapper
}
function viewport(desktop: boolean) {
  vi.stubGlobal('matchMedia', vi.fn(() => ({
    matches: desktop, addEventListener: vi.fn(), removeEventListener: vi.fn(),
  })))
}
const toggle = '[data-testid="web-search-toggle"]'
const children = '[data-testid="web-search-event"]'
const sources = '[data-testid="web-search-sources-toggle"]'

beforeEach(() => viewport(true))
afterEach(() => {
  wrappers.splice(0).forEach(wrapper => wrapper.unmount())
  vi.unstubAllGlobals()
})

describe('UsageTable web search with real DataTable', () => {
  it('has no disclosure for omitted/empty events and requires explicit admin opt-in', () => {
    expect(render([row(1), row(2, [])]).find(toggle).exists()).toBe(false)
    expect(render([row(1, [event(1)])], false).find(toggle).exists()).toBe(false)
    const wrapper = mount(UsageTable, { props: { data: [row(1, [event(1)])], columns } })
    wrappers.push(wrapper)
    expect(wrapper.find(toggle).exists()).toBe(false)
  })

  it.each([1, 3])('expands/collapses %i calls adjacent to parent without duplicating usage cells', async (count) => {
    const data = [row(1, Array.from({ length: count }, (_, i) => event(i + 1))), row(2)]
    const before = JSON.stringify(data)
    const wrapper = render(data)
    expect(wrapper.get(toggle).text()).toBe(`Web Search × ${count}`)
    expect(wrapper.get(toggle).attributes('aria-expanded')).toBe('false')
    expect(wrapper.findAll(children)).toHaveLength(0)
    await wrapper.get(toggle).trigger('click')
    expect(wrapper.get(toggle).attributes('aria-expanded')).toBe('true')
    expect(wrapper.findAll(children)).toHaveLength(count)
    const rows = wrapper.findAll('tbody > tr')
    expect(rows[0].attributes('data-row-id')).toBe('1')
    expect(rows[count + 1].attributes('data-row-id')).toBe('2')
    for (const child of wrapper.findAll(children)) {
      expect(child.element.tagName).toBe('TR')
      expect(child.get('td').attributes('colspan')).toBe(String(columns.length))
      expect(child.find('[data-column-key]').exists()).toBe(false)
      expect(child.text()).not.toContain('0.123456')
      expect(child.text()).toContain('completed')
      expect(child.find('time').exists()).toBe(true)
    }
    expect(wrapper.findAll('[data-testid="telemetry-token-throughput"]')).toHaveLength(2)
    expect(wrapper.findAll('.telemetry-cost-value')).toHaveLength(2)
    expect(wrapper.findAll('[data-testid="telemetry-latency"]')).toHaveLength(2)
    expect(wrapper.findComponent(DataTable).props('data')).toEqual(data)
    expect(JSON.stringify(data)).toBe(before)
    await wrapper.get(toggle).trigger('click')
    expect(wrapper.findAll(children)).toHaveLength(0)
  })

  it('keeps source disclosure independent per event and resets it when parent collapses', async () => {
    const wrapper = render([row(1, [event(1), event(2)])])
    await wrapper.get(toggle).trigger('click')
    expect(wrapper.findAll('a')).toHaveLength(0)
    await wrapper.findAll(sources)[0].trigger('click')
    expect(wrapper.findAll('a').map(a => a.text())).toEqual(['Source 1'])
    expect(wrapper.findAll(sources)[1].attributes('aria-expanded')).toBe('false')
    await wrapper.findAll(sources)[1].trigger('click')
    await wrapper.findAll(sources)[0].trigger('click')
    expect(wrapper.findAll('a').map(a => a.text())).toEqual(['Source 2'])
    expect(wrapper.get('a').attributes()).toMatchObject({ target: '_blank', rel: 'noopener noreferrer' })
    await wrapper.get(toggle).trigger('click')
    await wrapper.get(toggle).trigger('click')
    expect(wrapper.findAll('a')).toHaveLength(0)
  })

  it('escapes malicious text and makes only absolute HTTP(S) sources clickable', async () => {
    const malicious = '<img src=x onerror=alert(1)>'
    const search = event(1)
    search.query = malicious
    search.sources = ['https://example.com/safe', 'http://example.com/ok', 'javascript:alert(1)',
      'data:text/html,<script>alert(1)</script>', '//example.com', '/relative', 'not a url', 'file:///tmp/test']
      .map(url => ({ url, title: malicious }))
    search.source_count = search.sources.length
    const wrapper = render([row(1, [search])])
    await wrapper.get(toggle).trigger('click')
    await wrapper.get(sources).trigger('click')
    expect(wrapper.get(children).text()).toContain(malicious)
    expect(wrapper.findAll('img, script')).toHaveLength(0)
    expect(wrapper.findAll('a').map(a => a.attributes('href'))).toEqual(['https://example.com/safe', 'http://example.com/ok'])
    expect(wrapper.findAll('li')).toHaveLength(8)
  })

  it('resets disclosures across same-id data replacement and paging', async () => {
    const wrapper = render([row(1, [event(1)])])
    await wrapper.get(toggle).trigger('click')
    await wrapper.get(sources).trigger('click')
    await wrapper.setProps({ data: [row(1, [event(2)])] })
    expect(wrapper.get(toggle).attributes('aria-expanded')).toBe('false')
    expect(wrapper.findAll(children)).toHaveLength(0)
    await wrapper.get(toggle).trigger('click')
    expect(wrapper.findAll('a')).toHaveLength(0)
    await wrapper.setProps({ data: [row(2)] })
    expect(wrapper.findAll(children)).toHaveLength(0)
    expect(wrapper.findAll(toggle)).toHaveLength(0)
  })

  it('handles zero sources without an empty source disclosure', async () => {
    const search = { ...event(1), sources: [], source_count: 0 }
    const wrapper = render([row(1, [search])])
    await wrapper.get(toggle).trigger('click')
    expect(wrapper.get(children).text()).toContain('Sources: 0')
    expect(wrapper.find(sources).exists()).toBe(false)
  })

  it('expands parents independently and collapses everything when admin opt-in is removed', async () => {
    const wrapper = render([row(1, [event(1)]), row(2, [event(2)])])
    await wrapper.findAll(toggle)[0].trigger('click')
    await wrapper.findAll(toggle)[1].trigger('click')
    expect(wrapper.findAll(children)).toHaveLength(2)
    await wrapper.findAll(toggle)[0].trigger('click')
    expect(wrapper.findAll(children)).toHaveLength(1)
    expect(wrapper.get(children).text()).toContain('query 2')
    await wrapper.setProps({ showWebSearchEvents: false })
    expect(wrapper.findAll(children)).toHaveLength(0)
    expect(wrapper.findAll(toggle)).toHaveLength(0)
  })

  it('keeps detail rows with their parent under client sorting', async () => {
    const second = { ...row(2), created_at: '2026-01-01T00:00:00Z' }
    const wrapper = render([row(1, [event(1)]), second])
    await wrapper.get(toggle).trigger('click')
    await wrapper.get('th[data-column-key="created_at"]').trigger('click')
    expect(wrapper.findAll('tr[data-row-id]').map(tr => tr.attributes('data-row-id'))).toEqual(['2', '1'])
    const parent = wrapper.get('tr[data-row-id="1"]').element
    expect(parent.nextElementSibling?.getAttribute('data-testid')).toBe('web-search-event')
  })

  it('renders mobile details as valid card content, not table rows', async () => {
    viewport(false)
    const wrapper = render([row(1, [event(1), event(2)])])
    await wrapper.get(toggle).trigger('click')
    expect(wrapper.findAll('table, tr, td')).toHaveLength(0)
    expect(wrapper.findAll(children)).toHaveLength(2)
    await wrapper.findAll(sources)[1].trigger('click')
    expect(wrapper.get('a').text()).toBe('Source 2')
  })

  it('disables virtual measurement for expanded details above 100 parent rows', async () => {
    const data = Array.from({ length: 105 }, (_, i) => row(i + 1, [event(i + 1)]))
    const wrapper = render(data)
    const table = wrapper.findComponent(DataTable)
    expect((table.vm as unknown as { shouldVirtualize: boolean }).shouldVirtualize).toBe(true)
    await wrapper.findAll(toggle)[0].trigger('click')
    expect((table.vm as unknown as { shouldVirtualize: boolean }).shouldVirtualize).toBe(false)
    expect(wrapper.findAll('tr[data-row-id]')).toHaveLength(105)
    expect(wrapper.findAll(children)).toHaveLength(1)
    expect(wrapper.findAll('tr[aria-hidden="true"]')).toHaveLength(0)
    await wrapper.findAll(toggle)[0].trigger('click')
    expect((table.vm as unknown as { shouldVirtualize: boolean }).shouldVirtualize).toBe(true)
    expect(table.props('data')).toHaveLength(105)
  })
})
