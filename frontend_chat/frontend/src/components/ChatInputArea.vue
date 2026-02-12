<template>
  <div class="border-t border-border bg-card p-2 sm:p-6">
    <div class="w-full mx-auto p-2 sm:px-0">
      <ChatInput
        :is-loading="isLoading"
        :agents="agents"
        :model-value="currentAgentId"
        @update:model-value="emit('update:agent', $event)"
        @send="(text, directory) => emit('send', text, directory)"
        @select-directory="emit('select-directory', $event)"
      />
    </div>
  </div>
</template>

<script setup>
/**
 * ChatInputArea 组件 - 输入区域包装器
 * 包装 ChatInput 组件，提供统一的输入区域布局
 */
import ChatInput from "./ChatInput.vue"

/**
 * 组件属性定义
 */
defineProps({
  /** 是否正在加载 */
  isLoading: {
    type: Boolean,
    default: false,
  },
  /** 可用的 Agent 列表 */
  agents: {
    type: Array,
    default: () => [],
  },
  /** 当前选中的 Agent ID */
  currentAgentId: {
    type: String,
    default: "",
  },
})

/** 组件事件定义 */
const emit = defineEmits(["send", "update:agent", "select-directory"])
</script>
