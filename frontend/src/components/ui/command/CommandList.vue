<script setup>
import { ComboboxContent } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  asChild: { type: Boolean, default: false },
  as: { type: null, default: undefined },
  class: { type: String, default: '' },
})
</script>

<template>
  <ComboboxContent
    v-bind="props"
    :class="cn('max-h-[300px] overflow-y-auto overflow-x-hidden', props.class)"
  >
    <slot />
  </ComboboxContent>
</template>
