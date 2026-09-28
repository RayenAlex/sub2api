import { afterEach, describe, expect, it, vi } from 'vitest'
import { apiClient } from '@/api/client'
import { getCodexModelReasoningSettings, updateCodexModelReasoningSettings } from '@/api/admin/settings'

describe('Codex model reasoning settings API', () => {
  afterEach(() => vi.restoreAllMocks())

  it('loads from the dedicated admin endpoint', async () => {
    const response = { rules: [{ model: 'gpt-6-sol', supported_reasoning_levels: ['none', 'max'], default_reasoning_level: 'max' }] }
    const get = vi.spyOn(apiClient, 'get').mockResolvedValue({ data: response })
    expect(await getCodexModelReasoningSettings()).toEqual(response)
    expect(get).toHaveBeenCalledWith('/admin/settings/codex-model-reasoning')
  })

  it('replaces only these rules, including empty reset, through its dedicated PUT', async () => {
    const put = vi.spyOn(apiClient, 'put').mockResolvedValue({ data: { rules: [] } })
    expect(await updateCodexModelReasoningSettings({ rules: [] })).toEqual({ rules: [] })
    expect(put).toHaveBeenCalledWith('/admin/settings/codex-model-reasoning', { rules: [] })
  })
})
