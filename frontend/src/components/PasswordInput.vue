<script setup lang="ts">
import { ref } from 'vue'
import { Eye, EyeOff } from 'lucide-vue-next'
import { Input } from '@/components/ui/input'

defineProps<{
  id?: string
  placeholder?: string
  autocomplete?: string
}>()

const modelValue = defineModel<string>({ default: '' })
const visible = ref(false)
</script>

<template>
  <div class="relative">
    <Input
      :id="id"
      v-model="modelValue"
      :type="visible ? 'text' : 'password'"
      :placeholder="placeholder ?? 'Password'"
      :autocomplete="autocomplete ?? 'off'"
      spellcheck="false"
      class="pr-9 font-mono text-xs"
    />
    <button
      type="button"
      class="absolute inset-y-0 right-0 flex w-9 items-center justify-center text-muted-foreground transition-colors outline-none hover:text-foreground focus-visible:text-foreground"
      :aria-label="visible ? 'Hide password' : 'Show password'"
      @click="visible = !visible"
    >
      <EyeOff v-if="visible" class="size-4" />
      <Eye v-else class="size-4" />
    </button>
  </div>
</template>
