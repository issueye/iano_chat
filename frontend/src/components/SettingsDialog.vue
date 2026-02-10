<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent
            class="sm:max-w-[1200px] max-w-[95vw] max-h-[90vh] overflow-hidden flex bg-card"
        >
            <DialogHeader class="sr-only">
                <DialogTitle>设置</DialogTitle>
                <DialogDescription>管理您的供应商、Agents 和 Tools 配置</DialogDescription>
            </DialogHeader>

            <div class="flex sm:w-48 min-h-[calc(90vh-4rem)] min-w-[70%]">
                <div class="w-48 border-r border-border pr-4 flex flex-col flex-shrink-0">
                    <div class="mb-4 px-1">
                        <h2 class="font-semibold text-sm text-foreground">设置</h2>
                        <p class="text-xs text-muted-foreground">管理配置</p>
                    </div>

                    <nav class="space-y-1 flex-1" role="tablist" aria-label="设置分类">
                        <button
                            v-for="tab in tabs"
                            :key="tab.value"
                            type="button"
                            class="w-full flex items-center gap-2 px-3 py-2 rounded-lg text-sm transition-all duration-200"
                            :class="[
                                activeTab === tab.value
                                    ? 'bg-primary text-primary-foreground shadow-sm'
                                    : 'text-muted-foreground hover:text-foreground hover:bg-muted',
                            ]"
                            role="tab"
                            :aria-selected="activeTab === tab.value"
                            @click="activeTab = tab.value"
                        >
                            <component :is="tab.icon" class="h-4 w-4 flex-shrink-0" />
                            <span class="truncate">{{ tab.label }}</span>
                        </button>
                    </nav>

                    <div class="pt-4 border-t border-border mt-auto">
                        <p class="text-xs text-muted-foreground mb-2">快捷键</p>
                        <div class="space-y-1 text-xs text-muted-foreground">
                            <div class="flex justify-between">
                                <span>上一个</span>
                                <kbd class="px-1.5 py-0.5 bg-muted rounded text-[10px]">⌘[</kbd>
                            </div>
                            <div class="flex justify-between">
                                <span>下一个</span>
                                <kbd class="px-1.5 py-0.5 bg-muted rounded text-[10px]">⌘]</kbd>
                            </div>
                            <div class="flex justify-between">
                                <span>关闭</span>
                                <kbd class="px-1.5 py-0.5 bg-muted rounded text-[10px]">Esc</kbd>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="flex-1 pl-4 overflow-hidden flex flex-col">
                    <ProvidersPanel
                        v-if="activeTab === 'providers'"
                        :items="providers"
                        :loading="loading"
                        @add="handleAdd"
                        @configure="handleConfigure"
                        @edit="handleEdit"
                        @delete="handleDelete"
                    />
                    <AgentsPanel
                        v-else-if="activeTab === 'agents'"
                        :items="agents"
                        :loading="loading"
                        @add="handleAdd"
                        @configure="handleConfigure"
                        @edit="handleEdit"
                        @delete="handleDelete"
                    />
                    <ToolsPanel
                        v-else-if="activeTab === 'tools'"
                        :items="tools"
                        :loading="loading"
                        @add="handleAdd"
                        @configure="handleConfigure"
                        @edit="handleEdit"
                        @delete="handleDelete"
                    />
                </div>
            </div>
        </DialogContent>
    </Dialog>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from "vue";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import ProvidersPanel from "./settings/ProvidersPanel.vue";
import AgentsPanel from "./settings/AgentsPanel.vue";
import ToolsPanel from "./settings/ToolsPanel.vue";
import { Building2, Bot, Wrench } from "lucide-vue-next";

const props = defineProps({
    open: { type: Boolean, default: false },
});

const emit = defineEmits(["update:open", "add", "configure", "edit", "delete"]);

const activeTab = ref("providers");
const loading = ref(false);
const providers = ref([]);
const agents = ref([]);
const tools = ref([]);

const tabs = [
    { value: "providers", label: "供应商", icon: Building2 },
    { value: "agents", label: "Agents", icon: Bot },
    { value: "tools", label: "Tools", icon: Wrench },
];

function handleOpenChange(value) {
    emit("update:open", value);
    if (value) {
        fetchData();
    }
}

async function fetchData() {
    loading.value = true;
    try {
        const endpoints = {
            providers: "/api/providers",
            agents: "/api/agents",
            tools: "/api/tools",
        };
        const response = await fetch(endpoints[activeTab.value]);
        if (response.ok) {
            const result = await response.json();
            if (result.data) {
                if (activeTab.value === "providers") {
                    providers.value = result.data;
                } else if (activeTab.value === "agents") {
                    agents.value = result.data;
                } else if (activeTab.value === "tools") {
                    tools.value = result.data;
                }
            }
        }
    } catch (error) {
        console.warn("Failed to fetch data:", error);
    } finally {
        loading.value = false;
    }
}

function handleAdd() {
    emit("add", activeTab.value);
}

function handleConfigure(item) {
    emit("configure", { type: activeTab.value, item });
}

function handleEdit(item) {
    emit("edit", { type: activeTab.value, item });
}

function handleDelete(item) {
    emit("delete", { type: activeTab.value, item });
}

function handleKeyboard(e) {
    if (!props.open) return;
    if ((e.metaKey || e.ctrlKey) && e.key === "[") {
        e.preventDefault();
        const currentIndex = tabs.findIndex((t) => t.value === activeTab.value);
        if (currentIndex > 0) {
            activeTab.value = tabs[currentIndex - 1].value;
            fetchData();
        }
    }
    if ((e.metaKey || e.ctrlKey) && e.key === "]") {
        e.preventDefault();
        const currentIndex = tabs.findIndex((t) => t.value === activeTab.value);
        if (currentIndex < tabs.length - 1) {
            activeTab.value = tabs[currentIndex + 1].value;
            fetchData();
        }
    }
}

onMounted(() => {
    window.addEventListener("keydown", handleKeyboard);
});

onUnmounted(() => {
    window.removeEventListener("keydown", handleKeyboard);
});
</script>
