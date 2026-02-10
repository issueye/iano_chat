import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useChatStore = defineStore('chat', () => {
  // State
  const messages = ref([])
  const sessions = ref([])
  const currentSessionId = ref(null)
  const isLoading = ref(false)
  const error = ref(null)

  // Getters
  const currentMessages = computed(() => {
    return messages.value.filter(m => m.session_id === currentSessionId.value)
  })

  const currentSession = computed(() => {
    return sessions.value.find(s => s.id === currentSessionId.value)
  })

  // Actions
  function setCurrentSession(sessionId) {
    currentSessionId.value = sessionId
  }

  function addMessage(message) {
    messages.value.push({
      id: Date.now().toString(),
      session_id: currentSessionId.value,
      created_at: new Date().toISOString(),
      ...message
    })
  }

  function updateMessage(id, updates) {
    const index = messages.value.findIndex(m => m.id === id)
    if (index !== -1) {
      messages.value[index] = { ...messages.value[index], ...updates }
    }
  }

  function setLoading(loading) {
    isLoading.value = loading
  }

  function setError(err) {
    error.value = err
  }

  function clearError() {
    error.value = null
  }

  // API Actions
  async function fetchSessions() {
    try {
      const response = await fetch('/api/sessions')
      const data = await response.json()
      if (data.code === 200) {
        sessions.value = data.data
      }
    } catch (err) {
      error.value = err.message
    }
  }

  async function createSession(title) {
    try {
      const response = await fetch('/api/sessions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          key_id: 'default',
          title: title || '新会话'
        })
      })
      const data = await response.json()
      if (data.code === 201) {
        sessions.value.unshift(data.data)
        currentSessionId.value = data.data.id
        return data.data
      }
    } catch (err) {
      error.value = err.message
    }
  }

  async function sendMessage(content) {
    if (!currentSessionId.value) {
      await createSession()
    }

    // Add user message
    const userMessage = {
      type: 'user',
      content: JSON.stringify({ text: content }),
      status: 'completed'
    }
    addMessage(userMessage)

    setLoading(true)
    clearError()

    try {
      // Create assistant message placeholder
      const assistantMessage = {
        type: 'assistant',
        content: JSON.stringify({ text: '' }),
        status: 'streaming'
      }
      addMessage(assistantMessage)
      const assistantMessageId = messages.value[messages.value.length - 1].id

      // Call API
      const response = await fetch('/api/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          session_id: currentSessionId.value,
          message: content
        })
      })

      if (!response.ok) {
        throw new Error('Failed to send message')
      }

      const reader = response.body.getReader()
      const decoder = new TextDecoder()
      let fullContent = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            const data = line.slice(6)
            if (data === '[DONE]') {
              updateMessage(assistantMessageId, {
                status: 'completed'
              })
              break
            }
            try {
              const parsed = JSON.parse(data)
              if (parsed.content) {
                fullContent += parsed.content
                updateMessage(assistantMessageId, {
                  content: JSON.stringify({ text: fullContent })
                })
              }
            } catch (e) {
              // Ignore parse errors
            }
          }
        }
      }

      updateMessage(assistantMessageId, {
        status: 'completed',
        content: JSON.stringify({ text: fullContent })
      })

    } catch (err) {
      error.value = err.message
      // Update last message to failed status
      const lastMessage = messages.value[messages.value.length - 1]
      if (lastMessage && lastMessage.type === 'assistant') {
        updateMessage(lastMessage.id, { status: 'failed' })
      }
    } finally {
      setLoading(false)
    }
  }

  return {
    messages,
    sessions,
    currentSessionId,
    isLoading,
    error,
    currentMessages,
    currentSession,
    setCurrentSession,
    addMessage,
    updateMessage,
    setLoading,
    setError,
    clearError,
    fetchSessions,
    createSession,
    sendMessage
  }
})
