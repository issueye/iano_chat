<template>
  <FormDialog
    v-model:open="dialogOpen"
    :title="isEdit ? '编辑 Agent' : '添加 Agent'"
    :data="agent"
    id-key="id"
    :fields="fields"
    content-class="sm:max-w-[600px]"
    data-dialog="agent-form"
    :on-submit="handleSubmit"
    @success="$emit('success')"
  />
</template>

<script setup>
import { ref, computed, watch } from 'vue';
import { FormDialog } from '@/components/ui/form-dialog';

const props = defineProps({
  open: { type: Boolean, default: false },
  agent: { type: Object, default: null },
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
      { key: 'name', label: '名称', placeholder: '例如：客服助手', required: true },
      { key: 'type', label: '类型', placeholder: '例如：chat', required: true, default: 'chat' },
    ],
  },
  { key: 'description', label: '描述', placeholder: '描述该 Agent 的功能' },
  { key: 'model', label: '模型', placeholder: '例如：gpt-4' },
  { key: 'system_prompt', label: '系统提示词', type: 'textarea', rows: 4, placeholder: '输入系统提示词...' },
  { key: 'tools', label: 'Tools', placeholder: '工具列表，用逗号分隔' },
  { key: 'status', label: '状态', type: 'switch', switchLabel: '启用该 Agent', default: 'active', trueValue: 'active', falseValue: 'inactive' },
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

/** 监听 agent 变化 */
watch(() => props.agent, (newAgent) => {
  if (newAgent && newAgent.id) {
    currentId.value = newAgent.id;
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
  const url = isEditMode ? `/api/agents/${id}` : '/api/agents';
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
