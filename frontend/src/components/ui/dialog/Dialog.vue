<template>
    <DialogRoot :open="open" @update:open="handleOpenChange">
        <slot />
    </DialogRoot>
</template>

<script setup>
import { DialogRoot } from "reka-ui";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:open"]);

function handleOpenChange(value) {
    emit("update:open", value);
}
</script>
