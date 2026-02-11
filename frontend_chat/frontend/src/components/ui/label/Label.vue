<script setup>
import { useVModel } from "@vueuse/core"
import { cn } from "@/lib/utils"

const props = defineProps({
  for: { type: String, default: '' },
  class: { type: [String, Array, Object], default: '' }
})

const emits = defineEmits(['update:modelValue'])
</script>

<template>
  <label :for="for" :class="cn('text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70', props.class)">
    <slot />
  </label>
</template>
