<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-2">
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

    <div class="grid gap-4 md:grid-cols-3">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">供应商总数</CardTitle>
          <Building2 class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ providerStore.totalCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">已配置</CardTitle>
          <CheckCircle2 class="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">{{ providerStore.totalCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">状态</CardTitle>
          <Activity class="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">正常</div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>供应商列表</CardTitle>
        <CardDescription>管理所有 API 提供商的配置信息</CardDescription>
      </CardHeader>
      <CardContent>
        <DataTable
          :data="providerStore.providers"
          :columns="columns"
          :loading="providerStore.loading"
        >
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
          
          <template #base_url="{ value }">
            <span class="text-muted-foreground text-sm max-w-[200px] truncate block">
              {{ value }}
            </span>
          </template>
          
          <template #created_at="{ value }">
            <span class="text-muted-foreground text-sm">{{ formatDatetime(value) }}</span>
          </template>
          
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

    <ProviderFormDialog
      v-model:open="formDialogOpen"
      :provider="editingItem"
      @success="providerStore.fetchAll()"
    />

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
import { ref, onMounted } from "vue"
import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { DataTable } from "@/components/ui/data-table"
import { Tooltip } from "@/components/ui/tooltip"
import { AlertDialog } from "@/components/ui/alert-dialog"
import ProviderFormDialog from "./components/ProviderFormDialog.vue"
import { Plus, Pencil, Trash2, Building2, CheckCircle2, Activity } from "lucide-vue-next"
import { formatDatetime } from "@/lib/utils"
import { useProviderStore } from "@/stores"

const providerStore = useProviderStore()

const columns = [
  { key: "name", title: "名称", width: "200px" },
  { key: "base_url", title: "API Base URL" },
  { key: "model", title: "模型名称", width: "150px", align: "center" },
  { key: "created_at", title: "创建时间", width: "180px", slot: "created_at" },
  { title: "操作", slot: "actions", width: "100px", align: "center" },
]

const formDialogOpen = ref(false)
const editingItem = ref(null)
const deleteDialogOpen = ref(false)
const deletingItem = ref(null)

function handleAdd() {
  editingItem.value = null
  formDialogOpen.value = true
}

function handleEdit(item) {
  editingItem.value = item
  formDialogOpen.value = true
}

function handleDelete(item) {
  deletingItem.value = item
  deleteDialogOpen.value = true
}

async function executeDelete() {
  if (!deletingItem.value?.id) return

  try {
    await providerStore.remove(deletingItem.value.id)
  } catch (error) {
    alert(error.message || "删除失败")
  } finally {
    deletingItem.value = null
  }
}

onMounted(() => {
  providerStore.fetchAll()
})
</script>
