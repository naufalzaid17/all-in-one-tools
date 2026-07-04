<script setup lang="ts">
import { ref } from 'vue'
import { AlertCircle, Download, QrCode } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { apiPost } from '@/lib/api'

interface QrResult {
  image: string
  size: number
}

const content = ref('')
const size = ref('256')
const result = ref<QrResult | null>(null)
const error = ref('')
const busy = ref(false)

async function generate() {
  error.value = ''
  busy.value = true
  try {
    result.value = await apiPost<QrResult>('/api/v1/tools/qrcode/generate', {
      content: content.value,
      size: Number(size.value),
    })
  } catch (e) {
    result.value = null
    error.value = e instanceof Error ? e.message : 'Something went wrong'
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="mx-auto flex max-w-3xl flex-col gap-4">
    <Card class="gap-3 py-4">
      <CardHeader class="px-4">
        <CardTitle class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
          Content
        </CardTitle>
        <CardDescription>URL or text to encode. Returned as a base64 PNG data URI.</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-3 px-4">
        <div class="flex flex-col gap-1.5">
          <Label for="qr-content" class="text-xs text-muted-foreground">URL / text</Label>
          <Input
            id="qr-content"
            v-model="content"
            placeholder="https://example.com"
            spellcheck="false"
            class="font-mono text-xs"
            @keydown.enter="content.trim() && generate()"
          />
        </div>
        <div class="flex items-end gap-3">
          <div class="flex flex-col gap-1.5">
            <Label for="qr-size" class="text-xs text-muted-foreground">Size</Label>
            <select
              id="qr-size"
              v-model="size"
              class="h-9 rounded-md border border-input bg-transparent px-2 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 dark:bg-input/30"
            >
              <option value="128">128 px</option>
              <option value="256">256 px</option>
              <option value="512">512 px</option>
            </select>
          </div>
          <Button size="sm" class="h-9" :disabled="busy || !content.trim()" @click="generate">
            <QrCode /> Generate
          </Button>
        </div>
      </CardContent>
    </Card>

    <div
      v-if="error"
      class="flex items-center gap-2 rounded-md border border-destructive/50 bg-destructive/10 px-3 py-2 text-sm text-destructive"
      role="alert"
    >
      <AlertCircle class="size-4 shrink-0" />
      {{ error }}
    </div>

    <Card v-if="result" class="gap-3 py-4">
      <CardHeader class="flex-row items-center justify-between px-4">
        <CardTitle class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
          Result
        </CardTitle>
        <Button variant="outline" size="sm" as-child>
          <a :href="result.image" download="qrcode.png">
            <Download /> Download PNG
          </a>
        </Button>
      </CardHeader>
      <CardContent class="flex justify-center px-4 py-2">
        <img
          :src="result.image"
          :width="result.size"
          :height="result.size"
          alt="Generated QR code"
          class="rounded-md border bg-white p-2"
        />
      </CardContent>
    </Card>
  </div>
</template>
