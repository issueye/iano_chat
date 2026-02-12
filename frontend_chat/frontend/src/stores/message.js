/**
 * 消息管理 Store
 * 管理聊天消息的发送、接收、更新等操作
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { API_BASE } from './config'
import { useSessionStore } from './session'
import { useAgentStore } from './agent'

export const useMessageStore = defineStore('message', () => {
  /** 消息列表 */
  const messages = ref([])
  /** 是否正在加载 */
  const isLoading = ref(false)
  /** 错误信息 */
  const error = ref(null)
  /** 用于取消请求的 AbortController */
  const abortController = ref(null)

  /**
   * 当前会话的消息列表
   */
  const currentMessages = computed(() => {
    const sessionStore = useSessionStore()
    return messages.value.filter(m => String(m.session_id) === String(sessionStore.currentSessionId))
  })

  /**
   * 添加消息
   * @param message - 消息对象
   */
  function addMessage(message) {
    const sessionStore = useSessionStore()
    messages.value.push({
      id: message.id || Date.now().toString(),
      session_id: message.session_id || sessionStore.currentSessionId,
      created_at: message.created_at || new Date().toISOString(),
      ...message
    })
  }

  /**
   * 更新消息
   * @param id - 消息 ID
   * @param updates - 更新内容
   */
  function updateMessage(id, updates) {
    const index = messages.value.findIndex(m => String(m.id) === String(id))
    if (index !== -1) {
      messages.value[index] = { ...messages.value[index], ...updates }
    }
  }

  /**
   * 设置加载状态
   * @param loading - 是否加载中
   */
  function setLoading(loading) {
    isLoading.value = loading
  }

  /**
   * 设置错误信息
   * @param err - 错误信息
   */
  function setError(err) {
    error.value = err
  }

  /**
   * 清除错误信息
   */
  function clearError() {
    error.value = null
  }

  /**
   * 清空当前会话的消息
   */
  function clearCurrentSession() {
    const sessionStore = useSessionStore()
    messages.value = messages.value.filter(m => String(m.session_id) !== String(sessionStore.currentSessionId))
  }

  /**
   * 取消流式传输
   */
  function cancelStreaming() {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
    }
    setLoading(false)
  }

  /**
   * 获取指定会话的消息
   * @param sessionId - 会话 ID
   */
  async function fetchMessagesBySession(sessionId) {
    try {
      const response = await fetch(`${API_BASE}/messages/session?session_id=${sessionId}`)
      const data = await response.json()
      if (data.code === 200) {
        const sessionMessages = data.data || []
        messages.value = messages.value.filter(m => String(m.session_id) !== String(sessionId))
        messages.value.push(...sessionMessages)
      }
    } catch (err) {
      setError(err.message)
    }
  }

  /**
   * 发送消息（流式响应）
   * @param content - 消息内容
   * @param directory - 可选的目录路径
   */
  async function sendMessage(content, directory) {
    const sessionStore = useSessionStore()
    const agentStore = useAgentStore()

    if (!sessionStore.currentSessionId) {
      await sessionStore.createSession()
    }

    const userMessage = {
      id: Date.now().toString() + '_user',
      session_id: String(sessionStore.currentSessionId),
      type: 'user',
      content: JSON.stringify({ text: content }),
      status: 'completed'
    }

    addMessage(userMessage)
    setLoading(true)
    clearError()

    const assistantMessageId = Date.now().toString() + '_assistant'
    const assistantMessage = {
      id: assistantMessageId,
      session_id: String(sessionStore.currentSessionId),
      type: 'assistant',
      content: JSON.stringify({ text: '', tool_calls: [] }),
      status: 'streaming'
    }
    addMessage(assistantMessage)

    try {
      const streamResponse = await fetch(`${API_BASE}/chat/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: String(sessionStore.currentSessionId),
          agent_id: agentStore.currentAgentId,
          message: content,
          directory: directory || undefined
        })
      })

      if (!streamResponse.ok) {
        throw new Error(`HTTP error! status: ${streamResponse.status}`)
      }

      const reader = streamResponse.body.getReader()
      const decoder = new TextDecoder()
      let accumulatedContent = ''
      let accumulatedToolCalls = []

      // 用于跟踪当前事件类型
      let currentEventType = 'message'

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            // 记录当前事件类型
            currentEventType = line.slice(7).trim()
          } else if (line.startsWith('data: ')) {
            try {
              const eventData = JSON.parse(line.slice(6))

              // 根据事件类型处理数据
              if (currentEventType === 'message') {
                if (eventData.content) {
                  accumulatedContent += eventData.content
                }
                if (eventData.error) {
                  setError(eventData.error)
                  updateMessage(assistantMessageId, { status: 'failed' })
                }
                // 更新消息显示
                updateMessage(assistantMessageId, {
                  content: JSON.stringify({
                    text: accumulatedContent,
                    tool_calls: accumulatedToolCalls
                  })
                })
              } else if (currentEventType === 'tool_call') {
                // 处理工具调用事件
                if (eventData.id && eventData.name) {
                  accumulatedToolCalls.push({
                    id: eventData.id,
                    type: 'function',
                    function: {
                      name: eventData.name,
                      arguments: eventData.arguments
                    }
                  })
                  // 更新消息显示工具调用
                  updateMessage(assistantMessageId, {
                    content: JSON.stringify({
                      text: accumulatedContent,
                      tool_calls: accumulatedToolCalls
                    })
                  })
                }
              }
            } catch (e) {
              // 忽略不完整 JSON 的解析错误
            }
          } else if (line.trim() === '') {
            // 空行表示事件结束，重置事件类型
            currentEventType = 'message'
          }
        }
      }

      updateMessage(assistantMessageId, { status: 'completed' })

    } catch (err) {
      setError(err.message)
      updateMessage(assistantMessageId, { status: 'failed' })
    } finally {
      setLoading(false)
    }
  }

  /**
   * 发送消息（非流式响应）
   * @param content - 消息内容
   */
  async function sendMessageNonStreaming(content) {
    const sessionStore = useSessionStore()
    const agentStore = useAgentStore()

    if (!sessionStore.currentSessionId) {
      await sessionStore.createSession()
    }

    const userMessage = {
      id: Date.now().toString() + '_user',
      session_id: String(sessionStore.currentSessionId),
      type: 'user',
      content: JSON.stringify({ text: content }),
      status: 'completed'
    }

    addMessage(userMessage)
    setLoading(true)
    clearError()

    try {
      const response = await fetch(`${API_BASE}/chat`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: String(sessionStore.currentSessionId),
          agent_id: agentStore.currentAgentId,
          message: content
        })
      })
      const data = await response.json()
      if (data.code === 200 && data.data) {
        const assistantMessage = {
          id: Date.now().toString() + '_assistant',
          session_id: String(sessionStore.currentSessionId),
          type: 'assistant',
          content: JSON.stringify({ text: data.data.content }),
          status: 'completed'
        }
        addMessage(assistantMessage)
      } else {
        throw new Error(data.message || 'Chat failed')
      }
    } catch (err) {
      setError(err.message)
    } finally {
      setLoading(false)
    }
  }

  return {
    messages,
    isLoading,
    error,
    currentMessages,
    addMessage,
    updateMessage,
    setLoading,
    setError,
    clearError,
    clearCurrentSession,
    cancelStreaming,
    fetchMessagesBySession,
    sendMessage,
    sendMessageNonStreaming,
  }
})
