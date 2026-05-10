import { ref, shallowRef } from 'vue'

export interface MonitorState {
  active: boolean
  error: string | null
  analyser: AnalyserNode | null
  channels: number
  rate: number
}

export function useMonitor() {
  const active = ref(false)
  const error = ref<string | null>(null)
  const analyser = shallowRef<AnalyserNode | null>(null)

  let ws: WebSocket | null = null
  let ctx: AudioContext | null = null
  let analyserNode: AnalyserNode | null = null
  let nextPlayAt = 0
  let channels = 2
  let rate = 48000

  function start(target: string, ch: number, sampleRate: number) {
    stop()
    error.value = null
    channels = ch
    rate = sampleRate

    const params = new URLSearchParams({ target, channels: String(ch), rate: String(sampleRate) })
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    ws = new WebSocket(`${proto}://${location.host}/ws/monitor?${params}`)
    ws.binaryType = 'arraybuffer'

    ws.onmessage = (ev) => {
      if (typeof ev.data === 'string') {
        // JSON message: either format metadata or error
        try {
          const msg = JSON.parse(ev.data as string) as Record<string, unknown>
          if (msg.error) {
            error.value = String(msg.error)
            active.value = false
            cleanupAudio()
            return
          }
          // Format metadata — set up AudioContext now
          channels = (msg.channels as number) || ch
          rate = (msg.rate as number) || sampleRate
          setupAudio()
          active.value = true
        } catch {
          // ignore malformed
        }
        return
      }

      // Binary PCM frame — s16le interleaved
      if (!ctx || !analyserNode) return
      const raw = new Int16Array(ev.data as ArrayBuffer)
      const frames = raw.length / channels
      const audioBuffer = ctx.createBuffer(channels, frames, rate)
      for (let c = 0; c < channels; c++) {
        const ch_data = audioBuffer.getChannelData(c)
        for (let i = 0; i < frames; i++) {
          ch_data[i] = raw[i * channels + c] / 32768
        }
      }

      const src = ctx.createBufferSource()
      src.buffer = audioBuffer
      src.connect(analyserNode)
      src.connect(ctx.destination)

      const now = ctx.currentTime
      if (nextPlayAt < now) nextPlayAt = now + 0.05 // small buffer to avoid glitches
      src.start(nextPlayAt)
      nextPlayAt += audioBuffer.duration
    }

    ws.onerror = () => {
      error.value = 'WebSocket connection error'
      active.value = false
      cleanupAudio()
    }

    ws.onclose = () => {
      if (active.value) {
        // Unexpected close
        error.value = error.value ?? 'Connection closed unexpectedly'
      }
      active.value = false
      cleanupAudio()
    }
  }

  function setupAudio() {
    ctx = new AudioContext({ sampleRate: rate })
    analyserNode = ctx.createAnalyser()
    analyserNode.fftSize = 2048
    analyserNode.connect(ctx.destination)
    analyser.value = analyserNode
    nextPlayAt = 0
  }

  function cleanupAudio() {
    analyser.value = null
    analyserNode = null
    ctx?.close().catch(() => {})
    ctx = null
    nextPlayAt = 0
  }

  function stop() {
    active.value = false
    ws?.close()
    ws = null
    cleanupAudio()
  }

  return { active, error, analyser, start, stop }
}
