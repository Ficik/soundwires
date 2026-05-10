import { ref, onMounted, onUnmounted } from 'vue'
import { usePipeWireStore } from '../stores/pipewire'
import type { WsEvent } from '../stores/pipewire'

export type ConnectionState = 'connecting' | 'connected' | 'disconnected'

export function useSocket() {
  const state = ref<ConnectionState>('disconnected')
  const store = usePipeWireStore()
  let ws: WebSocket | null = null
  let retryTimer: ReturnType<typeof setTimeout> | null = null

  function connect() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url = `${proto}://${location.host}/ws`
    state.value = 'connecting'
    ws = new WebSocket(url)

    ws.onopen = () => { state.value = 'connected' }

    ws.onmessage = (e: MessageEvent) => {
      try {
        const event = JSON.parse(e.data as string) as WsEvent
        store.applyEvent(event)
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      state.value = 'disconnected'
      retryTimer = setTimeout(connect, 2000)
    }

    ws.onerror = () => {
      ws?.close()
    }
  }

  onMounted(connect)

  onUnmounted(() => {
    if (retryTimer) clearTimeout(retryTimer)
    ws?.close()
  })

  return { state }
}
