<script setup>
import { computed, ref, watch } from 'vue'

import { resolvePageJump } from './adminTablePagination.helpers'

const props = defineProps({
  currentPage: {
    type: Number,
    default: 1
  },
  pageSize: {
    type: Number,
    default: 20
  },
  total: {
    type: Number,
    default: 0
  },
  layout: {
    type: String,
    default: 'total, prev, pager, next'
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:currentPage', 'update:pageSize', 'current-change', 'size-change'])

const jumpPage = ref('1')

const jumpState = computed(() =>
  resolvePageJump(jumpPage.value, {
    currentPage: props.currentPage,
    pageSize: props.pageSize,
    total: props.total
  })
)

const jumpDisabled = computed(() => props.disabled || jumpState.value.disabled)

watch(
  () => props.currentPage,
  (value) => {
    const normalized = Number.isInteger(Number(value)) && Number(value) > 0 ? Number(value) : 1
    jumpPage.value = String(normalized)
  },
  { immediate: true }
)

function handleCurrentChange(page) {
  emit('update:currentPage', page)
  emit('current-change', page)
}

function handlePageSizeChange(size) {
  emit('update:pageSize', size)
  emit('size-change', size)
}

function resetJumpInput() {
  jumpPage.value = jumpState.value.displayValue
}

function submitJump() {
  const result = jumpState.value
  jumpPage.value = result.displayValue
  if (props.disabled || result.disabled || !result.shouldJump) {
    return
  }
  emit('update:currentPage', result.page)
  emit('current-change', result.page)
}
</script>

<template>
  <div class="admin-table-pagination">
    <el-pagination
      :current-page="currentPage"
      :page-size="pageSize"
      :layout="layout"
      :total="total"
      :disabled="disabled"
      @current-change="handleCurrentChange"
      @size-change="handlePageSizeChange"
    />

    <div class="admin-table-pagination__jump">
      <span class="admin-table-pagination__label">页码</span>
      <el-input
        v-model="jumpPage"
        class="admin-table-pagination__input"
        inputmode="numeric"
        :disabled="jumpDisabled"
        @keyup.enter="submitJump"
        @blur="resetJumpInput"
      />
      <el-button :disabled="jumpDisabled" @click="submitJump">跳转</el-button>
    </div>
  </div>
</template>

<style scoped>
.admin-table-pagination {
  display: flex;
  justify-content: flex-end;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  width: 100%;
}

.admin-table-pagination__jump {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.admin-table-pagination__label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}

.admin-table-pagination__input {
  width: 88px;
}
</style>
