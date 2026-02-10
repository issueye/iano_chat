<template>
    <div>
        <slot />
    </div>
</template>

<script setup>
import { provide, ref, watch } from 'vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(['update:open', 'open']);

const isOpen = ref(props.open);

watch(() => props.open, (newVal) => {
    isOpen.value = newVal;
    if (newVal) {
        emit('open');
    }
});

function closeModal() {
    isOpen.value = false;
    emit('update:open', false);
}

provide('dialog', {
    isOpen: isOpen,
    closeModal,
});
</script>
