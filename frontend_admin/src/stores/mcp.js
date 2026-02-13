import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { mcpApi } from '@/api'

export const useMCPStore = defineStore('mcp', () => {
  const servers = ref([])
  const loading = ref(false)
  const error = ref(null)

  const totalCount = computed(() => servers.value.length)

  const connectedCount = computed(() =>
    servers.value.filter(s => s.status === 'connected').length
  )

  const disconnectedCount = computed(() =>
    servers.value.filter(s => s.status === 'disconnected').length
  )

  const errorCount = computed(() =>
    servers.value.filter(s => s.status === 'error').length
  )

  const fetchAllServers = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpApi.getAllServers()
      servers.value = result.data || []
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const fetchServerById = async (id) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpApi.getServerById(id)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const createServer = async (data) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpApi.createServer(data)
      await fetchAllServers()
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const updateServer = async (id, data) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpApi.updateServer(id, data)
      await fetchAllServers()
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const deleteServer = async (id) => {
    loading.value = true
    error.value = null
    try {
      await mcpApi.deleteServer(id)
      await fetchAllServers()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const connectServer = async (serverId) => {
    loading.value = true
    error.value = null
    try {
      await mcpApi.connectServer(serverId)
      await fetchAllServers()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const disconnectServer = async (id) => {
    loading.value = true
    error.value = null
    try {
      await mcpApi.disconnectServer(id)
      await fetchAllServers()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const getServerTools = async (id) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpApi.getServerTools(id)
      return result.data || []
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  const pingServer = async (id) => {
    loading.value = true
    error.value = null
    try {
      await mcpApi.pingServer(id)
      return true
    } catch (e) {
      error.value = e.message
      return false
    } finally {
      loading.value = false
    }
  }

  const callTool = async (serverId, toolName, args) => {
    loading.value = true
    error.value = null
    try {
      const result = await mcpApi.callTool(serverId, toolName, args)
      return result.data
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    servers,
    loading,
    error,
    totalCount,
    connectedCount,
    disconnectedCount,
    errorCount,
    fetchAllServers,
    fetchServerById,
    createServer,
    updateServer,
    deleteServer,
    connectServer,
    disconnectServer,
    getServerTools,
    pingServer,
    callTool,
  }
})
