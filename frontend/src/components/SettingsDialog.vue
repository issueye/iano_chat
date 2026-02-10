<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent
            class="sm:max-w-[1200px] max-w-[95vw] max-h-[90vh] overflow-hidden flex bg-card"
        >
            <DialogHeader class="sr-only">
                <DialogTitle>设置</DialogTitle>
                <DialogDescription>管理您的供应商、Agents 和 Tools 配置</DialogDescription>
            </DialogHeader>

            <div class="flex w-full h-full">
                <Transition name="slide-right" mode="out-in">
                    <div
                        v-if="!showConfig"
                        class="flex w-full"
                        key="main-panel"
                    >
                        <div class="w-48 border-r border-border pr-4 flex flex-col flex-shrink-0">
                            <div class="mb-4 px-1">
                                <h2 class="font-semibold text-sm text-foreground">
                                    设置
                                </h2>
                                <p class="text-xs text-muted-foreground">
                                    管理配置
                                </p>
                            </div>

                            <nav
                                class="space-y-1 flex-1"
                                role="tablist"
                                aria-label="设置分类"
                            >
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
                                    :aria-controls="`panel-${tab.value}`"
                                    @click="activeTab = tab.value"
                                >
                                    <component
                                        :is="tab.icon"
                                        class="h-4 w-4 flex-shrink-0"
                                    />
                                    <span class="truncate">{{ tab.label }}</span>
                                </button>
                            </nav>
                        </div>

                        <div
                            :id="`panel-${activeTab}`"
                            class="flex-1 pl-4 overflow-hidden flex flex-col"
                            role="tabpanel"
                            :aria-labelledby="`tab-${activeTab}`"
                        >
                            <div class="flex items-center justify-between mb-4 flex-shrink-0">
                                <div class="relative flex-1 max-w-xs">
                                    <Search
                                        class="absolute left-2.5 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground"
                                    />
                                    <Input
                                        v-model="searchQuery"
                                        type="search"
                                        :placeholder="`搜索${currentTabLabel}...`"
                                        class="pl-9 h-9"
                                        @keydown.esc="searchQuery = ''"
                                    />
                                </div>
                                <Button
                                    variant="default"
                                    size="sm"
                                    class="ml-2"
                                    @click="handleAdd"
                                >
                                    <Plus class="h-4 w-4 mr-1.5" />
                                    添加{{ currentTabLabel }}
                                </Button>
                            </div>

                            <ScrollArea class="flex-1 -mr-4 pr-4">
                                <div class="space-y-2 min-h-[200px]">
                                    <TransitionGroup name="list">
                                        <SettingsListItem
                                            v-for="item in filteredItems"
                                            :key="item.id"
                                            :title="item.name"
                                            :subtitle="getItemSubtitle(item)"
                                            :icon="currentTabIcon"
                                            :badge="getItemBadge(item)"
                                            @click="handleItemClick(item)"
                                            @configure="handleConfigure(item)"
                                        />
                                    </TransitionGroup>

                                    <Transition name="fade" mode="out-in">
                                        <div
                                            v-if="loading"
                                            key="loading"
                                            class="flex items-center justify-center py-12"
                                        >
                                            <div
                                                class="flex flex-col items-center gap-2"
                                            >
                                                <div
                                                    class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"
                                                />
                                                <p
                                                    class="text-sm text-muted-foreground"
                                                >
                                                    加载中...
                                                </p>
                                            </div>
                                        </div>

                                        <div
                                            v-else-if="filteredItems.length === 0"
                                            key="empty"
                                            class="flex flex-col items-center justify-center py-12 text-center"
                                        >
                                            <div
                                                class="rounded-full bg-muted p-3 mb-3"
                                            >
                                                <component
                                                    :is="currentTabIcon"
                                                    class="h-6 w-6 text-muted-foreground"
                                                />
                                            </div>
                                            <p
                                                class="text-sm font-medium text-foreground mb-1"
                                            >
                                                {{ searchQuery ? '未找到匹配项' : `暂无${currentTabLabel}` }}
                                            </p>
                                            <p
                                                class="text-xs text-muted-foreground mb-3"
                                            >
                                                {{
                                                    searchQuery
                                                        ? '请尝试其他搜索词'
                                                        : `点击上方按钮添加新的${currentTabLabel}`
                                                }}
                                            </p>
                                            <Button
                                                v-if="searchQuery"
                                                variant="ghost"
                                                size="sm"
                                                @click="searchQuery = ''"
                                            >
                                                清除搜索
                                            </Button>
                                        </div>
                                    </Transition>
                                </div>
                            </ScrollArea>
                        </div>
                    </div>

                    <div
                        v-else
                        key="config-panel"
                        class="flex-1 pl-4 overflow-hidden flex flex-col"
                    >
                        <div class="flex items-center gap-2 mb-4 flex-shrink-0">
                            <Button
                                variant="ghost"
                                size="sm"
                                @click="showConfig = false"
                            >
                                <ArrowLeft class="h-4 w-4 mr-1" />
                                返回
                            </Button>
                            <h3 class="font-semibold text-sm">
                                配置 {{ configItem?.name }}
                            </h3>
                        </div>
                        <ScrollArea class="flex-1 -mr-4 pr-4">
                            <Card class="mb-4">
                                <CardHeader>
                                    <CardTitle class="text-sm">基本信息</CardTitle>
                                </CardHeader>
                                <CardContent class="space-y-3">
                                    <div class="grid grid-cols-2 gap-3">
                                        <div class="space-y-1">
                                            <Label class="text-xs">名称</Label>
                                            <Input
                                                v-model="configForm.name"
                                                placeholder="输入名称"
                                            />
                                        </div>
                                        <div class="space-y-1">
                                            <Label class="text-xs">类型</Label>
                                            <Input
                                                v-model="configForm.type"
                                                placeholder="输入类型"
                                                disabled
                                            />
                                        </div>
                                    </div>
                                    <div class="space-y-1">
                                        <Label class="text-xs">描述</Label>
                                        <Input
                                            v-model="configForm.description"
                                            placeholder="输入描述"
                                        />
                                    </div>
                                </CardContent>
                            </Card>
                            <Card v-if="configItem?.type === 'tool'" class="mb-4">
                                <CardHeader>
                                    <CardTitle class="text-sm">Tool 配置</CardTitle>
                                </CardHeader>
                                <CardContent class="space-y-3">
                                    <div class="space-y-1">
                                        <Label class="text-xs">参数配置</Label>
                                        <Textarea
                                            v-model="configForm.config"
                                            placeholder="JSON 格式"
                                            rows="4"
                                        />
                                    </div>
                                </CardContent>
                            </Card>
                            <Card class="mb-4">
                                <CardHeader class="flex flex-row items-center justify-between">
                                    <CardTitle class="text-sm">高级设置</CardTitle>
                                    <Switch v-model="advancedEnabled" />
                                </CardHeader>
                                <Transition name="expand">
                                    <CardContent v-if="advancedEnabled" class="space-y-3 pt-0">
                                        <div class="space-y-1">
                                            <Label class="text-xs">自定义参数</Label>
                                            <Textarea
                                                v-model="configForm.extra"
                                                placeholder="额外配置"
                                                rows="3"
                                            />
                                        </div>
                                    </CardContent>
                                </Transition>
                            </Card>
                            <div class="flex justify-end gap-2">
                                <Button variant="outline" @click="showConfig = false">
                                    取消
                                </Button>
                                <Button @click="saveConfig">
                                    保存配置
                                </Button>
                            </div>
                        </ScrollArea>
                    </div>
                </Transition>
            </div>
        </DialogContent>
    </Dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from "vue";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogDescription } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Card, CardHeader, CardTitle, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Switch } from "@/components/ui/switch";
import SettingsListItem from "@/components/SettingsListItem.vue";
import {
    Building2,
    Bot,
    Wrench,
    Plus,
    Search,
    ArrowLeft,
    Settings2,
} from "lucide-vue-next";

const props = defineProps({
    open: { type: Boolean, default: false },
});

const emit = defineEmits(["update:open", "add", "configure"]);

const activeTab = ref("providers");
const searchQuery = ref("");
const loading = ref(false);
const showConfig = ref(false);
const configItem = ref(null);
const advancedEnabled = ref(false);

const configForm = ref({
    name: "",
    type: "",
    description: "",
    config: "",
    extra: "",
});

const tabs = [
    { value: "providers", label: "供应商", icon: Building2 },
    { value: "agents", label: "Agents", icon: Bot },
    { value: "tools", label: "Tools", icon: Wrench },
];

const dataMap = {
    providers: ref([]),
    agents: ref([]),
    tools: ref([]),
};

const currentTabLabel = computed(() => {
    const tab = tabs.find((t) => t.value === activeTab.value);
    return tab?.label || "";
});

const currentTabIcon = computed(() => {
    const tab = tabs.find((t) => t.value === activeTab.value);
    return tab?.icon || Settings2;
});

const currentItems = computed(() => {
    return dataMap[activeTab.value]?.value || [];
});

const filteredItems = computed(() => {
    let items = currentItems.value;
    if (searchQuery.value) {
        const query = searchQuery.value.toLowerCase();
        items = items.filter(
            (item) =>
                item.name?.toLowerCase().includes(query) ||
                item.description?.toLowerCase().includes(query) ||
                item.category?.toLowerCase().includes(query)
        );
    }
    return items;
});

function getItemSubtitle(item) {
    if (activeTab.value === "providers") {
        return `${item.models || 0} 个模型`;
    }
    return item.description || item.category || "";
}

function getItemBadge(item) {
    if (activeTab.value === "tools") {
        return item.category || "";
    }
    if (activeTab.value === "providers") {
        return `${item.models || 0}`;
    }
    return "";
}

function handleOpenChange(value) {
    emit("update:open", value);
    if (value) {
        searchQuery.value = "";
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
                dataMap[activeTab.value].value = result.data;
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

function handleItemClick(item) {
    handleConfigure(item);
}

function handleConfigure(item) {
    configItem.value = item;
    configForm.value = {
        name: item.name || "",
        type: item.type || activeTab.value.slice(0, -1),
        description: item.description || "",
        config: item.config || "",
        extra: item.extra || "",
    };
    showConfig.value = true;
}

function saveConfig() {
    emit("configure", { item: configItem.value, form: configForm.value });
    showConfig.value = false;
}

function handleKeyboard(e) {
    if (!props.open) return;
    if ((e.metaKey || e.ctrlKey) && e.key === "[") {
        e.preventDefault();
        const currentIndex = tabs.findIndex(
            (t) => t.value === activeTab.value
        );
        if (currentIndex > 0) {
            activeTab.value = tabs[currentIndex - 1].value;
        }
    }
    if ((e.metaKey || e.ctrlKey) && e.key === "]") {
        e.preventDefault();
        const currentIndex = tabs.findIndex(
            (t) => t.value === activeTab.value
        );
        if (currentIndex < tabs.length - 1) {
            activeTab.value = tabs[currentIndex + 1].value;
        }
    }
}

watch(activeTab, () => {
    searchQuery.value = "";
    fetchData();
});

onMounted(() => {
    window.addEventListener("keydown", handleKeyboard);
});

onUnmounted(() => {
    window.removeEventListener("keydown", handleKeyboard);
});
</script>

<style scoped>
.list-enter-active,
.list-leave-active {
    transition: all 0.2s ease;
}
.list-enter-from,
.list-leave-to {
    opacity: 0;
    transform: translateX(-10px);
}

.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.15s ease;
}
.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}

.expand-enter-active,
.expand-leave-active {
    transition: all 0.2s ease;
    overflow: hidden;
}
.expand-enter-from,
.expand-leave-to {
    opacity: 0;
    max-height: 0;
}

.slide-right-enter-active,
.slide-right-leave-active {
    transition: all 0.2s ease;
}
.slide-right-enter-from {
    opacity: 0;
    transform: translateX(-10px);
}
.slide-right-leave-to {
    opacity: 0;
    transform: translateX(10px);
}
</style>
