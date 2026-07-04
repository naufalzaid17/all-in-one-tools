<script setup lang="ts">
import { computed, ref } from 'vue'
import { File as FileIcon, Upload, X } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<{
    /** Allow selecting more than one file. */
    multiple?: boolean
    /** Accept filter, e.g. ".pdf" or ".md,.markdown". */
    accept?: string
    /** Short hint rendered under the title, e.g. "PDF files only". */
    hint?: string
  }>(),
  { multiple: false, accept: '', hint: '' },
)

const files = defineModel<File[]>({ default: () => [] })

const input = ref<HTMLInputElement>()
const dragOver = ref(false)

const acceptedExtensions = computed(() =>
  props.accept
    .split(',')
    .map((e) => e.trim().toLowerCase())
    .filter((e) => e.startsWith('.')),
)

function matchesAccept(file: File): boolean {
  if (acceptedExtensions.value.length === 0) return true
  const name = file.name.toLowerCase()
  return acceptedExtensions.value.some((ext) => name.endsWith(ext))
}

function addFiles(list: FileList | null) {
  if (!list) return
  const incoming = Array.from(list).filter(matchesAccept)
  if (incoming.length === 0) return
  files.value = props.multiple ? [...files.value, ...incoming] : incoming.slice(0, 1)
}

function onDrop(e: DragEvent) {
  dragOver.value = false
  addFiles(e.dataTransfer?.files ?? null)
}

function onPick(e: Event) {
  const target = e.target as HTMLInputElement
  addFiles(target.files)
  target.value = '' // allow re-selecting the same file
}

function removeAt(index: number) {
  files.value = files.value.filter((_, i) => i !== index)
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
</script>

<template>
  <div class="flex flex-col gap-2">
    <button
      type="button"
      :class="cn(
        'flex w-full cursor-pointer flex-col items-center justify-center gap-1.5 rounded-lg border-2 border-dashed px-4 py-8 text-center transition-colors outline-none',
        'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
        dragOver
          ? 'border-primary bg-primary/5'
          : 'border-input hover:border-muted-foreground/50 hover:bg-muted/40',
      )"
      @click="input?.click()"
      @dragover.prevent="dragOver = true"
      @dragleave.prevent="dragOver = false"
      @drop.prevent="onDrop"
    >
      <Upload class="size-5 text-muted-foreground" />
      <div class="text-sm font-medium">
        Drop {{ multiple ? 'files' : 'a file' }} here or <span class="text-primary underline underline-offset-2">browse</span>
      </div>
      <div v-if="hint" class="text-xs text-muted-foreground">{{ hint }}</div>
    </button>
    <input
      ref="input"
      type="file"
      class="hidden"
      :multiple="multiple"
      :accept="accept || undefined"
      @change="onPick"
    />

    <ul v-if="files.length" class="flex flex-col gap-1">
      <li
        v-for="(file, i) in files"
        :key="`${file.name}-${i}`"
        class="flex items-center gap-2.5 rounded-md border bg-muted/40 py-1.5 pr-1.5 pl-3"
      >
        <FileIcon class="size-3.5 shrink-0 text-muted-foreground" />
        <span class="min-w-0 flex-1 truncate font-mono text-xs" :title="file.name">{{ file.name }}</span>
        <span class="shrink-0 text-[11px] text-muted-foreground tabular-nums">{{ formatSize(file.size) }}</span>
        <Button
          variant="ghost"
          size="icon"
          class="size-6 text-muted-foreground hover:text-destructive"
          :aria-label="`Remove ${file.name}`"
          @click.stop="removeAt(i)"
        >
          <X class="size-3.5" />
        </Button>
      </li>
    </ul>
  </div>
</template>
