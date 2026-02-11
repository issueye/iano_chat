<template>
  <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :class="cn('h-4 w-4', props.class)">
    <path d="M20 6 9 17l-5-5" />
  </svg>
</template>

<script setup>
import { cn } from "@/lib/utils"

const props = defineProps({
  class: { type: [String, Array, Object], default: '' }
})
</script>
