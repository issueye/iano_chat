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

          <!-- Agent Selector -->
          <Popover v-model:open="agentPopoverOpen">
            <PopoverTrigger as-child>
              <Button
                variant="ghost"
                class="h-8 px-2 gap-1.5 hover:bg-muted transition-colors"
                :disabled="isLoading"
              >
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
                  <ChevronDown class="h-3.5 w-3.5 text-muted-foreground" />
                </div>
              </Button>
            </PopoverTrigger>
            <PopoverContent class="w-[280px] p-0" align="start" side="top" :side-offset="8">
              <div class="p-2 border-b">
                <p class="text-xs font-medium text-muted-foreground px-2 py-1">选择 Agent</p>
              </div>
              <Command>
                <CommandInput placeholder="搜索 Agent..." />
                <CommandList>
                  <CommandEmpty>没有找到匹配的 Agent</CommandEmpty>
                  <CommandGroup>
                    <CommandItem
                      v-for="agent in agents"
                      :key="agent.id"
                      :value="agent.id"
                      @select="selectAgent(agent.id)"
                      class="flex items-start gap-3 py-3 px-2 cursor-pointer"
                    >
                      <div
                        class="w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium shrink-0 mt-0.5"
                        :class="agent.color || 'bg-muted text-muted-foreground'"
                      >
                        {{ agent.name?.[0]?.toUpperCase() || 'A' }}
                      </div>
                      <div class="flex-1 min-w-0">
                        <div class="flex items-center gap-2">
                          <span class="text-sm font-medium">{{ agent.name }}</span>
                          <Check
                            v-if="selectedAgentId === agent.id"
                            class="h-3.5 w-3.5 text-primary"
                          />
                        </div>
                        <p class="text-xs text-muted-foreground mt-0.5 line-clamp-2">
                          {{ agent.description }}
                        </p>
                      </div>
                    </CommandItem>
                  </CommandGroup>
                </CommandList>
              </Command>
            </PopoverContent>
          </Popover>
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
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover"
import { Command, CommandEmpty, CommandGroup, CommandInput, CommandItem, CommandList } from "@/components/ui/command"
import { Paperclip, Send, Loader2, ChevronDown, Check } from "lucide-vue-next"

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
const agentPopoverOpen = ref(false)

watch(() => props.modelValue, (val) => {
  selectedAgentId.value = val
})

watch(selectedAgentId, (val) => {
  emit("update:modelValue", val)
})

const currentAgent = computed(() => {
  return props.agents.find(a => a.id === selectedAgentId.value)
})

const canSend = computed(() => {
  return inputText.value.trim() && !props.isLoading
})

function selectAgent(agentId) {
  selectedAgentId.value = agentId
  agentPopoverOpen.value = false
}

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

<style scoped>
.line-clamp-2 {
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
</style>
