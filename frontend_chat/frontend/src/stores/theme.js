/**
 * 主题管理 Store
 * 管理应用的暗色/亮色模式切换
 */
import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  /** 是否为暗色模式 */
  const isDark = ref(false)

  /**
   * 初始化主题
   * 从 localStorage 读取用户偏好，或使用系统偏好
   */
  function initTheme() {
    const stored = localStorage.getItem('theme')
    if (stored) {
      isDark.value = stored === 'dark'
    } else {
      isDark.value = window.matchMedia('(prefers-color-scheme: dark)').matches
    }
    applyTheme()
  }

  /**
   * 应用主题到 DOM
   */
  function applyTheme() {
    if (isDark.value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  /**
   * 切换主题
   */
  function toggleTheme() {
    isDark.value = !isDark.value
    localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
    applyTheme()
  }

  /**
   * 设置主题
   * @param dark - 是否为暗色模式
   */
  function setTheme(dark) {
    isDark.value = dark
    localStorage.setItem('theme', dark ? 'dark' : 'light')
    applyTheme()
  }

  watch(isDark, () => {
    applyTheme()
  })

  return {
    isDark,
    initTheme,
    toggleTheme,
    setTheme,
  }
})
