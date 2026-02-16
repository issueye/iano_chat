/**
 * 聊天 Store 兼容层
 * 整合 session、agent、message 三个 store，提供向后兼容的接口
 */
import { defineStore, storeToRefs } from 'pinia'
import { useSessionStore } from './session'
import { useAgentStore } from './agent'
import { useMessageStore } from './message'

/**
 * 统一的聊天 Store
 * 保持与原有 chat.js 的接口兼容性
 */
export const useChatStore = defineStore('chat', () => {
  const sessionStore = useSessionStore()
  const agentStore = useAgentStore()
  const messageStore = useMessageStore()

  const { sessions, currentSessionId, searchKeyword, currentSession, filteredSessions } = storeToRefs(sessionStore)
  const { agents, currentAgentId, mainAgents, currentAgent } = storeToRefs(agentStore)
  const { messages, isLoading, error, currentMessages, connectionStatus } = storeToRefs(messageStore)

  return {
    sessions,
    currentSessionId,
    searchKeyword,
    currentSession,
    filteredSessions,
    agents,
    currentAgentId,
    mainAgents,
    currentAgent,
    messages,
    isLoading,
    error,
    currentMessages,
    connectionStatus,

    setCurrentSession: sessionStore.setCurrentSession,
    setSearchKeyword: sessionStore.setSearchKeyword,
    fetchSessions: sessionStore.fetchSessions,
    createSession: sessionStore.createSession,
    deleteSession: sessionStore.deleteSession,
    updateSessionTitle: sessionStore.updateSessionTitle,

    setCurrentAgent: agentStore.setCurrentAgent,
    fetchAgents: agentStore.fetchAgents,

    addMessage: messageStore.addMessage,
    updateMessage: messageStore.updateMessage,
    setLoading: messageStore.setLoading,
    setError: messageStore.setError,
    clearError: messageStore.clearError,
    clearCurrentSession: messageStore.clearCurrentSession,
    cancelStreaming: messageStore.cancelStreaming,
    fetchMessagesBySession: messageStore.fetchMessagesBySession,
    sendMessage: messageStore.sendMessage,
    sendMessageNonStreaming: messageStore.sendMessageNonStreaming,
  }
})

export { useSessionStore } from './session'
export { useAgentStore } from './agent'
export { useMessageStore } from './message'
export { API_BASE, WS_BASE } from './config'
