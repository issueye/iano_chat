const API_BASE = '/api'

class ApiError extends Error {
  constructor(message, status, data) {
    super(message)
    this.status = status
    this.data = data
  }
}

async function request(url, options = {}) {
  const config = {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  }

  if (options.body && typeof options.body === 'object') {
    config.body = JSON.stringify(options.body)
  }

  const response = await fetch(`${API_BASE}${url}`, config)
  const result = await response.json()

  if (!response.ok) {
    throw new ApiError(result.message || '请求失败', response.status, result)
  }

  return result
}

export const api = {
  get: (url) => request(url, { method: 'GET' }),
  post: (url, body) => request(url, { method: 'POST', body }),
  put: (url, body) => request(url, { method: 'PUT', body }),
  delete: (url) => request(url, { method: 'DELETE' }),
}

export const providerApi = {
  getAll: () => api.get('/providers'),
  getById: (id) => api.get(`/providers/${id}`),
  create: (data) => api.post('/providers', data),
  update: (id, data) => api.put(`/providers/${id}`, data),
  delete: (id) => api.delete(`/providers/${id}`),
}

export const agentApi = {
  getAll: () => api.get('/agents'),
  getById: (id) => api.get(`/agents/${id}`),
  getByType: (type) => api.get(`/agents/type?type=${type}`),
  getStats: () => api.get('/agents/stats'),
  getInstances: () => api.get('/agents/instances'),
  create: (data) => api.post('/agents', data),
  update: (id, data) => api.put(`/agents/${id}`, data),
  delete: (id) => api.delete(`/agents/${id}`),
  reload: (id) => api.post(`/agents/${id}/reload`),
  addTool: (id, toolId) => api.post(`/agents/${id}/tools`, { tool_id: toolId }),
  removeTool: (id, toolName) => api.delete(`/agents/${id}/tools/${toolName}`),
}

export const toolApi = {
  getAll: () => api.get('/tools'),
  getById: (id) => api.get(`/tools/${id}`),
  getByType: (type) => api.get(`/tools/type?type=${type}`),
  getByStatus: (status) => api.get(`/tools/status?status=${status}`),
  create: (data) => api.post('/tools', data),
  update: (id, data) => api.put(`/tools/${id}`, data),
  updateConfig: (id, config) => api.put(`/tools/${id}/config`, { config }),
  delete: (id) => api.delete(`/tools/${id}`),
  registerToAgent: (toolId, agentId) => api.post(`/tools/${toolId}/register?agent_id=${agentId}`),
  test: (id) => api.get(`/tools/${id}/test`),
}

export const sessionApi = {
  getAll: () => api.get('/sessions'),
  getById: (id) => api.get(`/sessions/${id}`),
  getByStatus: (status) => api.get(`/sessions/status?status=${status}`),
  create: (data) => api.post('/sessions', data),
  update: (id, data) => api.put(`/sessions/${id}`, data),
  delete: (id) => api.delete(`/sessions/${id}`),
  getConfig: (id) => api.get(`/sessions/${id}/config`),
  updateConfig: (id, config) => api.put(`/sessions/${id}/config`, config),
}

export const messageApi = {
  getAll: () => api.get('/messages'),
  getById: (id) => api.get(`/messages/${id}`),
  getBySessionId: (sessionId) => api.get(`/messages/session?session_id=${sessionId}`),
  getByType: (type) => api.get(`/messages/type?type=${type}`),
  create: (data) => api.post('/messages', data),
  update: (id, data) => api.put(`/messages/${id}`, data),
  delete: (id) => api.delete(`/messages/${id}`),
  deleteBySessionId: (sessionId) => api.delete(`/messages?session_id=${sessionId}`),
  addFeedback: (id, feedback) => api.post(`/messages/${id}/feedback`, feedback),
}

export const chatApi = {
  chat: (data) => api.post('/chat', data),
  streamChat: (data) => api.post('/chat/stream', data),
  clearSession: (sessionId) => api.delete(`/chat/session/${sessionId}`),
  getConversation: (sessionId) => api.get(`/chat/conversation?session_id=${sessionId}`),
  getPoolStats: () => api.get('/chat/pool-stats'),
}

export { ApiError }
