<template>
  <div class="relative w-full overflow-auto">
    <table :class="cn('w-full caption-bottom text-sm', props.class)">
      <slot />
    </table>
  </div>
</template>

<script setup>
import { cn } from "@/lib/utils"

const props = defineProps({
  class: { type: [String, Array, Object], default: '' }
})
</script>
