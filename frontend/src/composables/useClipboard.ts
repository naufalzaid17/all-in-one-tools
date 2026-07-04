import { ref } from 'vue'

/**
 * Clipboard helper with a short-lived `copied` flag for button feedback.
 */
export function useClipboard(resetAfterMs = 1500) {
  const copied = ref(false)
  let timer: ReturnType<typeof setTimeout> | undefined

  async function copy(text: string) {
    try {
      await navigator.clipboard.writeText(text)
      copied.value = true
      clearTimeout(timer)
      timer = setTimeout(() => (copied.value = false), resetAfterMs)
    } catch {
      // Clipboard access can be denied; fail silently.
    }
  }

  return { copied, copy }
}
