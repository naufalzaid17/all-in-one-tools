<script setup lang="ts">
import { ref } from 'vue'
import { AlertCircle, Fingerprint } from 'lucide-vue-next'
import CopyButton from '@/components/CopyButton.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Textarea } from '@/components/ui/textarea'
import { apiPost } from '@/lib/api'

interface HashResult {
  hashes: Record<string, string>
  inputBytes: number
}

const algorithmLabels: Record<string, string> = {
  sha256: 'SHA-256',
  md5: 'MD5',
}

const input = ref('')
const result = ref<HashResult | null>(null)
const error = ref('')
const busy = ref(false)

async function generate() {
  error.value = ''
  busy.value = true
  try {
    result.value = await apiPost<HashResult>('/api/v1/tools/hash/generate', {
      input: input.value,
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
          Input
        </CardTitle>
        <CardDescription>Text to digest — hashing runs server-side over the raw UTF-8 bytes.</CardDescription>
      </CardHeader>
      <CardContent class="flex flex-col gap-3 px-4">
        <Textarea
          v-model="input"
          placeholder="Type or paste text…"
          spellcheck="false"
          class="min-h-32 resize-y font-mono text-xs leading-relaxed"
        />
        <div class="flex items-center gap-2">
          <Button size="sm" :disabled="busy" @click="generate">
            <Fingerprint /> Generate hashes
          </Button>
          <Badge v-if="result" variant="secondary" class="font-mono text-[10px]">
            {{ result.inputBytes.toLocaleString() }} input bytes
          </Badge>
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
      <CardHeader class="px-4">
        <CardTitle class="text-xs font-medium tracking-wider text-muted-foreground uppercase">
          Digests
        </CardTitle>
      </CardHeader>
      <CardContent class="flex flex-col gap-2 px-4">
        <div
          v-for="(digest, algo) in result.hashes"
          :key="algo"
          class="flex items-center gap-3 rounded-md border bg-muted/40 px-3 py-2"
        >
          <Badge variant="outline" class="w-20 justify-center font-mono text-[10px]">
            {{ algorithmLabels[algo] ?? algo }}
          </Badge>
          <code class="min-w-0 flex-1 truncate font-mono text-xs" :title="digest">{{ digest }}</code>
          <CopyButton :text="digest" />
        </div>
      </CardContent>
    </Card>
  </div>
</template>
