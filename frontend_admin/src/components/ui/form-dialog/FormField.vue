<template>
  <div class="space-y-2">
    <Label :for="fieldId">{{ field.label }}</Label>
    
    <!-- Switch 类型 -->
    <div v-if="field.type === 'switch'" class="flex items-center gap-3">
      <Switch
        :id="fieldId"
        :model-value="modelValue === (field.trueValue || 'active')"
        @update:model-value="handleSwitchChange"
      />
      <Label :for="fieldId" class="text-sm font-normal">{{ field.switchLabel || field.label }}</Label>
    </div>
    
    <!-- Textarea 类型 -->
    <textarea
      v-else-if="field.type === 'textarea'"
      :id="fieldId"
      :value="modelValue"
      @input="handleInput"
      :rows="field.rows || 4"
      :placeholder="field.placeholder"
      :required="field.required"
      :class="textareaClass"
      v-bind="field.props"
    ></textarea>
    
    <!-- Number 类型 -->
    <Input
      v-else-if="field.type === 'number'"
      :id="fieldId"
      :model-value="modelValue"
      @update:model-value="handleNumberInput"
      type="number"
      :step="field.step"
      :min="field.min"
      :max="field.max"
      :placeholder="field.placeholder"
      :required="field.required"
      v-bind="field.props"
    />
    
    <!-- 默认 Input 类型 -->
    <Input
      v-else
      :id="fieldId"
      :model-value="modelValue"
      @update:model-value="$emit('update:modelValue', $event)"
      :type="field.type || 'text'"
      :placeholder="field.placeholder"
      :required="field.required"
      v-bind="field.props"
    />
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';

const props = defineProps({
  /** 字段配置 */
  field: { type: Object, required: true },
  /** 字段值 */
  modelValue: { type: [String, Number, Boolean], default: '' },
});

const emit = defineEmits(['update:modelValue']);

/** 字段 ID */
const fieldId = computed(() => props.field.id || props.field.key);

/** Textarea 类名 */
const textareaClass = `
  flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm 
  shadow-sm placeholder:text-muted-foreground focus-visible:outline-none 
  focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed 
  disabled:opacity-50 resize-none
`;

/**
 * 处理输入事件
 * @param {Event} event - 输入事件
 */
function handleInput(event) {
  emit('update:modelValue', event.target.value);
}

/**
 * 处理数字输入
 * @param {string|number} value - 输入值
 */
function handleNumberInput(value) {
  const num = value === '' ? '' : Number(value);
  emit('update:modelValue', num);
}

/**
 * 处理 Switch 切换
 * @param {boolean} checked - 是否选中
 */
function handleSwitchChange(checked) {
  const trueValue = props.field.trueValue || 'active';
  const falseValue = props.field.falseValue || 'inactive';
  emit('update:modelValue', checked ? trueValue : falseValue);
}
</script>
