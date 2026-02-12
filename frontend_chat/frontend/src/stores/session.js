/**
 * 会话管理 Store
 * 管理聊天会话的创建、切换、删除等操作
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { API_BASE, handleApiResponse } from './config'

export const useSessionStore = defineStore('session', () => {
  /** 会话列表 */
  const sessions = ref([])
  /** 当前会话 ID */
  const currentSessionId = ref(null)
  /** 搜索关键词 */
  const searchKeyword = ref('')

  /**
   * 当前会话对象
   */
  const currentSession = computed(() => {
    return sessions.value.find(s => String(s.id) === String(currentSessionId.value))
  })

  /**
   * 过滤后的会话列表（根据搜索关键词）
   */
  const filteredSessions = computed(() => {
    if (!searchKeyword.value.trim()) {
      return sessions.value
    }
    const keyword = searchKeyword.value.toLowerCase()
    return sessions.value.filter(session => 
      session.title?.toLowerCase().includes(keyword)
    )
  })

  /**
   * 设置当前会话
   * @param sessionId - 会话 ID
   */
  function setCurrentSession(sessionId) {
    currentSessionId.value = sessionId
  }

  /**
   * 设置搜索关键词
   * @param keyword - 搜索关键词
   */
  function setSearchKeyword(keyword) {
    searchKeyword.value = keyword
  }

  /**
   * 获取所有会话
   */
  async function fetchSessions() {
    try {
      const response = await fetch(`${API_BASE}/sessions`)
      const data = await response.json()
      if (data.code === 200) {
        sessions.value = data.data || []
      }
    } catch (err) {
      console.error('获取会话列表失败:', err)
    }
  }

  /**
   * 创建新会话
   * @param title - 会话标题
   * @returns 创建的会话对象
   */
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
      console.error('创建会话失败:', err)
    }
  }

  /**
   * 删除会话
   * @param sessionId - 会话 ID
   */
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
        }
      }
    } catch (err) {
      console.error('删除会话失败:', err)
    }
  }

  /**
   * 更新会话标题
   * @param sessionId - 会话 ID
   * @param title - 新标题
   */
  async function updateSessionTitle(sessionId, title) {
    try {
      const response = await fetch(`${API_BASE}/sessions/${sessionId}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title })
      })
      const data = await response.json()
      if (data.code === 200) {
        const index = sessions.value.findIndex(s => s.id === sessionId)
        if (index !== -1) {
          sessions.value[index].title = title
        }
      }
    } catch (err) {
      console.error('更新会话标题失败:', err)
    }
  }

  return {
    sessions,
    currentSessionId,
    searchKeyword,
    currentSession,
    filteredSessions,
    setCurrentSession,
    setSearchKeyword,
    fetchSessions,
    createSession,
    deleteSession,
    updateSessionTitle,
  }
})
