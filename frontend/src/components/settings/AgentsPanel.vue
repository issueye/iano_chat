<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-lg font-semibold">Agents</h3>
        <p class="text-sm text-muted-foreground">管理 AI Agent 配置</p>
      </div>
      <Button size="sm" @click="$emit('add')">
        <Plus class="h-4 w-4 mr-1" />
        添加 Agent
      </Button>
    </div>

    <div class="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>名称</TableHead>
            <TableHead>类型</TableHead>
            <TableHead>模型</TableHead>
            <TableHead>Tools</TableHead>
            <TableHead>状态</TableHead>
            <TableHead class="w-24">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-if="loading">
            <TableCell colspan="6" class="text-center py-8">
              <div class="flex items-center justify-center gap-2">
                <div
                  class="animate-spin rounded-full h-5 w-5 border-b-2 border-primary"
                ></div>
                <span>加载中...</span>
              </div>
            </TableCell>
          </TableRow>
          <template v-else-if="items.length > 0">
            <TableRow v-for="item in items" :key="item.id">
              <TableCell>
                <div class="flex items-center gap-3">
                  <div
                    class="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center"
                  >
                    <Bot class="h-4 w-4 text-muted-foreground" />
                  </div>
                  <div>
                    <p class="font-medium">{{ item.name }}</p>
                    <p
                      class="text-xs text-muted-foreground truncate max-w-[200px]"
                    >
                      {{ item.description }}
                    </p>
                  </div>
                </div>
              </TableCell>
              <TableCell>
                <Badge variant="outline">{{ item.type }}</Badge>
              </TableCell>
              <TableCell class="text-sm text-muted-foreground">
                {{ item.model || "-" }}
              </TableCell>
              <TableCell>
                <Tooltip :content="item.tools">
                  <Badge variant="secondary">
                    {{ item.tools_count || 0 }} 个
                  </Badge>
                </Tooltip>
              </TableCell>
              <TableCell>
                <Badge
                  :variant="item.status === 'active' ? 'default' : 'outline'"
                >
                  {{ item.status === "active" ? "启用" : "禁用" }}
                </Badge>
              </TableCell>
              <TableCell>
                <div class="flex items-center gap-1">
                  <Tooltip content="配置">
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      @click="$emit('configure', item)"
                    >
                      <Settings2 class="h-4 w-4" />
                    </Button>
                  </Tooltip>
                  <Tooltip content="编辑">
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      @click="$emit('edit', item)"
                    >
                      <Pencil class="h-4 w-4" />
                    </Button>
                  </Tooltip>
                  <Tooltip content="删除">
                    <Button
                      variant="ghost"
                      size="icon-sm"
                      class="text-destructive"
                      @click="$emit('delete', item)"
                    >
                      <Trash2 class="h-4 w-4" />
                    </Button>
                  </Tooltip>
                </div>
              </TableCell>
            </TableRow>
          </template>
          <TableRow v-else>
            <TableCell
              colspan="6"
              class="text-center py-8 text-muted-foreground"
            >
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
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableCell,
  TableHead,
} from "@/components/ui/table";
import { Tooltip } from "@/components/ui/tooltip";
import { Plus, Settings2, Pencil, Trash2, Bot } from "lucide-vue-next";

const props = defineProps({
  items: { type: Array, default: () => [] },
  loading: { type: Boolean, default: false },
});

const emit = defineEmits(["add", "configure", "edit", "delete"]);
</script>
