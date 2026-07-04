import { ref } from 'vue'
import type { DownloadResult } from '@/lib/api'
import { saveBlob } from '@/lib/api'

/**
 * Shared state machine for file-processing tools: tracks the busy flag
 * for spinners, surfaces errors, triggers the browser download on
 * success and remembers the last produced file name.
 */
export function useToolAction() {
  const busy = ref(false)
  const error = ref('')
  const lastFile = ref('')

  async function run(fn: () => Promise<DownloadResult>) {
    if (busy.value) return
    error.value = ''
    lastFile.value = ''
    busy.value = true
    try {
      const result = await fn()
      saveBlob(result)
      lastFile.value = result.filename
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Something went wrong'
    } finally {
      busy.value = false
    }
  }

  return { busy, error, lastFile, run }
}
