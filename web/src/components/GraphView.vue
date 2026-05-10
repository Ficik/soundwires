<script setup lang="ts">
import { computed, markRaw } from 'vue'
import { VueFlow, useVueFlow } from '@vue-flow/core'
import type { NodeMouseEvent, NodeTypesObject } from '@vue-flow/core'
import { Background } from '@vue-flow/background'
import { MiniMap } from '@vue-flow/minimap'
import { Controls } from '@vue-flow/controls'
import CustomNode from './CustomNode.vue'
import { usePipeWireStore } from '../stores/pipewire'

const store = usePipeWireStore()
const { fitView } = useVueFlow()

const nodeTypes: NodeTypesObject = { custom: markRaw(CustomNode) as NodeTypesObject['custom'] }

const nodes = computed(() => store.laidOutNodes)
const edges = computed(() => store.flowEdges)

function onNodeClick(event: NodeMouseEvent) {
  store.selectedId = parseInt(event.node.id)
}

function refitView() {
  setTimeout(() => fitView({ padding: 0.2 }), 50)
}
</script>

<template>
  <VueFlow
    :nodes="nodes"
    :edges="edges"
    :node-types="nodeTypes"
    :default-edge-options="{ animated: true }"
    fit-view-on-init
    class="pw-graph"
    @node-click="(e) => onNodeClick(e)"
    @nodes-initialized="refitView"
  >
    <Background pattern-color="#334155" :gap="24" />
    <MiniMap
      node-color="#334155"
      :style="{ background: '#1e293b', border: '1px solid #334155' }"
    />
    <Controls :style="{ background: '#1e293b', border: '1px solid #334155' }" />
  </VueFlow>
</template>

<style>
@layer components {
  .pw-graph {
    flex: 1;
    background: #0f1117;
  }
  .vue-flow__controls-button {
    background: #1e293b;
    border: 1px solid #334155;
    fill: #94a3b8;
  }
  .vue-flow__controls-button:hover {
    background: #334155;
  }
}
</style>
