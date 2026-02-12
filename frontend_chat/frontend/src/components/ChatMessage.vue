<template>
  <div
    :class="[
      'flex w-full animate-fade-in mb-6 last:mb-0',
      message.type === 'user' ? 'justify-end' : 'justify-start',
    ]"
  >
    <div
      :class="[
        'group flex gap-3 sm:gap-4 max-w-[90%] sm:max-w-[80%] lg:max-w-[75%]',
        message.type === 'user' ? 'flex-row-reverse' : 'flex-row',
      ]"
    >
      <!-- Avatar -->
      <Avatar :class="message.type === 'user' ? 'bg-primary' : 'bg-secondary'">
        <AvatarFallback>
          <User
            v-if="message.type === 'user'"
            class="w-4 h-4 text-primary-foreground"
          />
          <Bot v-else class="w-4 h-4 text-foreground" />
        </AvatarFallback>
      </Avatar>

      <!-- Message Content -->
      <div class="flex flex-col gap-1.5">
        <!-- Name and Time -->
        <div
          :class="[
            'flex items-center gap-2 text-xs',
            message.type === 'user' ? 'justify-end' : 'justify-start',
          ]"
        >
          <span class="font-medium text-foreground">
            {{ message.type === "user" ? "我" : "AI 助手" }}
          </span>
          <span class="text-muted-foreground">
            {{ formatTime(message.created_at) }}
          </span>
        </div>

        <!-- Message Bubble -->
        <div
          :class="[
            'relative px-4 py-3 sm:px-5 sm:py-3.5 rounded-2xl shadow-sm',
            message.type === 'user'
              ? 'bg-primary text-primary-foreground rounded-tr-sm'
              : 'bg-card border border-border text-foreground rounded-tl-sm',
          ]"
        >
          <!-- Content -->
          <div class="text-sm leading-relaxed text-inherit">
            <template v-if="messageContent.text">
              <MarkdownRenderer :content="messageContent.text" />
            </template>
            <template v-else-if="message.status === 'streaming'">
              <span class="inline-flex items-center gap-1">
                <span
                  class="w-1.5 h-1.5 bg-current rounded-full animate-bounce"
                  style="animation-delay: 0s"
                ></span>
                <span
                  class="w-1.5 h-1.5 bg-current rounded-full animate-bounce"
                  style="animation-delay: 0.2s"
                ></span>
                <span
                  class="w-1.5 h-1.5 bg-current rounded-full animate-bounce"
                  style="animation-delay: 0.4s"
                ></span>
              </span>
            </template>
            <template v-else-if="message.status === 'failed'">
              <span class="flex items-center gap-1">
                <AlertCircle class="w-4 h-4" />
                发送失败
              </span>
            </template>
          </div>

          <!-- Tool Calls -->
          <div v-if="messageContent.tool_calls?.length" class="mt-3 space-y-2">
            <div
              v-for="tool in messageContent.tool_calls"
              :key="tool.id"
              class="bg-muted rounded-lg p-3 text-xs"
            >
              <div class="flex items-center gap-2 font-medium">
                <Wrench class="w-4 h-4 text-foreground" />
                {{ tool.function.name }}
              </div>
              <div
                class="mt-2 opacity-70 font-mono text-[10px] bg-black/5 rounded p-2 truncate"
              >
                {{ tool.function.arguments }}
              </div>
            </div>
          </div>

          <!-- Reasoning -->
          <div v-if="messageContent.reasoning_content" class="mt-3">
            <div
              class="text-xs opacity-70 italic border-l-2 border-current pl-3 py-1"
            >
              {{ messageContent.reasoning_content }}
            </div>
          </div>

          <!-- Actions -->
          <div
            v-if="
              message.type === 'assistant' && message.status === 'completed'
            "
            class="absolute -bottom-8 left-0 flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity"
          >
            <Tooltip content="复制">
              <Button
                variant="ghost"
                size="icon"
                class="h-7 w-7 hover:bg-muted"
                @click="copyMessage"
              >
                <Copy class="h-3.5 w-3.5" />
              </Button>
            </Tooltip>

            <Tooltip content="点赞">
              <Button
                variant="ghost"
                size="icon"
                class="h-7 w-7 hover:bg-muted"
              >
                <ThumbsUp class="h-3.5 w-3.5" />
              </Button>
            </Tooltip>

            <Tooltip content="点踩">
              <Button
                variant="ghost"
                size="icon"
                class="h-7 w-7 hover:bg-muted"
              >
                <ThumbsDown class="h-3.5 w-3.5" />
              </Button>
            </Tooltip>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
/**
 * ChatMessage 组件 - 聊天消息显示
 * 显示单条消息，包括用户消息和 AI 助手消息
 */
import { computed } from "vue"
import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Button } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import {
  User,
  Bot,
  Copy,
  ThumbsUp,
  ThumbsDown,
  AlertCircle,
  Wrench,
} from "lucide-vue-next"
import MarkdownRenderer from "./MarkdownRenderer.vue"

/**
 * 组件属性定义
 */
const props = defineProps({
  /** 消息对象 */
  message: {
    type: Object,
    required: true,
  },
  /** 是否为最后一条消息 */
  isLast: {
    type: Boolean,
    default: false,
  },
})

/**
 * 解析消息内容
 * 将 JSON 格式的消息内容解析为对象
 */
const messageContent = computed(() => {
  try {
    return JSON.parse(props.message.content) || {}
  } catch {
    return { text: props.message.content }
  }
})

/**
 * 格式化时间显示
 * @param isoString - ISO 格式的时间字符串
 * @returns 格式化后的时间文本 (HH:MM)
 */
function formatTime(isoString) {
  if (!isoString) return ""
  const date = new Date(isoString)
  return date.toLocaleTimeString("zh-CN", {
    hour: "2-digit",
    minute: "2-digit",
  })
}

/**
 * 复制消息内容到剪贴板
 */
function copyMessage() {
  const text = messageContent.value.text || ""
  navigator.clipboard.writeText(text)
}
</script>

<style scoped>
.animate-fade-in {
  animation: fadeIn 0.3s ease-out;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(8px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 确保用户消息中的 Markdown 内容继承主题颜色 */
:deep(.markdown-content) {
  color: inherit;
}

:deep(.markdown-content *) {
  color: inherit;
}
</style>
