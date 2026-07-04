<script setup lang="ts">
import { AlertCircle, CheckCircle2 } from 'lucide-vue-next'

defineProps<{
  /** Error message; takes precedence when both are set. */
  error?: string
  /** Name of the file that was just downloaded. */
  downloaded?: string
}>()
</script>

<template>
  <div
    v-if="error"
    class="flex items-center gap-2 rounded-md border border-destructive/50 bg-destructive/10 px-3 py-2 text-sm text-destructive"
    role="alert"
  >
    <AlertCircle class="size-4 shrink-0" />
    {{ error }}
  </div>
  <div
    v-else-if="downloaded"
    class="flex items-center gap-2 rounded-md border border-emerald-600/40 bg-emerald-500/10 px-3 py-2 text-sm text-emerald-600 dark:text-emerald-400"
    role="status"
  >
    <CheckCircle2 class="size-4 shrink-0" />
    <span>Downloaded <span class="font-mono text-xs">{{ downloaded }}</span></span>
  </div>
</template>
