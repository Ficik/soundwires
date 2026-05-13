import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { Node as FlowNode, Edge } from '@vue-flow/core'
import dagre from 'dagre'

export type PWObjectType =
  | 'PipeWire:Interface:Node'
  | 'PipeWire:Interface:Port'
  | 'PipeWire:Interface:Link'
  | 'PipeWire:Interface:Device'
  | 'PipeWire:Interface:Client'
  | string

export interface PWObject {
  id: number
  type: PWObjectType
  permissions: string[]
  info?: Record<string, unknown>
  props?: Record<string, unknown>
}

export interface WsEvent {
  type: 'init' | 'added' | 'changed' | 'removed'
  objects?: PWObject[]
  object?: PWObject
  id?: number
}

function getProps(obj: PWObject): Record<string, unknown> {
  // PW objects may have props at top level or inside info
  const info = obj.info as Record<string, unknown> | undefined
  const infoProps = info?.['props'] as Record<string, unknown> | undefined
  const topProps = obj.props as Record<string, unknown> | undefined
  return { ...(topProps ?? {}), ...(infoProps ?? {}) }
}

function mediaClass(obj: PWObject): string {
  return (getProps(obj)['media.class'] as string) ?? ''
}

function nodeName(obj: PWObject): string {
  const p = getProps(obj)
  return (
    (p['node.description'] as string) ??
    (p['node.nick'] as string) ??
    (p['node.name'] as string) ??
    (p['port.name'] as string) ??
    `#${obj.id}`
  )
}

function isMediaNode(obj: PWObject): boolean {
  if (obj.type !== 'PipeWire:Interface:Node') return false
  const mc = mediaClass(obj)
  return mc !== ''
}

function isPort(obj: PWObject): boolean {
  return obj.type === 'PipeWire:Interface:Port'
}

function isLink(obj: PWObject): boolean {
  return obj.type === 'PipeWire:Interface:Link'
}

const NODE_WIDTH = 240
const NODE_HEIGHT = 100

function layoutWithDagre(nodes: FlowNode[], edges: Edge[]): FlowNode[] {
  const g = new dagre.graphlib.Graph()
  g.setDefaultEdgeLabel(() => ({}))
  g.setGraph({ rankdir: 'LR', ranksep: 120, nodesep: 40 })

  for (const n of nodes) {
    g.setNode(n.id, { width: NODE_WIDTH, height: NODE_HEIGHT })
  }
  for (const e of edges) {
    g.setEdge(e.source, e.target)
  }

  dagre.layout(g)

  return nodes.map(n => {
    const pos = g.node(n.id)
    return {
      ...n,
      position: { x: pos.x - NODE_WIDTH / 2, y: pos.y - NODE_HEIGHT / 2 },
    }
  })
}

export const usePipeWireStore = defineStore('pipewire', () => {
  const objects = ref(new Map<number, PWObject>())
  const showInternal = ref(true)
  const selectedId = ref<number | null>(null)

  function applyEvent(event: WsEvent) {
    if (event.type === 'init' && event.objects) {
      objects.value.clear()
      for (const obj of event.objects) {
        objects.value.set(obj.id, obj)
      }
    } else if ((event.type === 'added' || event.type === 'changed') && event.object) {
      objects.value.set(event.object.id, event.object)
    } else if (event.type === 'removed' && event.id !== undefined) {
      objects.value.delete(event.id)
      if (selectedId.value === event.id) selectedId.value = null
    }
  }

  const mediaNodes = computed(() =>
    [...objects.value.values()].filter(isMediaNode),
  )

  const allNodes = computed(() =>
    showInternal.value
      ? [...objects.value.values()].filter(o => o.type === 'PipeWire:Interface:Node')
      : mediaNodes.value,
  )

  const ports = computed(() =>
    [...objects.value.values()].filter(isPort),
  )

  const links = computed(() =>
    [...objects.value.values()].filter(isLink),
  )

  // Build the set of visible node IDs for edge filtering
  const visibleNodeIds = computed(() => new Set(allNodes.value.map(n => n.id)))

  // Ports grouped by node id
  const portsByNode = computed(() => {
    const map = new Map<number, PWObject[]>()
    for (const port of ports.value) {
      const nodeId = (getProps(port)['node.id'] as number) ?? -1
      if (!map.has(nodeId)) map.set(nodeId, [])
      map.get(nodeId)!.push(port)
    }
    return map
  })

  const flowNodes = computed((): FlowNode[] => {
    return allNodes.value.map(obj => {
      const mc = mediaClass(obj)
      const info = obj.info as Record<string, unknown> | undefined
      const state = (info?.['state'] as string) ?? 'unknown'
      const inputPorts = (portsByNode.value.get(obj.id) ?? []).filter(
        p => (getProps(p)['port.direction'] as string) === 'in',
      )
      const outputPorts = (portsByNode.value.get(obj.id) ?? []).filter(
        p => (getProps(p)['port.direction'] as string) === 'out',
      )

      return {
        id: String(obj.id),
        type: 'custom',
        position: { x: 0, y: 0 },
        data: {
          obj,
          label: nodeName(obj),
          mediaClass: mc,
          state,
          inputPorts,
          outputPorts,
        },
      }
    })
  })

  const flowEdges = computed((): Edge[] => {
    const edges: Edge[] = []
    for (const link of links.value) {
      const info = link.info as Record<string, unknown> | undefined
      if (!info) continue
      const srcNode = info['output-node-id'] as number
      const srcPort = info['output-port-id'] as number
      const dstNode = info['input-node-id'] as number
      const dstPort = info['input-port-id'] as number
      if (!visibleNodeIds.value.has(srcNode) || !visibleNodeIds.value.has(dstNode)) continue
      edges.push({
        id: `link-${link.id}`,
        source: String(srcNode),
        sourceHandle: `port-${srcPort}`,
        target: String(dstNode),
        targetHandle: `port-${dstPort}`,
        animated: true,
        style: { stroke: '#60a5fa', strokeWidth: 2 },
      })
    }

    // Synthetic loopback edges: sink and source nodes sharing the same node.link-group
    // are two sides of one loopback — connect them so the graph shows
    // stream → sink → source → next node instead of two disconnected clusters.
    const linkGroups = new Map<string, PWObject[]>()
    for (const node of allNodes.value) {
      const linkGroup = getProps(node)['node.link-group'] as string | undefined
      if (linkGroup === undefined) continue
      if (!linkGroups.has(linkGroup)) linkGroups.set(linkGroup, [])
      linkGroups.get(linkGroup)!.push(node)
    }
    for (const group of linkGroups.values()) {
      if (group.length < 2) continue
      const sinks = group.filter(n => ['Audio/Sink', 'Stream/Input/Audio'].includes(mediaClass(n)))
      const sources = group.filter(n => ['Audio/Source', 'Stream/Output/Audio'].includes(mediaClass(n)))
      for (const sink of sinks) {
        for (const source of sources) {
          edges.push({
            id: `loopback-${sink.id}-${source.id}`,
            source: String(sink.id),
            target: String(source.id),
            animated: true,
            style: { stroke: '#a78bfa', strokeWidth: 2, strokeDasharray: '6,3' },
          })
        }
      }
    }

    return edges
  })

  const laidOutNodes = computed(() => layoutWithDagre(flowNodes.value, flowEdges.value))

  const selected = computed(() =>
    selectedId.value !== null ? objects.value.get(selectedId.value) : null,
  )

  // Returns the best identifier for pw-cat --target.
  // node.name is what WirePlumber matches on; object.serial is the fallback.
  // The top-level PWObject.id (object.id) can differ from object.serial in
  // PipeWire 1.x and must NOT be used for routing.
  function pwTarget(obj: PWObject): string {
    const p = getProps(obj)
    const name = p['node.name'] as string | undefined
    if (name) return name
    const serial = p['object.serial']
    return String(serial ?? obj.id)
  }

  return {
    objects,
    showInternal,
    selectedId,
    selected,
    allNodes,
    ports,
    links,
    portsByNode,
    flowNodes,
    flowEdges,
    laidOutNodes,
    applyEvent,
    getProps,
    nodeName,
    mediaClass,
    pwTarget,
  }
})
