<template>
  <Dialog :open="open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[425px]">
      <DialogTitle>{{ title }}</DialogTitle>
      <DialogDescription v-if="description">
        {{ description }}
      </DialogDescription>
      <div class="flex justify-end gap-2 mt-4">
        <Button variant="outline" @click="handleCancel">
          {{ cancelText }}
        </Button>
        <Button :variant="variant === 'destructive' ? 'destructive' : 'default'" @click="handleConfirm">
          {{ confirmText }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { Dialog, DialogContent, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: "确认" },
  description: { type: String, default: "" },
  confirmText: { type: String, default: "确认" },
  cancelText: { type: String, default: "取消" },
  variant: { type: String, default: "default" }, // default, destructive
});

const emit = defineEmits(["update:open", "confirm"]);

function handleOpenChange(value) {
  emit("update:open", value);
}

function handleCancel() {
  emit("update:open", false);
}

function handleConfirm() {
  emit("confirm");
  emit("update:open", false);
}
</script>
