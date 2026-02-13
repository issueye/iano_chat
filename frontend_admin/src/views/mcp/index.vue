<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-2xl font-bold tracking-tight">MCP 服务器管理</h2>
        <p class="text-muted-foreground">
          管理 MCP (Model Context Protocol) 服务器配置，连接和管理工具
        </p>
      </div>
      <Button @click="handleAdd">
        <Plus class="h-4 w-4 mr-2" />
        添加 MCP 服务器
      </Button>
    </div>

    <div class="grid gap-4 md:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">服务器总数</CardTitle>
          <Server class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ mcpStore.totalCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">已连接</CardTitle>
          <CheckCircle2 class="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">{{ mcpStore.connectedCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">未连接</CardTitle>
          <XCircle class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-muted-foreground">{{ mcpStore.disconnectedCount }}</div>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">错误</CardTitle>
          <AlertTriangle class="h-4 w-4 text-red-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-red-600">{{ mcpStore.errorCount }}</div>
        </CardContent>
      </Card>
    </div>

    <Card>
      <CardHeader>
        <CardTitle>MCP 服务器列表</CardTitle>
        <CardDescription>管理所有 MCP 服务器配置和连接状态</CardDescription>
      </CardHeader>
      <CardContent>
        <DataTable
          :data="mcpStore.servers"
          :columns="columns"
          :loading="mcpStore.loading"
        >
          <template #name="{ row }">
            <div class="flex items-center gap-3">
              <div class="w-8 h-8 rounded-lg bg-secondary flex items-center justify-center">
                <Server class="h-4 w-4 text-muted-foreground" />
              </div>
              <div>
                <p class="font-medium">{{ row.name }}</p>
                <p class="text-xs text-muted-foreground truncate max-w-[180px]">
                  {{ row.desc }}
                </p>
              </div>
            </div>
          </template>
          
          <template #transport="{ value }">
            <Badge :variant="getTransportVariant(value)">{{ getTransportLabel(value) }}</Badge>
          </template>
          
          <template #status="{ value }">
            <Badge :variant="getStatusVariant(value)">
              {{ getStatusLabel(value) }}
            </Badge>
          </template>
          
          <template #enabled="{ value }">
            <Badge :variant="value ? 'default' : 'outline'">
              {{ value ? "启用" : "禁用" }}
            </Badge>
          </template>
          
          <template #tools_count="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || 0 }}</span>
          </template>
          
          <template #version="{ value }">
            <span class="text-sm text-muted-foreground">{{ value || "-" }}</span>
          </template>
          
          <template #created_at="{ value }">
            <span class="text-muted-foreground text-sm">{{ formatDatetime(value) }}</span>
          </template>
          
          <template #actions="{ row }">
            <div class="flex items-center justify-center gap-1">
              <Tooltip :content="row.status === 'connected' ? '断开连接' : '连接'">
                <Button 
                  variant="ghost" 
                  size="icon-sm" 
                  @click="handleConnect(row)"
                  :class="{ 'text-green-500': row.status !== 'connected', 'text-orange-500': row.status === 'connected' }"
                >
                  <Wifi v-if="row.status !== 'connected'" class="h-4 w-4" />
                  <WifiOff v-else class="h-4 w-4" />
                </Button>
              </Tooltip>
              <Tooltip content="查看工具">
                <Button variant="ghost" size="icon-sm" @click="handleViewTools(row)">
                  <Wrench class="h-4 w-4" />
                </Button>
              </Tooltip>
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

    <MCPFormDialog
      v-model:open="formDialogOpen"
      :server="editingItem"
      @success="mcpStore.fetchAllServers()"
    />

    <MCPToolsDialog
      v-model:open="toolsDialogOpen"
      :server="selectedServer"
    />

    <AlertDialog
      v-model:open="deleteDialogOpen"
      :title="`删除 ${deletingItem?.name || ''}`"
      :description="`确定要删除 MCP 服务器「${deletingItem?.name}」吗？此操作无法撤销。`"
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
import MCPFormDialog from "./components/MCPFormDialog.vue"
import MCPToolsDialog from "./components/MCPToolsDialog.vue"
import { Plus, Pencil, Trash2, Server, CheckCircle2, XCircle, AlertTriangle, Wifi, WifiOff, Wrench } from "lucide-vue-next"
import { formatDatetime } from "@/lib/utils"
import { useMCPStore } from "@/stores"

const mcpStore = useMCPStore()

const columns = [
  { key: "name", title: "名称" },
  { key: "transport", title: "传输类型", width: "100px", align: "center" },
  { key: "status", title: "状态", width: "100px", align: "center" },
  { key: "enabled", title: "启用", width: "80px", align: "center" },
  { key: "tools_count", title: "工具数", width: "80px", align: "center" },
  { key: "version", title: "版本", width: "100px", align: "center" },
  { key: "created_at", title: "创建时间", slot: "created_at", width: "160px" },
  { title: "操作", slot: "actions", width: "160px", align: "center" },
]

const formDialogOpen = ref(false)
const editingItem = ref(null)
const deleteDialogOpen = ref(false)
const deletingItem = ref(null)
const toolsDialogOpen = ref(false)
const selectedServer = ref(null)

function getTransportVariant(transport) {
  const variants = {
    stdio: "default",
    sse: "secondary",
    http: "outline",
  }
  return variants[transport] || "outline"
}

function getTransportLabel(transport) {
  const labels = {
    stdio: "Stdio",
    sse: "SSE",
    http: "HTTP",
  }
  return labels[transport] || transport
}

function getStatusVariant(status) {
  const variants = {
    connected: "default",
    disconnected: "outline",
    error: "destructive",
  }
  return variants[status] || "outline"
}

function getStatusLabel(status) {
  const labels = {
    connected: "已连接",
    disconnected: "未连接",
    error: "错误",
  }
  return labels[status] || status
}

function handleAdd() {
  editingItem.value = null
  formDialogOpen.value = true
}

function handleEdit(item) {
  editingItem.value = item
  formDialogOpen.value = true
}

async function handleConnect(item) {
  try {
    if (item.status === 'connected') {
      await mcpStore.disconnectServer(item.id)
    } else {
      await mcpStore.connectServer(item.id)
    }
  } catch (error) {
    alert(error.message || "操作失败")
  }
}

function handleViewTools(item) {
  selectedServer.value = item
  toolsDialogOpen.value = true
}

function handleDelete(item) {
  deletingItem.value = item
  deleteDialogOpen.value = true
}

async function executeDelete() {
  if (!deletingItem.value?.id) return

  try {
    await mcpStore.deleteServer(deletingItem.value.id)
  } catch (error) {
    alert(error.message || "删除失败")
  } finally {
    deletingItem.value = null
  }
}

onMounted(() => {
  mcpStore.fetchAllServers()
})
</script>
