<template>
  <thead :class="cn('[&_tr]:border-b', props.class)">
    <slot />
  </thead>
</template>

<script setup>
import { cn } from "@/lib/utils"

const props = defineProps({
  class: { type: [String, Array, Object], default: '' }
})
</script>
