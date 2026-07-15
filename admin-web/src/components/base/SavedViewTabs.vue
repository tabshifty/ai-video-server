<script setup>
import { computed } from 'vue'
import { ElMessageBox } from 'element-plus'
import { Delete, EditPen, MoreFilled, Plus, RefreshRight } from '@element-plus/icons-vue'
import { CUSTOM_VIEW_ID } from './savedView.helpers'

const props = defineProps({
  items: { type: Array, required: true },
  activeId: { type: String, required: true },
  editableSourceId: { type: String, default: '' }
})
const emit = defineEmits(['select', 'save', 'update', 'rename', 'remove'])

const visibleItems = computed(() => props.activeId === CUSTOM_VIEW_ID
  ? [...props.items, { id: CUSTOM_VIEW_ID, label: '自定义', builtIn: true, transient: true }]
  : props.items)
const activeItem = computed(() =>
  props.items.find((item) => item.id === props.activeId) || null
)
const editableSource = computed(() =>
  props.items.find((item) => item.id === props.editableSourceId && !item.builtIn) || null
)

function isDismissed(error) {
  return error === 'cancel' || error === 'close'
}

async function requestSave() {
  try {
    const { value } = await ElMessageBox.prompt('请输入视图名称', '保存视图', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: (text) => String(text || '').trim() !== '' || '请输入视图名称'
    })
    emit('save', String(value).trim())
  } catch (error) {
    if (!isDismissed(error)) throw error
  }
}

async function requestRename() {
  if (!activeItem.value || activeItem.value.builtIn) return

  try {
    const { value } = await ElMessageBox.prompt('请输入新的视图名称', '重命名视图', {
      inputValue: activeItem.value.label,
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputValidator: (text) => String(text || '').trim() !== '' || '请输入视图名称'
    })
    emit('rename', { id: activeItem.value.id, label: String(value).trim() })
  } catch (error) {
    if (!isDismissed(error)) throw error
  }
}

async function requestRemove() {
  if (!activeItem.value || activeItem.value.builtIn) return

  try {
    await ElMessageBox.confirm('确认删除这个保存视图？', '删除视图', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
    emit('remove', activeItem.value.id)
  } catch (error) {
    if (!isDismissed(error)) throw error
  }
}
</script>

<template>
  <section class="saved-view-tabs" aria-label="保存视图">
    <el-tabs
      :model-value="activeId"
      class="saved-view-tabs__tabs"
      @tab-change="(id) => emit('select', id)"
    >
      <el-tab-pane
        v-for="item in visibleItems"
        :key="item.id"
        :name="item.id"
        :label="item.label"
      />
    </el-tabs>

    <div class="saved-view-tabs__actions">
      <el-button
        v-if="activeId === CUSTOM_VIEW_ID"
        :icon="Plus"
        @click="requestSave"
      >
        另存为视图
      </el-button>
      <el-button
        v-if="activeId === CUSTOM_VIEW_ID && editableSource"
        :icon="RefreshRight"
        @click="emit('update', editableSource.id)"
      >
        更新视图
      </el-button>
      <el-dropdown v-if="activeItem && !activeItem.builtIn" trigger="click">
        <el-tooltip content="视图操作" placement="top">
          <el-button :icon="MoreFilled" circle aria-label="视图操作" />
        </el-tooltip>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item :icon="EditPen" @click="requestRename">
              重命名
            </el-dropdown-item>
            <el-dropdown-item :icon="Delete" divided @click="requestRemove">
              删除
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </section>
</template>

<style scoped>
.saved-view-tabs {
  display: flex;
  max-width: 100%;
  min-width: 0;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
  border-bottom: 1px solid var(--line-soft);
  letter-spacing: 0;
}

.saved-view-tabs__tabs {
  min-width: 0;
  flex: 1 1 auto;
}

.saved-view-tabs__actions {
  display: inline-flex;
  min-width: 0;
  flex: 0 0 auto;
  flex-wrap: wrap;
  align-items: center;
  gap: var(--space-2);
}

.saved-view-tabs__actions :deep(.el-button) {
  height: 32px;
  min-height: 32px;
  margin-left: 0;
  letter-spacing: 0;
}

.saved-view-tabs__actions :deep(.el-button.is-circle) {
  width: 32px;
  padding: 0;
}

:deep(.el-tabs__header) {
  margin: 0;
}

:deep(.el-tabs__item) {
  max-width: 12rem;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  letter-spacing: 0;
}

@media (max-width: 63.9375rem) {
  .saved-view-tabs {
    align-items: stretch;
    flex-direction: column;
  }

  .saved-view-tabs__tabs {
    overflow-x: auto;
  }

  .saved-view-tabs__actions :deep(.el-button) {
    height: 44px;
    min-height: 44px;
  }

  .saved-view-tabs__actions :deep(.el-button.is-circle) {
    width: 44px;
  }
}
</style>
