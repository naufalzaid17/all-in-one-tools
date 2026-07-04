<script setup lang="ts">
import { computed } from 'vue'
import { Moon, Sun } from 'lucide-vue-next'
import { RouterView, useRoute } from 'vue-router'
import AppSidebar from '@/components/layout/AppSidebar.vue'
import { Button } from '@/components/ui/button'
import { toolGroups } from '@/lib/tools'
import { useTheme } from '@/composables/useTheme'

const route = useRoute()
const { theme, toggle } = useTheme()

const activeTool = computed(() =>
  toolGroups.flatMap((g) => g.items).find((t) => t.path === route.path),
)
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-background text-foreground">
    <AppSidebar />

    <div class="flex min-w-0 flex-1 flex-col">
      <!-- Header -->
      <header class="flex h-12 shrink-0 items-center justify-between border-b px-5">
        <div class="min-w-0">
          <h1 class="truncate text-sm font-semibold tracking-tight">
            {{ activeTool?.name ?? 'All-in-One Tools' }}
          </h1>
          <p v-if="activeTool" class="sr-only">{{ activeTool.description }}</p>
        </div>
        <div class="flex items-center gap-2">
          <span v-if="activeTool" class="hidden text-xs text-muted-foreground sm:block">
            {{ activeTool.description }}
          </span>
          <Button
            variant="ghost"
            size="icon"
            class="size-8"
            :aria-label="theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'"
            @click="toggle"
          >
            <Sun v-if="theme === 'dark'" class="size-4" />
            <Moon v-else class="size-4" />
          </Button>
        </div>
      </header>

      <!-- Active tool -->
      <main class="flex-1 overflow-y-auto p-5">
        <RouterView />
      </main>
    </div>
  </div>
</template>
