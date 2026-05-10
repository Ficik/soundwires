<script setup lang="ts">
import { computed } from 'vue'
import { Handle, Position } from '@vue-flow/core'
import type { NodeProps } from '@vue-flow/core'
import type { PWObject } from '../stores/pipewire'

export interface CustomNodeData {
  obj: PWObject
  label: string
  mediaClass: string
  state: string
  inputPorts: PWObject[]
  outputPorts: PWObject[]
}

const props = defineProps<NodeProps<CustomNodeData>>()

const borderColor = computed(() => {
  const mc = props.data.mediaClass
  if (mc.includes('Sink')) return '#f87171'     // red
  if (mc.includes('Source')) return '#4ade80'   // green
  if (mc.includes('Filter')) return '#a78bfa'   // purple
  return '#60a5fa'                              // blue default
})

const stateColor = computed(() => {
  switch (props.data.state) {
    case 'running': return '#4ade80'
    case 'idle': return '#94a3b8'
    case 'error': return '#f87171'
    default: return '#64748b'
  }
})

const shortClass = computed(() => {
  const mc = props.data.mediaClass
  return mc.replace('Audio/', '').replace('Video/', 'V/').replace('Midi/', 'M/')
})
</script>

<template>
  <div
    class="pw-node"
    :style="{ borderColor }"
    @click="() => {}"
  >
    <!-- Input port handles on the left -->
    <Handle
      v-for="port in data.inputPorts"
      :id="`port-${port.id}`"
      :key="port.id"
      type="target"
      :position="Position.Left"
      :style="{ top: `${50}%` }"
    />

    <div class="pw-node-header">
      <span class="pw-node-class" :style="{ color: borderColor }">{{ shortClass }}</span>
      <span class="pw-node-state" :style="{ background: stateColor }" />
    </div>
    <div class="pw-node-label">{{ data.label }}</div>
    <div class="pw-node-id">#{{ data.obj.id }}</div>

    <!-- Output port handles on the right -->
    <Handle
      v-for="port in data.outputPorts"
      :id="`port-${port.id}`"
      :key="port.id"
      type="source"
      :position="Position.Right"
      :style="{ top: `${50}%` }"
    />
  </div>
</template>

<style>
@layer components {
  .pw-node {
    background: #1e293b;
    border: 2px solid #60a5fa;
    border-radius: 8px;
    padding: 10px 16px;
    min-width: 160px;
    max-width: 240px;
    cursor: pointer;
    user-select: none;
    box-shadow: 0 4px 12px rgba(0,0,0,0.4);
    transition: box-shadow 0.15s;
  }
  .pw-node:hover {
    box-shadow: 0 6px 20px rgba(0,0,0,0.6);
  }
  .pw-node-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;
  }
  .pw-node-class {
    font-size: 10px;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .pw-node-state {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .pw-node-label {
    font-size: 13px;
    font-weight: 600;
    color: #e2e8f0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .pw-node-id {
    font-size: 10px;
    color: #64748b;
    margin-top: 2px;
  }
}
</style>
