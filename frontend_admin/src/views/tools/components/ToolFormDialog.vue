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
import { ref, computed, watch } from 'vue'
import { FormDialog } from '@/components/ui/form-dialog'
import { toolApi } from '@/api'

const props = defineProps({
  open: { type: Boolean, default: false },
  tool: { type: Object, default: null },
})

const emit = defineEmits(['update:open', 'success'])

const dialogOpen = ref(props.open)
const currentId = ref(null)

const isEdit = computed(() => !!currentId.value)

const toolTypeOptions = [
  { label: '内置工具', value: 'builtin' },
  { label: '自定义工具', value: 'custom' },
  { label: '外部工具', value: 'external' },
  { label: '插件工具', value: 'plugin' },
]

const fields = [
  {
    grid: 2,
    fields: [
      { key: 'name', label: '名称', placeholder: '例如：文件搜索', required: true },
      {
        key: 'type',
        label: '类型',
        type: 'select',
        options: toolTypeOptions,
        required: true,
        default: 'custom',
      },
    ],
  },
  { key: 'desc', label: '描述', type: 'textarea', rows: 2, placeholder: '描述该工具的功能' },
  { key: 'returns', label: '返回值', placeholder: '描述返回值格式' },
  {
    grid: 2,
    fields: [
      { key: 'version', label: '版本', placeholder: '例如：1.0.0', default: '1.0.0' },
      { key: 'author', label: '作者', placeholder: '作者名称' },
    ],
  },
  { key: 'parameters', label: '参数定义 (JSON)', type: 'textarea', rows: 4, placeholder: '[{"name": "query", "type": "string", "desc": "搜索关键词", "required": true}]' },
  { key: 'config', label: '配置 (JSON)', type: 'textarea', rows: 4, placeholder: '{"key": "value"}', default: '{}' },
  { key: 'example', label: '使用示例', type: 'textarea', rows: 3, placeholder: '使用示例...' },
]

watch(() => props.open, (val) => {
  dialogOpen.value = val
})

watch(dialogOpen, (val) => {
  emit('update:open', val)
  if (!val) {
    currentId.value = null
  }
})

watch(() => props.tool, (newTool) => {
  if (newTool && newTool.id) {
    currentId.value = newTool.id
  } else {
    currentId.value = null
  }
}, { immediate: true })

function beforeSubmit(formData) {
  try {
    if (formData.config) {
      JSON.parse(formData.config)
    }
    if (formData.parameters) {
      JSON.parse(formData.parameters)
    }
  } catch (e) {
    alert('JSON 格式错误: ' + e.message)
    return false
  }
  return true
}

async function handleSubmit(formData, isEditMode, id) {
  const data = {
    ...formData,
    config: formData.config || '{}',
    parameters: formData.parameters || '',
  }
  if (isEditMode) {
    await toolApi.update(id, data)
  } else {
    await toolApi.create(data)
  }
}
</script>
