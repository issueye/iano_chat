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
          <!-- Content Blocks (按顺序渲染) -->
          <div v-if="messageContent.blocks?.length" class="text-sm leading-relaxed text-inherit space-y-3">
            <template v-for="(block, index) in messageContent.blocks" :key="index">
              <!-- 文本块 -->
              <div v-if="block.type === 'text' && block.text">
                <MarkdownRenderer :content="block.text" />
              </div>
              
              <!-- 工具调用块 -->
              <div v-else-if="block.type === 'tool_call' && block.tool_call" 
                   class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30 border border-blue-200 dark:border-blue-800 rounded-lg p-3 text-xs overflow-hidden">
                <div class="flex items-center gap-2 font-medium text-blue-700 dark:text-blue-300">
                  <div class="w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center">
                    <Wrench class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" />
                  </div>
                  <span class="capitalize">{{ formatToolName(block.tool_call.function?.name) }}</span>
                </div>
                <div class="mt-2 space-y-1">
                  <div
                    v-for="(value, key) in parseToolArguments(block.tool_call.function?.arguments)"
                    :key="key"
                    class="flex items-start gap-2 text-[11px]"
                  >
                    <span class="text-blue-600 dark:text-blue-400 font-medium shrink-0">{{ key }}:</span>
                    <span class="text-gray-600 dark:text-gray-400 break-all font-mono bg-white/50 dark:bg-black/20 rounded px-1.5 py-0.5">{{ formatParamValue(value) }}</span>
                  </div>
                </div>
              </div>
            </template>
          </div>

          <!-- Fallback: 旧格式兼容 -->
          <div v-else class="text-sm leading-relaxed text-inherit">
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

          <!-- Fallback: 旧格式工具调用 -->
          <div v-if="!messageContent.blocks?.length && messageContent.tool_calls?.length" class="mt-3 space-y-2">
            <div
              v-for="tool in messageContent.tool_calls"
              :key="tool.id"
              class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-950/30 dark:to-indigo-950/30 border border-blue-200 dark:border-blue-800 rounded-lg p-3 text-xs overflow-hidden"
            >
              <div class="flex items-center gap-2 font-medium text-blue-700 dark:text-blue-300">
                <div class="w-6 h-6 rounded-full bg-blue-100 dark:bg-blue-900 flex items-center justify-center">
                  <Wrench class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" />
                </div>
                <span class="capitalize">{{ formatToolName(tool.function.name) }}</span>
              </div>
              <div class="mt-2 space-y-1">
                <div
                  v-for="(value, key) in parseToolArguments(tool.function.arguments)"
                  :key="key"
                  class="flex items-start gap-2 text-[11px]"
                >
                  <span class="text-blue-600 dark:text-blue-400 font-medium shrink-0">{{ key }}:</span>
                  <span class="text-gray-600 dark:text-gray-400 break-all font-mono bg-white/50 dark:bg-black/20 rounded px-1.5 py-0.5">{{ formatParamValue(value) }}</span>
                </div>
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

/**
 * 格式化工具名称显示
 * 将下划线分隔的名称转换为更友好的显示
 * @param name - 工具名称
 * @returns 格式化后的名称
 */
function formatToolName(name) {
  if (!name) return '未知工具'
  
  // 工具名称映射表
  const toolNameMap = {
    'file_write': '写入文件',
    'file_read': '读取文件',
    'file_list': '列出文件',
    'file_delete': '删除文件',
    'file_info': '文件信息',
    'grep_search': '搜索内容',
    'grep_replace': '替换内容',
    'command_execute': '执行命令',
    'shell_execute': '执行 Shell',
    'web_search': '网络搜索',
    'http_request': 'HTTP 请求',
    'archive_create': '创建压缩包',
    'archive_extract': '解压文件',
    'process_list': '进程列表',
    'env_get': '获取环境变量',
    'env_set': '设置环境变量',
    'system_info': '系统信息',
    'ping': '网络 Ping',
    'dns_lookup': 'DNS 查询',
    'http_headers': 'HTTP 头信息'
  }
  
  return toolNameMap[name] || name.replace(/_/g, ' ')
}

/**
 * 解析工具参数
 * 将 JSON 字符串解析为对象
 * @param args - JSON 格式的参数字符串
 * @returns 解析后的参数对象
 */
function parseToolArguments(args) {
  if (!args) return {}
  try {
    const parsed = JSON.parse(args)
    return parsed || {}
  } catch {
    // 如果解析失败，尝试处理转义的字符串
    try {
      const unescaped = args
        .replace(/\\n/g, '\n')
        .replace(/\\t/g, '\t')
        .replace(/\\"/g, '"')
      return JSON.parse(unescaped) || {}
    } catch {
      return { '参数': args }
    }
  }
}

/**
 * 格式化参数值显示
 * 对长文本进行截断处理
 * @param value - 参数值
 * @returns 格式化后的值
 */
function formatParamValue(value) {
  if (value === null || value === undefined) return ''
  
  const str = String(value)
  
  // 如果内容包含换行符，只显示第一行并添加省略号
  if (str.includes('\n')) {
    const firstLine = str.split('\n')[0].trim()
    return firstLine.length > 50 ? firstLine.substring(0, 50) + '...' : firstLine + '...'
  }
  
  // 如果内容太长，截断显示
  if (str.length > 100) {
    return str.substring(0, 100) + '...'
  }
  
  return str
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
