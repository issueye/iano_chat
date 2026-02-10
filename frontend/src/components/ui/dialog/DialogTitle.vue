<template>
    <div class="flex justify-between items-center border-b border-border p-6">
        <h2 :class="['text-lg font-semibold leading-none tracking-tight', customClass]">
            <slot />
        </h2>
        <button
            type="button"
            class="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:pointer-events-none data-[state=open]:bg-accent data-[state=open]:text-muted-foreground"
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

const dialog = inject('dialog');
const customClass = computed(() => props.class);

function handleClose() {
    dialog?.closeModal?.();
}
</script>
