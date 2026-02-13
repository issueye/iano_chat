<template>
  <Dialog :open="open" @update:open="handleOpenChange">
    <DialogContent
      :class="['max-h-[80vh] overflow-hidden flex flex-col', contentClass]"
      :data-dialog="dataDialog"
    >
      <DialogTitle class="flex items-center justify-between p-6 border-b" @click="handleCancel">
        {{ title }}
      </DialogTitle>
      <DialogDescription v-if="description" class="sr-only">
        {{ description }}
      </DialogDescription>
      
      <div class="flex-1 overflow-y-auto scrollbar-stable p-6">
        <slot />
      </div>
      
      <div v-if="showFooter" class="flex justify-end gap-2 p-4 border-t">
        <Button
          v-if="showCancel"
          type="button"
          variant="outline"
          @click="handleCancel"
          :disabled="loading"
        >
          {{ cancelText }}
        </Button>
        <slot name="footer">
          <Button
            v-if="showConfirm"
            type="button"
            @click="handleConfirm"
            :disabled="loading"
          >
            {{ loading ? loadingText : confirmText }}
          </Button>
        </slot>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { watch } from "vue";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";

const props = defineProps({
  /** 弹窗开关状态 */
  open: { type: Boolean, default: false },
  /** 弹窗标题 */
  title: { type: String, required: true },
  /** 弹窗描述 */
  description: { type: String, default: "" },
  /** 弹窗内容类名 */
  contentClass: { type: String, default: "" },
  /** 弹窗标识 */
  dataDialog: { type: String, default: "dialog" },
  /** 确认按钮文字 */
  confirmText: { type: String, default: "确定" },
  /** 取消按钮文字 */
  cancelText: { type: String, default: "取消" },
  /** 加载中文字 */
  loadingText: { type: String, default: "处理中..." },
  /** 是否显示底部按钮 */
  showFooter: { type: Boolean, default: true },
  /** 是否显示取消按钮 */
  showCancel: { type: Boolean, default: true },
  /** 是否显示确认按钮 */
  showConfirm: { type: Boolean, default: true },
  /** 加载状态 */
  loading: { type: Boolean, default: false },
});

const emit = defineEmits(["update:open", "confirm", "cancel"]);

const handleOpenChange = (value) => {
  emit("update:open", value);
};

const handleCancel = () => {
  emit("cancel");
  emit("update:open", false);
};

const handleConfirm = () => {
  emit("confirm");
};
</script>
