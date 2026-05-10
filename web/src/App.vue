<script setup lang="ts">
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
