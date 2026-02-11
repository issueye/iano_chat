<template>
  <FormDialog
    v-model:open="dialogOpen"
    :title="isEdit ? '编辑供应商' : '添加供应商'"
    :data="provider"
    id-key="id"
    :fields="fields"
    content-class="sm:max-w-[500px]"
    data-dialog="provider-form"
    :on-submit="handleSubmit"
    @success="$emit('success')"
  />
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { FormDialog } from '@/components/ui/form-dialog';

const props = defineProps({
  open: { type: Boolean, default: false },
  provider: { type: Object, default: null },
});

const emit = defineEmits(['update:open', 'success']);

const dialogOpen = ref(props.open);
const currentId = ref(null);

/** 是否为编辑模式 */
const isEdit = computed(() => !!currentId.value);

/** 表单字段配置 */
const fields = [
  { key: 'name', label: '名称', placeholder: '例如：OpenAI', required: true },
  { key: 'base_url', label: 'API Base URL', placeholder: '例如：https://api.openai.com/v1', required: true },
  { key: 'api_key', label: 'API Key', type: 'password', placeholder: '输入 API Key', required: true },
  { key: 'model', label: '默认模型', placeholder: '例如：gpt-4', required: true },
  {
    grid: 2,
    fields: [
      { key: 'temperature', label: 'Temperature', type: 'number', step: 0.1, min: 0, max: 2, default: 0.7 },
      { key: 'max_tokens', label: 'Max Tokens', type: 'number', min: 1, default: 4096 },
    ],
  },
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

/** 监听 provider 变化 */
watch(() => props.provider, (newProvider) => {
  if (newProvider && newProvider.id) {
    currentId.value = newProvider.id;
  } else {
    currentId.value = null;
  }
}, { immediate: true });

/**
 * 提交表单
 * @param {Object} formData - 表单数据
 * @param {boolean} isEditMode - 是否为编辑模式
 * @param {string|number} id - 编辑时的 ID
 */
async function handleSubmit(formData, isEditMode, id) {
  const url = isEditMode ? `/api/providers/${id}` : '/api/providers';
  const method = isEditMode ? 'PUT' : 'POST';

  const response = await fetch(url, {
    method,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(formData),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || '保存失败');
  }
}
</script>
