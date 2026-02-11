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
            <SelectTrigger class="h-8 w-auto min-w-[140px] border-0 bg-transparent hover:bg-muted px-2 gap-1.5">
              <div class="flex items-center gap-2">
                <div
                  class="w-6 h-6 rounded-full flex items-center justify-center text-xs font-medium"
                  :class="currentAgent?.color || 'bg-muted text-muted-foreground'"
                >
                  {{ currentAgent?.name?.[0]?.toUpperCase() || 'A' }}
                </div>
                <span class="text-sm text-foreground truncate max-w-[100px]">
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

          <div>
            <input
              ref="directoryInput"
              type="file"
              webkitdirectory
              directory
              class="hidden"
              @change="handleDirectorySelect"
            />
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
import { ref, computed } from "vue"
import { Button } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import { Select, SelectContent, SelectItem, SelectTrigger } from "@/components/ui/select"
import { Paperclip, Send, Loader2, FolderOpen, X } from "lucide-vue-next"

const props = defineProps({
  isLoading: {
    type: Boolean,
    default: false,
  },
  agents: {
    type: Array,
    default: () => [],
  },
  modelValue: {
    type: String,
    default: "",
  },
})

const emit = defineEmits(["send", "update:modelValue", "select-directory"])

const inputText = ref("")
const textareaRef = ref(null)
const directoryInput = ref(null)
const selectedDirectory = ref("")

const currentAgent = computed(() => {
  return props.agents.find(a => a.id === props.modelValue)
})

const canSend = computed(() => {
  return inputText.value.trim() && !props.isLoading
})

function autoResize() {
  const textarea = textareaRef.value
  if (textarea) {
    textarea.style.height = "auto"
    textarea.style.height = Math.min(textarea.scrollHeight, 128) + "px"
  }
}

function handleEnter(event) {
  if (event.shiftKey) {
    return
  }
  sendMessage()
}

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

function selectDirectory() {
  if (directoryInput.value) {
    directoryInput.value.click()
  }
}

function handleDirectorySelect(event) {
  const files = event.target.files
  if (files && files.length > 0) {
    const file = files[0]
    // 获取目录完整绝对路径
    const fullPath = file.path ? file.path.substring(0, file.path.lastIndexOf(file.name) - 1) : file.webkitRelativePath
    selectedDirectory.value = fullPath
    // 立即将绝对路径传给后端
    emit("select-directory", fullPath)
  }
}

function clearDirectory() {
  selectedDirectory.value = ""
  if (directoryInput.value) {
    directoryInput.value.value = ""
  }
}
</script>
