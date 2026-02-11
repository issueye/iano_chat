<script setup>
import { computed, inject } from "vue";

const props = defineProps({
    value: { type: String, required: true },
});

const { activeTab } = inject("TabsContext");

const isActive = computed(() => activeTab.value === props.value);
</script>

<template>
    <div v-show="isActive" class="flex-1 overflow-auto">
        <slot />
    </div>
</template>
