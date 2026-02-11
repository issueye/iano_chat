<template>
  <div class="markdown-body" ref="contentRef">
    <div v-html="renderedContent"></div>
  </div>
</template>

<script setup>
import { computed, ref, watch, nextTick } from 'vue';
import { marked } from 'marked';
import Prism from 'prismjs';

const props = defineProps({
  content: {
    type: String,
    default: '',
  },
});

const contentRef = ref(null);
let codeBlockId = 0;

function escapeHtml(text) {
  if (typeof text !== 'string') return '';
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
}

const renderer = new marked.Renderer();

renderer.code = function(code, language) {
  const id = `code-block-${codeBlockId++}`;
  
  let codeText = '';
  let lang = 'text';
  
  if (typeof code === 'string') {
    codeText = code;
    lang = language || 'text';
  } else if (code && typeof code === 'object') {
    codeText = code.text || '';
    lang = code.lang || language || 'text';
  }
  
  // 高亮代码
  let highlighted = escapeHtml(codeText);
  if (lang !== 'text' && Prism.languages[lang]) {
    try {
      highlighted = Prism.highlight(codeText, Prism.languages[lang], lang);
    } catch (e) {
      highlighted = escapeHtml(codeText);
    }
  }
  
  // 生成行号和代码行
  const lines = codeText.split('\n');
  const lineCount = lines[lines.length - 1] === '' ? lines.length - 1 : lines.length;
  
  // 将高亮后的代码按行分割
  const highlightedLines = highlighted.split('\n');
  
  let codeRows = '';
  for (let i = 0; i < lineCount; i++) {
    const lineContent = highlightedLines[i] || '';
    codeRows += `<tr>
      <td class="line-num" data-line="${i + 1}"></td>
      <td class="line-code">${lineContent}</td>
    </tr>`;
  }
  
  const escapedCode = escapeHtml(codeText);
  
  return `
    <div class="code-block" data-code="${escapedCode}">
      <div class="code-header">
        <div class="code-dots">
          <span class="dot red"></span>
          <span class="dot yellow"></span>
          <span class="dot green"></span>
        </div>
        <span class="code-lang">${escapeHtml(lang)}</span>
        <button class="copy-btn" type="button">
          <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="9" y="9" width="13" height="13" rx="2"/>
            <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/>
          </svg>
          <span class="copy-text">复制</span>
        </button>
      </div>
      <div class="code-container">
        <table class="code-table"><tbody>${codeRows}</tbody></table>
      </div>
    </div>
  `;
};

renderer.codespan = function(code) {
  const text = typeof code === 'string' ? code : (code?.text || '');
  return `<code class="inline-code">${escapeHtml(text)}</code>`;
};

marked.setOptions({
  renderer: renderer,
  gfm: true,
  breaks: true,
});

const renderedContent = computed(() => {
  if (!props.content) return '';
  codeBlockId = 0;
  try {
    return marked.parse(props.content);
  } catch (e) {
    console.error('Markdown parse error:', e);
    return escapeHtml(props.content);
  }
});

function handleCopy(e) {
  const btn = e.target.closest('.copy-btn');
  if (!btn) return;
  
  const codeBlock = btn.closest('.code-block');
  const encodedCode = codeBlock?.dataset?.code;
  
  if (encodedCode) {
    const code = encodedCode
      .replace(/&quot;/g, '"')
      .replace(/&amp;/g, '&')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&#039;/g, "'");
    
    navigator.clipboard.writeText(code).then(() => {
      btn.classList.add('copied');
      btn.querySelector('.copy-text').textContent = '已复制';
      setTimeout(() => {
        btn.classList.remove('copied');
        btn.querySelector('.copy-text').textContent = '复制';
      }, 2000);
    });
  }
}

watch(() => props.content, () => {
  nextTick(() => {
    if (contentRef.value) {
      contentRef.value.removeEventListener('click', handleCopy);
      contentRef.value.addEventListener('click', handleCopy);
    }
  });
}, { immediate: true });
</script>

<style>
.markdown-body {
  font-size: 14px;
  line-height: 1.7;
  color: inherit;
}

/* 标题 */
.markdown-body :deep(h1) { font-size: 1.75em; font-weight: 700; margin: 1em 0 0.5em; }
.markdown-body :deep(h2) { font-size: 1.5em; font-weight: 600; margin: 1em 0 0.5em; }
.markdown-body :deep(h3) { font-size: 1.25em; font-weight: 600; margin: 1em 0 0.5em; }
.markdown-body :deep(h4) { font-size: 1.1em; font-weight: 600; margin: 0.8em 0 0.4em; }
.markdown-body :deep(h5), .markdown-body :deep(h6) { font-size: 1em; font-weight: 600; margin: 0.6em 0 0.3em; }

/* 段落和列表 */
.markdown-body :deep(p) { margin: 0 0 1em; }
.markdown-body :deep(ul), .markdown-body :deep(ol) { margin: 0 0 1em; padding-left: 1.5em; }
.markdown-body :deep(li) { margin: 0.25em 0; }

/* 代码块 */
.markdown-body :deep(.code-block) {
  margin: 1em 0;
  border-radius: 10px;
  overflow: hidden;
  background: #1a1b26;
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', Consolas, monospace;
}

/* 头部 */
.markdown-body :deep(.code-header) {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #16161e;
}

.markdown-body :deep(.code-dots) {
  display: flex;
  gap: 8px;
}

.markdown-body :deep(.dot) {
  width: 12px;
  height: 12px;
  border-radius: 50%;
}

.markdown-body :deep(.dot.red) { background: #f56565; }
.markdown-body :deep(.dot.yellow) { background: #ecc94b; }
.markdown-body :deep(.dot.green) { background: #48bb78; }

.markdown-body :deep(.code-lang) {
  flex: 1;
  font-size: 12px;
  font-weight: 500;
  color: #565f89;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.markdown-body :deep(.copy-btn) {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 500;
  color: #565f89;
  background: #1a1b26;
  border: 1px solid #292e42;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.markdown-body :deep(.copy-btn:hover) {
  color: #a9b1d6;
  background: #24283b;
  border-color: #3d4359;
}

.markdown-body :deep(.copy-btn.copied) {
  color: #9ece6a;
  background: rgba(158, 206, 106, 0.1);
  border-color: #9ece6a;
}

/* 代码容器 */
.markdown-body :deep(.code-container) {
  overflow-x: auto;
}

.markdown-body :deep(.code-table) {
  width: 100%;
  border-collapse: collapse;
  margin: 0;
}

.markdown-body :deep(.code-table tr) {
  height: 24px;
}

.markdown-body :deep(.line-num) {
  width: 1%;
  min-width: 50px;
  padding: 0 8px 0 16px;
  text-align: right;
  color: #3d4359;
  font-size: 13px;
  user-select: none;
  vertical-align: top;
  background: #1a1b26;
}

.markdown-body :deep(.line-num::before) {
  content: attr(data-line);
}

.markdown-body :deep(.line-code) {
  padding: 0 16px;
  font-size: 13px;
  line-height: 24px;
  color: #a9b1d6;
  white-space: pre;
  vertical-align: top;
  background: #1a1b26;
}

/* Tokyo Night 语法高亮 */
.markdown-body :deep(.token.comment) { color: #565f89; font-style: italic; }
.markdown-body :deep(.token.keyword) { color: #bb9af7; }
.markdown-body :deep(.token.string) { color: #9ece6a; }
.markdown-body :deep(.token.number) { color: #ff9e64; }
.markdown-body :deep(.token.function) { color: #7aa2f7; }
.markdown-body :deep(.token.class-name) { color: #7dcfff; }
.markdown-body :deep(.token.operator) { color: #89ddff; }
.markdown-body :deep(.token.punctuation) { color: #89ddff; }
.markdown-body :deep(.token.variable) { color: #c0caf5; }
.markdown-body :deep(.token.property) { color: #73daca; }
.markdown-body :deep(.token.tag) { color: #f7768e; }
.markdown-body :deep(.token.attr-name) { color: #bb9af7; }
.markdown-body :deep(.token.attr-value) { color: #9ece6a; }

/* 内联代码 */
.markdown-body :deep(.inline-code) {
  padding: 2px 6px;
  font-size: 0.9em;
  font-family: inherit;
  background: rgba(102, 119, 153, 0.2);
  border-radius: 4px;
  color: #7aa2f7;
}

/* 引用 */
.markdown-body :deep(blockquote) {
  margin: 0 0 1em;
  padding: 0.5em 1em;
  border-left: 3px solid #7aa2f7;
  background: rgba(122, 162, 247, 0.05);
  color: #9aa5ce;
}

/* 链接 */
.markdown-body :deep(a) {
  color: #7aa2f7;
  text-decoration: none;
}
.markdown-body :deep(a:hover) {
  text-decoration: underline;
}

/* 表格 */
.markdown-body :deep(table:not(.code-table)) {
  width: 100%;
  margin: 1em 0;
  border-collapse: collapse;
  border-radius: 8px;
  overflow: hidden;
}

.markdown-body :deep(th),
.markdown-body :deep(td) {
  padding: 10px 14px;
  border: 1px solid #292e42;
}

.markdown-body :deep(th) {
  background: #16161e;
  font-weight: 600;
}

/* 分割线 */
.markdown-body :deep(hr) {
  height: 1px;
  margin: 1.5em 0;
  background: #292e42;
  border: none;
}

/* 图片 */
.markdown-body :deep(img) {
  max-width: 100%;
  border-radius: 8px;
}

/* 浅色模式 */
@media (prefers-color-scheme: light) {
  .markdown-body :deep(.code-block) {
    background: #fafbfc;
  }
  
  .markdown-body :deep(.code-header) {
    background: #f1f3f5;
  }
  
  .markdown-body :deep(.dot.red) { background: #ff5f57; }
  .markdown-body :deep(.dot.yellow) { background: #febc2e; }
  .markdown-body :deep(.dot.green) { background: #28c840; }
  
  .markdown-body :deep(.code-lang) {
    color: #6a737d;
  }
  
  .markdown-body :deep(.copy-btn) {
    color: #6a737d;
    background: #fafbfc;
    border-color: #d1d5da;
  }
  
  .markdown-body :deep(.copy-btn:hover) {
    color: #24292e;
    background: #f1f3f5;
  }
  
  .markdown-body :deep(.copy-btn.copied) {
    color: #22863a;
    background: rgba(34, 134, 58, 0.1);
    border-color: #22863a;
  }
  
  .markdown-body :deep(.line-num) {
    color: #cfd4da;
    background: #fafbfc;
  }
  
  .markdown-body :deep(.line-code) {
    color: #24292e;
    background: #fafbfc;
  }
  
  .markdown-body :deep(.token.comment) { color: #6a737d; }
  .markdown-body :deep(.token.keyword) { color: #d73a49; }
  .markdown-body :deep(.token.string) { color: #032f62; }
  .markdown-body :deep(.token.number) { color: #005cc5; }
  .markdown-body :deep(.token.function) { color: #6f42c1; }
  .markdown-body :deep(.token.class-name) { color: #005cc5; }
  .markdown-body :deep(.token.operator) { color: #d73a49; }
  .markdown-body :deep(.token.punctuation) { color: #24292e; }
  .markdown-body :deep(.token.variable) { color: #e36209; }
  .markdown-body :deep(.token.property) { color: #005cc5; }
  .markdown-body :deep(.token.tag) { color: #22863a; }
  .markdown-body :deep(.token.attr-name) { color: #6f42c1; }
  .markdown-body :deep(.token.attr-value) { color: #032f62; }
  
  .markdown-body :deep(.inline-code) {
    background: rgba(27, 31, 35, 0.05);
    color: #d73a49;
  }
  
  .markdown-body :deep(blockquote) {
    border-left-color: #0366d6;
    background: rgba(3, 102, 214, 0.05);
    color: #586069;
  }
  
  .markdown-body :deep(a) {
    color: #0366d6;
  }
  
  .markdown-body :deep(th),
  .markdown-body :deep(td) {
    border-color: #d1d5da;
  }
  
  .markdown-body :deep(th) {
    background: #f1f3f5;
  }
  
  .markdown-body :deep(hr) {
    background: #d1d5da;
  }
}
</style>
