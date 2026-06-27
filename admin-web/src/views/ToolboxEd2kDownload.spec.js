import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxEd2kDownload.vue', import.meta.url), 'utf8')

describe('ToolboxEd2kDownload', () => {
  it('keeps the download workbench backed by the admin API with hash dedupe and history lockout', () => {
    expect(source).toContain('ED2K 下载工作台')
    expect(source).toContain('任务标题默认取链接里的文件名')
    expect(source).toContain('当前资源已命中历史任务，不允许重建下载任务')
    expect(source).toContain('永久删除')
    expect(source).toContain('resourceHash')
    expect(source).toContain('titleInput')
    expect(source).toContain('getAdminEd2kDownloadTasks')
    expect(source).toContain('createAdminEd2kDownloadTasks')
    expect(source).toContain('retryAdminEd2kDownloadTask')
    expect(source).toContain('deleteAdminEd2kDownloadTask')
    expect(source).toContain('loadingTasks')
    expect(source).toContain('submitting')
    expect(source).toContain('历史任务存在就直接命中')
    expect(source).toContain('parseEd2kLinks')
    expect(source).toContain('syncTitleFromSingleLink')
    expect(source).not.toContain('components/Layout.vue')
    expect(source).not.toMatch(/<Layout[>\s]/)
  })

  it('keeps the workbench focused on task submission, task list, detail, and history sections', () => {
    expect(source).toContain('提交链接')
    expect(source).toContain('下载任务')
    expect(source).toContain('任务详情')
    expect(source).toContain('状态反馈')
    expect(source).toContain('任务操作')
    expect(source).toContain('文件区')
    expect(source).toContain('历史记录')
    expect(source).toContain('重试任务')
    expect(source).toContain('刷新任务')
    expect(source).toContain('后端工作台里管理下载任务')
  })
})
