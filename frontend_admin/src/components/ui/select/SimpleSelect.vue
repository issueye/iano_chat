<script setup>
import { computed } from 'vue'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './index'

/**
 * 简单选择器组件
 * 外部只需提供数据，内部自动渲染
 */
const props = defineProps({
  /** 当前选中的值 */
  modelValue: { type: [String, Number], default: '' },
  /** 选项数据数组 */
  options: {
    type: Array,
    default: () => [],
    // 支持格式: [{ label: '显示文本', value: '值' }, ...] 或 ['值1', '值2', ...]
  },
  /** 占位提示文本 */
  placeholder: { type: String, default: '请选择' },
  /** 是否禁用 */
  disabled: { type: Boolean, default: false },
  /** 选项标签字段名（当 options 为对象数组时） */
  labelField: { type: String, default: 'label' },
  /** 选项值字段名（当 options 为对象数组时） */
  valueField: { type: String, default: 'value' },
})

const emit = defineEmits(['update:modelValue', 'change'])

/**
 * 处理选项数据，统一转换为 { label, value } 格式
 */
const normalizedOptions = computed(() => {
  return props.options.map(option => {
    // 如果是字符串或数字，直接作为 value 和 label
    if (typeof option === 'string' || typeof option === 'number') {
      return { label: String(option), value: option }
    }
    // 如果是对象，根据配置的字段名提取
    return {
      label: option[props.labelField],
      value: option[props.valueField],
    }
  })
})

/**
 * 处理值变化
 * @param {string|number} value - 选中的值
 */
function handleChange(value) {
  emit('update:modelValue', value)
  emit('change', value)
}
</script>

<template>
  <Select :model-value="modelValue" @update:model-value="handleChange" :disabled="disabled">
    <SelectTrigger>
      <SelectValue :placeholder="placeholder" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem
        v-for="option in normalizedOptions"
        :key="option.value"
        :value="option.value"
      >
        {{ option.label }}
      </SelectItem>
    </SelectContent>
  </Select>
</template>
