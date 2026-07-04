<script setup lang="ts">
import { ref } from 'vue'
import { FileDown, Loader2, Lock, LockOpen, Merge } from 'lucide-vue-next'
import FileDropZone from '@/components/FileDropZone.vue'
import PasswordInput from '@/components/PasswordInput.vue'
import ToolAlert from '@/components/ToolAlert.vue'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { apiUpload } from '@/lib/api'
import { useToolAction } from '@/composables/useToolAction'

const { busy, error, lastFile, run } = useToolAction()

const mergeFiles = ref<File[]>([])
const compressFiles = ref<File[]>([])
const encryptFiles = ref<File[]>([])
const encryptPassword = ref('')
const decryptFiles = ref<File[]>([])
const decryptPassword = ref('')

const merge = () => run(() => apiUpload('/api/v1/tools/pdf/merge', mergeFiles.value))
const compress = () => run(() => apiUpload('/api/v1/tools/pdf/compress', compressFiles.value))
const encrypt = () =>
  run(() => apiUpload('/api/v1/tools/pdf/encrypt', encryptFiles.value, { password: encryptPassword.value }))
const decrypt = () =>
  run(() => apiUpload('/api/v1/tools/pdf/decrypt', decryptFiles.value, { password: decryptPassword.value }))
</script>

<template>
  <div class="mx-auto flex max-w-3xl flex-col gap-4">
    <Tabs default-value="merge">
      <TabsList class="w-full">
        <TabsTrigger value="merge"><Merge /> Merge</TabsTrigger>
        <TabsTrigger value="compress"><FileDown /> Compress</TabsTrigger>
        <TabsTrigger value="encrypt"><Lock /> Protect</TabsTrigger>
        <TabsTrigger value="decrypt"><LockOpen /> Unlock</TabsTrigger>
      </TabsList>

      <TabsContent value="merge">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Merge PDFs</CardTitle>
            <CardDescription>Combine two or more PDFs into a single document, in the order listed.</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="mergeFiles" multiple accept=".pdf" hint="PDF files only — at least two" />
            <Button size="sm" class="self-start" :disabled="busy || mergeFiles.length < 2" @click="merge">
              <Loader2 v-if="busy" class="animate-spin" />
              <Merge v-else />
              Merge {{ mergeFiles.length || '' }} files
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="compress">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Compress PDF</CardTitle>
            <CardDescription>Optimize the internal structure to reduce file size (lossless).</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="compressFiles" accept=".pdf" hint="One PDF file" />
            <Button size="sm" class="self-start" :disabled="busy || compressFiles.length !== 1" @click="compress">
              <Loader2 v-if="busy" class="animate-spin" />
              <FileDown v-else />
              Compress
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="encrypt">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Password-protect PDF</CardTitle>
            <CardDescription>Encrypt the document with AES-256. Opening it will require the password.</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="encryptFiles" accept=".pdf" hint="One PDF file" />
            <div class="flex max-w-sm flex-col gap-1.5">
              <Label for="pdf-encrypt-pw" class="text-xs text-muted-foreground">Password (min 4 characters)</Label>
              <PasswordInput id="pdf-encrypt-pw" v-model="encryptPassword" autocomplete="new-password" />
            </div>
            <Button
              size="sm"
              class="self-start"
              :disabled="busy || encryptFiles.length !== 1 || encryptPassword.length < 4"
              @click="encrypt"
            >
              <Loader2 v-if="busy" class="animate-spin" />
              <Lock v-else />
              Protect
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="decrypt">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Unlock PDF</CardTitle>
            <CardDescription>Remove the password from a protected PDF (requires the current password).</CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="decryptFiles" accept=".pdf" hint="One protected PDF file" />
            <div class="flex max-w-sm flex-col gap-1.5">
              <Label for="pdf-decrypt-pw" class="text-xs text-muted-foreground">Current password</Label>
              <PasswordInput id="pdf-decrypt-pw" v-model="decryptPassword" autocomplete="current-password" />
            </div>
            <Button
              size="sm"
              class="self-start"
              :disabled="busy || decryptFiles.length !== 1 || !decryptPassword"
              @click="decrypt"
            >
              <Loader2 v-if="busy" class="animate-spin" />
              <LockOpen v-else />
              Unlock
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <ToolAlert :error="error" :downloaded="lastFile" />
  </div>
</template>
