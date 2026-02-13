<script setup>
import { cn } from "@/lib/utils"
import { SelectIcon, SelectTrigger } from 'reka-ui'
import { ChevronDown } from 'lucide-vue-next'

const props = defineProps({
  class: { type: String, default: "" },
  disabled: { type: Boolean, default: false },
})
</script>

<template>
  <SelectTrigger
    v-bind="$attrs"
    :disabled="disabled"
    :class="cn(
      'flex h-10 w-full items-center justify-between rounded-md border border-input bg-transparent px-4 py-2.5 text-base shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50 [&>span]:line-clamp-1',
      props.class
    )"
  >
    <slot />
    <SelectIcon as-child>
      <ChevronDown class="h-5 w-5 opacity-50" />
    </SelectIcon>
  </SelectTrigger>
</template>
