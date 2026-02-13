<template>
  <div class="space-y-2">
    <Label :for="fieldId">{{ field.label }}</Label>
    
    <!-- Select 类型 -->
    <SimpleSelect
      v-if="field.type === 'select'"
      :id="fieldId"
      :model-value="modelValue"
      @update:model-value="$emit('update:modelValue', $event)"
      :options="field.options || []"
      :placeholder="field.placeholder || '请选择'"
      :multiple="field.multiple || false"
    />
    
    <!-- Switch 类型 -->
    <div v-else-if="field.type === 'switch'" class="flex items-center gap-3">
      <Switch
        :id="fieldId"
        :model-value="modelValue === (field.trueValue || true)"
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
import { computed } from 'vue'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { SimpleSelect } from '@/components/ui/select'

const props = defineProps({
  field: { type: Object, required: true },
  modelValue: { type: [String, Number, Boolean, Array], default: '' },
})

const emit = defineEmits(['update:modelValue'])

const fieldId = computed(() => props.field.id || props.field.key)

const textareaClass = `
  flex w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm 
  shadow-sm placeholder:text-muted-foreground focus-visible:outline-none 
  focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed 
  disabled:opacity-50 resize-none
`

function handleInput(event) {
  emit('update:modelValue', event.target.value)
}

function handleNumberInput(value) {
  const num = value === '' ? '' : Number(value)
  emit('update:modelValue', num)
}

function handleSwitchChange(checked) {
  const trueValue = props.field.trueValue !== undefined ? props.field.trueValue : true
  const falseValue = props.field.falseValue !== undefined ? props.field.falseValue : false
  emit('update:modelValue', checked ? trueValue : falseValue)
}
</script>
