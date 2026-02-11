<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-2">
      <h2 class="text-2xl font-bold tracking-tight">欢迎使用后台管理系统</h2>
      <p class="text-muted-foreground">
        在这里管理您的供应商、Agents 和 Tools 配置
      </p>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">供应商总数</CardTitle>
          <Building2 class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ providerStore.totalCount }}</div>
          <p class="text-xs text-muted-foreground">API 提供商配置</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Agents 总数</CardTitle>
          <Bot class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ agentStore.totalCount }}</div>
          <p class="text-xs text-muted-foreground">AI Agent 配置</p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">Tools 总数</CardTitle>
          <Wrench class="h-4 w-4 text-muted-foreground" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold">{{ toolStore.totalCount }}</div>
          <p class="text-xs text-muted-foreground">
            已启用 {{ toolStore.enabledCount }}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader class="flex flex-row items-center justify-between space-y-0 pb-2">
          <CardTitle class="text-sm font-medium">系统状态</CardTitle>
          <Activity class="h-4 w-4 text-green-500" />
        </CardHeader>
        <CardContent>
          <div class="text-2xl font-bold text-green-600">正常</div>
          <p class="text-xs text-muted-foreground">所有服务运行中</p>
        </CardContent>
      </Card>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <Card class="cursor-pointer hover:bg-accent/50 transition-colors" @click="$router.push('/providers')">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Building2 class="h-5 w-5" />
            供应商管理
          </CardTitle>
          <CardDescription>
            管理 API 提供商配置，包括 OpenAI、Claude 等
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="secondary" class="w-full">
            进入管理
            <ArrowRight class="h-4 w-4 ml-2" />
          </Button>
        </CardContent>
      </Card>
      <Card class="cursor-pointer hover:bg-accent/50 transition-colors" @click="$router.push('/agents')">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Bot class="h-5 w-5" />
            Agents 管理
          </CardTitle>
          <CardDescription>
            管理 AI Agent 配置，设置系统提示词和工具
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="secondary" class="w-full">
            进入管理
            <ArrowRight class="h-4 w-4 ml-2" />
          </Button>
        </CardContent>
      </Card>
      <Card class="cursor-pointer hover:bg-accent/50 transition-colors" @click="$router.push('/tools')">
        <CardHeader>
          <CardTitle class="flex items-center gap-2">
            <Wrench class="h-5 w-5" />
            Tools 管理
          </CardTitle>
          <CardDescription>
            管理工具扩展配置，添加自定义功能
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button variant="secondary" class="w-full">
            进入管理
            <ArrowRight class="h-4 w-4 ml-2" />
          </Button>
        </CardContent>
      </Card>
    </div>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Building2,
  Bot,
  Wrench,
  Activity,
  ArrowRight,
} from 'lucide-vue-next'
import { useProviderStore, useAgentStore, useToolStore } from '@/stores'

const router = useRouter()

const providerStore = useProviderStore()
const agentStore = useAgentStore()
const toolStore = useToolStore()

onMounted(async () => {
  await Promise.all([
    providerStore.fetchAll(),
    agentStore.fetchAll(),
    toolStore.fetchAll(),
  ])
})
</script>
