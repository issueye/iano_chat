<template>
  <header
    class="sticky top-0 z-30 flex items-center justify-between px-4 sm:px-6 py-3 sm:py-4 border-b border-border bg-card/90 backdrop-blur-md"
  >
    <div class="flex items-center gap-3">
      <Button
        variant="ghost"
        size="icon"
        class="lg:hidden hover:bg-muted"
        @click="emit('toggle-sidebar')"
      >
        <Menu class="h-5 w-5 text-muted-foreground" />
      </Button>
      <div class="flex items-center gap-1">
        <h1 class="font-semibold text-base text-foreground">
          {{ title || "新会话" }}
        </h1>
        <p class="text-xs text-muted-foreground">
          {{ messageCount }} 条消息
        </p>
      </div>
    </div>

    <div class="flex items-center gap-3">
      <!-- 连接状态 -->
      <div class="flex items-center gap-1.5">
        <span
          :class="[
            'w-2 h-2 rounded-full',
            statusConfig.bgColor
          ]"
        ></span>
        <span :class="['text-xs', statusConfig.textColor]">
          {{ statusConfig.text }}
        </span>
      </div>

      <Button
        variant="ghost"
        size="icon"
        class="hover:bg-muted"
        title="清空对话"
        @click="emit('clear-chat')"
      >
        <Trash2 class="h-4 w-4 text-muted-foreground" />
      </Button>
      <Button
        variant="ghost"
        size="icon"
        class="hover:bg-muted"
        :title="isDark ? '切换到亮色模式' : '切换到暗色模式'"
        @click="emit('toggle-theme')"
      >
        <Sun v-if="isDark" class="h-4 w-4 text-muted-foreground" />
        <Moon v-else class="h-4 w-4 text-muted-foreground" />
      </Button>
    </div>
  </header>
</template>

<script setup>
/**
 * ChatHeader 组件 - 聊天界面头部
 * 显示当前会话标题、消息数量和操作按钮
 */
import { computed } from "vue"
import { Button } from "@/components/ui/button"
import { Menu, Trash2, Sun, Moon } from "lucide-vue-next"

/**
 * 组件属性定义
 */
const props = defineProps({
  /** 会话标题 */
  title: {
    type: String,
    default: "",
  },
  /** 消息数量 */
  messageCount: {
    type: Number,
    default: 0,
  },
  /** 是否为暗色模式 */
  isDark: {
    type: Boolean,
    default: false,
  },
  /** 连接状态 */
  connectionStatus: {
    type: String,
    default: "disconnected",
  },
})

/** 组件事件定义 */
const emit = defineEmits(["toggle-sidebar", "clear-chat", "toggle-theme"])

const statusConfig = computed(() => {
  switch (props.connectionStatus) {
    case "connected":
      return {
        text: "已连接",
        bgColor: "bg-green-500",
        textColor: "text-green-500",
      }
    case "connecting":
      return {
        text: "连接中",
        bgColor: "bg-yellow-500 animate-pulse",
        textColor: "text-yellow-500",
      }
    case "reconnecting":
      return {
        text: "重连中",
        bgColor: "bg-yellow-500 animate-pulse",
        textColor: "text-yellow-500",
      }
    case "disconnected":
    default:
      return {
        text: "已断开",
        bgColor: "bg-red-500",
        textColor: "text-red-500",
      }
  }
})
</script>
