<template>
    <div class="space-y-4">
        <div class="flex items-center justify-between">
            <div>
                <h3 class="text-lg font-semibold">供应商</h3>
                <p class="text-sm text-muted-foreground">管理 API 提供商配置</p>
            </div>
            <Button size="sm" @click="$emit('add')">
                <Plus class="h-4 w-4 mr-1" />
                添加供应商
            </Button>
        </div>

        <div class="rounded-md border">
            <Table>
                <TableHeader>
                    <TableRow>
                        <TableHead>名称</TableHead>
                        <TableHead>模型数量</TableHead>
                        <TableHead>状态</TableHead>
                        <TableHead>创建时间</TableHead>
                        <TableHead class="w-24">操作</TableHead>
                    </TableRow>
                </TableHeader>
                <TableBody>
                    <TableRow v-if="loading">
                        <TableCell colspan="5" class="text-center py-8">
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
                        >
                            <TableCell>
                                <div class="flex items-center gap-3">
                                    <div class="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center">
                                        <Building2 class="h-4 w-4 text-muted-foreground" />
                                    </div>
                                    <div>
                                        <p class="font-medium">{{ item.name }}</p>
                                        <p class="text-xs text-muted-foreground">{{ item.api_base }}</p>
                                    </div>
                                </div>
                            </TableCell>
                            <TableCell>
                                <Badge variant="secondary">{{ item.models_count || 0 }}</Badge>
                            </TableCell>
                            <TableCell>
                                <Badge
                                    :variant="item.status === 'active' ? 'default' : 'outline'"
                                >
                                    {{ item.status === 'active' ? '启用' : '禁用' }}
                                </Badge>
                            </TableCell>
                            <TableCell class="text-muted-foreground text-sm">
                                {{ formatDate(item.created_at) }}
                            </TableCell>
                            <TableCell>
                                <div class="flex items-center gap-1">
                                    <Tooltip>
                                        <TooltipTrigger as-child>
                                            <Button variant="ghost" size="icon-sm" @click="$emit('edit', item)">
                                                <Pencil class="h-4 w-4" />
                                            </Button>
                                        </TooltipTrigger>
                                        <TooltipContent>编辑</TooltipContent>
                                    </Tooltip>
                                    <Tooltip>
                                        <TooltipTrigger as-child>
                                            <Button variant="ghost" size="icon-sm" class="text-destructive" @click="$emit('delete', item)">
                                                <Trash2 class="h-4 w-4" />
                                            </Button>
                                        </TooltipTrigger>
                                        <TooltipContent>删除</TooltipContent>
                                    </Tooltip>
                                </div>
                            </TableCell>
                        </TableRow>
                    </template>
                    <TableRow v-else>
                        <TableCell colspan="5" class="text-center py-8 text-muted-foreground">
                            暂无数据
                        </TableCell>
                    </TableRow>
                </TableBody>
            </Table>
        </div>
    </div>
</template>

<script setup>
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Table, TableHeader, TableBody, TableRow, TableCell, TableHead } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Plus, Pencil, Trash2, Building2 } from "lucide-vue-next";

const props = defineProps({
    items: { type: Array, default: () => [] },
    loading: { type: Boolean, default: false },
});

const emit = defineEmits(["add", "edit", "delete"]);

function formatDate(dateStr) {
    if (!dateStr) return "-";
    const date = new Date(dateStr);
    return date.toLocaleDateString("zh-CN");
}
</script>
