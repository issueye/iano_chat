<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent
            class="sm:max-w-[1200px] overflow-hidden flex bg-card"
        >
            <DialogTitle class="text-mauve12 m-0 text-[17px] font-semibold">
                设置
            </DialogTitle>
            <div class="flex sm:w-48 min-h-[calc(70vh-4rem)] p-6" style="width: 100%;">
                <div class="w-48 border-r border-border pr-4 flex flex-col shrink-0">
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
                            <component :is="tab.icon" class="h-4 w-4 shrink-0" />
                            <span class="truncate">{{ tab.label }}</span>
                        </button>
                    </nav>
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
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
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
