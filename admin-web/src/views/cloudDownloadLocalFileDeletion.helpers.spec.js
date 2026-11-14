import { describe, expect, it } from 'vitest'
import {
  canConfirmLocalFileDeletion,
  canDeleteCloudDownloadLocalFile,
  localFileDeletionSizeLabel
} from './cloudDownloadLocalFileDeletion.helpers'

describe('115 本地暂存删除 helpers', () => {
  it('只允许已完成、未删除且仍有本地路径的任务', () => {
    expect(canDeleteCloudDownloadLocalFile({ status: 'completed', local_path: '/downloads/a.rar' })).toBe(true)
    expect(canDeleteCloudDownloadLocalFile({ status: 'downloading', local_path: '/downloads/a.rar' })).toBe(false)
    expect(canDeleteCloudDownloadLocalFile({ status: 'failed', local_path: '/downloads/a.rar' })).toBe(false)
    expect(canDeleteCloudDownloadLocalFile({ status: 'completed', local_path: '' })).toBe(false)
    expect(canDeleteCloudDownloadLocalFile({
      status: 'completed',
      local_path: '/downloads/a.rar',
      local_file_deleted_at: '2026-11-14T02:00:00Z'
    })).toBe(false)
  })

  it('活动导入拒绝删除，失败或已结束的导入不阻塞', () => {
    for (const status of ['queued', 'hashing', 'copying', 'creating']) {
      expect(canDeleteCloudDownloadLocalFile({
        status: 'completed',
        local_path: '/downloads/a.rar',
        import: { status }
      })).toBe(false)
    }
    expect(canDeleteCloudDownloadLocalFile({
      status: 'completed',
      local_path: '/downloads/a.rar',
      import: { status: 'failed' }
    })).toBe(true)
    expect(canDeleteCloudDownloadLocalFile({
      status: 'completed',
      local_path: '/downloads/a.rar',
      import: { status: 'skipped_existing' }
    })).toBe(true)
  })

  it('确认框大小标签区分读取中、已删除、路径异常与文件缺失', () => {
    expect(localFileDeletionSizeLabel(null)).toBe('--')
    expect(localFileDeletionSizeLabel({ deleted: true })).toBe('已删除')
    expect(localFileDeletionSizeLabel({ managed: false, exists: false })).toBe('路径不在下载目录')
    expect(localFileDeletionSizeLabel({ managed: true, exists: false })).toBe('文件已不存在')
    expect(localFileDeletionSizeLabel({ managed: true, exists: true, size: 2048 })).toBe('2.0 KB')
  })

  it('非受管路径禁止确认删除', () => {
    expect(canConfirmLocalFileDeletion(null)).toBe(true)
    expect(canConfirmLocalFileDeletion({ managed: true, exists: true })).toBe(true)
    expect(canConfirmLocalFileDeletion({ managed: false, exists: false })).toBe(false)
  })
})
