<template>
  <div class="space-y-6">
    <!-- 页面标题 -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">供应商管理</h2>
        <p class="text-muted-foreground">
          管理 API 提供商配置，包括添加、编辑、删除供应商信息
        </p>
      </div>
      <Button @click="handleAdd">
        <Plus class="h-4 w-4 mr-2" />
        添加供应商
      </Button>
    </div>

    <!-- 统计卡片 -->
    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">供应商总数</CardTitle>
          <Building2 class="h-4 w-4 text-muted-foreground" />
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
    </div>

    <!-- 数据表格 -->
    <Card>
      <CardHeader>
        <CardTitle>供应商列表</CardTitle>
        <CardDescription>管理所有 API 提供商的配置信息</CardDescription>
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
                <Building2 class="h-4 w-4 text-muted-foreground" />
              </div>
              <div>
                <p class="font-medium">{{ row.name }}</p>
              </div>
            </div>
          </template>
          
          <!-- API Base URL 列 -->
          <template #api_base="{ value }">
            <span class="text-muted-foreground text-sm max-w-[200px] truncate block">
              {{ value }}
            </span>
          </template>
          
          <!-- 模型数量列 -->
          <template #models_count="{ value }">
            <Badge variant="secondary">{{ value || 0 }}</Badge>
          </template>
          
          <!-- 状态列 -->
          <template #status="{ value }">
            <Badge :variant="value === 'active' ? 'default' : 'outline'">
              {{ value === "active" ? "启用" : "禁用" }}
            </Badge>
          </template>
          
          <!-- 创建时间列 -->
          <template #created_at="{ value }">
            <span class="text-muted-foreground text-sm">{{ formatDate(value) }}</span>
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
    <ProviderFormDialog
      v-model:open="formDialogOpen"
      :provider="editingItem"
      @success="fetchData"
    />

    <!-- 删除确认弹窗 -->
    <AlertDialog
      v-model:open="deleteDialogOpen"
      :title="`删除 ${deletingItem?.name || ''}`"
      :description="`确定要删除供应商「${deletingItem?.name}」吗？此操作无法撤销。`"
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
import ProviderFormDialog from "./components/ProviderFormDialog.vue";
import { Plus, Pencil, Trash2, Building2, CheckCircle2, XCircle } from "lucide-vue-next";

// 表格列配置
const columns = [
  { key: "name", title: "名称", width: "200px" },
  { key: "api_base", title: "API Base URL" },
  { key: "models_count", title: "模型数量", width: "100px", align: "center" },
  { key: "status", title: "状态", width: "100px", align: "center" },
  { key: "created_at", title: "创建时间", width: "180px", slot: "created_at" },
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
  return {
    total,
    active,
    inactive: total - active
  };
});

/**
 * 格式化日期
 * @param {string} dateStr - 日期字符串
 * @returns {string} 格式化后的日期
 */
function formatDate(dateStr) {
  if (!dateStr) return "-";
  const date = new Date(dateStr);
  return date.toLocaleDateString("zh-CN") + " " + date.toLocaleTimeString("zh-CN");
}

/**
 * 获取供应商列表数据
 */
async function fetchData() {
  loading.value = true;
  try {
    const response = await fetch("/api/providers");
    if (response.ok) {
      const result = await response.json();
      if (result.data) {
        items.value = result.data;
      }
    }
  } catch (error) {
    console.warn("Failed to fetch providers:", error);
  } finally {
    loading.value = false;
  }
}

/**
 * 打开添加供应商弹窗
 */
function handleAdd() {
  editingItem.value = null;
  formDialogOpen.value = true;
}

/**
 * 打开编辑供应商弹窗
 * @param {Object} item - 供应商数据
 */
function handleEdit(item) {
  editingItem.value = item;
  formDialogOpen.value = true;
}

/**
 * 打开删除确认弹窗
 * @param {Object} item - 供应商数据
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
    const response = await fetch(`/api/providers/${deletingItem.value.id}`, {
      method: "DELETE",
    });

    if (response.ok) {
      fetchData();
    } else {
      const error = await response.json();
      alert(error.message || "删除失败");
    }
  } catch (error) {
    console.error("Failed to delete provider:", error);
    alert("删除失败，请检查网络连接");
  } finally {
    deletingItem.value = null;
  }
}

onMounted(() => {
  fetchData();
});
</script>
