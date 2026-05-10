<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'

const props = defineProps<{ analyser: AnalyserNode | null }>()

const canvas = ref<HTMLCanvasElement | null>(null)
let rafId = 0
const rmsDb = ref(-Infinity)

function draw() {
  const an = props.analyser
  const cv = canvas.value
  if (!an || !cv) return

  const buf = new Uint8Array(an.fftSize)
  an.getByteTimeDomainData(buf)

  // RMS in dBFS
  let sum = 0
  for (let i = 0; i < buf.length; i++) {
    const s = (buf[i] - 128) / 128
    sum += s * s
  }
  const rms = Math.sqrt(sum / buf.length)
  rmsDb.value = rms > 0 ? 20 * Math.log10(rms) : -Infinity

  const ctx = cv.getContext('2d')!
  const w = cv.width
  const h = cv.height
  ctx.clearRect(0, 0, w, h)

  // Waveform
  ctx.strokeStyle = '#60a5fa'
  ctx.lineWidth = 1.5
  ctx.beginPath()
  const step = w / buf.length
  for (let i = 0; i < buf.length; i++) {
    const y = ((buf[i] / 255) * h)
    i === 0 ? ctx.moveTo(0, y) : ctx.lineTo(i * step, y)
  }
  ctx.stroke()

  rafId = requestAnimationFrame(draw)
}

function startLoop() {
  cancelAnimationFrame(rafId)
  if (props.analyser) rafId = requestAnimationFrame(draw)
}

watch(() => props.analyser, (an) => {
  cancelAnimationFrame(rafId)
  if (an) startLoop()
}, { immediate: true })

onUnmounted(() => cancelAnimationFrame(rafId))

function dbToPercent(db: number): number {
  if (!isFinite(db)) return 0
  return Math.max(0, Math.min(1, (db + 60) / 60))
}

function dbColor(db: number): string {
  if (db > -6) return '#ef4444'
  if (db > -18) return '#facc15'
  return '#4ade80'
}
</script>

<template>
  <div class="lm-wrap">
    <canvas ref="canvas" class="lm-wave" width="320" height="48" />
    <div class="lm-bar-wrap">
      <div
        class="lm-bar"
        :style="{ width: `${dbToPercent(rmsDb) * 100}%`, background: dbColor(rmsDb) }"
      />
    </div>
    <div class="lm-db">{{ isFinite(rmsDb) ? rmsDb.toFixed(1) + ' dBFS' : '—' }}</div>
  </div>
</template>

<style>
@layer components {
  .lm-wrap {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }
  .lm-wave {
    width: 100%;
    height: 48px;
    background: #0f172a;
    border-radius: 4px;
    border: 1px solid #1e293b;
    display: block;
  }
  .lm-bar-wrap {
    height: 6px;
    background: #1e293b;
    border-radius: 3px;
    overflow: hidden;
  }
  .lm-bar {
    height: 100%;
    border-radius: 3px;
    transition: width 0.05s linear, background 0.1s;
  }
  .lm-db {
    font-size: 10px;
    color: #64748b;
    text-align: right;
  }
}
</style>
