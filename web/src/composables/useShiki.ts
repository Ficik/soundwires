import { shallowRef } from 'vue'
import type { Highlighter } from 'shiki'

const hl = shallowRef<Highlighter | null>(null)
let promise: Promise<void> | null = null

export function useShiki() {
  if (!promise) {
    promise = import('shiki').then(async ({ createHighlighter }) => {
      hl.value = await createHighlighter({
        langs: ['json'],
        themes: ['github-dark'],
      })
    })
  }
  return hl
}
