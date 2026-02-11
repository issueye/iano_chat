<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">Tools 管理</h2>
        <p class="text-muted-foreground">
          管理工具扩展配置，包括添加、编辑、删除 Tool 信息
        </p>
      </div>
      <Button @click="handleAdd">
        <Plus class="h-4 w-4 mr-2" />
        添加 Tool
      </Button>
    </div>

    <!-- 统计卡片 -->
    <div class="grid gap-4 md:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Tool 总数</CardTitle>
          <Wrench class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.total }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">已启用</CardTitle>
          <CheckCircle2 class="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">{{ stats.active }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">已禁用</CardTitle>
          <XCircle class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-muted-foreground">{{ stats.inactive }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">分类数</CardTitle>
          <Tag class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ stats.categories }}</div>
        </CardContent>
      </Card>
    </div>

    <!-- 数据表格 -->
    <Card>
      <CardHeader>
        <CardTitle>Tool 列表</CardTitle>
        <CardDescription>管理所有工具扩展的配置信息</CardDescription>
      </CardHeader>
      <CardContent>
        <DataTable
          :data="items"
          :columns="columns"
          :loading="loading"
        >
          <!-- 名称列 -->
          <template #name="{ row }">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center">
                <Wrench class="h-4 w-4 text-muted-foreground" />
              </div>
              <div>
                <p class="font-medium">{{ row.name }}</p>
                <p class="text-xs text-muted-foreground truncate max-w-[180px]">
                  {{ row.desc }}
                </p>
              </div>
            </div>
          </template>
          
          <!-- 类型列 -->
          <template #type="{ value }">
            <Badge :variant="getTypeVariant(value)">{{ value }}</Badge>
          </template>
          
          <!-- 分类列 -->
          <template #category="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || "-" }}</span>
          </template>
          
          <!-- 版本列 -->
          <template #version="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || "v1.0" }}</span>
          </template>
          
          <!-- 作者列 -->
          <template #author="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || "-" }}</span>
          </template>
          
          <!-- 状态列 -->
          <template #status="{ value }">
            <Badge :variant="value === 'active' ? 'default' : 'outline'">
              {{ value === "active" ? "启用" : "禁用" }}
            </Badge>
          </template>
          
          <!-- 操作列 -->
          <template #actions="{ row }">
            <div class="flex items-center gap-1">
              <Tooltip content="编辑">
                <Button variant="ghost" size="icon-sm" @click="handleEdit(row)">
                  <Pencil class="h-4 w-4" />
                </Button>
              </Tooltip>
              <Tooltip content="删除">
                <Button variant="ghost" size="icon-sm" class="text-destructive" @click="handleDelete(row)">
                  <Trash2 class="h-4 w-4" />
                </Button>
              </Tooltip>
            </div>
          </template>
        </DataTable>
      </CardContent>
    </Card>

    <!-- 添加/编辑弹窗 -->
    <ToolFormDialog
      v-model:open="formDialogOpen"
      :tool="editingItem"
      @success="fetchData"
    />

    <!-- 删除确认弹窗 -->
    <AlertDialog
      v-model:open="deleteDialogOpen"
      :title="`删除 ${deletingItem?.name || ''}`"
      :description="`确定要删除 Tool「${deletingItem?.name}」吗？此操作无法撤销。`"
      confirm-text="删除"
      cancel-text="取消"
      variant="destructive"
      @confirm="executeDelete"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { DataTable } from "@/components/ui/data-table";
import { Tooltip } from "@/components/ui/tooltip";
import { AlertDialog } from "@/components/ui/alert-dialog";
import ToolFormDialog from "./components/ToolFormDialog.vue";
import { Plus, Pencil, Trash2, Wrench, CheckCircle2, XCircle, Tag } from "lucide-vue-next";

// 表格列配置
const columns = [
  { key: "name", title: "名称", width: "220px" },
  { key: "type", title: "类型", width: "100px", align: "center" },
  { key: "category", title: "分类", width: "120px" },
  { key: "version", title: "版本", width: "80px", align: "center" },
  { key: "author", title: "作者", width: "120px" },
  { key: "status", title: "状态", width: "100px", align: "center" },
  { title: "操作", slot: "actions", width: "100px", align: "center" },
];

// 数据状态
const items = ref([]);
const loading = ref(false);
const formDialogOpen = ref(false);
const editingItem = ref(null);
const deleteDialogOpen = ref(false);
const deletingItem = ref(null);

// 统计数据
const stats = computed(() => {
  const total = items.value.length;
  const active = items.value.filter(item => item.status === 'active').length;
  const categories = new Set(items.value.map(item => item.category).filter(Boolean)).size;
  return {
    total,
    active,
    inactive: total - active,
    categories
  };
});

/**
 * 获取类型对应的 Badge 变体
 * @param {string} type - 工具类型
 * @returns {string} Badge 变体名称
 */
function getTypeVariant(type) {
  const variants = {
    function: "default",
    search: "secondary",
    file: "outline",
    api: "secondary",
  };
  return variants[type] || "outline";
}

/**
 * 获取 Tool 列表数据
 */
async function fetchData() {
  loading.value = true;
  try {
    const response = await fetch("/api/tools");
    if (response.ok) {
      const result = await response.json();
      if (result.data) {
        items.value = result.data;
      }
    }
  } catch (error) {
    console.warn("Failed to fetch tools:", error);
  } finally {
    loading.value = false;
  }
}

/**
 * 打开添加 Tool 弹窗
 */
function handleAdd() {
  editingItem.value = null;
  formDialogOpen.value = true;
}

/**
 * 打开编辑 Tool 弹窗
 * @param {Object} item - Tool 数据
 */
function handleEdit(item) {
  editingItem.value = item;
  formDialogOpen.value = true;
}

/**
 * 打开删除确认弹窗
 * @param {Object} item - Tool 数据
 */
function handleDelete(item) {
  deletingItem.value = item;
  deleteDialogOpen.value = true;
}

/**
 * 执行删除操作
 */
async function executeDelete() {
  if (!deletingItem.value?.id) return;

  try {
    const response = await fetch(`/api/tools/${deletingItem.value.id}`, {
      method: "DELETE",
    });

    if (response.ok) {
      fetchData();
    } else {
      const error = await response.json();
      alert(error.message || "删除失败");
    }
  } catch (error) {
    console.error("Failed to delete tool:", error);
    alert("删除失败，请检查网络连接");
  } finally {
    deletingItem.value = null;
  }
}

onMounted(() => {
  fetchData();
});
</script>
