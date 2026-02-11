<template>
    <div class="flex flex-col h-full">
        <slot />
    </div>
</template>

<script setup>
import { provide, ref, computed, watch } from "vue";

const props = defineProps({
    modelValue: { type: String, default: "" },
});

const emit = defineEmits(["update:modelValue"]);

const activeTab = ref(props.modelValue || "");

watch(
    () => props.modelValue,
    (val) => {
        if (val !== undefined) activeTab.value = val;
    },
);

function selectTab(value) {
    activeTab.value = value;
    emit("update:modelValue", value);
}

provide("TabsContext", {
    activeTab: computed(() => activeTab.value),
    selectTab,
});
</script>
