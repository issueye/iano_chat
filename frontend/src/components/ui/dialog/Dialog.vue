<template>
    <div>
        <slot />
    </div>
</template>

<script setup>
import { provide, computed } from 'vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(['update:open']);

const isOpen = computed({
    get: () => props.open,
    set: (value) => emit('update:open', value),
});

function closeModal() {
    isOpen.value = false;
}

provide('dialog', {
    isOpen,
    closeModal,
});
</script>
