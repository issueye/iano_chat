<template>
  <XMarkdown :markdown="content" />
</template>

<script setup>
import { XMarkdown } from 'vue-element-plus-x';

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
});
</script>

<style>
/* 自定义 markdown 样式 */
:deep(.ep-md) {
  font-size: 14px;
  line-height: 1.7;
}

/* 标题 */
:deep(.ep-md h1) { font-size: 1.75em; font-weight: 700; margin: 1em 0 0.5em; }
:deep(.ep-md h2) { font-size: 1.5em; font-weight: 600; margin: 1em 0 0.5em; }
:deep(.ep-md h3) { font-size: 1.25em; font-weight: 600; margin: 1em 0 0.5em; }
:deep(.ep-md h4) { font-size: 1.1em; font-weight: 600; margin: 0.8em 0 0.4em; }

/* 段落 */
:deep(.ep-md p) { margin: 0 0 1em; }

/* 列表 */
:deep(.ep-md ul), :deep(.ep-md ol) { margin: 0 0 1em; padding-left: 1.5em; }
:deep(.ep-md li) { margin: 0.25em 0; }

/* 内联代码 */
:deep(.ep-md-code-inline) {
  padding: 2px 6px;
  font-size: 0.9em;
  background: rgba(0, 0, 0, 0.08);
  border-radius: 4px;
}

/* 代码块 */
:deep(.ep-md-code-block) {
  margin: 12px 0;
  border-radius: 8px;
  overflow: hidden;
}

:deep(.ep-md-code-block pre) {
  margin: 0;
  padding: 16px;
}

:deep(.ep-md-code-block code) {
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
  font-size: 13px;
  line-height: 1.6;
}

/* 引用 */
:deep(.ep-md blockquote) {
  margin: 0 0 1em;
  padding: 0.5em 1em;
  border-left: 3px solid var(--primary);
  background: rgba(0, 0, 0, 0.02);
}

/* 链接 */
:deep(.ep-md a) {
  color: var(--primary);
  text-decoration: none;
}
:deep(.ep-md a:hover) {
  text-decoration: underline;
}

/* 表格 */
:deep(.ep-md table) {
  width: 100%;
  margin: 1em 0;
  border-collapse: collapse;
}

:deep(.ep-md th),
:deep(.ep-md td) {
  padding: 10px 14px;
  border: 1px solid var(--border);
}

:deep(.ep-md th) {
  background: rgba(0, 0, 0, 0.02);
  font-weight: 600;
}

/* 分割线 */
:deep(.ep-md hr) {
  height: 1px;
  margin: 1.5em 0;
  background: var(--border);
  border: none;
}

/* 图片 */
:deep(.ep-md img) {
  max-width: 100%;
  border-radius: 8px;
}
</style>
