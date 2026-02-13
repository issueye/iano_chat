<script setup>
import { cn } from "@/lib/utils";
import { DialogTitle } from "reka-ui";
import { X } from 'lucide-vue-next'

const props = defineProps({
  class: { type: String, default: "" },
});

const emit = defineEmits(['close'])
</script>

<template>
  <DialogTitle
    :class="
      cn('text-lg font-semibold leading-none tracking-tight', props.class)
    "
  >
    <slot />
    <button
      class="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:opacity-100 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
      @click="emit('close')"
      >
      <X class="h-4 w-4" />
      <span class="sr-only">关闭</span>
    </button>
  </DialogTitle>
</template>
