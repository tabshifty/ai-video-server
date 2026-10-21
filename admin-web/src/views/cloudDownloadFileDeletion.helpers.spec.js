import { describe, expect, it } from 'vitest'
import {
  canDeleteCloudDownloadFile,
  cloudDownloadFileName,
  cloudFileDeletionActionLabel,
  cloudFileDeletionMeta,
  hasActiveCloudFileDeletions
} from './cloudDownloadFileDeletion.helpers'

describe('115 云盘源文件清理 helpers', () => {
  it('只允许已完成且本地文件仍存在的任务删除', () => {
    expect(canDeleteCloudDownloadFile({ status: 'completed', local_path: '/tmp/a.zip' })).toBe(true)
    expect(canDeleteCloudDownloadFile({ status: 'downloading', local_path: '/tmp/a.zip' })).toBe(false)
    expect(canDeleteCloudDownloadFile({ status: 'completed', local_path: '' })).toBe(false)
    expect(canDeleteCloudDownloadFile({ status: 'completed', local_path: '/tmp/a.zip', file_deletion: { status: 'deleting' } })).toBe(false)
    expect(canDeleteCloudDownloadFile({ status: 'completed', local_path: '/tmp/a.zip', file_deletion: { status: 'deleted' } })).toBe(false)
    expect(canDeleteCloudDownloadFile({ status: 'completed', local_path: '/tmp/a.zip', file_deletion: { status: 'failed' } })).toBe(true)
  })

  it('展示移入回收站状态与失败原因', () => {
    expect(cloudFileDeletionMeta({ file_deletion: { status: 'deleting' } })).toMatchObject({ label: '移入回收站中', type: 'warning' })
    expect(cloudFileDeletionMeta({ file_deletion: { status: 'deleted' } })).toMatchObject({ label: '已移入回收站', type: 'success' })
    expect(cloudFileDeletionMeta({ file_deletion: { status: 'failed', error_message: '115 登录已过期' } })).toMatchObject({
      label: '移入失败',
      type: 'danger',
      detail: '115 登录已过期'
    })
    expect(cloudFileDeletionMeta({})).toBeNull()
  })

  it('失败或超时后显示重试动作，并识别活动清理用于轮询', () => {
    expect(cloudFileDeletionActionLabel({ file_deletion: { status: 'failed' } })).toBe('重试移入回收站')
    const stale = new Date(Date.now() - 6 * 60 * 1000).toISOString()
    expect(cloudFileDeletionActionLabel({ file_deletion: { status: 'deleting', updated_at: stale } })).toBe('重试移入回收站')
    expect(cloudFileDeletionMeta({ file_deletion: { status: 'deleting', updated_at: stale } })).toMatchObject({ label: '移入超时', type: 'warning' })
    expect(cloudFileDeletionActionLabel({})).toBe('移入 115 回收站')
    expect(hasActiveCloudFileDeletions([{ file_deletion: { status: 'deleted' } }, { file_deletion: { status: 'deleting' } }])).toBe(true)
    expect(hasActiveCloudFileDeletions([{ file_deletion: { status: 'deleted' } }])).toBe(false)
  })

  it('确认框优先展示 115 文件名，回退到本地文件名', () => {
    expect(cloudDownloadFileName({ declared_name: 'archive.zip', local_path: '/tmp/other.zip' })).toBe('archive.zip')
    expect(cloudDownloadFileName({ local_path: '/tmp/archive.zip' })).toBe('archive.zip')
    expect(cloudDownloadFileName({})).toBe('--')
  })
})
