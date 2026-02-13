<template>
  <div class="space-y-6">
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

    <div class="grid gap-4 md:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Tool 总数</CardTitle>
          <Wrench class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ toolStore.totalCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">已启用</CardTitle>
          <CheckCircle2 class="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">{{ toolStore.enabledCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">已禁用</CardTitle>
          <XCircle class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-muted-foreground">{{ toolStore.disabledCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">类型数</CardTitle>
          <Tag class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ typeCount }}</div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Tool 列表</CardTitle>
        <CardDescription>管理所有工具扩展的配置信息</CardDescription>
      </CardHeader>
      <CardContent>
        <DataTable
          :data="toolStore.tools"
          :columns="columns"
          :loading="toolStore.loading"
        >
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
          
          <template #type="{ value }">
            <Badge :variant="getTypeVariant(value)">{{ getTypeLabel(value) }}</Badge>
          </template>
          
          <template #status="{ value }">
            <Badge :variant="value === 'enabled' ? 'default' : 'outline'">
              {{ value === "enabled" ? "启用" : "禁用" }}
            </Badge>
          </template>
          
          <template #version="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || "-" }}</span>
          </template>
          
          <template #author="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || "-" }}</span>
          </template>
          
          <template #created_at="{ value }">
            <span class="text-muted-foreground text-sm">{{ formatDatetime(value) }}</span>
          </template>
          
          <template #actions="{ row }">
            <div class="flex items-center justify-center gap-1">
              <Tooltip content="编辑">
                <Button variant="ghost" size="icon-sm" @click="handleEdit(row)">
                  <Pencil class="h-4 w-4" />
                </Button>
              </Tooltip>
              <Tooltip content="测试">
                <Button variant="ghost" size="icon-sm" @click="handleTest(row)">
                  <Play class="h-4 w-4" />
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

    <ToolFormDialog
      v-model:open="formDialogOpen"
      :tool="editingItem"
      @success="toolStore.fetchAll()"
    />

    <AlertDialog
      v-model:open="deleteDialogOpen"
      :title="`删除 ${deletingItem?.name || ''}`"
      :description="`确定要删除工具 ${deletingItem?.name} 吗？此操作无法撤销。`"
      confirmText="删除"
      cancelText="取消"
      variant="destructive"
      @confirm="executeDelete"
    >
      <p class="text-muted-foreground">
        确定要删除工具「{{ deletingItem?.name }}」吗？此操作无法撤销。
      </p>
    </AlertDialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from "vue"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
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
import ToolFormDialog from "./components/ToolFormDialog.vue"
import { Plus, Pencil, Trash2, Wrench, CheckCircle2, XCircle, Tag, Play } from "lucide-vue-next"
import { formatDatetime } from "@/lib/utils"
import { useToolStore } from "@/stores"

const toolStore = useToolStore()

const columns = [
  { key: "name", title: "名称" },
  { key: "type", title: "类型", width: "100px", align: "center" },
  { key: "status", title: "状态", width: "80px", align: "center" },
  { key: "version", title: "版本", width: "100px", align: "center" },
  { key: "author", title: "作者", width: "120px" },
  { key: "created_at", title: "创建时间", width: "160px" },
  { title: "操作", slot: "actions", width: "120px", align: "center" },
]

const formDialogOpen = ref(false)
const editingItem = ref(null)
const deleteDialogOpen = ref(false)
const deletingItem = ref(null)

const typeCount = computed(() => {
  const types = new Set(toolStore.tools.map(t => t.type).filter(Boolean))
  return types.size
})

function getTypeVariant(type) {
  const variants = {
    builtin: "default",
    custom: "secondary",
    external: "outline",
    plugin: "secondary",
  }
  return variants[type] || "outline"
}

function getTypeLabel(type) {
  const labels = {
    builtin: "内置",
    custom: "自定义",
    external: "外部",
    plugin: "插件",
  }
  return labels[type] || type
}

function handleAdd() {
  editingItem.value = null
  formDialogOpen.value = true
}

function handleEdit(item) {
  editingItem.value = item
  formDialogOpen.value = true
}

async function handleTest(item) {
  try {
    const result = await toolStore.test(item.id)
    alert(`测试成功: ${result.message || "工具定义加载成功"}`)
  } catch (error) {
    alert(error.message || "测试失败")
  }
}

function handleDelete(item) {
  deletingItem.value = item
  deleteDialogOpen.value = true
}

async function executeDelete() {
  if (!deletingItem.value?.id) return

  try {
    await toolStore.remove(deletingItem.value.id)
  } catch (error) {
    alert(error.message || "删除失败")
  } finally {
    deletingItem.value = null
  }
}

onMounted(() => {
  toolStore.fetchAll()
})
</script>
