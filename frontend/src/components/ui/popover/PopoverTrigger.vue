<script setup>
import { PopoverTrigger } from 'reka-ui'

const props = defineProps({
  asChild: { type: Boolean, default: false },
})
</script>

<template>
  <PopoverTrigger v-bind="props">
    <slot />
  </PopoverTrigger>
</template>
