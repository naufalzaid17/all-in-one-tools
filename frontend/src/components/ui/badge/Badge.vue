<script setup lang="ts">
import type { PrimitiveProps } from 'reka-ui'
import type { HTMLAttributes } from 'vue'
import { Primitive } from 'reka-ui'
import type { BadgeVariants } from '.'
import { badgeVariants } from '.'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<PrimitiveProps & { variant?: BadgeVariants['variant']; class?: HTMLAttributes['class'] }>(),
  { as: 'span' },
)
</script>

<template>
  <Primitive
    data-slot="badge"
    :as="as"
    :as-child="asChild"
    :class="cn(badgeVariants({ variant }), props.class)"
  >
    <slot />
  </Primitive>
</template>
