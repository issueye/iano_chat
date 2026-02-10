<template>
    <div class="flex justify-between items-center p-4 border-b border-border">
        <h2 :class="['text-lg font-semibold leading-none tracking-tight', customClass]">
            <slot />
        </h2>
        <button
            type="button"
            class="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none"
            @click="handleClose"
        >
            <X class="h-4 w-4" />
            <span class="sr-only">关闭</span>
        </button>
    </div>
</template>

<script setup>
import { inject, computed } from 'vue';
import { X } from 'lucide-vue-next';

const props = defineProps({
    class: {
        type: String,
        default: '',
    },
});

const dialog = inject('dialog', null);
const customClass = computed(() => props.class);

function handleClose() {
    console.log('DialogTitle handleClose called');
    if (dialog && dialog.closeModal) {
        dialog.closeModal();
    } else {
        console.warn('dialog or closeModal not available');
    }
}
</script>
