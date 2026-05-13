<script setup lang="ts">
import { ref } from 'vue'
import { useSocket } from './composables/useSocket'
import { usePipeWireStore } from './stores/pipewire'
import GraphView from './components/GraphView.vue'
import NodeDetail from './components/NodeDetail.vue'
import RecordPanel from './components/RecordPanel.vue'
import PlayPanel from './components/PlayPanel.vue'

const { state } = useSocket()
const store = usePipeWireStore()

const statusColor: Record<string, string> = {
  connected: '#4ade80',
  connecting: '#facc15',
  disconnected: '#f87171',
}

const reloading = ref(false)
async function reload() {
  if (reloading.value) return
  reloading.value = true
  try {
    await fetch('/api/reload', { method: 'POST' })
  } catch {
    // ignore — server will re-emit init on the next poll if it succeeded
  } finally {
    reloading.value = false
  }
}
</script>

<template>
  <div class="app-layout">
    <!-- Topbar -->
    <header class="topbar">
      <span class="topbar-logo">Soundwires</span>
      <span class="topbar-status" :style="{ color: statusColor[state] }">
        ● {{ state }}
      </span>
      <label class="topbar-toggle">
        <input v-model="store.showInternal" type="checkbox" />
        Show internal nodes
      </label>
      <span class="topbar-count">
        {{ store.allNodes.length }} nodes · {{ store.links.length }} links
      </span>
      <button
        class="topbar-reload"
        :disabled="reloading"
        title="Restart pw-dump tracking and rebuild the graph"
        @click="reload"
      >
        {{ reloading ? '⟳ reloading…' : '⟳ reload' }}
      </button>
    </header>

    <!-- Main split -->
    <div class="main-split">
      <GraphView class="graph-area" />

      <!-- Sidebar -->
      <aside class="sidebar">
        <NodeDetail />
        <RecordPanel />
        <PlayPanel />
      </aside>
    </div>
  </div>
</template>

<style>
@layer components {
  .app-layout {
    display: flex;
    flex-direction: column;
    height: 100vh;
    overflow: hidden;
  }
  .topbar {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 8px 16px;
    background: #0f172a;
    border-bottom: 1px solid #1e293b;
    flex-shrink: 0;
    font-size: 13px;
  }
  .topbar-logo {
    font-weight: 800;
    font-size: 15px;
    color: #60a5fa;
    letter-spacing: -0.02em;
  }
  .topbar-status { font-size: 12px; }
  .topbar-toggle {
    display: flex;
    align-items: center;
    gap: 6px;
    color: #94a3b8;
    cursor: pointer;
    font-size: 12px;
  }
  .topbar-count { color: #475569; font-size: 12px; margin-left: auto; }
  .topbar-reload {
    background: #1e293b;
    color: #94a3b8;
    border: 1px solid #334155;
    border-radius: 4px;
    padding: 4px 10px;
    font-size: 12px;
    cursor: pointer;
    font-family: inherit;
  }
  .topbar-reload:hover:not(:disabled) { background: #334155; color: #e2e8f0; }
  .topbar-reload:disabled { opacity: 0.5; cursor: progress; }

  .main-split {
    display: flex;
    flex: 1;
    overflow: hidden;
  }
  .graph-area {
    flex: 1;
    min-width: 0;
  }
  .sidebar {
    width: 420px;
    flex-shrink: 0;
    background: #0f172a;
    border-left: 1px solid #1e293b;
    padding: 4px;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
}
</style>
