<template>
    <Dialog :open="open" @update:open="handleOpenChange" @open="handleOpen">
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

    <ProviderFormDialog
        v-model:open="providerFormOpen"
        :provider="editingProvider"
        @success="fetchData"
    />

    <AlertDialog
        v-model:open="deleteDialogOpen"
        :title="deleteDialogTitle"
        :description="deleteDialogDescription"
        confirm-text="删除"
        cancel-text="取消"
        variant="destructive"
        @confirm="executeDelete"
    />
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from "vue";
import { Dialog, DialogContent, DialogTitle } from "@/components/ui/dialog";
import { AlertDialog } from "@/components/ui/alert-dialog";
import ProvidersPanel from "./settings/ProvidersPanel.vue";
import AgentsPanel from "./settings/AgentsPanel.vue";
import ToolsPanel from "./settings/ToolsPanel.vue";
import ProviderFormDialog from "./settings/ProviderFormDialog.vue";
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

const providerFormOpen = ref(false);
const editingProvider = ref(null);

// 删除确认对话框状态
const deleteDialogOpen = ref(false);
const deleteItem = ref({ name: '' });
const deleteDialogTitle = computed(() => deleteItem.value?.name ? `删除 ${deleteItem.value.name}` : '删除');
const deleteDialogDescription = computed(() => deleteItem.value?.name ? `确定要删除 "${deleteItem.value.name}" 吗？此操作无法撤销。` : '确定要删除此项吗？此操作无法撤销。');

const tabs = [
    { value: "providers", label: "供应商", icon: Building2 },
    { value: "agents", label: "Agents", icon: Bot },
    { value: "tools", label: "Tools", icon: Wrench },
];

function handleDelete(item) {
    deleteItem.value = item;
    deleteDialogOpen.value = true;
}

async function executeDelete() {
    if (!deleteItem.value?.id) return;
    
    try {
        const endpoints = {
            providers: `/api/providers/${deleteItem.value.id}`,
            agents: `/api/agents/${deleteItem.value.id}`,
            tools: `/api/tools/${deleteItem.value.id}`,
        };
        
        const response = await fetch(endpoints[activeTab.value], {
            method: 'DELETE',
        });
        
        if (response.ok) {
            fetchData();
        } else {
            const error = await response.json();
            alert(error.message || '删除失败');
        }
    } catch (error) {
        console.error('Failed to delete item:', error);
        alert('删除失败，请检查网络连接');
    } finally {
        deleteItem.value = { name: '' };
    }
}

function handleOpenChange(value) {
    emit("update:open", value);
}

function handleOpen() {
    activeTab.value = "providers";
    fetchData();
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
        console.log('response', response);
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
    if (activeTab.value === 'providers') {
        editingProvider.value = null;
        providerFormOpen.value = true;
    } else {
        emit("add", activeTab.value);
    }
}

function handleConfigure(item) {
    emit("configure", { type: activeTab.value, item });
}

function handleEdit(item) {
    if (activeTab.value === 'providers') {
        editingProvider.value = item;
        providerFormOpen.value = true;
    } else {
        emit("edit", { type: activeTab.value, item });
    }
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
