<template>
    <TooltipTrigger>
        <slot />
        <TooltipContent :placement="placement">{{ content }}<slot name="content" /></TooltipContent>
    </TooltipTrigger>
</template>

<script setup>
import { computed } from 'vue';
import { TooltipTrigger, TooltipContent } from '.';

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
</script>
