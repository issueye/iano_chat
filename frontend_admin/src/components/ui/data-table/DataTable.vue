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
import { Tooltip } from '@/components/ui/tooltip'

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
    //   { key: 'name', title: '名称', width: '200px', align: 'left', ellipsis: true, tooltip: true },
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

/**
 * 获取列宽类名
 * @param {string} width - 宽度值
 * @returns {string} Tailwind 类名
 */
function getWidthClass(width) {
  // 如果是百分比宽度，转换为 flex-basis 样式
  if (typeof width === 'string' && width.includes('%')) {
    return ''
  }
  // 如果是固定像素宽度，使用内联样式，不添加类名
  return ''
}

/**
 * 获取列样式
 * @param {Object} column - 列配置
 * @returns {Object} 样式对象
 */
function getColumnStyle(column) {
  if (!column.width) return {}
  return { width: column.width, 'min-width': column.width }
}

/**
 * 获取省略类名
 * @param {Object} column - 列配置
 * @returns {string} 类名
 */
function getEllipsisClass(column) {
  if (column.ellipsis !== false) {
    // 默认启用省略，除非显式设置为 false
    return 'truncate'
  }
  return ''
}

/**
 * 是否需要显示 tooltip
 * @param {Object} column - 列配置
 * @returns {boolean}
 */
function needTooltip(column) {
  // 如果显式设置了 tooltip: true，或者设置了 ellipsis: true 但没有显式关闭 tooltip
  return column.tooltip === true || (column.ellipsis !== false && column.tooltip !== false)
}
</script>

<template>
  <div class="rounded-md border">
    <Table class="w-full table-fixed">
      <TableHeader>
        <TableRow>
          <TableHead
            v-for="column in columns"
            :key="column.key || column.slot"
            :class="[getAlignClass(column), column.width ? getWidthClass(column.width) : '']"
            :style="getColumnStyle(column)"
          >
            <div :class="['truncate', getEllipsisClass(column)]">
              {{ column.title }}
            </div>
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
              :class="[getAlignClass(column), column.width ? getWidthClass(column.width) : '']"
              :style="getColumnStyle(column)"
            >
              <!-- 使用插槽 -->
              <slot
                v-if="column.slot"
                :name="column.slot"
                :row="item"
                :value="getCellValue(item, column)"
                :column="column"
              >
                <Tooltip v-if="needTooltip(column)" :content="String(getCellValue(item, column) || '')">
                  <div :class="['truncate', getEllipsisClass(column)]">
                    {{ getCellValue(item, column) }}
                  </div>
                </Tooltip>
                <div v-else :class="['truncate', getEllipsisClass(column)]">
                  {{ getCellValue(item, column) }}
                </div>
              </slot>
              <!-- 默认显示 -->
              <template v-else>
                <Tooltip v-if="needTooltip(column)" :content="String(getCellValue(item, column) || '')">
                  <div :class="['truncate', getEllipsisClass(column)]">
                    {{ getCellValue(item, column) }}
                  </div>
                </Tooltip>
                <div v-else :class="['truncate', getEllipsisClass(column)]">
                  {{ getCellValue(item, column) }}
                </div>
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
