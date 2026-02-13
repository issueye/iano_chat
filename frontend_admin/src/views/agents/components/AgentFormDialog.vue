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
import { ref, computed, watch, onMounted } from 'vue'
import { FormDialog } from '@/components/ui/form-dialog'
import { agentApi, providerApi, mcpApi } from '@/api'

const props = defineProps({
  open: { type: Boolean, default: false },
  agent: { type: Object, default: null },
})

const emit = defineEmits(['update:open', 'success'])

const dialogOpen = ref(props.open)
const currentId = ref(null)
const providerOptions = ref([])
const mcpServerOptions = ref([])

const isEdit = computed(() => !!currentId.value)

const agentTypeOptions = [
  { label: '主 Agent', value: 'main' },
  { label: '子 Agent', value: 'sub' },
  { label: '自定义', value: 'custom' },
]

const fields = computed(() => [
  {
    grid: 2,
    fields: [
      { key: 'name', label: '名称', placeholder: '例如：客服助手', required: true },
      {
        key: 'type',
        label: '类型',
        type: 'select',
        options: agentTypeOptions,
        required: true,
        default: 'main',
      },
    ],
  },
  { key: 'description', label: '描述', placeholder: '描述该 Agent 的功能' },
  {
    key: 'provider_id',
    label: '供应商',
    type: 'select',
    options: providerOptions.value,
    placeholder: '选择供应商',
  },
  {
    grid: 2,
    fields: [
      { key: 'model', label: '模型', placeholder: '例如：gpt-4o-mini' },
      {
        key: 'is_sub_agent',
        label: '子 Agent',
        type: 'switch',
        switchLabel: '作为子 Agent',
        default: false,
      },
    ],
  },
  { key: 'instructions', label: '系统指令', type: 'textarea', rows: 4, placeholder: '输入系统指令...' },
  { key: 'tools', label: 'Tools', type: 'textarea', rows: 2, placeholder: '工具名称列表，JSON 数组格式' },
  {
    key: 'mcp_server_ids',
    label: 'MCP 服务器',
    type: 'select',
    multiple: true,
    options: mcpServerOptions.value,
    placeholder: '选择 MCP 服务器',
  },
])

watch(() => props.open, (val) => {
  dialogOpen.value = val
})

watch(dialogOpen, (val) => {
  emit('update:open', val)
  if (!val) {
    currentId.value = null
  }
})

watch(() => props.agent, (newAgent) => {
  if (newAgent && newAgent.id) {
    currentId.value = newAgent.id
  } else {
    currentId.value = null
  }
}, { immediate: true })

async function fetchProviders() {
  try {
    const result = await providerApi.getAll()
    providerOptions.value = (result.data || []).map(p => ({
      label: p.name,
      value: p.id,
    }))
  } catch (e) {
    console.error('Failed to fetch providers:', e)
  }
}

async function fetchMCPServers() {
  try {
    const result = await mcpApi.getAllServers()
    mcpServerOptions.value = (result.data || []).map(s => ({
      label: s.name,
      value: s.id,
    }))
  } catch (e) {
    console.error('Failed to fetch MCP servers:', e)
  }
}

async function handleSubmit(formData, isEditMode, id) {
  if (isEditMode) {
    await agentApi.update(id, formData)
  } else {
    await agentApi.create(formData)
  }
}

onMounted(() => {
  fetchProviders()
  fetchMCPServers()
})
</script>
