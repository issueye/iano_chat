<template>
  <div class="text-center py-12 sm:py-16 px-4">
    <div
      class="w-14 h-14 sm:w-16 sm:h-16 mx-auto mb-5 sm:mb-6 rounded-2xl bg-secondary flex items-center justify-center"
    >
      <Sparkles class="w-7 h-7 sm:w-8 sm:h-8 text-foreground" />
    </div>

    <h2 class="text-xl sm:text-2xl font-semibold mb-2 text-foreground">
      有什么可以帮您的？
    </h2>
    <p class="text-muted-foreground mb-8 sm:mb-10 text-sm sm:text-base">
      开始一个新对话，或从左侧选择一个会话
    </p>

    <div
      class="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-sm sm:max-w-md mx-auto px-4 sm:px-0"
    >
      <button
        v-for="action in quickActions"
        :key="action.text"
        class="group flex items-center gap-3 p-3 sm:p-4 rounded-xl bg-card border border-border hover:border-primary hover:bg-secondary transition-all duration-200 text-left"
        @click="emit('quick-action', action.text)"
      >
        <div
          class="w-9 h-9 sm:w-10 sm:h-10 rounded-lg bg-secondary flex items-center justify-center group-hover:bg-muted transition-colors shrink-0"
        >
          <component
            :is="action.icon"
            class="w-4 h-4 sm:w-5 sm:h-5 text-muted-foreground group-hover:text-foreground transition-colors"
          />
        </div>
        <span class="text-sm font-medium text-foreground truncate">
          {{ action.text }}
        </span>
      </button>
    </div>

    <div
      class="mt-8 sm:mt-10 flex flex-wrap justify-center gap-2 text-xs text-muted-foreground px-4"
    >
      <span class="px-3 py-1.5 bg-muted rounded-full">支持代码高亮</span>
      <span class="px-3 py-1.5 bg-muted rounded-full">多轮对话</span>
      <span class="px-3 py-1.5 bg-muted rounded-full">快速提问</span>
    </div>
  </div>
</template>

<script setup>
/**
 * ChatWelcome 组件 - 欢迎界面
 * 显示欢迎信息和快捷操作按钮
 */
import { Sparkles, Code, FileText, Lightbulb, MessageSquare } from "lucide-vue-next"

/** 快捷操作列表 */
const quickActions = [
  { icon: Code, text: "帮我写一段代码" },
  { icon: FileText, text: "帮我写一篇文章" },
  { icon: Lightbulb, text: "给我一些创意" },
  { icon: MessageSquare, text: "随便聊聊" },
]

/** 组件事件定义 */
const emit = defineEmits(["quick-action"])
</script>
