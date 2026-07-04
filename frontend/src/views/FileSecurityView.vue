<script setup lang="ts">
import { ref } from 'vue'
import { Loader2, LockKeyhole, LockOpen, ShieldCheck } from 'lucide-vue-next'
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

const encryptFiles = ref<File[]>([])
const encryptPassword = ref('')
const decryptFiles = ref<File[]>([])
const decryptPassword = ref('')

const encrypt = () =>
  run(() => apiUpload('/api/v1/tools/security/encrypt', encryptFiles.value, { password: encryptPassword.value }))
const decrypt = () =>
  run(() => apiUpload('/api/v1/tools/security/decrypt', decryptFiles.value, { password: decryptPassword.value }))
</script>

<template>
  <div class="mx-auto flex max-w-3xl flex-col gap-4">
    <Tabs default-value="encrypt">
      <TabsList class="w-full">
        <TabsTrigger value="encrypt"><LockKeyhole /> Encrypt</TabsTrigger>
        <TabsTrigger value="decrypt"><LockOpen /> Decrypt</TabsTrigger>
      </TabsList>

      <TabsContent value="encrypt">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Encrypt a file</CardTitle>
            <CardDescription>
              Any file type, any size. AES-256-GCM with an Argon2id-derived key; processing is
              streamed, so large files are fine. Output: <span class="font-mono text-xs">yourfile.ext.enc</span>
            </CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="encryptFiles" hint="Any file" />
            <div class="flex max-w-sm flex-col gap-1.5">
              <Label for="enc-pw" class="text-xs text-muted-foreground">Password (min 6 characters)</Label>
              <PasswordInput id="enc-pw" v-model="encryptPassword" autocomplete="new-password" />
            </div>
            <div class="flex items-center gap-2 text-xs text-muted-foreground">
              <ShieldCheck class="size-3.5 shrink-0" />
              There is no recovery — a lost password means a lost file.
            </div>
            <Button
              size="sm"
              class="self-start"
              :disabled="busy || encryptFiles.length !== 1 || encryptPassword.length < 6"
              @click="encrypt"
            >
              <Loader2 v-if="busy" class="animate-spin" />
              <LockKeyhole v-else />
              {{ busy ? 'Encrypting…' : 'Encrypt' }}
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="decrypt">
        <Card class="gap-3 py-4">
          <CardHeader class="px-4">
            <CardTitle>Decrypt a file</CardTitle>
            <CardDescription>
              Restores a file encrypted by this tool. The password is verified cryptographically —
              a wrong password never produces corrupt output.
            </CardDescription>
          </CardHeader>
          <CardContent class="flex flex-col gap-3 px-4">
            <FileDropZone v-model="decryptFiles" hint="A .enc file produced by this tool" />
            <div class="flex max-w-sm flex-col gap-1.5">
              <Label for="dec-pw" class="text-xs text-muted-foreground">Password</Label>
              <PasswordInput id="dec-pw" v-model="decryptPassword" autocomplete="current-password" />
            </div>
            <Button
              size="sm"
              class="self-start"
              :disabled="busy || decryptFiles.length !== 1 || !decryptPassword"
              @click="decrypt"
            >
              <Loader2 v-if="busy" class="animate-spin" />
              <LockOpen v-else />
              {{ busy ? 'Decrypting…' : 'Decrypt' }}
            </Button>
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>

    <ToolAlert :error="error" :downloaded="lastFile" />
  </div>
</template>
