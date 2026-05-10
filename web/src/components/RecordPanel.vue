<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { usePipeWireStore } from '../stores/pipewire'
import { useMonitor } from '../composables/useMonitor'
import LevelMeter from './LevelMeter.vue'

const store = usePipeWireStore()
const { active, error, analyser, start, stop } = useMonitor()

const channels = ref(2)
const rate = ref(48000)

const activeNode = ref<{ id: number; name: string; pwTarget: string } | null>(null)

const target = computed(() => {
  if (activeNode.value) return activeNode.value
  if (store.selected) return {
    id: store.selected.id,
    name: store.nodeName(store.selected),
    pwTarget: store.pwTarget(store.selected),
  }
  return null
})

watch(active, v => { if (!v) activeNode.value = null })

function toggle() {
  if (active.value) {
    stop()
    activeNode.value = null
  } else if (target.value) {
    activeNode.value = { ...target.value }
    start(target.value.pwTarget, channels.value, rate.value)
  }
}
</script>

<template>
  <div class="panel-section">
    <div class="panel-title">Monitor Output</div>

    <div v-if="!target" class="panel-hint">Select a node in the graph</div>
    <div v-else class="panel-target">
      <span class="panel-target-name">{{ target.name }}</span>
      <span class="panel-target-id">#{{ target.id }}</span>
    </div>

    <div class="panel-row">
      <label class="panel-label">Ch</label>
      <select v-model="channels" class="panel-select-sm" :disabled="active">
        <option :value="1">Mono</option>
        <option :value="2">Stereo</option>
      </select>
      <label class="panel-label">Rate</label>
      <select v-model="rate" class="panel-select-sm" :disabled="active">
        <option :value="44100">44100</option>
        <option :value="48000">48000</option>
        <option :value="96000">96000</option>
      </select>
    </div>

    <button
      class="panel-btn"
      :class="active ? 'panel-btn-stop' : 'panel-btn-start'"
      :disabled="!target && !active"
      @click="toggle"
    >
      {{ active ? 'Stop' : 'Monitor' }}
    </button>

    <LevelMeter v-if="active" :analyser="analyser" />

    <div v-if="error" class="panel-error">{{ error }}</div>
    <div v-if="active && !error" class="panel-recording">● Monitoring…</div>
  </div>
</template>
