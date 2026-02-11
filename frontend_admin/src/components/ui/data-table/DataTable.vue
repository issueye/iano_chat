<script setup>
import { computed } from 'vue'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'

/**
 * 数据表格组件
 * 外部提供数据和列配置，内部自动渲染
 */
const props = defineProps({
  /** 表格数据数组 */
  data: {
    type: Array,
    default: () => [],
  },
  /** 列配置数组 */
  columns: {
    type: Array,
    required: true,
    // 格式: [
    //   { key: 'name', title: '名称', width: '200px', align: 'left' },
    //   { key: 'status', title: '状态', slot: 'status' },
    //   { title: '操作', slot: 'actions', width: '120px' }
    // ]
  },
  /** 是否加载中 */
  loading: {
    type: Boolean,
    default: false,
  },
  /** 唯一标识字段名 */
  rowKey: {
    type: String,
    default: 'id',
  },
  /** 空数据提示文本 */
  emptyText: {
    type: String,
    default: '暂无数据',
  },
})

/**
 * 计算列数（用于空数据单元格合并）
 */
const columnCount = computed(() => props.columns.length)

/**
 * 获取单元格对齐方式
 * @param {Object} column - 列配置
 * @returns {string} 对齐类名
 */
function getAlignClass(column) {
  const alignMap = {
    left: 'text-left',
    center: 'text-center',
    right: 'text-right',
  }
  return alignMap[column.align] || 'text-left'
}

/**
 * 获取单元格内容
 * @param {Object} item - 行数据
 * @param {Object} column - 列配置
 * @returns {any} 单元格内容
 */
function getCellValue(item, column) {
  if (!column.key) return null
  const keys = column.key.split('.')
  let value = item
  for (const key of keys) {
    value = value?.[key]
  }
  return value
}
</script>

<template>
  <div class="rounded-md border">
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead
            v-for="column in columns"
            :key="column.key || column.slot"
            :class="getAlignClass(column)"
            :style="column.width ? { width: column.width } : {}"
          >
            {{ column.title }}
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <!-- 加载状态 -->
        <TableRow v-if="loading">
          <TableCell :colspan="columnCount" class="text-center py-8">
            <div class="flex items-center justify-center gap-2">
              <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-primary"></div>
              <span>加载中...</span>
            </div>
          </TableCell>
        </TableRow>
        <!-- 数据行 -->
        <template v-else-if="data.length > 0">
          <TableRow v-for="item in data" :key="item[rowKey]">
            <TableCell
              v-for="column in columns"
              :key="column.key || column.slot"
              :class="getAlignClass(column)"
            >
              <!-- 使用插槽 -->
              <slot
                v-if="column.slot"
                :name="column.slot"
                :row="item"
                :value="getCellValue(item, column)"
                :column="column"
              >
                {{ getCellValue(item, column) }}
              </slot>
              <!-- 默认显示 -->
              <template v-else>
                {{ getCellValue(item, column) }}
              </template>
            </TableCell>
          </TableRow>
        </template>
        <!-- 空数据 -->
        <TableRow v-else>
          <TableCell :colspan="columnCount" class="text-center py-8 text-muted-foreground">
            {{ emptyText }}
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  </div>
</template>
