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

          <Select v-model="selectedAgentId" :disabled="isLoading">
            <SelectTrigger class="h-8 w-auto min-w-[120px] max-w-[200px] border-0 bg-transparent hover:bg-muted px-2">
              <div class="flex items-center gap-1.5">
                <Bot class="h-3.5 w-3.5 text-muted-foreground" />
                <SelectValue placeholder="选择 Agent" class="text-xs" />
              </div>
            </SelectTrigger>
            <SelectContent>
              <SelectItem
                v-for="agent in agents"
                :key="agent.id"
                :value="agent.id"
              >
                <div class="flex items-center gap-2">
                  <span class="text-sm">{{ agent.name }}</span>
                  <span v-if="agent.description" class="text-xs text-muted-foreground truncate max-w-[150px]">
                    {{ agent.description }}
                  </span>
                </div>
              </SelectItem>
            </SelectContent>
          </Select>
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
import { ref, computed, watch } from "vue"
import { Button } from "@/components/ui/button"
import { Tooltip } from "@/components/ui/tooltip"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Paperclip, Send, Loader2, Bot } from "lucide-vue-next"

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

const emit = defineEmits(["send", "update:modelValue"])

const inputText = ref("")
const textareaRef = ref(null)
const selectedAgentId = ref(props.modelValue)

watch(() => props.modelValue, (val) => {
  selectedAgentId.value = val
})

watch(selectedAgentId, (val) => {
  emit("update:modelValue", val)
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

  emit("send", text)
  inputText.value = ""

  const textarea = textareaRef.value
  if (textarea) {
    textarea.style.height = "auto"
  }
}
</script>
