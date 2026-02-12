<template>
  <div class="markdown-content" :data-theme="currentTheme">
    <XMarkdown
      :markdown="content"
      :themes="{ light: 'github-light', dark: 'github-dark' }"
      :codeXProps="{ enableCodeLineNumber: true }"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { XMarkdown } from 'vue-element-plus-x';
import { useThemeStore } from '@/stores/theme';

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
});

const themeStore = useThemeStore();

/**
 * 当前主题
 */
const currentTheme = computed(() => themeStore.isDark ? 'dark' : 'light');
</script>

<style scoped>
/* Markdown 内容基础样式 */
.markdown-content {
  color: inherit;
}

/* 确保代码块在主题切换时正确显示 */
.markdown-content :deep(pre) {
  background-color: var(--color-muted);
  border-radius: var(--radius-md);
  padding: 1rem;
  overflow-x: auto;
}

.markdown-content :deep(code) {
  font-family: ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace;
  font-size: 0.875em;
}

/* 行内代码样式 */
.markdown-content :deep(:not(pre) > code) {
  background-color: var(--color-muted);
  padding: 0.2em 0.4em;
  border-radius: var(--radius-sm);
}

/* 链接样式 */
.markdown-content :deep(a) {
  color: var(--color-primary);
  text-decoration: underline;
}

/* 列表样式 */
.markdown-content :deep(ul),
.markdown-content :deep(ol) {
  padding-left: 1.5rem;
  margin: 0.5rem 0;
}

.markdown-content :deep(li) {
  margin: 0.25rem 0;
}

/* 段落样式 */
.markdown-content :deep(p) {
  margin: 0.75rem 0;
  line-height: 1.6;
}

.markdown-content :deep(p:first-child) {
  margin-top: 0;
}

.markdown-content :deep(p:last-child) {
  margin-bottom: 0;
}
</style>
