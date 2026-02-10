<template>
    <div
        ref="triggerRef"
        class="relative inline-block"
        @mouseenter="handleMouseEnter"
        @mouseleave="handleMouseLeave"
        @focusin="handleMouseEnter"
        @focusout="handleMouseLeave"
    >
        <slot />
    </div>
</template>

<script setup>
import { ref, provide } from 'vue';

const triggerRef = ref(null);
const isOpen = ref(false);

provide('tooltipOpen', isOpen);
provide('tooltipTrigger', triggerRef);

function handleMouseEnter() {
    isOpen.value = true;
}

function handleMouseLeave() {
    isOpen.value = false;
}
</script>
