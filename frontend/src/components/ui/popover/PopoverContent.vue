<script setup>
import { PopoverContent, PopoverPortal } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  side: { type: String, default: 'bottom' },
  sideOffset: { type: Number, default: 4 },
  align: { type: String, default: 'start' },
  alignOffset: { type: Number, default: 0 },
  avoidCollisions: { type: Boolean, default: true },
  collisionBoundary: { type: [Object, Array], default: () => [] },
  collisionPadding: { type: [Number, Object], default: 0 },
  arrowPadding: { type: Number, default: 0 },
  sticky: { type: String, default: 'partial' },
  hideWhenDetached: { type: Boolean, default: false },
  updatePositionStrategy: { type: String, default: 'optimized' },
  prioritizePosition: { type: Boolean, default: false },
  asChild: { type: Boolean, default: false },
  as: { type: null, default: undefined },
  class: { type: String, default: '' },
})
</script>

<template>
  <PopoverPortal>
    <PopoverContent
      v-bind="props"
      :class="
        cn(
          'z-50 w-72 rounded-md border bg-popover p-4 text-popover-foreground shadow-md outline-none data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2',
          props.class,
        )
      "
    >
      <slot />
    </PopoverContent>
  </PopoverPortal>
</template>
