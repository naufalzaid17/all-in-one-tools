<script setup lang="ts">
import { ref } from 'vue'
import { FileCode2, Loader2, Merge } from 'lucide-vue-next'
import FileDropZone from '@/components/FileDropZone.vue'
import ToolAlert from '@/components/ToolAlert.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { apiUpload } from '@/lib/api'
import { useToolAction } from '@/composables/useToolAction'

const { busy, error, lastFile, run } = useToolAction()

const mergeFiles = ref<File[]>([])
const htmlFiles = ref<File[]>([])
const standalone = ref(true)

const merge = () => run(() => apiUpload('/api/v1/tools/markdown/merge', mergeFiles.value))
const toHtml = () =>
  run(() =>
    apiUpload('/api/v1/tools/markdown/to-html', htmlFiles.value, { standalone: String(standalone.value) }),
  )
</script>

<template>
  <div class="mx-auto flex max-w-3xl flex-col gap-4">
    <Tabs default-value="merge">
      <TabsList class="w-full">
        <TabsTrigger value="merge"><Merge /> Merge to master context</TabsTrigger>
        <TabsTrigger value="html"><FileCode2 /> MD → HTML</TabsTrigger>
      </TabsList>

      <TabsContent value="merge">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Merge Markdown files</CardTitle>
            <CardDescription>
              Concatenates .md files into one master context document. Each source is wrapped in
              begin/end comment markers so AI agents (and humans) can locate section boundaries.
            </CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="mergeFiles" multiple accept=".md,.markdown,.txt" hint=".md files — at least two, merged in listed order" />
            <Button size="sm" class="self-start" :disabled="busy || mergeFiles.length < 2" @click="merge">
              <Loader2 v-if="busy" class="animate-spin" />
              <Merge v-else />
              Merge {{ mergeFiles.length || '' }} files
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="html">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Convert to HTML</CardTitle>
            <CardDescription>Renders GitHub-flavoured Markdown (tables, task lists, strikethrough, autolinks).</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="htmlFiles" accept=".md,.markdown,.txt" hint="One Markdown file" />
            <Label class="flex cursor-pointer items-center gap-2 text-xs text-muted-foreground">
              <input v-model="standalone" type="checkbox" class="size-3.5 accent-primary" />
              Standalone document (embed styles; uncheck for a raw HTML fragment)
            </Label>
            <Button size="sm" class="self-start" :disabled="busy || htmlFiles.length !== 1" @click="toHtml">
              <Loader2 v-if="busy" class="animate-spin" />
              <FileCode2 v-else />
              Convert
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <ToolAlert :error="error" :downloaded="lastFile" />
  </div>
</template>
