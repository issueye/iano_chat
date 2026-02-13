<template>
  <FormDialog
    v-model:open="dialogOpen"
    :title="dialogTitle"
    :fields="fields"
    contentClass="sm:max-w-[800px]"
    dataDialog="mcp-form"
    :on-submit="handleSubmit"
    @success="$emit('success')"
  />
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { FormDialog } from '@/components/ui/form-dialog'
import { mcpApi } from '@/api'

const props = defineProps({
  open: { type: Boolean, default: false },
  server: { type: Object, default: null },
})

const emit = defineEmits(['update:open', 'success'])

const dialogOpen = ref(props.open)
const currentId = ref(null)
const submitting = ref(false)
const form = ref({})

const isEdit = computed(() => !!currentId.value)

const dialogTitle = computed(() => isEdit.value ? '编辑 MCP 服务器' : '添加 MCP 服务器')

const transportOptions = [
  { label: 'Stdio (标准输入输出)', value: 'stdio' },
  { label: 'SSE (Server-Sent Events)', value: 'sse' },
  { label: 'HTTP', value: 'http' },
]

const fields = [
  {
    grid: 2,
    fields: [
      { key: 'name', label: '名称', placeholder: '例如：文件系统服务器', required: true },
      {
        key: 'transport',
        label: '传输类型',
        type: 'select',
        options: transportOptions,
        required: true,
        default: 'stdio',
      },
    ],
  },
  { key: 'desc', label: '描述', type: 'textarea', rows: 2, placeholder: '描述该 MCP 服务器的功能' },
  
  {
    key: 'command',
    label: '命令',
    type: 'textarea',
    rows: 1,
    placeholder: '例如：npx (用于 stdio 类型)',
    showOn: (form) => form.transport === 'stdio',
  },
  {
    key: 'args',
    label: '命令参数 (JSON)',
    type: 'textarea',
    rows: 2,
    placeholder: '例如：["-y", "@modelcontextprotocol/server-filesystem", "/tmp"]',
    showOn: (form) => form.transport === 'stdio',
  },
  {
    key: 'env',
    label: '环境变量 (JSON)',
    type: 'textarea',
    rows: 2,
    placeholder: '例如：{"KEY": "value"}',
    showOn: (form) => form.transport === 'stdio',
  },
  
  {
    key: 'url',
    label: 'URL',
    type: 'textarea',
    rows: 1,
    placeholder: '例如：http://localhost:8080/sse (用于 sse/http 类型)',
    showOn: (form) => form.transport === 'sse' || form.transport === 'http',
  },
  
  {
    grid: 2,
    fields: [
      { key: 'version', label: '版本', placeholder: '例如：1.0.0', default: '1.0.0' },
      { key: 'author', label: '作者', placeholder: '作者名称' },
    ],
  },
  { key: 'icon', label: '图标', placeholder: '例如：folder, database, server' },
]

watch(() => props.open, (val) => {
  dialogOpen.value = val
})

watch(dialogOpen, (val) => {
  emit('update:open', val)
})

watch(() => props.server, (newServer) => {
  if (newServer && newServer.id) {
    currentId.value = newServer.id
    form.value = { ...newServer }
  } else {
    currentId.value = null
    form.value = {}
    // 设置默认值
    fields.forEach(field => {
      if (field.grid) {
        field.fields.forEach(f => {
          if (f.default !== undefined) {
            form.value[f.key] = f.default
          }
        })
      } else if (field.default !== undefined) {
        form.value[field.key] = field.default
      }
    })
  }
}, { immediate: true })

function beforeSubmit(formData) {
  console.log('beforeSubmit', formData);
  
  try {
    if (formData.args) {
      JSON.parse(formData.args)
    }
    if (formData.env) {
      JSON.parse(formData.env)
    }
  } catch (e) {
    alert('JSON 格式错误: ' + e.message)
    return false
  }
  return true
}

async function handleSubmit(formData, isEdit, id) {
  const data = { ...formData }

  console.log('data', data);
  
  if (!beforeSubmit(data)) {
    return
  }
  
  submitting.value = true
  try {
    if (isEdit) {
      await mcpApi.updateServer(id, data)
    } else {
      await mcpApi.createServer(data)
    }
    dialogOpen.value = false
    emit('success')
  } catch (error) {
    alert(error.message || '保存失败')
  } finally {
    submitting.value = false
  }
}
</script>
