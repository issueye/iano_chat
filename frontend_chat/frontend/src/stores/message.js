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
  /** SSE 连接状态 */
  const connectionStatus = ref('disconnected')
  /** 用于取消请求的 AbortController */
  const abortController = ref(null)
  /** 重连配置 */
  const reconnectConfig = {
    maxRetries: 5,
    baseDelay: 1000,
    maxDelay: 10000,
    retryCount: 0
  }

  /**
   * 计算重连延迟（指数退避）
   * @param retryCount - 当前重试次数
   */
  function getRetryDelay(retryCount) {
    const delay = Math.min(
      reconnectConfig.baseDelay * Math.pow(2, retryCount),
      reconnectConfig.maxDelay
    )
    return delay + Math.random() * 500
  }

  /**
   * 设置连接状态
   * @param status - 连接状态
   */
  function setConnectionStatus(status) {
    connectionStatus.value = status
  }

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
    reconnectConfig.retryCount = reconnectConfig.maxRetries
    setConnectionStatus('disconnected')
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

    setLoading(true)
    clearError()
    reconnectConfig.retryCount = 0

    await streamChat(content, directory, sessionStore.currentSessionId, agentStore.currentAgentId)
  }

  /**
   * 流式聊天核心逻辑（支持重连，后端驱动消息创建）
   */
  async function streamChat(content, directory, sessionId, agentId) {
    let accumulatedContent = ''
    let accumulatedToolCalls = []
    let contentBlocks = []
    let assistantMessageId = null

    setConnectionStatus('connecting')

    try {
      abortController.value = new AbortController()
      
      const streamResponse = await fetch(`${API_BASE}/chat/stream`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: String(sessionId),
          agent_id: agentId,
          message: content,
          work_dir: directory || undefined
        }),
        signal: abortController.value.signal
      })

      if (!streamResponse.ok) {
        const errorText = await streamResponse.text()
        throw new Error(`HTTP ${streamResponse.status}: ${errorText}`)
      }

      const reader = streamResponse.body.getReader()
      const decoder = new TextDecoder()
      let currentEventType = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('event: ')) {
            currentEventType = line.slice(7).trim()
          } else if (line.startsWith('data: ')) {
            try {
              const eventData = JSON.parse(line.slice(6))

              if (currentEventType === 'message_created') {
                const msg = {
                  id: eventData.id,
                  session_id: eventData.session_id,
                  type: eventData.type,
                  content: eventData.content,
                  status: eventData.status || 'completed',
                  created_at: eventData.created_at
                }
                
                const existingIndex = messages.value.findIndex(m => m.id === msg.id)
                if (existingIndex === -1) {
                  messages.value.push(msg)
                }
                
                if (eventData.type === 'assistant') {
                  assistantMessageId = eventData.id
                  try {
                    const parsed = JSON.parse(eventData.content)
                    accumulatedContent = parsed.text || ''
                    accumulatedToolCalls = parsed.tool_calls || []
                    contentBlocks = parsed.blocks || []
                  } catch (e) {}
                }
              } else if (currentEventType === 'content_block' && assistantMessageId) {
                if (eventData.type === 'text' && eventData.text) {
                  accumulatedContent += eventData.text

                  const lastBlock = contentBlocks[contentBlocks.length - 1]
                  if (lastBlock && lastBlock.type === 'text') {
                    lastBlock.text += eventData.text
                  } else {
                    contentBlocks.push({ type: 'text', text: eventData.text })
                  }
                } else if (eventData.type === 'tool_call' && eventData.tool_call) {
                  const tc = eventData.tool_call
                  const toolCall = {
                    id: tc.id,
                    type: 'function',
                    function: {
                      name: tc.name,
                      arguments: tc.arguments
                    }
                  }
                  accumulatedToolCalls.push(toolCall)
                  contentBlocks.push({ type: 'tool_call', tool_call: toolCall })
                }

                updateMessage(assistantMessageId, {
                  content: JSON.stringify({
                    blocks: contentBlocks,
                    text: accumulatedContent,
                    tool_calls: accumulatedToolCalls
                  })
                })
              } else if (currentEventType === 'message_completed' && assistantMessageId) {
                updateMessage(assistantMessageId, {
                  status: eventData.status,
                  ...(eventData.content && { content: eventData.content })
                })
              } else if (currentEventType === 'error') {
                setError(eventData.error)
                setConnectionStatus('disconnected')
                if (assistantMessageId) {
                  updateMessage(assistantMessageId, { status: 'failed' })
                }
                setLoading(false)
                return
              }
            } catch (e) {
              // 忽略不完整 JSON 的解析错误
            }
          } else if (line.trim() === '') {
            currentEventType = ''
          }
        }
      }

      setConnectionStatus('connected')
      setLoading(false)

    } catch (err) {
      if (err.name === 'AbortError') {
        if (assistantMessageId) {
          updateMessage(assistantMessageId, { status: 'failed' })
        }
        setConnectionStatus('disconnected')
        setLoading(false)
        return
      }

      const isNetworkError = err.message.includes('Failed to fetch') || 
                             err.message.includes('NetworkError') ||
                             err.message.includes('ERR_NETWORK') ||
                             err.message.includes('ECONNRESET') ||
                             err.message.includes('ECONNREFUSED')

      if (isNetworkError && reconnectConfig.retryCount < reconnectConfig.maxRetries) {
        reconnectConfig.retryCount++
        const delay = getRetryDelay(reconnectConfig.retryCount - 1)
        setConnectionStatus('reconnecting')
        setError(`连接断开，正在重连 (${reconnectConfig.retryCount}/${reconnectConfig.maxRetries})...`)
        
        await new Promise(resolve => setTimeout(resolve, delay))
        
        return streamChat(content, directory, sessionId, agentId)
      }

      setError(err.message)
      setConnectionStatus('disconnected')
      if (assistantMessageId) {
        updateMessage(assistantMessageId, { status: 'failed' })
      }
      setLoading(false)
    }
  }

  /**
   * 断开 SSE 连接
   */
  function disconnect() {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
    }
    reconnectConfig.retryCount = reconnectConfig.maxRetries
    setConnectionStatus('disconnected')
    setLoading(false)
  }

  /**
   * 检查连接状态
   * @returns {string} 连接状态
   */
  function getConnectionStatus() {
    return connectionStatus.value
  }

  /**
   * 重置重连计数
   */
  function resetReconnectCount() {
    reconnectConfig.retryCount = 0
    setConnectionStatus('connected')
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

  /**
   * 发送消息反馈
   * @param messageId - 消息 ID
   * @param rating - 评分 ('like' 或 'dislike')
   * @param comment - 可选评论
   */
  async function sendFeedback(messageId, rating, comment = '') {
    try {
      const response = await fetch(`${API_BASE}/messages/${messageId}/feedback`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          rating,
          comment
        })
      })
      const data = await response.json()
      if (data.code === 200 && data.data) {
        updateMessage(messageId, {
          feedback_rating: data.data.feedback_rating,
          feedback_comment: data.data.feedback_comment,
          feedback_at: data.data.feedback_at
        })
        return true
      }
      return false
    } catch (err) {
      console.error('Failed to send feedback:', err)
      return false
    }
  }

  return {
    messages,
    isLoading,
    error,
    connectionStatus,
    currentMessages,
    addMessage,
    updateMessage,
    setLoading,
    setError,
    setConnectionStatus,
    clearError,
    clearCurrentSession,
    cancelStreaming,
    disconnect,
    getConnectionStatus,
    resetReconnectCount,
    fetchMessagesBySession,
    sendMessage,
    sendMessageNonStreaming,
    sendFeedback,
  }
})
