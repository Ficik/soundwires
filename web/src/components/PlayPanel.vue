<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { usePipeWireStore } from '../stores/pipewire'

const store = usePipeWireStore()

const freq = ref(440)
const amp = ref(0.5)
const channels = ref(2)
const rate = ref(48000)
const playing = ref(false)
const error = ref<string | null>(null)

// Tracks the node that is currently receiving sine (locked in on Start).
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

let ctrl: AbortController | null = null

async function start() {
  if (!target.value) return
  error.value = null
  activeNode.value = { ...target.value }
  playing.value = true
  ctrl = new AbortController()

  const params = new URLSearchParams({
    targetId: activeNode.value.pwTarget,
    freq: String(freq.value),
    amp: String(amp.value),
    channels: String(channels.value),
    rate: String(rate.value),
  })

  try {
    const res = await fetch(`/api/play?${params.toString()}`, { signal: ctrl.signal })
    if (!res.ok) {
      error.value = await res.text()
    }
  } catch (e) {
    if ((e as Error).name !== 'AbortError') {
      error.value = String(e)
    }
  } finally {
    playing.value = false
    activeNode.value = null
    ctrl = null
  }
}

function stop() {
  ctrl?.abort()
}

// Clear error when the target node changes (user selected a different node).
watch(() => store.selectedId, () => {
  if (!playing.value) error.value = null
})
</script>

<template>
  <div class="panel-section">
    <div class="panel-title">Inject Sine Wave</div>

    <div v-if="!target" class="panel-hint">Select a node in the graph</div>
    <div v-else class="panel-target">
      <span class="panel-target-name">{{ target.name }}</span>
      <span class="panel-target-id">#{{ target.id }}</span>
    </div>

    <div class="panel-row">
      <label class="panel-label">Hz</label>
      <input
        v-model.number="freq"
        type="range"
        min="50"
        max="4000"
        step="10"
        class="panel-range"
        :disabled="playing"
      />
      <span class="panel-val-label">{{ freq }}</span>
    </div>

    <div class="panel-row">
      <label class="panel-label">Amp</label>
      <input
        v-model.number="amp"
        type="range"
        min="0.05"
        max="1"
        step="0.05"
        class="panel-range"
        :disabled="playing"
      />
      <span class="panel-val-label">{{ amp.toFixed(2) }}</span>
    </div>

    <div class="panel-row">
      <label class="panel-label">Ch</label>
      <select v-model="channels" class="panel-select-sm" :disabled="playing">
        <option :value="1">Mono</option>
        <option :value="2">Stereo</option>
      </select>
      <label class="panel-label">Rate</label>
      <select v-model="rate" class="panel-select-sm" :disabled="playing">
        <option :value="44100">44100</option>
        <option :value="48000">48000</option>
        <option :value="96000">96000</option>
      </select>
    </div>

    <button
      class="panel-btn"
      :class="playing ? 'panel-btn-stop' : 'panel-btn-start'"
      :disabled="!target && !playing"
      @click="playing ? stop() : start()"
    >
      {{ playing ? 'Stop' : 'Inject' }}
    </button>

    <div v-if="error" class="panel-error">{{ error }}</div>
    <div v-if="playing" class="panel-recording">▶ Injecting @ {{ freq }} Hz…</div>
  </div>
</template>

<style>
@layer components {
  .panel-range {
    flex: 1;
    accent-color: #60a5fa;
  }
  .panel-val-label {
    font-size: 11px;
    color: #e2e8f0;
    min-width: 36px;
    text-align: right;
    flex-shrink: 0;
  }
}
</style>
