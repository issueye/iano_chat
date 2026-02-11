<template>
  <div
    class="markdown-body"
    v-html="renderedContent"
  />
</template>

<script setup>
import { computed, onMounted, watch, nextTick } from 'vue';
import MarkdownIt from 'markdown-it';
import hljs from 'highlight.js';
import 'highlight.js/styles/github-dark.css';

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
});

// 复制代码功能
function copyCode(button) {
  const code = button.closest('.code-block-wrapper').querySelector('.code-content code');
  if (code) {
    navigator.clipboard.writeText(code.textContent).then(() => {
      const originalText = button.innerHTML;
      button.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>';
      button.classList.add('copied');
      setTimeout(() => {
        button.innerHTML = originalText;
        button.classList.remove('copied');
      }, 2000);
    });
  }
}

const md = new MarkdownIt({
  html: true,
  linkify: true,
  typographer: true,
  breaks: true,
  highlight: function (str, lang) {
    // 检测语言
    let detectedLang = lang || 'text';
    let highlighted = '';
    
    if (lang && hljs.getLanguage(lang)) {
      try {
        highlighted = hljs.highlight(str, { language: lang }).value;
      } catch (__) {
        highlighted = md.utils.escapeHtml(str);
      }
    } else {
      highlighted = md.utils.escapeHtml(str);
    }
    
    // 生成行号
    const lines = str.split('\n');
    const lineNumbers = lines.map((_, i) => `<span class="line-number">${i + 1}</span>`).join('\n');
    
    return `<div class="code-block-wrapper">
      <div class="code-block-header">
        <span class="code-lang">${detectedLang}</span>
        <button class="copy-code-btn" title="复制代码">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"></rect><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"></path></svg>
          <span class="copy-text">复制</span>
        </button>
      </div>
      <div class="code-content-wrapper">
        <div class="line-numbers">${lineNumbers}</div>
        <div class="code-content">
          <pre class="hljs"><code class="language-${detectedLang}">${highlighted}</code></pre>
        </div>
      </div>
    </div>`;
  }
});

const renderedContent = computed(() => {
  if (!props.content) return '';
  return md.render(props.content);
});

// 暴露 copyCode 方法到组件实例
defineExpose({
  copyCode
});

// 为复制按钮添加事件监听
function attachCopyListeners() {
  nextTick(() => {
    const buttons = document.querySelectorAll('.copy-code-btn');
    buttons.forEach(btn => {
      btn.removeEventListener('click', handleCopyClick);
      btn.addEventListener('click', handleCopyClick);
    });
  });
}

function handleCopyClick(e) {
  const button = e.currentTarget;
  const code = button.closest('.code-block-wrapper').querySelector('.code-content code');
  if (code) {
    navigator.clipboard.writeText(code.textContent).then(() => {
      const originalHTML = button.innerHTML;
      button.innerHTML = '<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg><span class="copy-text">已复制</span>';
      button.classList.add('copied');
      setTimeout(() => {
        button.innerHTML = originalHTML;
        button.classList.remove('copied');
      }, 2000);
    });
  }
}

onMounted(() => {
  attachCopyListeners();
});

watch(() => props.content, () => {
  attachCopyListeners();
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
  margin: 0;
  overflow-x: auto;
  background: transparent;
}

.markdown-body :deep(pre code) {
  padding: 0;
  background-color: transparent;
  font-size: 0.85em;
  line-height: 1.6;
}

/* 代码块包装器 */
.markdown-body :deep(.code-block-wrapper) {
  margin: 16px 0;
  border-radius: 8px;
  overflow: hidden;
  background: #0d1117;
  border: 1px solid #30363d;
}

/* 代码块头部 */
.markdown-body :deep(.code-block-header) {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: #161b22;
  border-bottom: 1px solid #30363d;
  min-height: 40px;
}

.markdown-body :deep(.code-lang) {
  font-size: 12px;
  color: #8b949e;
  text-transform: uppercase;
  font-weight: 600;
  font-family: ui-monospace, SFMono-Regular, monospace;
}

/* 复制按钮 */
.markdown-body :deep(.copy-code-btn) {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border: 1px solid #30363d;
  border-radius: 6px;
  background: #21262d;
  color: #c9d1d9;
  cursor: pointer;
  transition: all 0.2s ease;
  font-size: 12px;
}

.markdown-body :deep(.copy-code-btn:hover) {
  background: #30363d;
  border-color: #8b949e;
}

.markdown-body :deep(.copy-code-btn.copied) {
  background: #238636;
  border-color: #238636;
  color: #fff;
}

.markdown-body :deep(.copy-text) {
  font-size: 12px;
}

/* 代码内容包装器 */
.markdown-body :deep(.code-content-wrapper) {
  display: flex;
  overflow-x: auto;
}

/* 行号 */
.markdown-body :deep(.line-numbers) {
  flex-shrink: 0;
  padding: 16px 12px;
  background: #0d1117;
  border-right: 1px solid #21262d;
  text-align: right;
  user-select: none;
}

.markdown-body :deep(.line-number) {
  display: block;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 0.85em;
  line-height: 1.6;
  color: #484f58;
  min-width: 20px;
}

/* 代码内容 */
.markdown-body :deep(.code-content) {
  flex: 1;
  overflow-x: auto;
  min-width: 0;
}

.markdown-body :deep(.code-content .hljs) {
  margin: 0;
  padding: 16px;
  background: #0d1117;
}

.markdown-body :deep(.code-content code) {
  display: block;
  font-family: ui-monospace, SFMono-Regular, 'SF Mono', Consolas, 'Liberation Mono', Menlo, monospace;
  font-size: 0.85em;
  line-height: 1.6;
}

/* 语法高亮颜色调整 */
.markdown-body :deep(.hljs) {
  background: #0d1117;
  color: #c9d1d9;
}

.markdown-body :deep(.hljs-keyword) { color: #ff7b72; }
.markdown-body :deep(.hljs-string) { color: #a5d6ff; }
.markdown-body :deep(.hljs-number) { color: #79c0ff; }
.markdown-body :deep(.hljs-function) { color: #d2a8ff; }
.markdown-body :deep(.hljs-comment) { color: #8b949e; font-style: italic; }
.markdown-body :deep(.hljs-operator) { color: #ff7b72; }
.markdown-body :deep(.hljs-punctuation) { color: #c9d1d9; }
.markdown-body :deep(.hljs-property) { color: #79c0ff; }
.markdown-body :deep(.hljs-tag) { color: #7ee787; }
.markdown-body :deep(.hljs-attr) { color: #79c0ff; }

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

/* 浅色模式适配 */
@media (prefers-color-scheme: light) {
  .markdown-body :deep(.code-block-wrapper) {
    background: #f6f8fa;
    border-color: #d0d7de;
  }
  
  .markdown-body :deep(.code-block-header) {
    background: #f3f4f6;
    border-bottom-color: #d0d7de;
  }
  
  .markdown-body :deep(.code-lang) {
    color: #57606a;
  }
  
  .markdown-body :deep(.copy-code-btn) {
    background: #fff;
    border-color: #d0d7de;
    color: #24292f;
  }
  
  .markdown-body :deep(.copy-code-btn:hover) {
    background: #f3f4f6;
    border-color: #bbb;
  }
  
  .markdown-body :deep(.copy-code-btn.copied) {
    background: #1a7f37;
    border-color: #1a7f37;
    color: #fff;
  }
  
  .markdown-body :deep(.line-numbers) {
    background: #f6f8fa;
    border-right-color: #d0d7de;
  }
  
  .markdown-body :deep(.line-number) {
    color: #6e7781;
  }
  
  .markdown-body :deep(.code-content .hljs),
  .markdown-body :deep(.hljs) {
    background: #f6f8fa;
    color: #24292f;
  }
  
  .markdown-body :deep(.hljs-keyword) { color: #cf222e; }
  .markdown-body :deep(.hljs-string) { color: #0a3069; }
  .markdown-body :deep(.hljs-number) { color: #0550ae; }
  .markdown-body :deep(.hljs-function) { color: #8250df; }
  .markdown-body :deep(.hljs-comment) { color: #6e7781; font-style: italic; }
  .markdown-body :deep(.hljs-operator) { color: #cf222e; }
  .markdown-body :deep(.hljs-punctuation) { color: #24292f; }
  .markdown-body :deep(.hljs-property) { color: #0550ae; }
  .markdown-body :deep(.hljs-tag) { color: #116329; }
  .markdown-body :deep(.hljs-attr) { color: #0550ae; }
}
</style>
