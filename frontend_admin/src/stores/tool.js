import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { toolApi } from '@/api'

export const useToolStore = defineStore('tool', () => {
  const tools = ref([])
  const loading = ref(false)
  const error = ref(null)

  const totalCount = computed(() => tools.value.length)

  const enabledCount = computed(() => 
    tools.value.filter(t => t.status === 'enabled').length
  )

  const disabledCount = computed(() => 
    tools.value.filter(t => t.status === 'disabled').length
  )

  const fetchAll = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await toolApi.getAll()
      tools.value = result.data || []
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
      const result = await toolApi.getById(id)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const fetchByType = async (type) => {
    loading.value = true
    error.value = null
    try {
      const result = await toolApi.getByType(type)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const create = async (data) => {
    loading.value = true
    error.value = null
    try {
      const result = await toolApi.create(data)
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
      const result = await toolApi.update(id, data)
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
      await toolApi.delete(id)
      await fetchAll()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const registerToAgent = async (toolId, agentId) => {
    try {
      const result = await toolApi.registerToAgent(toolId, agentId)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    }
  }

  const test = async (id) => {
    try {
      const result = await toolApi.test(id)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    }
  }

  return {
    tools,
    loading,
    error,
    totalCount,
    enabledCount,
    disabledCount,
    fetchAll,
    fetchById,
    fetchByType,
    create,
    update,
    remove,
    registerToAgent,
    test,
  }
})
