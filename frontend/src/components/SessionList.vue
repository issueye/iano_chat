<template>
  <div class="flex flex-col h-full bg-sidebar">
    <!-- Header -->
    <div class="p-4 sm:p-5 border-b border-sidebar-border">
      <div class="flex items-center gap-3 mb-4">
        <div class="w-9 h-9 rounded-xl bg-secondary flex items-center justify-center">
          <MessageSquare class="w-5 h-5 text-foreground" />
        </div>
        <div>
          <h2 class="font-semibold text-base text-sidebar-foreground">AI Chat</h2>
          <p class="text-xs text-sidebar-foreground/60">智能对话助手</p>
        </div>
      </div>

      <Button class="w-full gap-2 bg-primary hover:bg-primary/90 text-primary-foreground" @click="$emit('create')">
        <Plus class="w-4 h-4" />
        <span>新建会话</span>
      </Button>
    </div>

    <!-- Search -->
    <div class="px-4 sm:px-5 py-3">
      <div class="relative">
        <Search class="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
        <Input
          type="text"
          placeholder="搜索会话..."
          class="pl-9 pr-4 !py-2 bg-muted/50 border-0 focus:bg-muted focus:ring-1 focus:ring-ring/30"
        />
      </div>
    </div>

    <!-- Session List -->
    <ScrollArea class="flex-1 px-3 sm:px-4">
      <div class="px-3 py-2.5 text-xs font-medium text-muted-foreground uppercase tracking-wide">
        最近会话
      </div>

      <div class="space-y-1.5 mt-1">
        <div
          v-for="session in sessions"
          :key="session.id"
          @click="$emit('select', session.id)"
          :class="[
            'group flex items-center gap-3 p-2.5 sm:p-3 rounded-xl cursor-pointer transition-all duration-200',
            currentSessionId === session.id
              ? 'bg-secondary'
              : 'hover:bg-muted'
          ]"
        >
          <!-- Icon -->
          <div
            :class="[
              'w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 transition-colors',
              currentSessionId === session.id
                ? 'bg-muted text-foreground'
                : 'bg-muted text-muted-foreground'
            ]"
          >
            <MessageCircle class="w-4 h-4" />
          </div>

          <!-- Info -->
          <div class="flex-1 min-w-0">
            <div class="font-medium text-sm text-foreground truncate">
              {{ session.title || '新会话' }}
            </div>
            <div class="flex items-center gap-2 mt-0.5">
              <span class="text-xs text-muted-foreground">{{ formatTime(session.last_active_at) }}</span>
              <Badge v-if="session.message_count" variant="secondary" class="text-[10px] px-1.5 py-0 h-4 bg-muted text-muted-foreground border-0">
                {{ session.message_count }}
              </Badge>
            </div>
          </div>

          <!-- Actions -->
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="h-7 w-7 opacity-0 group-hover:opacity-100 transition-opacity hover:bg-muted"
                  @click.stop="$emit('delete', session.id)"
                >
                  <Trash2 class="w-4 h-4 text-muted-foreground group-hover:text-foreground" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>删除会话</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </div>

      <!-- Empty State -->
      <div v-if="!sessions.length" class="text-center py-10 px-4">
        <div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-muted flex items-center justify-center">
          <MessageCircle class="w-7 h-7 text-muted-foreground" />
        </div>
        <p class="text-sm text-muted-foreground font-medium">暂无会话</p>
        <p class="text-xs text-muted-foreground/70 mt-2">点击上方按钮创建新会话</p>
      </div>
    </ScrollArea>

    <!-- Footer -->
    <div class="p-4 sm:p-5 border-t border-sidebar-border">
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-2 h-2 rounded-full bg-emerald-500"></div>
          <span class="text-xs text-muted-foreground">系统正常</span>
        </div>
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" size="icon" class="h-8 w-8 hover:bg-muted">
                <Settings class="w-4 h-4 text-muted-foreground" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>设置</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
      </div>
    </div>
  </div>
</template>

<script setup>
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'
import {
  MessageSquare,
  Plus,
  Search,
  MessageCircle,
  Trash2,
  Settings
} from 'lucide-vue-next'

defineProps({
  sessions: {
    type: Array,
    default: () => []
  },
  currentSessionId: {
    type: String,
    default: null
  }
})

defineEmits(['select', 'create', 'delete'])

function formatTime(isoString) {
  if (!isoString) return ''
  const date = new Date(isoString)
  const now = new Date()
  const diff = now - date

  if (diff < 60000) return '刚刚'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}小时前`
  if (diff < 604800000) return `${Math.floor(diff / 86400000)}天前`

  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric' })
}
</script>
