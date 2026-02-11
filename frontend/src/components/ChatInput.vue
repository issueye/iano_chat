<template>
  <div class="relative w-full">
    <!-- Input Container -->
    <div
      class="relative flex flex-col gap-3 bg-card rounded-2xl border border-border shadow-sm p-3 transition-all duration-200 focus-within:ring-1 focus-within:ring-ring/30 focus-within:border-ring/50"
      :class="{ 'ring-1 ring-ring/20': isLoading }"
    >
      <!-- Textarea -->
      <textarea
        v-model="inputText"
        :disabled="isLoading"
        :placeholder="isLoading ? 'AI 正在思考...' : '输入消息...'"
        class="w-full bg-transparent border-0 resize-none max-h-32 min-h-20 py-2 px-1 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-0"
        rows="1"
        @keydown.enter.prevent="handleEnter"
        @input="autoResize"
        ref="textareaRef"
      />

      <!-- Bottom Bar -->
      <div class="flex items-center justify-between">
        <!-- Attachment Button -->
        <Tooltip content="添加附件">
          <Button
            variant="ghost"
            size="icon"
            class="shrink-0 rounded-xl hover:bg-muted transition-colors h-8 w-8"
          >
            <Paperclip class="h-4 w-4 text-muted-foreground" />
          </Button>
        </Tooltip>
        <!-- Send Button -->
        <Button
          :disabled="!canSend"
          class="rounded-xl transition-all duration-200 h-8 px-3"
          :class="
            canSend
              ? 'bg-primary hover:bg-primary/90 text-primary-foreground'
              : ''
          "
          @click="sendMessage"
        >
          <Send v-if="!isLoading" class="h-4 w-4 mr-1.5" />
          <Loader2 v-else class="h-4 w-4 animate-spin mr-1.5" />
          <span class="text-xs">发送</span>
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { Button } from "@/components/ui/button";
import { Tooltip } from "@/components/ui/tooltip";
import { Paperclip, Smile, Send, Loader2 } from "lucide-vue-next";

const props = defineProps({
  isLoading: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["send"]);

const inputText = ref("");
const textareaRef = ref(null);

const canSend = computed(() => {
  return inputText.value.trim() && !props.isLoading;
});

function autoResize() {
  const textarea = textareaRef.value;
  if (textarea) {
    textarea.style.height = "auto";
    textarea.style.height = Math.min(textarea.scrollHeight, 128) + "px";
  }
}

function handleEnter(event) {
  if (event.shiftKey) {
    return;
  }
  sendMessage();
}

function sendMessage() {
  const text = inputText.value.trim();
  if (!text || props.isLoading) return;

  emit("send", text);
  inputText.value = "";

  // Reset textarea height
  const textarea = textareaRef.value;
  if (textarea) {
    textarea.style.height = "auto";
  }
}
</script>
