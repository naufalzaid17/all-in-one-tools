<script setup lang="ts">
import { ref } from 'vue'
import { FileSpreadsheet, Loader2, Merge } from 'lucide-vue-next'
import FileDropZone from '@/components/FileDropZone.vue'
import ToolAlert from '@/components/ToolAlert.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { apiUpload } from '@/lib/api'
import { useToolAction } from '@/composables/useToolAction'

const { busy, error, lastFile, run } = useToolAction()

const mergeFiles = ref<File[]>([])
const csvFiles = ref<File[]>([])
const sheetName = ref('')

const merge = () => run(() => apiUpload('/api/v1/tools/excel/merge', mergeFiles.value))
const toCsv = () =>
  run(() =>
    apiUpload('/api/v1/tools/excel/to-csv', csvFiles.value, sheetName.value ? { sheet: sheetName.value } : {}),
  )
</script>

<template>
  <div class="mx-auto flex max-w-3xl flex-col gap-4">
    <Tabs default-value="merge">
      <TabsList class="w-full">
        <TabsTrigger value="merge"><Merge /> Merge workbooks</TabsTrigger>
        <TabsTrigger value="csv"><FileSpreadsheet /> XLSX → CSV</TabsTrigger>
      </TabsList>

      <TabsContent value="merge">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Merge workbooks</CardTitle>
            <CardDescription>
              Copies every sheet from every workbook into one file. Sheets are renamed
              "workbook - sheet"; values are preserved, formatting is flattened.
            </CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="mergeFiles" multiple accept=".xlsx" hint=".xlsx files — at least two" />
            <Button size="sm" class="self-start" :disabled="busy || mergeFiles.length < 2" @click="merge">
              <Loader2 v-if="busy" class="animate-spin" />
              <Merge v-else />
              Merge {{ mergeFiles.length || '' }} workbooks
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="csv">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Convert to CSV</CardTitle>
            <CardDescription>Exports one sheet as CSV. Leave the sheet name empty to use the first sheet.</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="csvFiles" accept=".xlsx" hint="One .xlsx file" />
            <div class="flex max-w-sm flex-col gap-1.5">
              <Label for="sheet-name" class="text-xs text-muted-foreground">Sheet name (optional)</Label>
              <Input id="sheet-name" v-model="sheetName" placeholder="First sheet" spellcheck="false" class="font-mono text-xs" />
            </div>
            <Button size="sm" class="self-start" :disabled="busy || csvFiles.length !== 1" @click="toCsv">
              <Loader2 v-if="busy" class="animate-spin" />
              <FileSpreadsheet v-else />
              Convert
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <ToolAlert :error="error" :downloaded="lastFile" />
  </div>
</template>
