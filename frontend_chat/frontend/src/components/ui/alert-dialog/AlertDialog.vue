<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent class="sm:max-w-[425px]">
            <DialogTitle>{{ title }}</DialogTitle>
            <p class="text-sm text-muted-foreground p-4">
                {{ description }}
            </p>
            <div class="flex justify-end gap-2 p-4">
                <Button variant="outline" @click="handleCancel">
                    {{ cancelText }}
                </Button>
                <Button :variant="variant" @click="handleConfirm">
                    {{ confirmText }}
                </Button>
            </div>
        </DialogContent>
    </Dialog>
</template>

<script setup>
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';

const props = defineProps({
    open: { type: Boolean, default: false },
    title: { type: String, default: '确认' },
    description: { type: String, default: '' },
    confirmText: { type: String, default: '确认' },
    cancelText: { type: String, default: '取消' },
    variant: { type: String, default: 'default' },
});

const emit = defineEmits(['update:open', 'confirm', 'cancel']);

function handleOpenChange(value) {
    emit('update:open', value);
}

function handleConfirm() {
    emit('confirm');
    emit('update:open', false);
}

function handleCancel() {
    emit('cancel');
    emit('update:open', false);
}
</script>
