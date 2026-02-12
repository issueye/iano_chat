<script setup>
import { ref } from "vue"

const props = defineProps()

const viewportRef = ref(null)

const scrollToBottom = (options) => {
  const el = viewportRef.value
  console.log('el', el);
  if (el) {
    el.scrollTo({
      top: el.scrollHeight,
      behavior: "smooth",
      ...options,
    })
  }
}

defineExpose({
  scrollToBottom,
  viewportRef,
})
</script>

<template>
  <div ref="viewportRef" class="relative h-full w-full rounded-[inherit] overflow-y-auto">
    <slot />
  </div>
</template>
