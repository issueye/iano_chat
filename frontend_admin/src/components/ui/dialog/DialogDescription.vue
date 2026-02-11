<script setup>
import { cn } from "@/lib/utils"
import { DialogDescription } from 'reka-ui'

const props = defineProps({
  class: { type: String, default: "" },
})
</script>

<template>
  <DialogDescription
    :class="cn('text-sm text-muted-foreground', props.class)"
  >
    <slot />
  </DialogDescription>
</template>
