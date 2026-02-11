<script setup>
import { SelectRoot, useForwardPropsEmits } from 'reka-ui'

const props = defineProps({
  modelValue: { type: [String, Number, Object], default: undefined },
  defaultValue: { type: [String, Number, Object], default: undefined },
  open: { type: Boolean, default: undefined },
  defaultOpen: { type: Boolean, default: false },
  disabled: { type: Boolean, default: false },
  name: { type: String, default: undefined },
  required: { type: Boolean, default: false },
})

const emit = defineEmits(['update:modelValue', 'update:open'])

const forwarded = useForwardPropsEmits(props, emit)
</script>

<template>
  <SelectRoot v-bind="forwarded">
    <slot />
  </SelectRoot>
</template>
