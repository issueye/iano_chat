<template>
  <ScrollArea class="flex-1 px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
    <div class="mx-auto space-y-6 pb-20">
      <slot name="welcome" />

      <template v-if="messages.length">
        <ChatMessage
          v-for="(message, index) in messages"
          :key="message.id"
          :message="message"
          :is-last="index === messages.length - 1"
        />
      </template>

      <div v-if="isLoading" class="flex justify-center py-4">
        <div class="flex items-center gap-2 text-muted-foreground">
          <Loader2 class="w-4 h-4 animate-spin text-primary" />
          <span class="text-sm">AI 正在思考...</span>
        </div>
      </div>

      <div
        v-if="error"
        class="p-4 rounded-lg bg-destructive/10 text-destructive text-sm flex items-center gap-2"
      >
        <AlertCircle class="w-4 h-4 shrink-0" />
        {{ error }}
      </div>
    </div>
  </ScrollArea>
</template>

<script setup>
/**
 * ChatMessages 组件 - 消息列表区域
 * 显示聊天消息列表、加载状态和错误信息
 */
import ChatMessage from "./ChatMessage.vue"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Loader2, AlertCircle } from "lucide-vue-next"

/**
 * 组件属性定义
 */
defineProps({
  /** 消息列表 */
  messages: {
    type: Array,
    default: () => [],
  },
  /** 是否正在加载 */
  isLoading: {
    type: Boolean,
    default: false,
  },
  /** 错误信息 */
  error: {
    type: String,
    default: null,
  },
})
</script>
