import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import zhSettings from '@/i18n/locales/zh/admin/settings'
import enSettings from '@/i18n/locales/en/admin/settings'
import CodexModelReasoningSettings from '../CodexModelReasoningSettings.vue'

const { getRules, saveRules, showError, showSuccess } = vi.hoisted(() => ({
  getRules: vi.fn(), saveRules: vi.fn(), showError: vi.fn(), showSuccess: vi.fn(),
}))
vi.mock('@/api/admin', () => ({ adminAPI: { settings: {
  getCodexModelReasoningSettings: getRules,
  updateCodexModelReasoningSettings: saveRules,
} } }))
vi.mock('@/stores', () => ({ useAppStore: () => ({ showError, showSuccess }) }))
vi.mock('vue-i18n', () => ({ useI18n: () => ({ t: (key: string) => key }) }))

const originalRule = {
  model: 'gpt-6-sol',
  supported_reasoning_levels: ['none', 'low', 'medium', 'high'],
  default_reasoning_level: 'medium',
}

async function open() {
  const wrapper = mount(CodexModelReasoningSettings, { global: { stubs: { Icon: true } } })
  await flushPromises()
  return wrapper
}

describe('Codex model reasoning settings', () => {
  beforeEach(() => {
    getRules.mockReset().mockResolvedValue({ rules: [originalRule] })
    saveRules.mockReset().mockImplementation(async (payload) => payload)
    showError.mockReset()
    showSuccess.mockReset()
  })

  it('loads exact model rules and displays every known effort as a selectable option', async () => {
    const wrapper = await open()
    expect(getRules).toHaveBeenCalledTimes(1)
    expect(wrapper.get<HTMLInputElement>('[data-testid="codex-reasoning-model-0"]').element.value).toBe('gpt-6-sol')
    expect(wrapper.get<HTMLSelectElement>('[data-testid="codex-reasoning-default-0"]').element.value).toBe('medium')
    expect(wrapper.get<HTMLInputElement>('[data-testid="codex-reasoning-effort-0-none"]').element.checked).toBe(true)
    expect(wrapper.get<HTMLInputElement>('[data-testid="codex-reasoning-effort-0-ultra"]').element.checked).toBe(false)
    expect(wrapper.get<HTMLSelectElement>('[data-testid="codex-reasoning-default-0"]').findAll('option').map(o => o.element.value)).toEqual(['none', 'low', 'medium', 'high'])
  })

  it('adds a row and sends trimmed exact slugs and selected efforts through its own endpoint', async () => {
    const wrapper = await open()
    await wrapper.get('[data-testid="codex-reasoning-add"]').trigger('click')
    await wrapper.get('[data-testid="codex-reasoning-model-1"]').setValue('  gpt-6-luna  ')
    await wrapper.get('[data-testid="codex-reasoning-effort-1-none"]').setValue(true)
    await wrapper.get('[data-testid="codex-reasoning-effort-1-max"]').setValue(true)
    await wrapper.get('[data-testid="codex-reasoning-default-1"]').setValue('max')
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    await flushPromises()
    expect(saveRules).toHaveBeenCalledWith({ rules: [originalRule, {
      model: 'gpt-6-luna',
      supported_reasoning_levels: ['none', 'max'],
      default_reasoning_level: 'max',
    }] })
    expect(showSuccess).toHaveBeenCalledTimes(1)
  })

  it('accepts an exact non-ASCII upstream model slug supported by the backend contract', async () => {
    const wrapper = await open()
    await wrapper.get('[data-testid="codex-reasoning-model-0"]').setValue('  gpt-6-日  ')
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    await flushPromises()
    expect(saveRules).toHaveBeenCalledWith({ rules: [{ ...originalRule, model: 'gpt-6-日' }] })
  })

  it('limits the rule set to 64 entries and prevents adding a 65th', async () => {
    const wrapper = await open()
    for (let i = 1; i < 64; i++) await wrapper.get('[data-testid="codex-reasoning-add"]').trigger('click')
    expect(wrapper.findAll('input[data-testid^="codex-reasoning-model-"]')).toHaveLength(64)
    expect(wrapper.get<HTMLButtonElement>('[data-testid="codex-reasoning-add"]').element.disabled).toBe(true)
  })

  it('deletes the last rule and saves empty rules to restore server defaults', async () => {
    const wrapper = await open()
    await wrapper.get('[data-testid="codex-reasoning-remove-0"]').trigger('click')
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    await flushPromises()
    expect(saveRules).toHaveBeenCalledWith({ rules: [] })
  })

  it.each(['', 'gpt-6-*', 'gpt-6-[ab]', 'gpt 6 sol', `gpt-${'a'.repeat(200)}`])('rejects invalid slug %s', async (model) => {
    const wrapper = await open()
    await wrapper.get('[data-testid="codex-reasoning-model-0"]').setValue(model)
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    expect(saveRules).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledTimes(1)
  })

  it('rejects duplicate trimmed slugs and empty effort sets', async () => {
    const wrapper = await open()
    await wrapper.get('[data-testid="codex-reasoning-add"]').trigger('click')
    await wrapper.get('[data-testid="codex-reasoning-model-1"]').setValue(' gpt-6-sol ')
    await wrapper.get('[data-testid="codex-reasoning-effort-1-low"]').setValue(true)
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    expect(saveRules).not.toHaveBeenCalled()
    await wrapper.get('[data-testid="codex-reasoning-model-1"]').setValue('gpt-6-luna')
    await wrapper.get('[data-testid="codex-reasoning-effort-1-low"]').setValue(false)
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    expect(saveRules).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledTimes(2)
  })

  it('rejects a default that is no longer selected', async () => {
    const wrapper = await open()
    await wrapper.get('[data-testid="codex-reasoning-default-0"]').setValue('high')
    await wrapper.get('[data-testid="codex-reasoning-effort-0-high"]').setValue(false)
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    expect(saveRules).not.toHaveBeenCalled()
    expect(showError).toHaveBeenCalledTimes(1)
  })

  it('does not save a destructive empty ruleset after a load failure', async () => {
    getRules.mockRejectedValueOnce(new Error('offline'))
    const wrapper = await open()
    expect(showError).toHaveBeenCalledTimes(1)
    expect(wrapper.get<HTMLButtonElement>('[data-testid="codex-reasoning-save"]').element.disabled).toBe(true)
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    expect(saveRules).not.toHaveBeenCalled()
  })

  it('keeps edits after a failed save to allow a retry', async () => {
    saveRules.mockRejectedValueOnce(new Error('offline'))
    const wrapper = await open()
    await wrapper.get('[data-testid="codex-reasoning-model-0"]').setValue('gpt-6-luna')
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    await flushPromises()
    expect(showError).toHaveBeenCalledTimes(1)
    expect(wrapper.get<HTMLInputElement>('[data-testid="codex-reasoning-model-0"]').element.value).toBe('gpt-6-luna')
    await wrapper.get('[data-testid="codex-reasoning-save"]').trigger('click')
    await flushPromises()
    expect(saveRules).toHaveBeenCalledTimes(2)
  })

  it('provides translated catalog-only labels and errors', () => {
    for (const locale of [zhSettings.settings, enSettings.settings]) {
      expect(locale.tabs.codexModelReasoning).toBeTruthy()
      for (const key of ['title', 'description', 'model', 'supportedLevels', 'defaultLevel', 'invalidModel', 'duplicateModel', 'emptyLevels', 'invalidDefault', 'loadError', 'saveError'] as const) {
        expect(locale.codexModelReasoning[key]).toBeTruthy()
      }
    }
    expect(zhSettings.settings.codexModelReasoning.description).toContain('清单')
    expect(enSettings.settings.codexModelReasoning.description.toLowerCase()).toContain('catalog')
  })
})
