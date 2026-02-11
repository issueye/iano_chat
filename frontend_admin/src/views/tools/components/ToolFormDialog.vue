<template>
  <Dialog :open="open" @update:open="handleOpenChange">
    <DialogContent class="sm:max-w-[600px]" data-dialog="tool-form">
      <DialogTitle>{{ dialogTitle }}</DialogTitle>
      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="name">名称</Label>
            <Input
              id="name"
              v-model="form.name"
              placeholder="例如：文件搜索"
              required
            />
          </div>
          <div class="space-y-2">
            <Label for="type">类型</Label>
            <Input
              id="type"
              v-model="form.type"
              placeholder="例如：function"
              required
            />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="desc">描述</Label>
          <Input
            id="desc"
            v-model="form.desc"
            placeholder="描述该工具的功能"
          />
        </div>
        <div class="grid grid-cols-2 gap-4">
          <div class="space-y-2">
            <Label for="category">分类</Label>
            <Input
              id="category"
              v-model="form.category"
              placeholder="例如：文件处理"
            />
          </div>
          <div class="space-y-2">
            <Label for="version">版本</Label>
            <Input
              id="version"
              v-model="form.version"
              placeholder="例如：v1.0.0"
            />
          </div>
        </div>
        <div class="space-y-2">
          <Label for="author">作者</Label>
          <Input
            id="author"
            v-model="form.author"
            placeholder="作者名称"
          />
        </div>
        <div class="space-y-2">
          <Label for="config">配置 (JSON)</Label>
          <textarea
            id="config"
            v-model="form.config"
            rows="4"
            class="flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 resize-none font-mono"
            placeholder='{"key": "value"}'
          ></textarea>
        </div>
        <div class="flex items-center gap-3">
          <Switch
            id="status"
            v-model="form.status"
            :true-value="'active'"
            :false-value="'inactive'"
          />
          <Label for="status" class="text-sm font-normal">启用该 Tool</Label>
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
  tool: { type: Object, default: null },
});

const emit = defineEmits(['update:open', 'success']);

const submitting = ref(false);
const currentToolId = ref(null);

const defaultForm = {
  name: '',
  type: 'function',
  desc: '',
  category: '',
  version: 'v1.0.0',
  author: '',
  config: '{}',
  status: 'active',
};

const form = ref({ ...defaultForm });

const isEdit = computed(() => !!currentToolId.value);
const dialogTitle = computed(() => isEdit.value ? '编辑 Tool' : '添加 Tool');

/**
 * 监听弹窗打开状态，初始化表单数据
 */
watch(() => props.open, (isOpen) => {
  if (isOpen) {
    if (props.tool && props.tool.id) {
      currentToolId.value = props.tool.id;
      form.value = {
        name: props.tool.name || '',
        type: props.tool.type || 'function',
        desc: props.tool.desc || '',
        category: props.tool.category || '',
        version: props.tool.version || 'v1.0.0',
        author: props.tool.author || '',
        config: props.tool.config ? JSON.stringify(props.tool.config, null, 2) : '{}',
        status: props.tool.status || 'active',
      };
    } else {
      currentToolId.value = null;
      form.value = { ...defaultForm };
    }
  }
}, { immediate: true });

/**
 * 监听 tool 属性变化
 */
watch(() => props.tool, (newTool) => {
  if (props.open && newTool) {
    currentToolId.value = newTool.id;
    form.value = {
      name: newTool.name || '',
      type: newTool.type || 'function',
      desc: newTool.desc || '',
      category: newTool.category || '',
      version: newTool.version || 'v1.0.0',
      author: newTool.author || '',
      config: newTool.config ? JSON.stringify(newTool.config, null, 2) : '{}',
      status: newTool.status || 'active',
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
      currentToolId.value = null;
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
    currentToolId.value = null;
    form.value = { ...defaultForm };
  }, 200);
}

/**
 * 提交表单
 */
async function handleSubmit() {
  submitting.value = true;
  try {
    // 验证 JSON 格式
    try {
      JSON.parse(form.value.config || '{}');
    } catch (e) {
      alert('配置 JSON 格式错误');
      submitting.value = false;
      return;
    }

    const url = isEdit.value
      ? `/api/tools/${currentToolId.value}`
      : '/api/tools';
    const method = isEdit.value ? 'PUT' : 'POST';

    const response = await fetch(url, {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        ...form.value,
        config: form.value.config || '{}',
      }),
    });

    if (response.ok) {
      emit('success');
      emit('update:open', false);
      setTimeout(() => {
        currentToolId.value = null;
        form.value = { ...defaultForm };
      }, 200);
    } else {
      const error = await response.json();
      alert(error.message || '操作失败');
    }
  } catch (error) {
    console.error('Failed to save tool:', error);
    alert('保存失败，请检查网络连接');
  } finally {
    submitting.value = false;
  }
}
</script>
