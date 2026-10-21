import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import CloudDownloadList from './CloudDownloadList.vue'

const source = readFileSync(new URL('./CloudDownloadList.vue', import.meta.url), 'utf8')
const script = source.match(/<script setup>[\s\S]*?<\/script>/)?.[0] || ''
const template = source.match(/<template>[\s\S]*<\/template>/)?.[0] || ''
const style = source.match(/<style scoped>[\s\S]*?<\/style>/)?.[0] || ''

describe('115 云下载列表页', () => {
  it('通过真实 SFC 编译并复用管理端接口', () => {
    expect(CloudDownloadList).toBeTruthy()
    expect(script).toContain("import AdminTablePagination from '../components/AdminTablePagination.vue'")
    expect(script).toContain("import Layout from '../components/Layout.vue'")
    expect(script).toContain("import CloudDownloadImportDialog from '../components/CloudDownloadImportDialog.vue'")
    expect(script).toContain('listCloudDownloads')
    expect(script).toContain('retryCloudDownload')
    expect(script).toContain('previewCloudDownloadImport')
    expect(script).toContain('getCloudBrowserState')
    expect(script).toContain('getCloudLoginQR')
    expect(script).toContain('extractApiErrorMessage')
    expect(script).toContain('deleteCloudDownloadCloudFile')
    expect(script).toContain('cloudFileDeletionMeta')
    expect(script).toContain('canDeleteCloudDownloadFile')
    expect(script).toContain('hasActiveCloudFileDeletions')
  })

  it('失败和需处理任务仍然可以重试下载', () => {
    expect(template).toContain("resolveCloudImportAction(row) === 'retry-download'")
    expect(template).toContain('@click="retryTask(row)"')
    expect(script).toContain('retryCloudDownload(row.id)')
  })

  it('完成且未导入的任务先 preview，再按类型打开导入弹窗', () => {
    expect(script).toContain('previewCloudDownloadImport(row.id)')
    expect(script).toContain("preview.kind !== 'video' && preview.kind !== 'archive'")
    expect(script).toContain('importDialogVisible.value = true')
    expect(template).toContain('入媒体库')
    expect(template).toContain(':loading="previewingImportId === row.id"')
    expect(template).toContain('<CloudDownloadImportDialog')
    expect(template).toContain(':task="importTask"')
    expect(template).toContain(':preview="importPreview"')
    expect(template).toContain('@submitted="handleImportSubmitted"')
    expect(script).toContain('importTask.value.import = importInfo')
  })

  it('按 import 状态展示导入中、已导入、已存在和重试导入', () => {
    expect(template).toContain('导入中')
    expect(template).toContain('已导入')
    expect(template).toContain('已存在')
    expect(template).toContain('重试导入')
    expect(template).toContain("resolveCloudImportAction(row) === 'importing'")
    expect(template).toContain("resolveCloudImportAction(row) === 'imported'")
    expect(template).toContain("resolveCloudImportAction(row) === 'existing'")
    expect(template).toContain("resolveCloudImportAction(row) === 'retry-import'")
    expect(template).toContain('row.import?.error_message')
    expect(template).toContain('<el-button size="small" disabled>导入中</el-button>')
  })

  it('失败导入通过同一 preview + 弹窗流程重试，而不是无请求体重试', () => {
    const retryImportBlock = template.match(
      /v-else-if="resolveCloudImportAction\(row\) === 'retry-import'"[\s\S]*?<\/template>/
    )?.[0] || ''

    expect(retryImportBlock).toContain('startImport(row)')
    expect(retryImportBlock).toContain('重试导入')
    expect(retryImportBlock).not.toContain('importCloudDownload')
    expect(script).toContain('existingKind !== preview.kind')
  })

  it('活动导入期间继续轮询，并支持跳转到入库结果', () => {
    expect(script).toContain('hasActiveImportTasks')
    expect(script).toContain('hasActiveImport')
    expect(script).toContain('shouldPoll')
    expect(script).toContain('hasActiveTasks.value || hasActiveImport.value')
    expect(script).toContain('导入中，每 5 秒自动刷新')
    expect(script).toContain('importDestination')
    expect(script).toContain('router.push(destination)')
    expect(script).toMatch(/handleImportSubmitted[\s\S]*syncPolling\(\)[\s\S]*await load\(\)/)
  })

  it('已完成任务通过更多菜单移入 115 回收站并显式确认', () => {
    expect(template).toContain('canDeleteCloudFileRow(row)')
    expect(script).toContain('MoreFilled')
    expect(template).toContain('aria-label="更多云下载操作"')
    expect(template).toContain('cloudFileDeletionActionLabel(row)')
    expect(template).toContain('handleCloudFileDeletionCommand')
    expect(template).toContain('移入 115 回收站')
    expect(template).toContain('deleteCloudFileDialogVisible')
    expect(template).toContain('deleteCloudFileError')
    expect(template).toContain('仅影响 115 云盘源文件')
    expect(template).toContain('不会删除本地文件或导入记录')
    expect(script).toContain('deleteCloudDownloadCloudFile(task.id)')
    expect(script).toContain('deleteCloudFileDialogVisible.value = true')
    expect(script).toContain('deleteCloudFileError.value = message')
    expect(script).toMatch(/confirmDeleteCloudFile[\s\S]*extractErrorMessage\(error/)
    expect(template).toContain(':close-on-click-modal="!deletingCloudFileId"')
    expect(template).toContain(':close-on-press-escape="!deletingCloudFileId"')
    expect(script).toContain('blockedCloudFileDeletionIds')
  })

  it('展示清理状态并在失败时提供重试，清理中继续轮询', () => {
    expect(template).toContain('cloudFileDeletionMeta(row)?.detail')
    expect(template).toContain("row.file_deletion?.status === 'deleting'")
    expect(template).toContain('移入中')
    expect(script).toContain('hasActiveCloudFileDeletions')
    expect(script).toContain('hasActiveFileDeletion.value')
    expect(script).toContain('云盘清理中，每 5 秒自动刷新')
  })

  it('操作列加宽，窄屏导入与下载按钮保持 44px 触控目标', () => {
    expect(template).toMatch(/label="操作" width="160"/)
    expect(style).toContain('.cloud-actions')
    expect(style).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*\.cloud-actions :deep\(\.el-button\) \{\s*min-width: 44px;\s*min-height: 44px;\s*\}/)
  })

  it('不新增论坛操作列', () => {
    expect([...template.matchAll(/label="操作"/g)]).toHaveLength(1)
    expect(template).not.toContain('发送到115')
  })

  it('展开行展示失败原因、落地路径与触发信息', () => {
    expect(template).toContain('row.error_message')
    expect(template).toContain('row.local_path')
    expect(template).toContain('row.pick_code')
    expect(template).toContain('row.retry_count')
  })

  it('保留自动刷新、二维码倒计时和响应式布局', () => {
    expect(script).toContain('POLL_INTERVAL_MS')
    expect(script).toContain('syncPolling')
    expect(script).toContain('startQrCountdown')
    expect(style).toContain('@media (max-width: 63.9375rem)')
    expect(style).toContain('.cloud-detail__path')
  })

  it('按 source 显示工具箱、手动或原帖来源', () => {
    expect(template).toContain('label="来源"')
    expect(template).toContain("row.source === 'toolbox'")
    expect(template).toContain("row.source === 'manual'")
    expect(template).toContain("row.post_title || '原帖已删除'")
  })
})
