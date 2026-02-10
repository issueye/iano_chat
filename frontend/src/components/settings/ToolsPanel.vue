<template>
    <div class="space-y-4">
        <div class="flex items-center justify-between">
            <div>
                <h3 class="text-lg font-semibold">Tools</h3>
                <p class="text-sm text-muted-foreground">管理工具扩展配置</p>
            </div>
            <Button size="sm" @click="$emit('add')">
                <Plus class="h-4 w-4 mr-1" />
                添加 Tool
            </Button>
        </div>

        <div class="rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead class="w-12">
                            <Checkbox
                                :checked="selectedItems.length === items.length && items.length > 0"
                                :indeterminate="selectedItems.length > 0 && selectedItems.length < items.length"
                                @update:checked="toggleSelectAll"
                            />
                        </TableHead>
                        <TableHead>名称</TableHead>
                        <TableHead>类型</TableHead>
                                                <TableHead>类型</TableHead>
                        <TableHead>分类</TableHead>
                        <TableHead>版本</TableHead>
                        <TableHead>作者</TableHead>
                        <TableHead>状态</TableHead>
                        <TableHead class="w-24">操作</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    <TableRow v-if="loading">
                        <TableCell colspan="8" class="text-center py-8">
                            <div class="flex items-center justify-center gap-2">
                                <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-primary"></div>
                                <span>加载中...</span>
                            </div>
                        </TableCell>
                    </TableRow>
                    <template v-else-if="items.length > 0">
                        <TableRow
                            v-for="item in items"
                            :key="item.id"
                            :class="{ 'bg-muted/50': selectedItems.includes(item.id) }"
                        >
                            <TableCell>
                                <Checkbox
                                    :checked="selectedItems.includes(item.id)"
                                    @update:checked="toggleSelect(item.id)"
                                />
                            </TableCell>
                            <TableCell>
                                <div class="flex items-center gap-3">
                                    <div class="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center">
                                        <Wrench class="h-4 w-4 text-muted-foreground" />
                                    </div>
                                    <div>
                                        <p class="font-medium">{{ item.name }}</p>
                                        <p class="text-xs text-muted-foreground truncate max-w-[180px]">{{ item.desc }}</p>
                                    </div>
                                </div>
                            </TableCell>
                            <TableCell>
                                <Badge :variant="getTypeVariant(item.type)">{{ item.type }}</Badge>
                            </TableCell>
                            <TableCell class="text-sm text-muted-foreground">
                                {{ item.category || "-" }}
                            </TableCell>
                            <TableCell class="text-sm text-muted-foreground">
                                {{ item.version || "v1.0" }}
                            </TableCell>
                            <TableCell class="text-sm text-muted-foreground">
                                {{ item.author || "-" }}
                            </TableCell>
                            <TableCell>
                                <Badge
                                    :variant="item.status === 'active' ? 'default' : 'outline'"
                                >
                                    {{ item.status === 'active' ? '启用' : '禁用' }}
                                </Badge>
                            </TableCell>
                            <TableCell>
                                <div class="flex items-center gap-1">
                                    <TooltipProvider>
                                        <Tooltip>
                                            <TooltipTrigger as-child>
                                                <Button variant="ghost" size="icon-sm" @click="$emit('configure', item)">
                                                    <Settings2 class="h-4 w-4" />
                                                </Button>
                                            </TooltipTrigger>
                                            <TooltipContent>配置</TooltipContent>
                                        </Tooltip>
                                    </TooltipProvider>
                                    <TooltipProvider>
                                        <Tooltip>
                                            <TooltipTrigger as-child>
                                                <Button variant="ghost" size="icon-sm" @click="$emit('edit', item)">
                                                    <Pencil class="h-4 w-4" />
                                                </Button>
                                            </TooltipTrigger>
                                            <TooltipContent>编辑</TooltipContent>
                                        </Tooltip>
                                    </TooltipProvider>
                                    <TooltipProvider>
                                        <Tooltip>
                                            <TooltipTrigger as-child>
                                                <Button variant="ghost" size="icon-sm" class="text-destructive" @click="$emit('delete', item)">
                                                    <Trash2 class="h-4 w-4" />
                                                </Button>
                                            </TooltipTrigger>
                                            <TooltipContent>删除</TooltipContent>
                                        </Tooltip>
                                    </TooltipProvider>
                                </div>
                            </TableCell>
                        </TableRow>
                    </template>
                    <TableRow v-else>
                        <TableCell colspan="8" class="text-center py-8 text-muted-foreground">
                            暂无数据
                        </TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        </div>

        <div v-if="selectedItems.length > 0" class="flex items-center justify-between p-3 bg-muted rounded-lg">
            <span class="text-sm text-muted-foreground">已选择 {{ selectedItems.length }} 项</span>
            <div class="flex gap-2">
                <Button variant="outline" size="sm" @click="selectedItems = []">取消选择</Button>
                <Button variant="destructive" size="sm" @click="$emit('batch-delete', selectedItems)">
                    批量删除
                </Button>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref } from "vue";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Table, TableHeader, TableBody, TableRow, TableCell, TableHead } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { Plus, Settings2, Pencil, Trash2, Wrench } from "lucide-vue-next";

const props = defineProps({
    items: { type: Array, default: () => [] },
    loading: { type: Boolean, default: false },
});

const emit = defineEmits(["add", "configure", "edit", "delete", "batch-delete"]);

const selectedItems = ref([]);

function toggleSelect(id) {
    const index = selectedItems.value.indexOf(id);
    if (index > -1) {
        selectedItems.value.splice(index, 1);
    } else {
        selectedItems.value.push(id);
    }
}

function toggleSelectAll(checked) {
    if (checked) {
        selectedItems.value = props.items.map((item) => item.id);
    } else {
        selectedItems.value = [];
    }
}

function getTypeVariant(type) {
    const variants = {
        function: "default",
        search: "secondary",
        file: "outline",
        api: "secondary",
    };
    return variants[type] || "outline";
}
</script>
