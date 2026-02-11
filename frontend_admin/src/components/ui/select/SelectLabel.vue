<script setup>
import { cn } from "@/lib/utils"
import { SelectLabel } from 'reka-ui'

const props = defineProps({
  class: { type: String, default: "" },
})
</script>

<template>
  <SelectLabel
    :class="cn('px-2 py-1.5 text-sm font-semibold', props.class)"
  >
    <slot />
  </SelectLabel>
</template>
