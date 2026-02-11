<script setup>
import { ComboboxRoot, useForwardPropsEmits } from 'reka-ui'
import { cn } from '@/lib/utils'

const props = defineProps({
  modelValue: { type: [String, Object], default: '' },
  defaultValue: { type: [String, Object], default: '' },
  open: { type: Boolean, default: false },
  defaultOpen: { type: Boolean, default: false },
  searchTerm: { type: String, default: '' },
  multiple: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  name: { type: String, default: undefined },
  dir: { type: String, default: undefined },
  filterFunction: { type: Function, default: undefined },
  displayValue: { type: Function, default: undefined },
  asChild: { type: Boolean, default: false },
  as: { type: null, default: undefined },
  class: { type: String, default: '' },
})

const emits = defineEmits(['update:modelValue', 'update:open', 'update:searchTerm'])

const forwarded = useForwardPropsEmits(props, emits)
</script>

<template>
  <ComboboxRoot
    v-bind="forwarded"
    :class="cn('flex h-full w-full flex-col overflow-hidden rounded-md bg-popover text-popover-foreground', props.class)"
  >
    <slot />
  </ComboboxRoot>
</template>
