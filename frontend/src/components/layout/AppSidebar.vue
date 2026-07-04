<script setup lang="ts">
import { TerminalSquare } from 'lucide-vue-next'
import { RouterLink } from 'vue-router'
import { toolGroups } from '@/lib/tools'
</script>

<template>
  <aside
    class="flex h-full w-60 shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground"
  >
    <!-- Brand -->
    <div class="flex h-12 items-center gap-2 border-b border-sidebar-border px-4">
      <TerminalSquare class="size-4.5 text-sidebar-foreground" />
      <div class="text-sm font-semibold tracking-tight">All-in-One Tools</div>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 space-y-4 overflow-y-auto px-2 py-3">
      <div v-for="group in toolGroups" :key="group.label">
        <div class="px-2 pb-1.5 text-[11px] font-medium tracking-wider text-muted-foreground uppercase">
          {{ group.label }}
        </div>
        <ul class="space-y-0.5">
          <li v-for="item in group.items" :key="item.path">
            <RouterLink
              :to="item.path"
              class="flex items-center gap-2.5 rounded-md px-2 py-1.5 text-sm text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"
              active-class="bg-sidebar-accent font-medium !text-sidebar-accent-foreground"
            >
              <component :is="item.icon" class="size-4 shrink-0" />
              <span class="truncate">{{ item.name }}</span>
            </RouterLink>
          </li>
        </ul>
      </div>
    </nav>

    <!-- Footer -->
    <div class="border-t border-sidebar-border px-4 py-2.5 text-[11px] text-muted-foreground">
      Local · single binary · Go + Vue
    </div>
  </aside>
</template>
