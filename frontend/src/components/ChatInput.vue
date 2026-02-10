<template>
  <div class="relative w-full">
    <!-- Input Container -->
    <div
      class="relative flex items-end gap-2 bg-card rounded-2xl border border-border shadow-sm p-2 transition-all duration-200 focus-within:ring-1 focus-within:ring-ring/30 focus-within:border-ring/50"
      :class="{ 'ring-1 ring-ring/20': isLoading }"
    >
      <!-- Attachment Button -->
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger as-child>
            <Button
              variant="ghost"
              size="icon"
              class="shrink-0 rounded-xl hover:bg-muted transition-colors h-9 w-9 sm:h-10 sm:w-10"
            >
              <Paperclip class="h-4 w-4 sm:h-5 sm:w-5 text-muted-foreground" />
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            <p>添加附件</p>
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>

      <!-- Textarea -->
      <textarea
        v-model="inputText"
        :disabled="isLoading"
        :placeholder="isLoading ? 'AI 正在思考...' : '输入消息...'"
        class="flex-1 bg-transparent border-0 resize-none max-h-32 min-h-[80px] py-2.5 px-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-0"
        rows="1"
        @keydown.enter.prevent="handleEnter"
        @input="autoResize"
        ref="textareaRef"
      />

      <!-- Actions -->
      <div class="flex items-center gap-1 shrink-0">
        <!-- Send Button -->
        <Button
          :disabled="!canSend"
          class="rounded-xl transition-all duration-200 h-9 w-9 sm:h-10 sm:w-10"
          :class="
            canSend
              ? 'bg-primary hover:bg-primary/90 text-primary-foreground'
              : ''
          "
          @click="sendMessage"
        >
          <Send v-if="!isLoading" class="h-4 w-4" />
          <Loader2 v-else class="h-4 w-4 animate-spin" />
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed } from "vue";
import { Button } from "@/components/ui/button";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
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
