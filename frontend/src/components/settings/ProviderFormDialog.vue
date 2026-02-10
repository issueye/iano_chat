<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent v-if="open" class="sm:max-w-[500px]" data-dialog="provider-form">
            <DialogTitle>{{ isEdit ? '编辑供应商' : '添加供应商' }}</DialogTitle>
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

const defaultForm = {
    name: '',
    base_url: '',
    api_key: '',
    model: '',
    temperature: 0.7,
    max_tokens: 4096,
};

const form = ref({ ...defaultForm });

const isEdit = computed(() => !!props.provider?.id);

watch(() => props.provider, (newProvider) => {
    if (newProvider) {
        form.value = {
            name: newProvider.name || '',
            base_url: newProvider.base_url || '',
            api_key: newProvider.api_key || '',
            model: newProvider.model || '',
            temperature: newProvider.temperature ?? 0.7,
            max_tokens: newProvider.max_tokens || 4096,
        };
    } else {
        form.value = { ...defaultForm };
    }
}, { immediate: true });

function handleOpenChange(value) {
    console.log('ProviderFormDialog handleOpenChange called with:', value);
    emit('update:open', value);
    console.log('ProviderFormDialog emitted update:open with:', value);
    if (!value) {
        form.value = { ...defaultForm };
    }
}

function handleCancel() {
    emit('update:open', false);
    form.value = { ...defaultForm };
}

async function handleSubmit() {
    submitting.value = true;
    try {
        const url = isEdit.value
            ? `/api/providers/${props.provider.id}`
            : '/api/providers';
        const method = isEdit.value ? 'PUT' : 'POST';

        const payload = { ...form.value };
        if (isEdit.value) {
            const updates = {};
            if (payload.name !== props.provider.name) updates.name = payload.name;
            if (payload.base_url !== props.provider.base_url) updates.base_url = payload.base_url;
            if (payload.api_key !== props.provider.api_key) updates.api_key = payload.api_key;
            if (payload.model !== props.provider.model) updates.model = payload.model;
            if (payload.temperature !== props.provider.temperature) updates.temperature = payload.temperature;
            if (payload.max_tokens !== props.provider.max_tokens) updates.max_tokens = payload.max_tokens;
            
            if (Object.keys(updates).length === 0) {
                emit('update:open', false);
                return;
            }
        }

        const response = await fetch(url, {
            method,
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(isEdit.value ? payload : payload),
        });

        if (response.ok) {
            emit('success');
            emit('update:open', false);
            form.value = { ...defaultForm };
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
