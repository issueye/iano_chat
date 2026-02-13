<script setup>
import { computed } from "vue";
import { Check, X } from "lucide-vue-next";
import {
  SelectContent,
  SelectItem,
  SelectItemIndicator,
  SelectItemText,
  SelectRoot,
  SelectTrigger,
  SelectValue,
} from "./index";

const props = defineProps({
  modelValue: { type: [String, Number, Array], default: "" },
  options: {
    type: Array,
    default: () => [],
  },
  placeholder: { type: String, default: "请选择" },
  disabled: { type: Boolean, default: false },
  labelField: { type: String, default: "label" },
  valueField: { type: String, default: "value" },
  multiple: { type: Boolean, default: false },
});

const emit = defineEmits(["update:modelValue", "change"]);

const normalizedOptions = computed(() => {
  return props.options.map((option) => {
    if (typeof option === "string" || typeof option === "number") {
      return { label: String(option), value: option };
    }
    return {
      label: option[props.labelField],
      value: option[props.valueField],
    };
  });
});

const selectedLabels = computed(() => {
  const value = props.modelValue;
  if (props.multiple && Array.isArray(value)) {
    if (value.length === 0) return "";
    return value
      .map((v) => {
        const opt = normalizedOptions.value.find((o) => o.value === v);
        return opt ? opt.label : v;
      })
      .join(", ");
  }
  if (value) {
    const opt = normalizedOptions.value.find((o) => o.value === value);
    return opt ? opt.label : value;
  }
  return "";
});

function handleValueChange(value) {
  if (props.multiple) {
    const current = props.modelValue || [];
    let newArray;
    if (current.includes(value)) {
      newArray = current.filter((v) => v !== value);
    } else {
      newArray = [...current, value];
    }
    emit("update:modelValue", newArray);
    emit("change", newArray);
  } else {
    emit("update:modelValue", value);
    emit("change", value);
  }
}

function handleMultiValueChange(values) {
  emit("update:modelValue", values);
  emit("change", values);
}

function clearSelection() {
  emit("update:modelValue", []);
  emit("change", []);
}
</script>

<template>
  <div class="relative">
    <SelectRoot
      :model-value="modelValue"
      @update:model-value="
        multiple ? handleMultiValueChange : handleValueChange
      "
      :multiple="multiple"
      :disabled="disabled"
    >
      <SelectTrigger :class="multiple ? 'w-full' : 'w-full'">
        <SelectValue :placeholder="placeholder">
          <span v-if="selectedLabels" class="truncate">
            {{ selectedLabels }}
          </span>
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        <div
          v-for="option in normalizedOptions"
          :key="option.value"
          :value="option.value"
          class="w-full"
          @click="handleValueChange(option.value)"
        >
          <!-- 鼠标 -->
          <div class="w-full flex items-center gap-2 cursor-pointer h-10 hover:bg-accent px-4 rounded-md">
            <!-- <Check class="h-4 w-4" /> -->
            <span class="font-medium text-foreground text-sm">{{ option.label }}</span>
          </div>
        </div>
      </SelectContent>
    </SelectRoot>

    <button
      v-if="multiple && selectedLabels"
      type="button"
      class="absolute right-8 top-1/2 -translate-y-1/2 h-6 w-6 flex items-center justify-center rounded-md hover:bg-accent z-10"
      @click.stop="clearSelection"
    >
      <X class="h-3 w-3" />
    </button>
  </div>
</template>
