<script setup>
import { cn } from "@/lib/utils"
import { SelectSeparator } from 'reka-ui'

const props = defineProps({
  class: { type: String, default: "" },
})
</script>

<template>
  <SelectSeparator
    :class="cn('-mx-1 my-1 h-px bg-muted', props.class)"
  />
</template>
