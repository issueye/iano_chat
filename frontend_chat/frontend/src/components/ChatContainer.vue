<template>
  <div class="flex h-screen bg-background relative overflow-hidden">
    <aside
      :class="[
        'fixed inset-y-0 left-0 z-50 w-72 bg-sidebar border-r border-sidebar-border transform transition-transform duration-300 ease-in-out lg:relative lg:transform-none',
        isSidebarOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0',
      ]"
    >
      <SessionList
        :sessions="chatStore.filteredSessions"
        :current-session-id="chatStore.currentSessionId"
        :search-keyword="chatStore.searchKeyword"
        @select="handleSessionSelect"
        @create="handleCreateSession"
        @delete="handleDeleteSession"
        @update:search-keyword="chatStore.setSearchKeyword"
      />
    </aside>

    <div
      v-if="isSidebarOpen"
      class="fixed inset-0 bg-black/20 z-40 lg:hidden backdrop-blur-sm"
      @click="isSidebarOpen = false"
    />

    <main class="flex-1 flex flex-col min-w-0 relative bg-background">
      <ChatHeader
        :title="chatStore.currentSession?.title"
        :message-count="chatStore.currentMessages.length"
        :is-dark="themeStore.isDark"
        @toggle-sidebar="isSidebarOpen = true"
        @clear-chat="clearChat"
        @toggle-theme="themeStore.toggleTheme"
      />

      <ChatMessages
        :messages="chatStore.currentMessages"
        :is-loading="chatStore.isLoading"
        :error="chatStore.error"
      >
        <template #welcome>
          <ChatWelcome v-if="!chatStore.currentMessages.length" @quick-action="sendQuickAction" />
        </template>
      </ChatMessages>

      <ChatInputArea
        :is-loading="chatStore.isLoading"
        :agents="chatStore.mainAgents"
        :current-agent-id="chatStore.currentAgentId"
        @send="handleSendMessage"
        @update:agent="chatStore.setCurrentAgent"
        @select-directory="handleSelectDirectory"
      />
    </main>
  </div>
</template>

<script setup>
/**
 * ChatContainer 组件 - 聊天主容器
 * 整合侧边栏、头部、消息列表和输入区域
 */
import { ref, computed, onMounted } from "vue"
import { useChatStore } from "@/stores/chat"
import { useThemeStore } from "@/stores/theme"
import SessionList from "./SessionList.vue"
import ChatHeader from "./ChatHeader.vue"
import ChatMessages from "./ChatMessages.vue"
import ChatWelcome from "./ChatWelcome.vue"
import ChatInputArea from "./ChatInputArea.vue"

const chatStore = useChatStore()
const themeStore = useThemeStore()

/** 侧边栏是否打开 */
const isSidebarOpen = ref(false)

/**
 * 组件挂载时初始化
 */
onMounted(() => {
  chatStore.fetchSessions()
  chatStore.fetchAgents()
  themeStore.initTheme()
})

/**
 * 创建新会话
 */
async function handleCreateSession() {
  await chatStore.createSession()
  isSidebarOpen.value = false
}

/**
 * 选择会话
 * @param sessionId - 会话 ID
 */
async function handleSessionSelect(sessionId) {
  chatStore.setCurrentSession(sessionId)
  await chatStore.fetchMessagesBySession(sessionId)
  isSidebarOpen.value = false
}

/**
 * 删除会话
 * @param sessionId - 会话 ID
 */
async function handleDeleteSession(sessionId) {
  if (confirm("确定要删除这个会话吗？")) {
    await chatStore.deleteSession(sessionId)
  }
}

/**
 * 发送消息
 * @param content - 消息内容
 * @param directory - 可选的目录路径
 */
async function handleSendMessage(content, directory) {
  await chatStore.sendMessage(content, directory)
}

/**
 * 发送快捷操作消息
 * @param text - 消息文本
 */
async function sendQuickAction(text) {
  await chatStore.sendMessage(text)
}

/**
 * 清空当前会话消息
 */
function clearChat() {
  chatStore.clearCurrentSession()
}

/**
 * 处理目录选择
 * @param path - 选中的目录路径
 */
function handleSelectDirectory(path) {
  console.log("Selected directory:", path)
}
</script>
