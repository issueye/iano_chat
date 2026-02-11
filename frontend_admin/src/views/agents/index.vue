<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">Agents 管理</h2>
        <p class="text-muted-foreground">
          管理 AI Agent 配置，包括添加、编辑、删除 Agent 信息
        </p>
      </div>
      <Button @click="handleAdd">
        <Plus class="h-4 w-4 mr-2" />
        添加 Agent
      </Button>
    </div>

    <div class="grid gap-4 md:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Agent 总数</CardTitle>
          <Bot class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ agentStore.totalCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">主 Agent</CardTitle>
          <CheckCircle2 class="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">{{ mainAgentCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">子 Agent</CardTitle>
          <Layers class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ subAgentCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">自定义</CardTitle>
          <Settings class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ customAgentCount }}</div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>Agent 列表</CardTitle>
        <CardDescription>管理所有 AI Agent 的配置信息</CardDescription>
      </CardHeader>
      <CardContent>
        <DataTable
          :data="agentStore.agents"
          :columns="columns"
          :loading="agentStore.loading"
        >
          <template #name="{ row }">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center">
                <Bot class="h-4 w-4 text-muted-foreground" />
              </div>
              <div>
                <p class="font-medium">{{ row.name }}</p>
                <p class="text-xs text-muted-foreground truncate max-w-[200px]">
                  {{ row.description }}
                </p>
              </div>
            </div>
          </template>
          
          <template #type="{ value }">
            <Badge :variant="getTypeVariant(value)">{{ getTypeLabel(value) }}</Badge>
          </template>
          
          <template #model="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || "-" }}</span>
          </template>
          
          <template #is_sub_agent="{ value }">
            <Badge :variant="value ? 'secondary' : 'outline'">
              {{ value ? "是" : "否" }}
            </Badge>
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
              <Tooltip content="重载">
                <Button variant="ghost" size="icon-sm" @click="handleReload(row)">
                  <RefreshCw class="h-4 w-4" />
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

    <AgentFormDialog
      v-model:open="formDialogOpen"
      :agent="editingItem"
      @success="agentStore.fetchAll()"
    />

    <AlertDialog
      v-model:open="deleteDialogOpen"
      :title="`删除 ${deletingItem?.name || ''}`"
      :description="`确定要删除 Agent「${deletingItem?.name}」吗？此操作无法撤销。`"
      confirm-text="删除"
      cancel-text="取消"
      variant="destructive"
      @confirm="executeDelete"
    />
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
import AgentFormDialog from "./components/AgentFormDialog.vue"
import { Plus, Pencil, Trash2, Bot, CheckCircle2, RefreshCw, Layers, Settings } from "lucide-vue-next"
import { formatDatetime } from "@/lib/utils"
import { useAgentStore } from "@/stores"

const agentStore = useAgentStore()

const columns = [
  { key: "name", title: "名称" },
  { key: "type", title: "类型", width: "100px", align: "center" },
  { key: "model", title: "模型", width: "150px" },
  { key: "is_sub_agent", title: "子Agent", width: "80px", align: "center" },
  { key: "created_at", title: "创建时间", width: "180px" },
  { title: "操作", slot: "actions", width: "120px", fixed: "right", align: "center" },
]

const formDialogOpen = ref(false)
const editingItem = ref(null)
const deleteDialogOpen = ref(false)
const deletingItem = ref(null)

const mainAgentCount = computed(() => 
  agentStore.agents.filter(a => a.type === 'main').length
)

const subAgentCount = computed(() => 
  agentStore.agents.filter(a => a.type === 'sub').length
)

const customAgentCount = computed(() => 
  agentStore.agents.filter(a => a.type === 'custom').length
)

function getTypeVariant(type) {
  const variants = {
    main: "default",
    sub: "secondary",
    custom: "outline",
  }
  return variants[type] || "outline"
}

function getTypeLabel(type) {
  const labels = {
    main: "主 Agent",
    sub: "子 Agent",
    custom: "自定义",
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

async function handleReload(item) {
  try {
    await agentStore.reload(item.id)
    alert("Agent 重载成功")
  } catch (error) {
    alert(error.message || "重载失败")
  }
}

function handleDelete(item) {
  deletingItem.value = item
  deleteDialogOpen.value = true
}

async function executeDelete() {
  if (!deletingItem.value?.id) return

  try {
    await agentStore.remove(deletingItem.value.id)
  } catch (error) {
    alert(error.message || "删除失败")
  } finally {
    deletingItem.value = null
  }
}

onMounted(() => {
  agentStore.fetchAll()
})
</script>
