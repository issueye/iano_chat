/**
 * Agent 管理 Store
 * 管理可用的 AI Agent 列表和当前选中的 Agent
 */
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { API_BASE } from './config'

export const useAgentStore = defineStore('agent', () => {
  /** Agent 列表 */
  const agents = ref([])
  /** 当前选中的 Agent ID */
  const currentAgentId = ref('default')

  /**
   * 主要 Agent 列表（类型为 main 的 Agent）
   */
  const mainAgents = computed(() => {
    return agents.value.filter(a => a.type === 'main')
  })

  /**
   * 当前选中的 Agent 对象
   */
  const currentAgent = computed(() => {
    return agents.value.find(a => a.id === currentAgentId.value)
  })

  /**
   * 设置当前 Agent
   * @param agentId - Agent ID
   */
  function setCurrentAgent(agentId) {
    currentAgentId.value = agentId
  }

  /**
   * 获取所有 Agent
   */
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
      console.error('获取 Agent 列表失败:', err)
    }
  }

  return {
    agents,
    currentAgentId,
    mainAgents,
    currentAgent,
    setCurrentAgent,
    fetchAgents,
  }
})
