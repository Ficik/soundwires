import { ref } from 'vue'

export function useAudio() {
  const recording = ref(false)
  const audioEl = ref<HTMLAudioElement | null>(null)
  const error = ref<string | null>(null)

  function startRecording(targetId: string, channels = 2, rate = 48000) {
    stopRecording()
    error.value = null

    const url = `/api/record?targetId=${targetId}&channels=${channels}&rate=${rate}`
    const audio = new Audio(url)
    audio.onerror = () => {
      error.value = 'Failed to start recording stream'
      recording.value = false
    }
    audioEl.value = audio
    audio.play().then(() => {
      recording.value = true
    }).catch(e => {
      error.value = String(e)
      recording.value = false
    })
  }

  function stopRecording() {
    if (audioEl.value) {
      audioEl.value.pause()
      audioEl.value.src = ''
      audioEl.value = null
    }
    recording.value = false
  }

  async function playFile(targetId: number, file: File): Promise<void> {
    const url = `/api/play?targetId=${targetId}`
    const resp = await fetch(url, {
      method: 'POST',
      body: file,
      headers: { 'Content-Type': file.type || 'audio/wav' },
    })
    if (!resp.ok) {
      throw new Error(`play failed: ${resp.status} ${resp.statusText}`)
    }
  }

  return { recording, error, startRecording, stopRecording, playFile }
}
