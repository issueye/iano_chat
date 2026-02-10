<template>
    <p :class="['text-sm text-muted-foreground', customClass]">
        <slot />
    </p>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
    class: {
        type: String,
        default: '',
    },
});

const customClass = computed(() => props.class);
</script>
