<template>
  <div
    class="markdown-body"
    v-html="renderedContent"
  />
</template>

<script setup>
import { computed } from 'vue';
import MarkdownIt from 'markdown-it';

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
});

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: true,
});

const renderedContent = computed(() => {
  if (!props.content) return '';
  return md.render(props.content);
});
</script>

<style>
.markdown-body {
  font-size: 14px;
  line-height: 1.6;
}

.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4),
.markdown-body :deep(h5),
.markdown-body :deep(h6) {
  margin-top: 16px;
  margin-bottom: 12px;
  font-weight: 600;
}

.markdown-body :deep(h1) { font-size: 1.5em; }
.markdown-body :deep(h2) { font-size: 1.3em; }
.markdown-body :deep(h3) { font-size: 1.1em; }

.markdown-body :deep(p) {
  margin-top: 0;
  margin-bottom: 12px;
}

.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  margin-top: 0;
  margin-bottom: 12px;
  padding-left: 24px;
}

.markdown-body :deep(li) {
  margin-bottom: 4px;
}

.markdown-body :deep(code) {
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 0.9em;
  padding: 2px 6px;
  border-radius: 4px;
  background-color: rgba(0, 0, 0, 0.08);
}

.markdown-body :deep(pre) {
  margin-top: 0;
  margin-bottom: 16px;
  padding: 12px 16px;
  overflow-x: auto;
  border-radius: 8px;
  background-color: rgba(0, 0, 0, 0.08);
}

.markdown-body :deep(pre code) {
  padding: 0;
  background-color: transparent;
  font-size: 0.85em;
}

.markdown-body :deep(blockquote) {
  margin: 0 0 16px 0;
  padding: 0 16px;
  border-left: 4px solid currentColor;
  opacity: 0.7;
}

.markdown-body :deep(a) {
  text-decoration: underline;
  opacity: 0.9;
}

.markdown-body :deep(a:hover) {
  opacity: 1;
}

.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin-bottom: 16px;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  padding: 8px 12px;
  border: 1px solid currentColor;
  opacity: 0.3;
}

.markdown-body :deep(th) {
  font-weight: 600;
  opacity: 0.5;
}

.markdown-body :deep(hr) {
  border: none;
  border-top: 1px solid currentColor;
  opacity: 0.3;
  margin: 16px 0;
}
</style>
