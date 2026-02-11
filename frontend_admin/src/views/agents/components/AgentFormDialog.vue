<template>
  <Dialog :open="open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[600px]" data-dialog="agent-form">
      <DialogTitle>{{ dialogTitle }}</DialogTitle>
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="name">名称</Label>
            <Input
              id="name"
              v-model="form.name"
              placeholder="例如：客服助手"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="type">类型</Label>
            <Input
              id="type"
              v-model="form.type"
              placeholder="例如：chat"
              required
            />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="description">描述</Label>
          <Input
            id="description"
            v-model="form.description"
            placeholder="描述该 Agent 的功能"
          />
        </div>
        <div class="space-y-2">
          <Label for="model">模型</Label>
          <Input
            id="model"
            v-model="form.model"
            placeholder="例如：gpt-4"
          />
        </div>
        <div class="space-y-2">
          <Label for="system_prompt">系统提示词</Label>
          <textarea
            id="system_prompt"
            v-model="form.system_prompt"
            rows="4"
            class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 resize-none"
            placeholder="输入系统提示词..."
          ></textarea>
        </div>
        <div class="space-y-2">
          <Label for="tools">Tools</Label>
          <Input
            id="tools"
            v-model="form.tools"
            placeholder="工具列表，用逗号分隔"
          />
        </div>
        <div class="flex items-center gap-3">
          <Switch
            id="status"
            v-model="form.status"
            :true-value="'active'"
            :false-value="'inactive'"
          />
          <Label for="status" class="text-sm font-normal">启用该 Agent</Label>
        </div>
        <div class="flex justify-end gap-2 pt-4">
          <Button type="button" variant="outline" @click="handleCancel">
            取消
          </Button>
          <Button type="submit" :disabled="submitting">
            {{ submitting ? '保存中...' : '保存' }}
          </Button>
        </div>
      </form>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { ref, watch, computed } from 'vue';
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';

const props = defineProps({
  open: { type: Boolean, default: false },
  agent: { type: Object, default: null },
});

const emit = defineEmits(['update:open', 'success']);

const submitting = ref(false);
const currentAgentId = ref(null);

const defaultForm = {
  name: '',
  type: 'chat',
  description: '',
  model: '',
  system_prompt: '',
  tools: '',
  status: 'active',
};

const form = ref({ ...defaultForm });

const isEdit = computed(() => !!currentAgentId.value);
const dialogTitle = computed(() => isEdit.value ? '编辑 Agent' : '添加 Agent');

/**
 * 监听弹窗打开状态，初始化表单数据
 */
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    if (props.agent && props.agent.id) {
      currentAgentId.value = props.agent.id;
      form.value = {
        name: props.agent.name || '',
        type: props.agent.type || 'chat',
        description: props.agent.description || '',
        model: props.agent.model || '',
        system_prompt: props.agent.system_prompt || '',
        tools: props.agent.tools || '',
        status: props.agent.status || 'active',
      };
    } else {
      currentAgentId.value = null;
      form.value = { ...defaultForm };
    }
  }
}, { immediate: true });

/**
 * 监听 agent 属性变化
 */
watch(() => props.agent, (newAgent) => {
  if (props.open && newAgent) {
    currentAgentId.value = newAgent.id;
    form.value = {
      name: newAgent.name || '',
      type: newAgent.type || 'chat',
      description: newAgent.description || '',
      model: newAgent.model || '',
      system_prompt: newAgent.system_prompt || '',
      tools: newAgent.tools || '',
      status: newAgent.status || 'active',
    };
  }
}, { deep: true });

/**
 * 处理弹窗状态变化
 * @param {boolean} value - 弹窗开关状态
 */
function handleOpenChange(value) {
  emit('update:open', value);
  if (!value) {
    setTimeout(() => {
      currentAgentId.value = null;
      form.value = { ...defaultForm };
    }, 200);
  }
}

/**
 * 取消操作
 */
function handleCancel() {
  emit('update:open', false);
  setTimeout(() => {
    currentAgentId.value = null;
    form.value = { ...defaultForm };
  }, 200);
}

/**
 * 提交表单
 */
async function handleSubmit() {
  submitting.value = true;
  try {
    const url = isEdit.value
      ? `/api/agents/${currentAgentId.value}`
      : '/api/agents';
    const method = isEdit.value ? 'PUT' : 'POST';

    const response = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(form.value),
    });

    if (response.ok) {
      emit('success');
      emit('update:open', false);
      setTimeout(() => {
        currentAgentId.value = null;
        form.value = { ...defaultForm };
      }, 200);
    } else {
      const error = await response.json();
      alert(error.message || '操作失败');
    }
  } catch (error) {
    console.error('Failed to save agent:', error);
    alert('保存失败，请检查网络连接');
  } finally {
    submitting.value = false;
  }
}
</script>
