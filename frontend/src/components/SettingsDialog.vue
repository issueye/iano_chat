<template>
    <Dialog :open="open" @update:open="handleOpenChange">
        <DialogContent
            class="sm:max-w-[600px] max-h-[80vh] overflow-hidden flex flex-col"
        >
            <DialogHeader>
                <DialogTitle class="flex items-center gap-2">
                    <Settings class="h-5 w-5" />
                    设置
                </DialogTitle>
                <DialogDescription>
                    管理您的供应商、Agent 和 Tools 配置
                </DialogDescription>
            </DialogHeader>

            <!-- Tabs -->
            <Tabs
                v-model="activeTab"
                class="flex-1 flex flex-col overflow-hidden"
            >
                <TabsList class="grid w-full grid-cols-3 mb-4">
                    <TabsTrigger value="providers" class="gap-1.5">
                        <Building2 class="h-4 w-4" />
                        供应商
                    </TabsTrigger>
                    <TabsTrigger value="agents" class="gap-1.5">
                        <Bot class="h-4 w-4" />
                        Agents
                    </TabsTrigger>
                    <TabsTrigger value="tools" class="gap-1.5">
                        <Wrench class="h-4 w-4" />
                        Tools
                    </TabsTrigger>
                </TabsList>

                <!-- Providers Tab -->
                <TabsContent
                    value="providers"
                    class="flex-1 overflow-auto mt-0"
                >
                    <div class="space-y-3">
                        <div
                            v-for="provider in providers"
                            :key="provider.id"
                            class="flex items-center justify-between p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors"
                        >
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-9 h-9 rounded-lg bg-secondary flex items-center justify-center"
                                >
                                    <Building2
                                        class="h-4 w-4 text-muted-foreground"
                                    />
                                </div>
                                <div>
                                    <p class="font-medium text-sm">
                                        {{ provider.name }}
                                    </p>
                                    <p class="text-xs text-muted-foreground">
                                        {{ provider.models }} 个模型
                                    </p>
                                </div>
                            </div>
                            <Button
                                variant="ghost"
                                size="sm"
                                class="text-muted-foreground"
                            >
                                配置
                            </Button>
                        </div>
                        <Button variant="outline" class="w-full" size="sm">
                            <Plus class="h-4 w-4 mr-1.5" />
                            添加供应商
                        </Button>
                    </div>
                </TabsContent>

                <!-- Agents Tab -->
                <TabsContent value="agents" class="flex-1 overflow-auto mt-0">
                    <div class="space-y-3">
                        <div
                            v-for="agent in agents"
                            :key="agent.id"
                            class="flex items-center justify-between p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors"
                        >
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-9 h-9 rounded-lg bg-secondary flex items-center justify-center"
                                >
                                    <Bot
                                        class="h-4 w-4 text-muted-foreground"
                                    />
                                </div>
                                <div>
                                    <p class="font-medium text-sm">
                                        {{ agent.name }}
                                    </p>
                                    <p class="text-xs text-muted-foreground">
                                        {{ agent.description }}
                                    </p>
                                </div>
                            </div>
                            <Button
                                variant="ghost"
                                size="sm"
                                class="text-muted-foreground"
                            >
                                配置
                            </Button>
                        </div>
                        <Button variant="outline" class="w-full" size="sm">
                            <Plus class="h-4 w-4 mr-1.5" />
                            添加 Agent
                        </Button>
                    </div>
                </TabsContent>

                <!-- Tools Tab -->
                <TabsContent value="tools" class="flex-1 overflow-auto mt-0">
                    <div class="space-y-3">
                        <div
                            v-for="tool in tools"
                            :key="tool.id"
                            class="flex items-center justify-between p-3 rounded-lg border border-border hover:bg-muted/50 transition-colors"
                        >
                            <div class="flex items-center gap-3">
                                <div
                                    class="w-9 h-9 rounded-lg bg-secondary flex items-center justify-center"
                                >
                                    <Wrench
                                        class="h-4 w-4 text-muted-foreground"
                                    />
                                </div>
                                <div>
                                    <p class="font-medium text-sm">
                                        {{ tool.name }}
                                    </p>
                                    <p class="text-xs text-muted-foreground">
                                        {{ tool.category }}
                                    </p>
                                </div>
                            </div>
                            <Button
                                variant="ghost"
                                size="sm"
                                class="text-muted-foreground"
                            >
                                配置
                            </Button>
                        </div>
                        <Button variant="outline" class="w-full" size="sm">
                            <Plus class="h-4 w-4 mr-1.5" />
                            添加 Tool
                        </Button>
                    </div>
                </TabsContent>
            </Tabs>
        </DialogContent>
    </Dialog>
</template>

<script setup>
import { ref } from "vue";
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Settings, Building2, Bot, Wrench, Plus } from "lucide-vue-next";

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
});

const emit = defineEmits(["update:open"]);

const activeTab = ref("providers");

function handleOpenChange(value) {
    emit("update:open", value);
}

// Mock data - 后续可以从 API 获取
const providers = ref([
    { id: 1, name: "OpenAI", models: 5 },
    { id: 2, name: "Anthropic", models: 3 },
    { id: 3, name: "DeepSeek", models: 2 },
]);

const agents = ref([
    { id: 1, name: "代码助手", description: "专注于编程和代码相关任务" },
    { id: 2, name: "写作助手", description: "帮助撰写文章、文档" },
    { id: 3, name: "数据分析", description: "处理和分析数据" },
]);

const tools = ref([
    { id: 1, name: "网络搜索", category: "搜索" },
    { id: 2, name: "文件读取", category: "文件" },
    { id: 3, name: "计算器", category: "实用工具" },
]);
</script>
