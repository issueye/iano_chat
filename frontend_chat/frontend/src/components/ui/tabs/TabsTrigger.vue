<script setup>
import { computed, inject } from "vue";

const props = defineProps({
    value: { type: String, required: true },
});

const { activeTab, selectTab } = inject("TabsContext");

const isActive = computed(() => activeTab.value === props.value);
</script>

<template>
    <button
        type="button"
        class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50"
        :class="[
            isActive
                ? 'bg-background text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground hover:bg-muted/50',
        ]"
        @click="selectTab(value)"
    >
        <slot />
    </button>
</template>
