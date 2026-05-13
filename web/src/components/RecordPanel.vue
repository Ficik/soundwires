<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { usePipeWireStore } from '../stores/pipewire'
import { useMonitor } from '../composables/useMonitor'
import LevelMeter from './LevelMeter.vue'

const store = usePipeWireStore()
const { active, error, analyser, muted, start, stop } = useMonitor()

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
      <button
        type="button"
        class="mute-btn"
        :class="{ 'mute-btn--on': muted }"
        :title="muted ? 'Unmute' : 'Mute'"
        :aria-pressed="muted"
        @click="muted = !muted"
      >
        <svg v-if="!muted" viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
          <path
            fill="currentColor"
            d="M3 10v4a1 1 0 0 0 1 1h3l4 4a1 1 0 0 0 1.7-.7V5.7A1 1 0 0 0 11 5L7 9H4a1 1 0 0 0-1 1Zm12.5 2a4 4 0 0 0-2-3.46v6.93A4 4 0 0 0 15.5 12Zm-2-7.07v2.06a7 7 0 0 1 0 13.02v2.06a9 9 0 0 0 0-17.14Z"
          />
        </svg>
        <svg v-else viewBox="0 0 24 24" width="16" height="16" aria-hidden="true">
          <path
            fill="currentColor"
            d="M3 10v4a1 1 0 0 0 1 1h3l4 4a1 1 0 0 0 1.7-.7V5.7A1 1 0 0 0 11 5L7 9H4a1 1 0 0 0-1 1Zm17.7 6.3-2.6-2.6 2.6-2.6a1 1 0 0 0-1.4-1.4L16.7 12.3 14.1 9.7a1 1 0 1 0-1.4 1.4l2.6 2.6-2.6 2.6a1 1 0 1 0 1.4 1.4l2.6-2.6 2.6 2.6a1 1 0 0 0 1.4-1.4Z"
          />
        </svg>
      </button>
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

<style>
@layer components {
  .mute-btn {
    margin-left: auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    padding: 0;
    border-radius: 6px;
    background: #1e293b;
    border: 1px solid #334155;
    color: #94a3b8;
    cursor: pointer;
    transition: background 0.12s, color 0.12s, border-color 0.12s;
  }
  .mute-btn:hover { background: #334155; color: #e2e8f0; }
  .mute-btn--on {
    background: #7f1d1d;
    border-color: #b91c1c;
    color: #fecaca;
  }
  .mute-btn--on:hover { background: #991b1b; color: #fff; }
}
</style>
