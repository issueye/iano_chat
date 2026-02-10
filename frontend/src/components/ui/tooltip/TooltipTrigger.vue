<template>
    <div
        ref="triggerRef"
        class="relative inline-block"
        @mouseenter="open = true"
        @mouseleave="open = false"
        @focusin="open = true"
        @focusout="open = false"
    >
        <slot />
    </div>
</template>

<script setup>
import { ref, provide, watch } from 'vue';

const open = ref(false);
const triggerRef = ref(null);

provide('tooltipOpen', open);
provide('tooltipTrigger', triggerRef);
</script>
