<script setup lang="ts">
import { Check, Copy } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { useClipboard } from '@/composables/useClipboard'

const props = defineProps<{ text: string }>()

const { copied, copy } = useClipboard()
</script>

<template>
  <Button
    variant="ghost"
    size="icon"
    class="size-7 text-muted-foreground hover:text-foreground"
    :disabled="!props.text"
    :aria-label="copied ? 'Copied' : 'Copy to clipboard'"
    @click="copy(props.text)"
  >
    <Check v-if="copied" class="size-3.5 text-emerald-500" />
    <Copy v-else class="size-3.5" />
  </Button>
</template>
