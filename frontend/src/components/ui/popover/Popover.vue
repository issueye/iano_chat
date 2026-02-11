<script setup>
import { PopoverRoot, useForwardPropsEmits } from 'reka-ui'

const props = defineProps({
  defaultOpen: { type: Boolean, default: false },
  open: { type: Boolean, default: undefined },
})

const emit = defineEmits(['update:open'])

const forwarded = useForwardPropsEmits(props, emit)
</script>

<template>
  <PopoverRoot v-bind="forwarded">
    <slot />
  </PopoverRoot>
</template>
