<template>
  <div class="card">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
        {{ t('admin.settings.codexModelReasoning.title') }}
      </h2>
      <p class="mt-1 text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.settings.codexModelReasoning.description') }}
      </p>
    </div>

    <div class="space-y-5 p-6">
      <p v-if="loading" role="status" class="text-sm text-gray-500 dark:text-gray-400">
        {{ t('common.loading') }}
      </p>
      <template v-else-if="loadFailed">
        <p role="alert" class="text-sm text-red-600 dark:text-red-400">
          {{ t('admin.settings.codexModelReasoning.loadError') }}
        </p>
        <button type="button" class="btn btn-secondary" @click="loadRules">
          {{ t('admin.settings.codexModelReasoning.retry') }}
        </button>
      </template>
      <template v-else>
        <p v-if="rules.length === 0" class="text-sm text-gray-500 dark:text-gray-400">
          {{ t('admin.settings.codexModelReasoning.empty') }}
        </p>
        <div
          v-for="(rule, index) in rules"
          :key="rule.id"
          class="space-y-4 rounded-lg border border-gray-200 p-4 dark:border-dark-600"
        >
          <div class="flex flex-wrap items-end gap-3">
            <label class="min-w-[14rem] flex-1 text-sm font-medium text-gray-700 dark:text-gray-200">
              {{ t('admin.settings.codexModelReasoning.model') }}
              <input
                v-model="rule.model"
                type="text"
                class="input mt-1 w-full"
                maxlength="200"
                autocomplete="off"
                :placeholder="t('admin.settings.codexModelReasoning.modelPlaceholder')"
                :data-testid="`codex-reasoning-model-${index}`"
              />
            </label>
            <button
              type="button"
              class="btn btn-secondary"
              :data-testid="`codex-reasoning-remove-${index}`"
              @click="rules.splice(index, 1)"
            >
              {{ t('admin.settings.codexModelReasoning.remove') }}
            </button>
          </div>
          <fieldset>
            <legend class="mb-2 text-sm font-medium text-gray-700 dark:text-gray-200">
              {{ t('admin.settings.codexModelReasoning.supportedLevels') }}
            </legend>
            <div class="flex flex-wrap gap-x-5 gap-y-2">
              <label
                v-for="effort in EFFORTS"
                :key="effort"
                class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200"
              >
                <input
                  type="checkbox"
                  :checked="rule.supported_reasoning_levels.includes(effort)"
                  :data-testid="`codex-reasoning-effort-${index}-${effort}`"
                  @change="toggleEffort(rule, effort, $event)"
                />
                {{ effort }}
              </label>
            </div>
          </fieldset>
          <label class="block text-sm font-medium text-gray-700 dark:text-gray-200">
            {{ t('admin.settings.codexModelReasoning.defaultLevel') }}
            <select
              v-model="rule.default_reasoning_level"
              class="input mt-1 w-full max-w-xs"
              :data-testid="`codex-reasoning-default-${index}`"
            >
              <option v-if="!rule.default_reasoning_level" value="" disabled>
                {{ t('admin.settings.codexModelReasoning.chooseDefault') }}
              </option>
              <option
                v-else-if="!rule.supported_reasoning_levels.includes(rule.default_reasoning_level)"
                :value="rule.default_reasoning_level"
                disabled
              >
                {{ t('admin.settings.codexModelReasoning.invalidDefault') }}
              </option>
              <option v-for="effort in selectedEfforts(rule)" :key="effort" :value="effort">
                {{ effort }}
              </option>
            </select>
          </label>
        </div>
        <div class="flex flex-wrap justify-between gap-3">
          <button
            type="button"
            class="btn btn-secondary"
            :disabled="saving || rules.length >= MAX_RULES"
            data-testid="codex-reasoning-add"
            @click="addRule"
          >
            {{ t('admin.settings.codexModelReasoning.add') }}
          </button>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="saving || loadFailed || loading"
            data-testid="codex-reasoning-save"
            @click="saveRules"
          >
            {{ saving ? t('admin.settings.saving') : t('admin.settings.codexModelReasoning.save') }}
          </button>
        </div>
      </template>
      <div v-if="loading || loadFailed" class="flex justify-end">
        <button
          type="button"
          disabled
          class="btn btn-primary"
          data-testid="codex-reasoning-save"
        >
          {{ t('admin.settings.codexModelReasoning.save') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { adminAPI } from '@/api/admin'
import type { CodexModelReasoningRule, CodexReasoningEffort } from '@/api/admin/settings'
import { useAppStore } from '@/stores'
import { extractApiErrorMessage } from '@/utils/apiError'

const { t } = useI18n()
const appStore = useAppStore()
const EFFORTS: CodexReasoningEffort[] = ['none', 'minimal', 'low', 'medium', 'high', 'xhigh', 'max', 'ultra']
const MAX_RULES = 64
const INVALID_MODEL_CHAR = /[?*\[\]\p{White_Space}\p{Cc}]/u

type DraftRule = Omit<CodexModelReasoningRule, 'default_reasoning_level'> & {
  id: number
  default_reasoning_level: CodexReasoningEffort | ''
}
const rules = ref<DraftRule[]>([])
const loading = ref(true)
const loadFailed = ref(false)
const saving = ref(false)
let nextId = 0

function selectedEfforts(rule: DraftRule): CodexReasoningEffort[] {
  return EFFORTS.filter(effort => rule.supported_reasoning_levels.includes(effort))
}

function toggleEffort(rule: DraftRule, effort: CodexReasoningEffort, event: Event) {
  const checked = (event.target as HTMLInputElement).checked
  rule.supported_reasoning_levels = checked
    ? EFFORTS.filter(item => item === effort || rule.supported_reasoning_levels.includes(item))
    : rule.supported_reasoning_levels.filter(item => item !== effort)
}

function addRule() {
  if (rules.value.length >= MAX_RULES) return
  rules.value.push({ id: nextId++, model: '', supported_reasoning_levels: [], default_reasoning_level: '' })
}

async function loadRules() {
  loading.value = true
  loadFailed.value = false
  try {
    const response = await adminAPI.settings.getCodexModelReasoningSettings()
    if (!Array.isArray(response.rules)) throw new Error('Invalid Codex reasoning rules')
    rules.value = response.rules.map(rule => ({
      id: nextId++, model: rule.model,
      supported_reasoning_levels: [...rule.supported_reasoning_levels],
      default_reasoning_level: rule.default_reasoning_level as CodexReasoningEffort,
    }))
  } catch (err) {
    loadFailed.value = true
    appStore.showError(extractApiErrorMessage(err, t('admin.settings.codexModelReasoning.loadError')))
  } finally {
    loading.value = false
  }
}

function validatedRules(): CodexModelReasoningRule[] | null {
  if (rules.value.length > MAX_RULES) {
    appStore.showError(t('admin.settings.codexModelReasoning.tooManyRules'))
    return null
  }
  const seen = new Set<string>()
  const result: CodexModelReasoningRule[] = []
  for (const rule of rules.value) {
    const model = rule.model.trim()
    if (!model || new TextEncoder().encode(model).length > 200 || INVALID_MODEL_CHAR.test(model)) {
      appStore.showError(t('admin.settings.codexModelReasoning.invalidModel'))
      return null
    }
    if (seen.has(model)) {
      appStore.showError(t('admin.settings.codexModelReasoning.duplicateModel'))
      return null
    }
    seen.add(model)
    if (rule.supported_reasoning_levels.length === 0) {
      appStore.showError(t('admin.settings.codexModelReasoning.emptyLevels'))
      return null
    }
    if (rule.supported_reasoning_levels.some(effort => !EFFORTS.includes(effort)) ||
        new Set(rule.supported_reasoning_levels).size !== rule.supported_reasoning_levels.length) {
      appStore.showError(t('admin.settings.codexModelReasoning.invalidLevels'))
      return null
    }
    const defaultLevel = rule.default_reasoning_level
    if (!defaultLevel || !rule.supported_reasoning_levels.includes(defaultLevel)) {
      appStore.showError(t('admin.settings.codexModelReasoning.invalidDefault'))
      return null
    }
    result.push({
      model,
      supported_reasoning_levels: selectedEfforts(rule),
      default_reasoning_level: defaultLevel,
    })
  }
  return result
}

async function saveRules() {
  if (loading.value || loadFailed.value || saving.value) return
  const validated = validatedRules()
  if (!validated) return
  saving.value = true
  try {
    await adminAPI.settings.updateCodexModelReasoningSettings({ rules: validated })
    rules.value.forEach((rule, index) => { rule.model = validated[index].model })
    appStore.showSuccess(t('admin.settings.codexModelReasoning.saved'))
  } catch (err) {
    appStore.showError(extractApiErrorMessage(err, t('admin.settings.codexModelReasoning.saveError')))
  } finally {
    saving.value = false
  }
}

onMounted(loadRules)
</script>
