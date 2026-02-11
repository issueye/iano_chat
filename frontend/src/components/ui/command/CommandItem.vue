<script setup>
import { ComboboxItem } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  value: { type: [String, Object], required: true },
  disabled: { type: Boolean, default: false },
  asChild: { type: Boolean, default: false },
  as: { type: null, default: undefined },
  class: { type: String, default: '' },
})

const emits = defineEmits(['select'])
</script>

<template>
  <ComboboxItem
    v-bind="props"
    :class="cn('relative flex cursor-default select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none data-[highlighted]:bg-accent data-[highlighted]:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50', props.class)"
    @select="emits('select')"
  >
    <slot />
  </ComboboxItem>
</template>
