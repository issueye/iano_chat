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
        <div class="rounded-md border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>名称</TableHead>
                <TableHead>API Base URL</TableHead>
                <TableHead>模型数量</TableHead>
                <TableHead>状态</TableHead>
                <TableHead>创建时间</TableHead>
                <TableHead class="w-24">操作</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-if="loading">
                <TableCell colspan="6" class="text-center py-8">
                  <div class="flex items-center justify-center gap-2">
                    <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-primary"></div>
                    <span>加载中...</span>
                  </div>
                </TableCell>
              </TableRow>
              <template v-else-if="items.length > 0">
                <TableRow v-for="item in items" :key="item.id">
                  <TableCell>
                    <div class="flex items-center gap-3">
                      <div class="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center">
                        <Building2 class="h-4 w-4 text-muted-foreground" />
                      </div>
                      <div>
                        <p class="font-medium">{{ item.name }}</p>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell class="text-muted-foreground text-sm max-w-[200px] truncate">
                    {{ item.base_url }}
                  </TableCell>
                  <TableCell>
                    <Badge variant="secondary">{{ item.models_count || 0 }}</Badge>
                  </TableCell>
                  <TableCell>
                    <Badge :variant="item.status === 'active' ? 'default' : 'outline'">
                      {{ item.status === "active" ? "启用" : "禁用" }}
                    </Badge>
                  </TableCell>
                  <TableCell class="text-muted-foreground text-sm">
                    {{ formatDate(item.created_at) }}
                  </TableCell>
                  <TableCell>
                    <div class="flex items-center gap-1">
                      <Tooltip content="编辑">
                        <Button variant="ghost" size="icon-sm" @click="handleEdit(item)">
                          <Pencil class="h-4 w-4" />
                        </Button>
                      </Tooltip>
                      <Tooltip content="删除">
                        <Button variant="ghost" size="icon-sm" class="text-destructive" @click="handleDelete(item)">
                          <Trash2 class="h-4 w-4" />
                        </Button>
                      </Tooltip>
                    </div>
                  </TableCell>
                </TableRow>
              </template>
              <TableRow v-else>
                <TableCell colspan="6" class="text-center py-8 text-muted-foreground">
                  暂无数据
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </div>
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
import {
  Table,
  TableHeader,
  TableBody,
  TableRow,
  TableCell,
  TableHead,
} from "@/components/ui/table";
import { Tooltip } from "@/components/ui/tooltip";
import { AlertDialog } from "@/components/ui/alert-dialog";
import ProviderFormDialog from "./components/ProviderFormDialog.vue";
import { Plus, Pencil, Trash2, Building2, CheckCircle2, XCircle } from "lucide-vue-next";

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
  return date.toLocaleDateString("zh-CN");
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
