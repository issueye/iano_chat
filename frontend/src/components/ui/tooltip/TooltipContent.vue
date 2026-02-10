<template>
    <Teleport to="body">
        <Transition
            enter-active-class="transition-all duration-200 ease-out"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition-all duration-150 ease-in"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
        >
            <div
                v-if="isOpen"
                ref="tooltipRef"
                :class="[
                    'fixed z-[100] px-2 py-1 text-xs font-medium rounded-md shadow-sm',
                    'bg-popover text-popover-foreground',
                    'border border-border',
                    placementClass
                ]"
                :style="tooltipStyle"
            >
                <slot><p v-if="content">{{ content }}</p></slot>
            </div>
        </Transition>
    </Teleport>
</template>

<script setup>
import { ref, computed, watch, inject, onMounted, onUnmounted, nextTick } from 'vue';

const props = defineProps({
    content: {
        type: String,
        default: '',
    },
    placement: {
        type: String,
        default: 'top',
    },
});

const tooltipRef = ref(null);
const isOpen = inject('tooltipOpen', ref(false));

const tooltipPos = ref({ top: 0, left: 0 });

const placementClass = computed(() => {
    switch (props.placement) {
        case 'top': return '-translate-x-1/2';
        case 'bottom': return '-translate-x-1/2';
        case 'left': return 'translate-y-[-50%]';
        case 'right': return 'translate-y-[-50%]';
        default: return '-translate-x-1/2';
    }
});

const tooltipStyle = computed(() => ({
    top: `${tooltipPos.value.top}px`,
    left: `${tooltipPos.value.left}px`,
}));

function updatePosition() {
    const trigger = document.activeElement;
    if (!trigger || !tooltipRef.value || !isOpen.value) return;

    const triggerRect = trigger.getBoundingClientRect();
    const tooltipRect = tooltipRef.value.getBoundingClientRect();
    const gutter = 8;

    let top = 0;
    let left = 0;

    switch (props.placement) {
        case 'top':
            top = triggerRect.top - gutter - tooltipRect.height;
            left = triggerRect.left + (triggerRect.width - tooltipRect.width) / 2;
            break;
        case 'bottom':
            top = triggerRect.bottom + gutter;
            left = triggerRect.left + (triggerRect.width - tooltipRect.width) / 2;
            break;
        case 'left':
            top = triggerRect.top + (triggerRect.height - tooltipRect.height) / 2;
            left = triggerRect.left - gutter - tooltipRect.width;
            break;
        case 'right':
            top = triggerRect.top + (triggerRect.height - tooltipRect.height) / 2;
            left = triggerRect.right + gutter;
            break;
    }

    tooltipPos.value = { top, left };
}

watch(isOpen, async (val) => {
    if (val) {
        await nextTick();
        updatePosition();
    }
});

onMounted(() => {
    window.addEventListener('scroll', updatePosition, true);
    window.addEventListener('resize', updatePosition);
});

onUnmounted(() => {
    window.removeEventListener('scroll', updatePosition, true);
    window.removeEventListener('resize', updatePosition);
});
</script>
