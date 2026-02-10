<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent class="sm:max-w-[500px]" data-dialog="provider-form">
            <DialogTitle>{{ dialogTitle }}</DialogTitle>
            <form @submit.prevent="handleSubmit" class="space-y-4 p-6">
                <div class="space-y-2">
                    <Label for="name">名称</Label>
                    <Input
                        id="name"
                        v-model="form.name"
                        placeholder="例如：OpenAI"
                        required
                    />
                </div>
                <div class="space-y-2">
                    <Label for="base_url">API Base URL</Label>
                    <Input
                        id="base_url"
                        v-model="form.base_url"
                        placeholder="例如：https://api.openai.com/v1"
                        required
                    />
                </div>
                <div class="space-y-2">
                    <Label for="api_key">API Key</Label>
                    <Input
                        id="api_key"
                        v-model="form.api_key"
                        type="password"
                        placeholder="输入 API Key"
                        required
                    />
                </div>
                <div class="space-y-2">
                    <Label for="model">默认模型</Label>
                    <Input
                        id="model"
                        v-model="form.model"
                        placeholder="例如：gpt-4"
                        required
                    />
                </div>
                <div class="grid grid-cols-2 gap-4">
                    <div class="space-y-2">
                        <Label for="temperature">Temperature</Label>
                        <Input
                            id="temperature"
                            v-model.number="form.temperature"
                            type="number"
                            step="0.1"
                            min="0"
                            max="2"
                            placeholder="0.7"
                        />
                    </div>
                    <div class="space-y-2">
                        <Label for="max_tokens">Max Tokens</Label>
                        <Input
                            id="max_tokens"
                            v-model.number="form.max_tokens"
                            type="number"
                            min="1"
                            placeholder="4096"
                        />
                    </div>
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

const props = defineProps({
    open: { type: Boolean, default: false },
    provider: { type: Object, default: null },
});

const emit = defineEmits(['update:open', 'submit', 'success']);

const submitting = ref(false);
const currentProviderId = ref(null);

const defaultForm = {
    name: '',
    base_url: '',
    api_key: '',
    model: '',
    temperature: 0.7,
    max_tokens: 4096,
};

const form = ref({ ...defaultForm });

const isEdit = computed(() => !!currentProviderId.value);
const dialogTitle = computed(() => isEdit.value ? '编辑供应商' : '添加供应商');

// 监听 open 变化，当弹窗打开时初始化表单
watch(() => props.open, (isOpen) => {
    if (isOpen) {
        // 弹窗打开时，根据 provider 初始化表单
        console.log('props.provider', props.provider);
        if (props.provider && props.provider.id) {
            currentProviderId.value = props.provider.id;
            form.value = {
                name: props.provider.name || '',
                base_url: props.provider.base_url || '',
                api_key: props.provider.api_key || '',
                model: props.provider.model || '',
                temperature: props.provider.temperature ?? 0.7,
                max_tokens: props.provider.max_tokens || 4096,
            };
        } else {
            currentProviderId.value = null;
            form.value = { ...defaultForm };
        }
    }
}, { immediate: true });

// 监听 provider 变化（当弹窗已经打开时）
watch(() => props.provider, (newProvider) => {
    if (props.open && newProvider) {
        currentProviderId.value = newProvider.id;
        form.value = {
            name: newProvider.name || '',
            base_url: newProvider.base_url || '',
            api_key: newProvider.api_key || '',
            model: newProvider.model || '',
            temperature: newProvider.temperature ?? 0.7,
            max_tokens: newProvider.max_tokens || 4096,
        };
    }
}, { deep: true });

function handleOpenChange(value) {
    emit('update:open', value);
    if (!value) {
        // 弹窗关闭时重置状态
        setTimeout(() => {
            currentProviderId.value = null;
            form.value = { ...defaultForm };
        }, 200);
    }
}

function handleCancel() {
    emit('update:open', false);
    setTimeout(() => {
        currentProviderId.value = null;
        form.value = { ...defaultForm };
    }, 200);
}

async function handleSubmit() {
    submitting.value = true;
    try {
        const url = isEdit.value
            ? `/api/providers/${currentProviderId.value}`
            : '/api/providers';
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
                currentProviderId.value = null;
                form.value = { ...defaultForm };
            }, 200);
        } else {
            const error = await response.json();
            alert(error.message || '操作失败');
        }
    } catch (error) {
        console.error('Failed to save provider:', error);
        alert('保存失败，请检查网络连接');
    } finally {
        submitting.value = false;
    }
}
</script>
