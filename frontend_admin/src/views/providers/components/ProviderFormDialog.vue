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
import { ref, computed, watch } from 'vue'
import { FormDialog } from '@/components/ui/form-dialog'
import { providerApi } from '@/api'

const props = defineProps({
  open: { type: Boolean, default: false },
  provider: { type: Object, default: null },
})

const emit = defineEmits(['update:open', 'success'])

const dialogOpen = ref(props.open)
const currentId = ref(null)

const isEdit = computed(() => !!currentId.value)

const fields = [
  { key: 'name', label: '名称', placeholder: '例如：OpenAI', required: true },
  { key: 'base_url', label: 'API Base URL', placeholder: '例如：https://api.openai.com/v1', required: true },
  { key: 'api_key', label: 'API Key', type: 'password', placeholder: '输入 API Key', required: true },
  { key: 'model', label: '默认模型', placeholder: '例如：gpt-4o-mini', required: true },
  {
    grid: 2,
    fields: [
      { key: 'temperature', label: 'Temperature', type: 'number', step: 0.1, min: 0, max: 2, default: 0.7 },
      { key: 'max_tokens', label: 'Max Tokens', type: 'number', min: 1, default: 4096 },
    ],
  },
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

watch(() => props.provider, (newProvider) => {
  if (newProvider && newProvider.id) {
    currentId.value = newProvider.id
  } else {
    currentId.value = null
  }
}, { immediate: true })

async function handleSubmit(formData, isEditMode, id) {
  if (isEditMode) {
    await providerApi.update(id, formData)
  } else {
    await providerApi.create(formData)
  }
}
</script>
