<template>
  <Dialog :open="open" @update:open="handleOpenChange">
    <DialogContent
      :class="['max-h-[80vh] overflow-hidden flex flex-col', contentClass]"
      :data-dialog="dataDialog"
    >
      <DialogTitle
        class="flex items-center justify-between p-6 border-b"
        @close="handleClose"
        >{{ title }}</DialogTitle
      >
      <DialogDescription v-if="description" class="sr-only">{{
        description
      }}</DialogDescription>
      <form
        @submit.prevent="handleSubmit"
        class="flex flex-col flex-1 overflow-hidden"
      >
        <div class="flex-1 overflow-y-auto scrollbar-stable space-y-4 p-6">
          <slot name="form" :form="form" :isEdit="isEdit">
            <!-- 默认表单渲染 -->
            <template v-for="field in fields" :key="field.key">
              <!-- 网格布局 -->
              <div
                v-if="field.grid"
                class="grid gap-4 px-2"
                :class="`grid-cols-${field.grid}`"
              >
                <template
                  v-for="gridField in field.fields"
                  :key="gridField.key"
                >
                  <FormField :field="gridField" v-model="form[gridField.key]" />
                </template>
              </div>
              <!-- 普通字段 -->
              <FormField
                class="px-2"
                v-else
                :field="field"
                v-model="form[field.key]"
              />
            </template>
          </slot>
        </div>
      </form>
      <!-- 底部按钮 -->
      <div class="flex justify-end gap-2 p-2 border-t">
        <Button
          type="button"
          variant="outline"
          @click="handleCancel"
          :disabled="submitting"
        >
          {{ cancelText }}
        </Button>
        <Button type="submit" :disabled="submitting" @click="handleSubmit">
          {{ submitting ? loadingText : confirmText }}
        </Button>
      </div>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { ref, watch, computed } from "vue";
import {
  Dialog,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import FormField from "./FormField.vue";

/**
 * 字段配置类型
 * @typedef {Object} FieldConfig
 * @property {string} key - 字段键名
 * @property {string} label - 字段标签
 * @property {string} type - 字段类型：text|number|textarea|password|switch
 * @property {string} [placeholder] - 占位符
 * @property {boolean} [required] - 是否必填
 * @property {string|number} [default] - 默认值
 * @property {Object} [props] - 额外属性
 */

const props = defineProps({
  /** 弹窗开关状态 */
  open: { type: Boolean, default: false },
  /** 弹窗标题 */
  title: { type: String, required: true },
  /** 编辑的数据对象 */
  data: { type: Object, default: null },
  /** 数据ID字段名 */
  idKey: { type: String, default: "id" },
  /** 表单字段配置 */
  fields: { type: Array, default: () => [] },
  /** 弹窗内容类名 */
  contentClass: { type: String, default: "sm:max-w-[800px]" },
  /** 弹窗标识 */
  dataDialog: { type: String, default: "form-dialog" },
  /** 弹窗描述 */
  description: { type: String, default: "" },
  /** 确认按钮文字 */
  confirmText: { type: String, default: "保存" },
  /** 取消按钮文字 */
  cancelText: { type: String, default: "取消" },
  /** 加载中文字 */
  loadingText: { type: String, default: "保存中..." },
  /** 提交前的验证函数，返回 false 阻止提交 */
  beforeSubmit: { type: Function, default: null },
  /** 提交处理函数，返回 Promise */
  onSubmit: { type: Function, required: true },
  /** 是否显示底部按钮 */
  showFooter: { type: Boolean, default: true },
});

const emit = defineEmits(["update:open", "success"]);

const submitting = ref(false);
const currentId = ref(null);
const form = ref({});

/** 是否为编辑模式 */
const isEdit = computed(() => !!currentId.value);

/**
 * 获取字段默认值
 * @param {FieldConfig} field - 字段配置
 * @returns {any} 默认值
 */
function getFieldDefault(field) {
  if (field.grid && field.fields) {
    const defaults = {};
    field.fields.forEach((f) => {
      defaults[f.key] = f.default ?? (f.type === "number" ? 0 : "");
    });
    return defaults;
  }
  return (
    field.default ??
    (field.type === "number" ? 0 : field.type === "switch" ? "active" : "")
  );
}

const handleClose = () => {
  emit("update:open", false);
};

/**
 * 初始化表单数据
 */
function initForm() {
  const defaultForm = {};
  props.fields.forEach((field) => {
    if (field.grid && field.fields) {
      field.fields.forEach((f) => {
        defaultForm[f.key] = getFieldDefault(f);
      });
    } else {
      defaultForm[field.key] = getFieldDefault(field);
    }
  });
  return defaultForm;
}

/**
 * 从数据对象中提取表单值
 * @param {Object} data - 数据对象
 */
function extractFormData(data) {
  const result = {};
  props.fields.forEach((field) => {
    if (field.grid && field.fields) {
      field.fields.forEach((f) => {
        result[f.key] = data[f.key] ?? getFieldDefault(f);
      });
    } else {
      result[field.key] = data[field.key] ?? getFieldDefault(field);
    }
  });
  return result;
}

/**
 * 监听弹窗打开状态，初始化表单数据
 */
watch(
  () => props.open,
  (isOpen) => {
    if (isOpen) {
      if (props.data && props.data[props.idKey]) {
        currentId.value = props.data[props.idKey];
        form.value = extractFormData(props.data);
      } else {
        currentId.value = null;
        form.value = initForm();
      }
    }
  },
  { immediate: true },
);

/**
 * 监听 data 属性变化
 */
watch(
  () => props.data,
  (newData) => {
    if (props.open && newData) {
      currentId.value = newData[props.idKey];
      form.value = extractFormData(newData);
    }
  },
  { deep: true },
);

/**
 * 处理弹窗状态变化
 * @param {boolean} value - 弹窗开关状态
 */
function handleOpenChange(value) {
  emit("update:open", value);
  if (!value) {
    setTimeout(() => {
      currentId.value = null;
      form.value = initForm();
    }, 200);
  }
}

/**
 * 取消操作
 */
function handleCancel() {
  emit("update:open", false);
  setTimeout(() => {
    currentId.value = null;
    form.value = initForm();
  }, 200);
}

/**
 * 提交表单
 */
async function handleSubmit() {
  // 执行前置验证
  if (props.beforeSubmit) {
    const valid = await props.beforeSubmit(form.value, isEdit.value);
    if (valid === false) return;
  }

  submitting.value = true;
  try {
    await props.onSubmit(form.value, isEdit.value, currentId.value);
    emit("success");
    emit("update:open", false);
    setTimeout(() => {
      currentId.value = null;
      form.value = initForm();
    }, 200);
  } catch (error) {
    console.error("Form submission failed:", error);
  } finally {
    submitting.value = false;
  }
}
</script>
