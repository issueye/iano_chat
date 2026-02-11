<script setup>
import { useVModel } from "@vueuse/core"
import { cn } from "@/lib/utils"
import { Check } from "lucide-vue-next"

const props = defineProps({
  modelValue: { type: [Boolean, String, Number], default: false },
  class: { type: [String, Array, Object], default: '' }
})

const emits = defineEmits(['update:modelValue'])

const modelValue = useVModel(props, "modelValue", emits, {
  passive: true,
})
</script>

<template>
  <button
    type="button"
    role="checkbox"
    :aria-checked="modelValue"
    :class="cn(
      'peer h-4 w-4 shrink-0 rounded-sm border border-primary ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 data-[state=checked]:bg-primary data-[state=checked]:text-primary-foreground',
      props.class
    )"
    @click="modelValue = !modelValue"
  >
    <span
      v-if="modelValue"
      class="flex items-center justify-center text-current"
    >
      <Check class="h-3 w-3" />
    </span>
  </button>
</template>
