import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const API_BASE = '/api'

export const useChatStore = defineStore('chat', () => {
  const messages = ref([])
  const sessions = ref([])
  const agents = ref([])
  const currentSessionId = ref(null)
  const currentAgentId = ref('default')
  const isLoading = ref(false)
  const error = ref(null)
  const abortController = ref(null)

  const currentMessages = computed(() => {
    return messages.value.filter(m => String(m.session_id) === String(currentSessionId.value))
  })

  const currentSession = computed(() => {
    return sessions.value.find(s => String(s.id) === String(currentSessionId.value))
  })

  const mainAgents = computed(() => {
    return agents.value.filter(a => a.type === 'main')
  })

  const currentAgent = computed(() => {
    return agents.value.find(a => a.id === currentAgentId.value)
  })

  function setCurrentSession(sessionId) {
    currentSessionId.value = sessionId
  }

  function setCurrentAgent(agentId) {
    currentAgentId.value = agentId
  }

  function addMessage(message) {
    messages.value.push({
      id: message.id || Date.now().toString(),
      session_id: message.session_id || currentSessionId.value,
      created_at: message.created_at || new Date().toISOString(),
      ...message
    })
  }

  function updateMessage(id, updates) {
    const index = messages.value.findIndex(m => String(m.id) === String(id))
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

  function clearCurrentSession() {
    messages.value = messages.value.filter(m => String(m.session_id) !== String(currentSessionId.value))
  }

  function cancelStreaming() {
    if (abortController.value) {
      abortController.value.abort()
      abortController.value = null
    }
    setLoading(false)
  }

  async function fetchSessions() {
    try {
      const response = await fetch(`${API_BASE}/sessions`)
      const data = await response.json()
      if (data.code === 200) {
        sessions.value = data.data || []
      }
    } catch (err) {
      setError(err.message)
    }
  }

  async function fetchAgents() {
    try {
      const response = await fetch(`${API_BASE}/agents/type?type=main`)
      const data = await response.json()
      if (data.code === 200) {
        agents.value = data.data || []
        if (agents.value.length > 0 && currentAgentId.value === 'default') {
          currentAgentId.value = agents.value[0].id
        }
      }
    } catch (err) {
      setError(err.message)
    }
  }

  async function createSession(title) {
    try {
      const response = await fetch(`${API_BASE}/sessions`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          title: title || '新会话'
        })
      })
      const data = await response.json()
      if (data.code === 200) {
        sessions.value.unshift(data.data)
        currentSessionId.value = data.data.id
        return data.data
      }
    } catch (err) {
      setError(err.message)
    }
  }

  async function switchSession(sessionId) {
    currentSessionId.value = sessionId
    await fetchMessagesBySession(sessionId)
  }

  async function deleteSession(sessionId) {
    try {
      const response = await fetch(`${API_BASE}/sessions/${sessionId}`, {
        method: 'DELETE'
      })
      const data = await response.json()
      if (data.code === 200) {
        sessions.value = sessions.value.filter(s => String(s.id) !== String(sessionId))
        if (String(currentSessionId.value) === String(sessionId)) {
          currentSessionId.value = null
          messages.value = []
        }
      }
    } catch (err) {
      setError(err.message)
    }
  }

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

  async function sendMessage(content) {
    if (!currentSessionId.value) {
      await createSession()
    }

    const userMessage = {
      id: Date.now().toString() + '_user',
      session_id: String(currentSessionId.value),
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
      session_id: String(currentSessionId.value),
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
          session_id: String(currentSessionId.value),
          agent_id: currentAgentId.value,
          message: content
        })
      })

      if (!streamResponse.ok) {
        throw new Error(`HTTP error! status: ${streamResponse.status}`)
      }

      const reader = streamResponse.body.getReader()
      const decoder = new TextDecoder()
      let accumulatedContent = ''
      let accumulatedToolCalls = []

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        const chunk = decoder.decode(value, { stream: true })
        const lines = chunk.split('\n')

        for (const line of lines) {
          if (line.startsWith('data: ')) {
            try {
              const eventData = JSON.parse(line.slice(6))
              if (eventData.content) {
                accumulatedContent += eventData.content
              }
              if (eventData.id && eventData.name) {
                accumulatedToolCalls.push({
                  id: eventData.id,
                  type: 'function',
                  function: {
                    name: eventData.name,
                    arguments: eventData.arguments
                  }
                })
              }
              updateMessage(assistantMessageId, {
                content: JSON.stringify({
                  text: accumulatedContent,
                  tool_calls: accumulatedToolCalls
                })
              })
              if (eventData.error) {
                setError(eventData.error)
                updateMessage(assistantMessageId, { status: 'failed' })
              }
            } catch (e) {
              // Ignore parse errors for incomplete JSON
            }
          } else if (line.startsWith('event: ')) {
            const eventType = line.slice(7)
            if (eventType === 'done') {
              updateMessage(assistantMessageId, { status: 'completed' })
            }
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

  async function sendMessageNonStreaming(content) {
    if (!currentSessionId.value) {
      await createSession()
    }

    const userMessage = {
      id: Date.now().toString() + '_user',
      session_id: String(currentSessionId.value),
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
          session_id: String(currentSessionId.value),
          agent_id: currentAgentId.value,
          message: content
        })
      })
      const data = await response.json()
      if (data.code === 200 && data.data) {
        const assistantMessage = {
          id: Date.now().toString() + '_assistant',
          session_id: String(currentSessionId.value),
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
    sessions,
    agents,
    currentSessionId,
    currentAgentId,
    isLoading,
    error,
    currentMessages,
    currentSession,
    mainAgents,
    currentAgent,
    setCurrentSession,
    setCurrentAgent,
    addMessage,
    updateMessage,
    setLoading,
    setError,
    clearError,
    clearCurrentSession,
    cancelStreaming,
    fetchSessions,
    fetchAgents,
    createSession,
    switchSession,
    deleteSession,
    fetchMessagesBySession,
    sendMessage,
    sendMessageNonStreaming
  }
})
