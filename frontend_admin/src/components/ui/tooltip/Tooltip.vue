<script setup>
import { TooltipArrow, TooltipContent, TooltipPortal, TooltipProvider, TooltipRoot, TooltipTrigger } from 'reka-ui'

const props = defineProps({
  content: {
    type: String,
    default: ''
  }
})
</script>

<template>
  <TooltipProvider>
    <TooltipRoot>
      <TooltipTrigger
        class="inline-flex items-center justify-center outline-none"
      >
        <slot></slot>
      </TooltipTrigger>
      <TooltipPortal>
        <TooltipContent
          class="data-[state=delayed-open]:data-[side=top]:animate-slideDownAndFade data-[state=delayed-open]:data-[side=right]:animate-slideLeftAndFade data-[state=delayed-open]:data-[side=left]:animate-slideRightAndFade data-[state=delayed-open]:data-[side=bottom]:animate-slideUpAndFade select-none rounded-md bg-white px-[15px] py-[10px] text-sm leading-none shadow-sm border will-change-[transform,opacity] z-[1000]"
          :side-offset="5"
        >
          {{ content }}
          <TooltipArrow
            class="fill-white stroke-gray-200"
            :width="12"
            :height="6"
          />
        </TooltipContent>
      </TooltipPortal>
    </TooltipRoot>
  </TooltipProvider>
</template>
