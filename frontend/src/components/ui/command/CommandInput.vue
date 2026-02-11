<script setup>
import { ComboboxInput } from 'reka-ui'
import { Search } from 'lucide-vue-next'
import { cn } from '@/lib/utils'

const props = defineProps({
  type: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  asChild: { type: Boolean, default: false },
  as: { type: null, default: undefined },
  class: { type: String, default: '' },
})
</script>

<template>
  <div class="flex items-center border-b px-3" cmd-input-wrapper="">
    <Search class="mr-2 h-4 w-4 shrink-0 opacity-50" />
    <ComboboxInput
      :class="cn('flex h-11 w-full rounded-md bg-transparent py-3 text-sm outline-none placeholder:text-muted-foreground disabled:cursor-not-allowed disabled:opacity-50', props.class)"
    >
      <slot />
    </ComboboxInput>
  </div>
</template>
