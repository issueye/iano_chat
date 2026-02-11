import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

const API_BASE = '/api'

function parseResponse(response) {
  return response.json().then(data => {
    if (data.code !== 200) {
      throw new Error(data.message || '请求失败')
    }
    return data.data
  })
}

export const useChatStore = defineStore('chat', () => {
  // State
  const messages = ref([])
  const sessions = ref([])
  const currentSessionId = ref(null)
  const isLoading = ref(false)
  const error = ref(null)

  // Getters
  const currentMessages = computed(() => {
    return messages.value.filter(m => String(m.session_id) === String(currentSessionId.value))
  })

  const currentSession = computed(() => {
    return sessions.value.find(s => String(s.id) === String(currentSessionId.value))
  })

  // Actions
  function setCurrentSession(sessionId) {
    currentSessionId.value = sessionId
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

  // Session API
  async function fetchSessions() {
    try {
      const response = await fetch(`${API_BASE}/sessions`)
      const data = await response.json()
      if (data.code === 200) {
        sessions.value = data.data || []
      }
    } catch (err) {
      error.value = err.message
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
      error.value = err.message
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
      error.value = err.message
    }
  }

  // Message API
  async function fetchMessagesBySession(sessionId) {
    try {
      const response = await fetch(`${API_BASE}/messages/session?session_id=${sessionId}`)
      const data = await response.json()
      if (data.code === 200) {
        messages.value = data.data || []
      }
    } catch (err) {
      error.value = err.message
    }
  }

  async function sendMessage(content) {
    if (!currentSessionId.value) {
      await createSession()
    }

    const userMessage = {
      session_id: String(currentSessionId.value),
      type: 'user',
      content: JSON.stringify({ text: content }),
      status: 'completed'
    }

    try {
      const response = await fetch(`${API_BASE}/messages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(userMessage)
      })
      const data = await response.json()
      if (data.code === 200) {
        messages.value.push(data.data)
      }
    } catch (err) {
      error.value = err.message
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
    clearCurrentSession,
    fetchSessions,
    createSession,
    switchSession,
    deleteSession,
    fetchMessagesBySession,
    sendMessage
  }
})
