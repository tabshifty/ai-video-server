import { ACTIVE_IMPORT_STATUSES, formatImportFileSize, isCloudLocalFileDeleted } from './cloudDownloadImport.helpers'

// 115 云下载本地暂存删除的纯函数集合。
// 只判断“能否手动删除 DOWNLOAD_DIR 里的文件”，不发起请求。

const TERMINAL_IMPORT_STATUSES = new Set(['imported', 'skipped_existing'])

export function canDeleteCloudDownloadLocalFile(task) {
  if (!task || task.status !== 'completed') return false
  if (isCloudLocalFileDeleted(task)) return false
  if (!String(task?.local_path || '').trim()) return false
  const importStatus = String(task?.import?.status || '')
  if (ACTIVE_IMPORT_STATUSES.includes(importStatus)) return false
  // 终态导入（imported/skipped_existing）已经删除本地源文件，不再提供入口；
  // failed 导入仍允许清理，活动导入由上面的判断拦住。
  if (TERMINAL_IMPORT_STATUSES.has(importStatus)) return false
  return true
}

export function localFileDeletionSizeLabel(preview) {
  if (!preview) return '--'
  if (preview.deleted) return '已删除'
  if (preview.managed === false) return '路径不在下载目录'
  if (!preview.exists) return '文件已不存在'
  return formatImportFileSize(preview.size)
}

export function canConfirmLocalFileDeletion(preview) {
  return !preview || preview.managed !== false
}
