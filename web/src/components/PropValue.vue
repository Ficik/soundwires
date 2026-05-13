<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue'
import { useShiki } from '../composables/useShiki'

const props = defineProps<{ value: unknown }>()

const hl = useShiki()
const open = ref(false)

// Returns the structured value to render as JSON (parsed if input was a JSON-ish string),
// or null if the value should render inline as a primitive.
const jsonValue = computed((): unknown => {
  const v = props.value
  if (Array.isArray(v) || (v !== null && typeof v === 'object')) return v
  if (typeof v === 'string' && (v.startsWith('{') || v.startsWith('['))) {
    try {
      const parsed: unknown = JSON.parse(v)
      if (typeof parsed === 'object' && parsed !== null) return parsed
    } catch { /* not valid JSON */ }
  }
  return null
})

const jsonStr = computed((): string | null => {
  const v = jsonValue.value
  return v === null ? null : JSON.stringify(v, null, 2)
})

const summary = computed((): string => {
  const v = jsonValue.value
  if (Array.isArray(v)) return `[ ] ${v.length} ${v.length === 1 ? 'item' : 'items'}`
  if (v !== null && typeof v === 'object') {
    const n = Object.keys(v).length
    return `{ } ${n} ${n === 1 ? 'key' : 'keys'}`
  }
  return ''
})

const highlighted = ref<string | null>(null)

watchEffect(() => {
  // Only highlight when expanded — saves work for huge collapsed blocks.
  if (!open.value) { highlighted.value = null; return }
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

  <!-- JSON: collapsible block, collapsed by default -->
  <div v-else class="pv-block">
    <button type="button" class="pv-toggle" @click="open = !open">
      <span class="pv-toggle-icon">{{ open ? '▼' : '▶' }}</span>
      <span class="pv-toggle-label">{{ summary }}</span>
    </button>
    <template v-if="open">
      <pre v-if="!highlighted" class="pv-raw">{{ jsonStr }}</pre>
      <div v-else class="pv-hl" v-html="highlighted" />
    </template>
  </div>
</template>

<style>
@layer components {
  .pv-inline {
    color: #e2e8f0;
    word-break: break-all;
  }
  .pv-block {
    min-width: 0;
  }
  .pv-toggle {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    background: transparent;
    border: none;
    padding: 2px 0;
    cursor: pointer;
    color: #64748b;
    font-size: 10px;
    font-family: ui-monospace, 'Cascadia Code', 'Source Code Pro', Menlo, monospace;
  }
  .pv-toggle:hover { color: #94a3b8; }
  .pv-toggle-icon {
    display: inline-block;
    width: 10px;
    font-size: 8px;
  }
  .pv-toggle-label { letter-spacing: 0.02em; }
  .pv-raw {
    font-size: 11px;
    color: #94a3b8;
    white-space: pre-wrap;
    word-break: break-all;
    margin: 4px 0 0 0;
    padding: 6px 8px;
    background: #0d1117;
    border-radius: 4px;
    border: 1px solid #1e293b;
  }
  .pv-hl {
    margin-top: 4px;
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
