/**
 * 配置管理模块
 * 提供应用全局配置和 API 基础地址管理
 */

/** API 基础地址 */
export const API_BASE = import.meta.env.VITE_API_BASE || 'http://127.0.0.1:8080/api'

/** WebSocket 基础地址 */
export const WS_BASE = import.meta.env.VITE_WS_BASE || 'ws://127.0.0.1:8080/ws'

/**
 * 构建完整的 API URL
 * @param path - API 路径
 * @returns 完整的 API URL
 */
export function buildApiUrl(path) {
  return `${API_BASE}${path}`
}

/**
 * 获取默认请求头
 * @returns 请求头对象
 */
export function getDefaultHeaders() {
  return {
    'Content-Type': 'application/json',
  }
}

/**
 * 处理 API 响应
 * @param response - fetch 响应对象
 * @returns 解析后的数据
 * @throws 当响应不成功时抛出错误
 */
export async function handleApiResponse(response) {
  const data = await response.json()
  if (data.code !== 200) {
    throw new Error(data.message || '请求失败')
  }
  return data
}

/**
 * 处理 API 错误
 * @param error - 错误对象
 * @returns 格式化的错误消息
 */
export function handleApiError(error) {
  if (error.response) {
    return `请求错误: ${error.response.status}`
  }
  return error.message || '未知错误'
}
