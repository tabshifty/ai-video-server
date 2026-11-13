import { ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { cleanupAdminEmptyCollections, previewAdminEmptyCollections } from '../api/admin'

export function useEmptyCollectionCleanup(kind, onCompleted) {
  const busy = ref(false)

  async function run() {
    if (busy.value) return
    busy.value = true
    let executing = false
    try {
      const preview = await previewAdminEmptyCollections(kind)
      if (preview.candidate_count === 0) {
        ElMessage.info('没有可清理的空合集，有成员或业务引用的合集会保留')
        return
      }
      const label = kind === 'image' ? '图片' : '视频'
      await ElMessageBox.confirm(
        `将永久删除 ${preview.candidate_count} 个空${label}合集，包含启用和停用合集，不受当前搜索、筛选和分页限制。有成员或业务引用的合集会保留，不会删除图片、视频或媒体文件。此操作不可恢复，是否继续？`,
        '清理空合集',
        { type: 'warning', confirmButtonText: '永久删除', cancelButtonText: '取消', closeOnClickModal: false }
      )
      executing = true
      const result = await cleanupAdminEmptyCollections(kind, preview.preview_token)
      ElMessage.success(`已清理 ${result.deleted_count} 个空合集，跳过 ${result.skipped_count} 个合集`)
      onCompleted()
    } catch (error) {
      if (error === 'cancel' || error === 'close') return
      ElMessage.error(executing
        ? '清理结果未确认，请刷新列表后重新预览'
        : (error?.message || '预览空合集失败，请稍后重试'))
    } finally {
      busy.value = false
    }
  }

  return { busy, run }
}
