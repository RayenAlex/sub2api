<template>
  <component
    :is="mobile ? 'div' : 'tr'"
    v-for="event in events"
    :key="event.id"
    data-testid="web-search-event"
    class="bg-gray-50 text-sm dark:bg-dark-800"
  >
    <component :is="mobile ? 'div' : 'td'" :colspan="mobile ? undefined : columnCount" class="px-6 py-3">
      <div class="space-y-2 border-l-2 border-primary-200 pl-4 dark:border-primary-800">
        <div class="flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-gray-500 dark:text-gray-400">
          <span>{{ t('usage.webSearch.sequence', { sequence: event.sequence }) }}</span>
          <time :datetime="event.created_at">{{ formatDateTime(event.created_at) }}</time>
          <span>{{ t('usage.webSearch.status') }}: {{ event.status }}</span>
        </div>
        <p class="whitespace-pre-wrap break-words text-gray-900 dark:text-gray-100">
          <span class="font-medium">{{ t('usage.webSearch.query') }}:</span> {{ event.query }}
        </p>
        <button
          v-if="event.sources.length"
          type="button"
          data-testid="web-search-sources-toggle"
          class="rounded text-xs text-primary-600 underline focus-visible:ring-2 focus-visible:ring-primary-500 dark:text-primary-400"
          :aria-expanded="expandedSources.has(event.id)"
          @click.stop="toggleSources(event.id)"
        >{{ t('usage.webSearch.sources', { count: event.source_count }) }}</button>
        <span v-else class="text-xs text-gray-500 dark:text-gray-400">
          {{ t('usage.webSearch.sources', { count: event.source_count }) }}
        </span>
        <ul v-if="expandedSources.has(event.id)" class="list-disc space-y-1 pl-5 text-xs">
          <li v-for="(source, index) in event.sources" :key="index" class="break-all">
            <a
              v-if="safeSourceUrl(source.url)"
              :href="safeSourceUrl(source.url)"
              target="_blank"
              rel="noopener noreferrer"
              class="text-primary-600 underline dark:text-primary-400"
              @click.stop
            >{{ source.title || source.url }}</a>
            <span v-else>{{ source.title || source.url }}</span>
          </li>
        </ul>
      </div>
    </component>
  </component>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import type { AdminWebSearchEvent } from '@/types'
import { formatDateTime } from '@/utils/format'

const props = defineProps<{
  events: AdminWebSearchEvent[]
  columnCount: number
  mobile: boolean
}>()
const { t } = useI18n()
const expandedSources = ref(new Set<number>())
const toggleSources = (id: number) => {
  const next = new Set(expandedSources.value)
  if (next.has(id)) next.delete(id)
  else next.add(id)
  expandedSources.value = next
}
watch(() => props.events, () => { expandedSources.value = new Set() }, { deep: true })

function safeSourceUrl(value: string): string | undefined {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:' ? url.href : undefined
  } catch {
    return undefined
  }
}
</script>
