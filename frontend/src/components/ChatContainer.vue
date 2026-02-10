<template>
  <div class="flex h-screen bg-background relative overflow-hidden">
    <!-- Session List Sidebar -->
    <aside
      :class="[
        'fixed inset-y-0 left-0 z-50 w-72 bg-sidebar border-r border-sidebar-border transform transition-transform duration-300 ease-in-out lg:relative lg:transform-none',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
      ]"
    >
      <SessionList
        :sessions="chatStore.sessions"
        :current-session-id="chatStore.currentSessionId"
        @select="handleSessionSelect"
        @create="handleCreateSession"
        @delete="handleDeleteSession"
      />
    </aside>

    <!-- Sidebar Overlay (Mobile) -->
    <div
      v-if="isSidebarOpen"
      class="fixed inset-0 bg-black/20 z-40 lg:hidden backdrop-blur-sm"
      @click="isSidebarOpen = false"
    />

    <!-- Main Chat Area -->
    <main class="flex-1 flex flex-col min-w-0 relative bg-background">
      <!-- Header -->
      <header
        class="sticky top-0 z-30 flex items-center justify-between px-4 sm:px-6 py-3 sm:py-4 border-b border-border bg-card/90 backdrop-blur-md"
      >
        <div class="flex items-center gap-3">
          <Button
            variant="ghost"
            size="icon"
            class="lg:hidden hover:bg-muted"
            @click="isSidebarOpen = true"
          >
            <Menu class="h-5 w-5 text-muted-foreground" />
          </Button>
          <div class="flex items-center gap-1">
            <h1 class="font-semibold text-base text-foreground">
              {{ currentSession?.title || "新会话" }}
            </h1>
            <p class="text-xs text-muted-foreground">
              {{ chatStore.messages.length }} 条消息
            </p>
          </div>
        </div>

        <div class="flex items-center gap-1">
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <Button
                  variant="ghost"
                  size="icon"
                  class="hover:bg-muted"
                  @click="clearChat"
                >
                  <Trash2 class="h-4 w-4 text-muted-foreground" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>清空对话</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>

          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger as-child>
                <Button variant="ghost" size="icon" class="hover:bg-muted">
                  <Settings class="h-4 w-4 text-muted-foreground" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>
                <p>设置</p>
              </TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </div>
      </header>

      <!-- Messages Area -->
      <ScrollArea class="flex-1 px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        <div class="max-w-2xl mx-auto space-y-6 pb-20">
          <!-- Welcome Screen -->
          <div
            v-if="!chatStore.messages.length"
            class="text-center py-12 sm:py-16 px-4"
          >
            <!-- Simple Icon -->
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

            <!-- Clean Quick Actions -->
            <div
              class="grid grid-cols-1 sm:grid-cols-2 gap-3 max-w-sm sm:max-w-md mx-auto px-4 sm:px-0"
            >
              <button
                v-for="action in quickActions"
                :key="action.text"
                class="group flex items-center gap-3 p-3 sm:p-4 rounded-xl bg-card border border-border hover:border-primary hover:bg-secondary transition-all duration-200 text-left"
                @click="sendQuickAction(action.text)"
              >
                <div
                  class="w-9 h-9 sm:w-10 sm:h-10 rounded-lg bg-secondary flex items-center justify-center group-hover:bg-muted transition-colors shrink-0"
                >
                  <component
                    :is="action.icon"
                    class="w-4 h-4 sm:w-5 sm:h-5 text-muted-foreground group-hover:text-foreground transition-colors"
                  />
                </div>
                <span class="text-sm font-medium text-foreground truncate">{{
                  action.text
                }}</span>
              </button>
            </div>

            <!-- Simple Tips -->
            <div
              class="mt-8 sm:mt-10 flex flex-wrap justify-center gap-2 text-xs text-muted-foreground px-4"
            >
              <span class="px-3 py-1.5 bg-muted rounded-full"
                >支持代码高亮</span
              >
              <span class="px-3 py-1.5 bg-muted rounded-full">多轮对话</span>
              <span class="px-3 py-1.5 bg-muted rounded-full">快速提问</span>
            </div>
          </div>

          <!-- Messages -->
          <template v-else>
            <ChatMessage
              v-for="(message, index) in chatStore.messages"
              :key="message.id"
              :message="message"
              :is-last="index === chatStore.messages.length - 1"
            />
          </template>

          <!-- Loading Indicator -->
          <div v-if="chatStore.isLoading" class="flex justify-center py-4">
            <div class="flex items-center gap-2 text-muted-foreground">
              <Loader2 class="w-4 h-4 animate-spin text-primary" />
              <span class="text-sm">AI 正在思考...</span>
            </div>
          </div>

          <!-- Error Message -->
          <div
            v-if="chatStore.error"
            class="p-4 rounded-lg bg-destructive/10 text-destructive text-sm flex items-center gap-2"
          >
            <AlertCircle class="w-4 h-4 shrink-0" />
            {{ chatStore.error }}
          </div>
        </div>
      </ScrollArea>

      <!-- Input Area -->
      <div class="border-t border-border bg-card p-4 sm:p-6">
        <div class="w-full mx-auto p-2 sm:px-0">
          <ChatInput
            :is-loading="chatStore.isLoading"
            @send="handleSendMessage"
          />
        </div>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { useChatStore } from "@/stores/chat";
import SessionList from "./SessionList.vue";
import ChatMessage from "./ChatMessage.vue";
import ChatInput from "./ChatInput.vue";
import { Button } from "@/components/ui/button";
import { ScrollArea } from "@/components/ui/scroll-area";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  Menu,
  Trash2,
  Settings,
  Sparkles,
  Loader2,
  AlertCircle,
  Code,
  FileText,
  Lightbulb,
  MessageSquare,
} from "lucide-vue-next";

const chatStore = useChatStore();
const isSidebarOpen = ref(false);

const currentSession = computed(() => {
  return chatStore.sessions.find((s) => s.id === chatStore.currentSessionId);
});

const quickActions = [
  { icon: Code, text: "帮我写一段代码" },
  { icon: FileText, text: "帮我写一篇文章" },
  { icon: Lightbulb, text: "给我一些创意" },
  { icon: MessageSquare, text: "随便聊聊" },
];

onMounted(() => {
  chatStore.fetchSessions();
});

async function handleCreateSession() {
  await chatStore.createSession();
  isSidebarOpen.value = false;
}

async function handleSessionSelect(sessionId) {
  await chatStore.switchSession(sessionId);
  isSidebarOpen.value = false;
}

async function handleDeleteSession(sessionId) {
  console.log("Delete session:", sessionId);
}

async function handleSendMessage(content) {
  await chatStore.sendMessage(content);
}

async function sendQuickAction(text) {
  await chatStore.sendMessage(text);
}

function clearChat() {
  chatStore.clearCurrentSession();
}
</script>
