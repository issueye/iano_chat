<template>
  <FormDialog
    v-model:open="dialogOpen"
    :title="isEdit ? '编辑 Tool' : '添加 Tool'"
    :data="tool"
    id-key="id"
    :fields="fields"
    content-class="sm:max-w-[600px]"
    data-dialog="tool-form"
    :before-submit="beforeSubmit"
    :on-submit="handleSubmit"
    @success="$emit('success')"
  />
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { FormDialog } from '@/components/ui/form-dialog';

const props = defineProps({
  open: { type: Boolean, default: false },
  tool: { type: Object, default: null },
});

const emit = defineEmits(['update:open', 'success']);

const dialogOpen = ref(props.open);
const currentId = ref(null);

/** 是否为编辑模式 */
const isEdit = computed(() => !!currentId.value);

/** 表单字段配置 */
const fields = [
  {
    grid: 2,
    fields: [
      { key: 'name', label: '名称', placeholder: '例如：文件搜索', required: true },
      { key: 'type', label: '类型', placeholder: '例如：function', required: true, default: 'function' },
    ],
  },
  { key: 'desc', label: '描述', placeholder: '描述该工具的功能' },
  {
    grid: 2,
    fields: [
      { key: 'category', label: '分类', placeholder: '例如：文件处理' },
      { key: 'version', label: '版本', placeholder: '例如：v1.0.0', default: 'v1.0.0' },
    ],
  },
  { key: 'author', label: '作者', placeholder: '作者名称' },
  { key: 'config', label: '配置 (JSON)', type: 'textarea', rows: 4, placeholder: '{"key": "value"}', default: '{}' },
  { key: 'status', label: '状态', type: 'switch', switchLabel: '启用该 Tool', default: 'active', trueValue: 'active', falseValue: 'inactive' },
];

/** 监听 open 属性变化 */
watch(() => props.open, (val) => {
  dialogOpen.value = val;
});

/** 监听 dialogOpen 变化，同步到父组件 */
watch(dialogOpen, (val) => {
  emit('update:open', val);
  if (!val) {
    currentId.value = null;
  }
});

/** 监听 tool 变化 */
watch(() => props.tool, (newTool) => {
  if (newTool && newTool.id) {
    currentId.value = newTool.id;
  } else {
    currentId.value = null;
  }
}, { immediate: true });

/**
 * 提交前验证
 * @param {Object} formData - 表单数据
 * @returns {boolean} 验证是否通过
 */
function beforeSubmit(formData) {
  // 验证 JSON 格式
  try {
    JSON.parse(formData.config || '{}');
  } catch (e) {
    alert('配置 JSON 格式错误');
    return false;
  }
  return true;
}

/**
 * 提交表单
 * @param {Object} formData - 表单数据
 * @param {boolean} isEditMode - 是否为编辑模式
 * @param {string|number} id - 编辑时的 ID
 */
async function handleSubmit(formData, isEditMode, id) {
  const url = isEditMode ? `/api/tools/${id}` : '/api/tools';
  const method = isEditMode ? 'PUT' : 'POST';

  const response = await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      ...formData,
      config: formData.config || '{}',
    }),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || '保存失败');
  }
}
</script>
