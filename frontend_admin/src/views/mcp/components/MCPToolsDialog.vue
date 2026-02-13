<template>
  <Dialog v-model:open="dialogOpen">
    <DialogContent class="max-w-lg max-h-[500px] overflow-hidden flex flex-col">
      <DialogTitle>{{ server?.name || 'MCP' }} 工具列表</DialogTitle>
      
      <div class="flex-1 overflow-y-auto space-y-3">
        <div v-if="loading" class="text-center py-4 text-muted-foreground">
          加载中...
        </div>
        
        <div v-else-if="tools.length === 0" class="text-center py-4 text-muted-foreground">
          暂无工具
        </div>
        
        <div v-else class="space-y-3">
          <div
            v-for="tool in tools"
            :key="tool.id"
            class="p-3 border rounded-lg hover:bg-accent/50 transition-colors"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <p class="font-medium truncate">{{ tool.name }}</p>
                <p class="text-sm text-muted-foreground line-clamp-2 mt-1">
                  {{ tool.description || '暂无描述' }}
                </p>
              </div>
            </div>
            
            <div v-if="tool.input_schema" class="mt-2">
              <details class="text-xs">
                <summary class="cursor-pointer text-muted-foreground hover:text-foreground">
                  查看参数
                </summary>
                <pre class="mt-2 p-2 bg-muted rounded overflow-x-auto">{{ formatJson(tool.input_schema) }}</pre>
              </details>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <Button variant="outline" @click="dialogOpen = false">关闭</Button>
      </template>
    </DialogContent>
  </Dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import { Dialog, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { useMCPStore } from '@/stores'

const props = defineProps({
  open: { type: Boolean, default: false },
  server: { type: Object, default: null },
})

const emit = defineEmits(['update:open'])

const dialogOpen = ref(props.open)
const tools = ref([])
const loading = ref(false)
const mcpStore = useMCPStore()

watch(() => props.open, (val) => {
  dialogOpen.value = val
  if (val && props.server?.id) {
    loadTools()
  }
})

watch(dialogOpen, (val) => {
  emit('update:open', val)
})

async function loadTools() {
  if (!props.server?.id) return
  
  loading.value = true
  try {
    tools.value = await mcpStore.getServerTools(props.server.id)
  } catch (error) {
    console.error('Failed to load tools:', error)
    tools.value = []
  } finally {
    loading.value = false
  }
}

function formatJson(str) {
  if (!str) return ''
  try {
    return JSON.stringify(JSON.parse(str), null, 2)
  } catch {
    return str
  }
}
</script>
