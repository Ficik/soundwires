<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue'
import { useShiki } from '../composables/useShiki'

const props = defineProps<{ value: unknown }>()

const hl = useShiki()

// Returns formatted JSON string if value should be rendered as JSON, null otherwise.
const jsonStr = computed((): string | null => {
  const v = props.value
  if (Array.isArray(v) || (v !== null && typeof v === 'object')) {
    return JSON.stringify(v, null, 2)
  }
  if (typeof v === 'string' && (v.startsWith('{') || v.startsWith('['))) {
    try {
      const parsed: unknown = JSON.parse(v)
      if (typeof parsed === 'object' && parsed !== null) {
        return JSON.stringify(parsed, null, 2)
      }
    } catch { /* not valid JSON */ }
  }
  return null
})

const highlighted = ref<string | null>(null)

watchEffect(() => {
  const src = jsonStr.value
  if (!src) { highlighted.value = null; return }
  const h = hl.value
  if (!h) { highlighted.value = null; return }
  highlighted.value = h.codeToHtml(src, { lang: 'json', theme: 'github-dark' })
})
</script>

<template>
  <!-- Primitive: render inline -->
  <span v-if="jsonStr === null" class="pv-inline">{{ value }}</span>

  <!-- JSON: render highlighted block, fall back to raw until Shiki loads -->
  <div v-else class="pv-block">
    <pre v-if="!highlighted" class="pv-raw">{{ jsonStr }}</pre>
    <div v-else class="pv-hl" v-html="highlighted" />
  </div>
</template>

<style>
@layer components {
  .pv-inline {
    color: #e2e8f0;
    word-break: break-all;
  }
  .pv-block {
    margin-top: 4px;
    width: 100%;
  }
  .pv-raw {
    font-size: 11px;
    color: #94a3b8;
    white-space: pre-wrap;
    word-break: break-all;
    margin: 0;
    padding: 6px 8px;
    background: #0d1117;
    border-radius: 4px;
    border: 1px solid #1e293b;
  }
  .pv-hl .shiki {
    background: #0d1117 !important;
    border-radius: 4px;
    border: 1px solid #1e293b;
    padding: 6px 8px;
    font-size: 11px;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-word;
  }
  .pv-hl .shiki code {
    font-family: ui-monospace, 'Cascadia Code', 'Source Code Pro', Menlo, monospace;
  }
}
</style>
