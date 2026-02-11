import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { agentApi } from '@/api'

export const useAgentStore = defineStore('agent', () => {
  const agents = ref([])
  const instances = ref([])
  const stats = ref(null)
  const loading = ref(false)
  const error = ref(null)

  const totalCount = computed(() => agents.value.length)

  const fetchAll = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await agentApi.getAll()
      agents.value = result.data || []
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const fetchById = async (id) => {
    loading.value = true
    error.value = null
    try {
      const result = await agentApi.getById(id)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const fetchInstances = async () => {
    try {
      const result = await agentApi.getInstances()
      instances.value = result.data || []
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    }
  }

  const fetchStats = async () => {
    try {
      const result = await agentApi.getStats()
      stats.value = result.data
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    }
  }

  const create = async (data) => {
    loading.value = true
    error.value = null
    try {
      const result = await agentApi.create(data)
      await fetchAll()
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const update = async (id, data) => {
    loading.value = true
    error.value = null
    try {
      const result = await agentApi.update(id, data)
      await fetchAll()
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const remove = async (id) => {
    loading.value = true
    error.value = null
    try {
      await agentApi.delete(id)
      await fetchAll()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const reload = async (id) => {
    loading.value = true
    error.value = null
    try {
      const result = await agentApi.reload(id)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const addTool = async (agentId, toolId) => {
    try {
      const result = await agentApi.addTool(agentId, toolId)
      await fetchAll()
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    }
  }

  const removeTool = async (agentId, toolName) => {
    try {
      const result = await agentApi.removeTool(agentId, toolName)
      await fetchAll()
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    }
  }

  return {
    agents,
    instances,
    stats,
    loading,
    error,
    totalCount,
    fetchAll,
    fetchById,
    fetchInstances,
    fetchStats,
    create,
    update,
    remove,
    reload,
    addTool,
    removeTool,
  }
})
