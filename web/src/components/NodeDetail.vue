<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { usePipeWireStore, type PWObject } from '../stores/pipewire'
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

function propRowsFor(o: PWObject): [string, unknown][] {
  return flatProps({ ...o.info, ...o.props })
    .filter(([k]) => k !== 'params')
    .sort(([a], [b]) => a.localeCompare(b))
}

const propRows = computed(() => obj.value ? propRowsFor(obj.value) : [])

const shortType = computed(() => obj.value?.type.replace('PipeWire:Interface:', '') ?? '')

interface PortGroup {
  port: PWObject
  direction: string
  channel: string
  name: string
}

const ports = computed((): PortGroup[] => {
  if (!obj.value) return []
  const list = store.portsByNode.get(obj.value.id) ?? []
  return list
    .map((p): PortGroup => {
      const props = store.getProps(p)
      return {
        port: p,
        direction: (props['port.direction'] as string) ?? '',
        channel: (props['audio.channel'] as string) ?? '',
        name: (props['port.name'] as string) ?? `#${p.id}`,
      }
    })
    .sort((a, b) => {
      if (a.direction !== b.direction) return a.direction === 'in' ? -1 : 1
      return a.name.localeCompare(b.name)
    })
})

const openPorts = ref(new Set<number>())
function togglePort(id: number) {
  if (openPorts.value.has(id)) openPorts.value.delete(id)
  else openPorts.value.add(id)
  openPorts.value = new Set(openPorts.value)
}

watch(() => obj.value?.id, () => { openPorts.value = new Set() })
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
      >
        <span class="detail-key">{{ key }}</span>
        <PropValue :value="val" />
      </div>
    </div>

    <div v-if="ports.length" class="detail-section">
      <div class="detail-section-title">Ports ({{ ports.length }})</div>
      <div class="detail-ports">
        <div
          v-for="p in ports"
          :key="p.port.id"
          class="port-card"
        >
          <button
            type="button"
            class="port-head"
            @click="togglePort(p.port.id)"
          >
            <span class="port-toggle">{{ openPorts.has(p.port.id) ? '▼' : '▶' }}</span>
            <span class="port-dir" :class="`port-dir--${p.direction}`">
              {{ p.direction === 'in' ? '→' : p.direction === 'out' ? '←' : '·' }}
              {{ p.direction || '?' }}
            </span>
            <span class="port-name">{{ p.name }}</span>
            <span v-if="p.channel" class="port-channel">{{ p.channel }}</span>
            <span class="port-id">#{{ p.port.id }}</span>
          </button>
          <div v-if="openPorts.has(p.port.id)" class="port-props">
            <div
              v-for="[key, val] in propRowsFor(p.port)"
              :key="key"
              class="detail-row"
            >
              <span class="detail-key">{{ key }}</span>
              <PropValue :value="val" />
            </div>
          </div>
        </div>
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
    display: flex;
    gap: 8px;
    align-items: baseline;
    justify-content: space-between;
  }
  .detail-row .pv-block {
    display: flex;
    flex-direction: column;
    align-items: flex-end;
    flex: 1;
    min-width: 0;
  }
  .detail-row .pv-raw,
  .detail-row .pv-hl {
    align-self: stretch;
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
  .detail-section {
    margin-top: 16px;
  }
  .detail-section-title {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: #60a5fa;
    font-weight: 700;
    margin-bottom: 6px;
  }
  .detail-ports {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .port-card {
    background: #0d1117;
    border: 1px solid #1e293b;
    border-radius: 4px;
  }
  .port-head {
    display: flex;
    align-items: center;
    gap: 8px;
    width: 100%;
    background: transparent;
    border: none;
    color: inherit;
    padding: 6px 8px;
    cursor: pointer;
    font-family: inherit;
    font-size: 11px;
    text-align: left;
  }
  .port-head:hover { background: #131c2b; }
  .port-toggle {
    color: #64748b;
    font-size: 8px;
    width: 10px;
    display: inline-block;
  }
  .port-dir {
    font-size: 10px;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-weight: 700;
    flex-shrink: 0;
  }
  .port-dir--in { color: #4ade80; }
  .port-dir--out { color: #f87171; }
  .port-name {
    color: #e2e8f0;
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .port-channel {
    color: #a78bfa;
    font-weight: 600;
    flex-shrink: 0;
  }
  .port-id {
    color: #475569;
    font-size: 10px;
    flex-shrink: 0;
  }
  .port-props {
    padding: 4px 8px 6px 8px;
    border-top: 1px solid #1e293b;
  }
}
</style>
