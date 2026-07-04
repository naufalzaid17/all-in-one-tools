<script setup lang="ts">
import { ref } from 'vue'
import { AlertCircle, Braces, Minimize2 } from 'lucide-vue-next'
import CopyButton from '@/components/CopyButton.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { apiPost } from '@/lib/api'

interface TransformResult {
  output: string
  bytes: number
}

const input = ref('')
const output = ref('')
const outputBytes = ref<number | null>(null)
const indent = ref('2')
const error = ref('')
const busy = ref(false)

async function transform(action: 'format' | 'minify') {
  error.value = ''
  busy.value = true
  try {
    const body =
      action === 'format'
        ? indent.value === 'tab'
          ? { input: input.value, useTabs: true }
          : { input: input.value, indent: Number(indent.value) }
        : { input: input.value }

    const result = await apiPost<TransformResult>(`/api/v1/tools/json/${action}`, body)
    output.value = result.output
    outputBytes.value = result.bytes
  } catch (e) {
    output.value = ''
    outputBytes.value = null
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="mx-auto flex max-w-6xl flex-col gap-4">
    <!-- Toolbar -->
    <div class="flex flex-wrap items-center gap-2">
      <Button size="sm" :disabled="busy || !input.trim()" @click="transform('format')">
        <Braces /> Format
      </Button>
      <Button size="sm" variant="secondary" :disabled="busy || !input.trim()" @click="transform('minify')">
        <Minimize2 /> Minify
      </Button>
      <div class="ml-auto flex items-center gap-2">
        <Label for="indent" class="text-xs text-muted-foreground">Indent</Label>
        <select
          id="indent"
          v-model="indent"
          class="h-8 rounded-md border border-input bg-transparent px-2 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 dark:bg-input/30"
        >
          <option value="2">2 spaces</option>
          <option value="4">4 spaces</option>
          <option value="tab">Tabs</option>
        </select>
      </div>
    </div>

    <!-- Error -->
    <div
      v-if="error"
      class="flex items-center gap-2 rounded-md border border-destructive/50 bg-destructive/10 px-3 py-2 text-sm text-destructive"
      role="alert"
    >
      <AlertCircle class="size-4 shrink-0" />
      {{ error }}
    </div>

    <!-- Input / Output panes -->
    <div class="grid min-h-0 grid-cols-1 gap-4 lg:grid-cols-2">
      <Card class="gap-3 py-4">
        <CardHeader class="px-4">
          <CardTitle class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
            Input
          </CardTitle>
        </CardHeader>
        <CardContent class="px-4">
          <Textarea
            v-model="input"
            placeholder='{"paste": "your JSON here"}'
            spellcheck="false"
            class="min-h-[420px] resize-y font-mono text-xs leading-relaxed"
          />
        </CardContent>
      </Card>

      <Card class="gap-3 py-4">
        <CardHeader class="flex-row items-center justify-between px-4">
          <CardTitle class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
            Output
          </CardTitle>
          <div class="flex items-center gap-1.5">
            <Badge v-if="outputBytes !== null" variant="secondary" class="font-mono text-[10px]">
              {{ outputBytes.toLocaleString() }} B
            </Badge>
            <CopyButton :text="output" />
          </div>
        </CardHeader>
        <CardContent class="px-4">
          <Textarea
            :model-value="output"
            readonly
            placeholder="Result appears here…"
            spellcheck="false"
            class="min-h-[420px] resize-y bg-muted/40 font-mono text-xs leading-relaxed"
          />
        </CardContent>
      </Card>
    </div>
  </div>
</template>
