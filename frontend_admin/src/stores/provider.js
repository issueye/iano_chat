import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { providerApi } from '@/api'

export const useProviderStore = defineStore('provider', () => {
  const providers = ref([])
  const loading = ref(false)
  const error = ref(null)

  const totalCount = computed(() => providers.value.length)

  const fetchAll = async () => {
    loading.value = true
    error.value = null
    try {
      const result = await providerApi.getAll()
      providers.value = result.data || []
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
      const result = await providerApi.getById(id)
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
      const result = await providerApi.create(data)
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
      const result = await providerApi.update(id, data)
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
      await providerApi.delete(id)
      await fetchAll()
    } catch (e) {
      error.value = e.message
      throw e
    } finally {
      loading.value = false
    }
  }

  return {
    providers,
    loading,
    error,
    totalCount,
    fetchAll,
    fetchById,
    create,
    update,
    remove,
  }
})
