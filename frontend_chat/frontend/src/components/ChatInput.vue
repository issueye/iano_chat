<template>
  <div class="relative w-full">
    <div
      class="relative flex flex-col gap-3 bg-card rounded-2xl border border-border shadow-sm p-3 transition-all duration-200 focus-within:ring-1 focus-within:ring-ring/30 focus-within:border-ring/50"
      :class="{ 'ring-1 ring-ring/20': isLoading }"
    >
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

      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <Tooltip content="添加附件">
            <Button
              variant="ghost"
              size="icon"
              class="shrink-0 rounded-xl hover:bg-muted transition-colors h-8 w-8"
            >
              <Paperclip class="h-4 w-4 text-muted-foreground" />
            </Button>
          </Tooltip>

          <div class="h-4 w-px bg-border"></div>

          <Select
            :model-value="modelValue"
            :disabled="isLoading"
            @update:model-value="emit('update:modelValue', $event)"
          >
            <SelectTrigger class="h-8 w-auto min-w-[140px] border border-border/80 bg-muted/40 hover:bg-muted/80 px-2 gap-1.5 rounded-lg shadow-sm">
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium shadow-sm"
                  :class="currentAgent?.color || 'bg-primary text-primary-foreground'"
                >
                  {{ currentAgent?.name?.[0]?.toUpperCase() || 'A' }}
                </div>
                <span class="text-sm font-medium text-foreground truncate max-w-[100px]">
                  {{ currentAgent?.name || '选择 Agent' }}
                </span>
              </div>
            </SelectTrigger>
            <SelectContent align="start" side="top" :side-offset="8">
              <SelectItem
                v-for="agent in agents"
                :key="agent.id"
                :value="agent.id"
              >
                <div class="flex items-center gap-3 py-1">
                  <div
                    class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium shrink-0"
                    :class="agent.color || 'bg-muted text-muted-foreground'"
                  >
                    {{ agent.name?.[0]?.toUpperCase() || 'A' }}
                  </div>
                  <div class="flex-1 min-w-0">
                    <span class="text-sm font-medium">{{ agent.name }}</span>
                  </div>
                </div>
              </SelectItem>
            </SelectContent>
          </Select>

          <div class="h-4 w-px bg-border"></div>

          <div
            class="flex items-center min-w-[140px] h-[34px] gap-1.5 px-2 py-1.5 rounded-lg bg-muted/50 border border-border/50 cursor-pointer hover:bg-muted transition-colors"
            @click="selectDirectory"
          >
            <FolderOpen class="h-4 w-4 text-muted-foreground shrink-0" />
            <span class="text-xs text-muted-foreground whitespace-nowrap">
              {{ selectedDirectory || '选择目录' }}
            </span>
            <X
              v-if="selectedDirectory"
              class="h-3 w-3 text-muted-foreground hover:text-foreground shrink-0"
              @click.stop="clearDirectory"
            />
          </div>
        </div>

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
/**
 * ChatInput 组件 - 聊天输入区域
 * 提供消息输入、Agent 选择、目录选择等功能
 */
import { ref, computed } from "vue"
import { Button } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select"
import { Paperclip, Send, Loader2, FolderOpen, X } from "lucide-vue-next"
import { SelectDirectory } from "@/lib/wails/go/main/App"

/**
 * 组件属性定义
 */
const props = defineProps({
  /** 是否正在加载中 */
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
  modelValue: {
    type: String,
    default: "",
  },
})

/** 组件事件定义 */
const emit = defineEmits(["send", "update:modelValue", "select-directory"])

/** 输入文本内容 */
const inputText = ref("")
/** textarea 元素引用 */
const textareaRef = ref(null)
/** 选中的目录路径 */
const selectedDirectory = ref("")

/**
 * 计算当前选中的 Agent 对象
 */
const currentAgent = computed(() => {
  return props.agents.find(a => a.id === props.modelValue)
})

/**
 * 计算是否可以发送消息
 */
const canSend = computed(() => {
  return inputText.value.trim() && !props.isLoading
})

/**
 * 自动调整 textarea 高度
 */
function autoResize() {
  const textarea = textareaRef.value
  if (textarea) {
    textarea.style.height = "auto"
    textarea.style.height = Math.min(textarea.scrollHeight, 128) + "px"
  }
}

/**
 * 处理回车键事件
 * @param event - 键盘事件对象
 */
function handleEnter(event) {
  if (event.shiftKey) {
    return
  }
  sendMessage()
}

/**
 * 发送消息
 */
function sendMessage() {
  const text = inputText.value.trim()
  if (!text || props.isLoading) return

  emit("send", text, selectedDirectory.value)
  inputText.value = ""

  const textarea = textareaRef.value
  if (textarea) {
    textarea.style.height = "auto"
  }
}

/**
 * 选择目录
 * 调用 Wails 后端打开目录选择对话框
 */
async function selectDirectory() {
  try {
    const path = await SelectDirectory()
    if (path) {
      selectedDirectory.value = path
      emit("select-directory", path)
    }
  } catch (error) {
    console.error("选择目录失败:", error)
  }
}

/**
 * 清除已选择的目录
 */
function clearDirectory() {
  selectedDirectory.value = ""
}
</script>
