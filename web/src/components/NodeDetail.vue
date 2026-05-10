<script setup lang="ts">
import { computed } from 'vue'
import { usePipeWireStore } from '../stores/pipewire'
import PropValue from './PropValue.vue'

const store = usePipeWireStore()
const obj = computed(() => store.selected)

function flatProps(o: unknown, prefix = ''): [string, unknown][] {
  if (!o || typeof o !== 'object' || Array.isArray(o)) return []
  const result: [string, unknown][] = []
  for (const [k, v] of Object.entries(o as Record<string, unknown>)) {
    const key = prefix ? `${prefix}.${k}` : k
    if (v && typeof v === 'object' && !Array.isArray(v)) {
      result.push(...flatProps(v, key))
    } else {
      result.push([key, v])
    }
  }
  return result
}

const propRows = computed(() => {
  if (!obj.value) return []
  return flatProps({ ...obj.value.info, ...obj.value.props })
    .filter(([k]) => k !== 'params')
    .sort(([a], [b]) => a.localeCompare(b))
})

const shortType = computed(() => obj.value?.type.replace('PipeWire:Interface:', '') ?? '')

function isJson(v: unknown): boolean {
  if (Array.isArray(v) || (v !== null && typeof v === 'object')) return true
  if (typeof v === 'string' && (v.startsWith('{') || v.startsWith('['))) {
    try { const p: unknown = JSON.parse(v); return typeof p === 'object' && p !== null }
    catch { return false }
  }
  return false
}
</script>

<template>
  <div v-if="obj" class="detail-panel">
    <div class="detail-type">{{ shortType }}</div>
    <div class="detail-name">{{ store.nodeName(obj) }}</div>
    <div class="detail-id">id: {{ obj.id }}</div>

    <div class="detail-props">
      <div
        v-for="[key, val] in propRows"
        :key="key"
        class="detail-row"
        :class="isJson(val) ? 'detail-row--block' : 'detail-row--inline'"
      >
        <span class="detail-key">{{ key }}</span>
        <PropValue :value="val" />
      </div>
    </div>
  </div>
  <div v-else class="detail-empty">
    Click a node to inspect it
  </div>
</template>

<style>
@layer components {
  .detail-panel {
    padding: 16px;
    overflow-y: auto;
    flex: 1;
    min-height: 0;
  }
  .detail-type {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #60a5fa;
    font-weight: 700;
  }
  .detail-name {
    font-size: 15px;
    font-weight: 700;
    color: #e2e8f0;
    margin-top: 4px;
    word-break: break-word;
  }
  .detail-id {
    font-size: 11px;
    color: #64748b;
    margin-bottom: 12px;
  }
  .detail-props {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .detail-row {
    font-size: 11px;
    padding: 3px 0;
    border-bottom: 1px solid #1e293b;
  }
  .detail-row--inline {
    display: flex;
    gap: 8px;
    align-items: baseline;
    justify-content: space-between;
  }
  .detail-row--block {
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .detail-key {
    color: #94a3b8;
    flex-shrink: 0;
    word-break: break-all;
  }
  .detail-empty {
    padding: 24px 16px;
    color: #475569;
    font-size: 13px;
    text-align: center;
  }
}
</style>
