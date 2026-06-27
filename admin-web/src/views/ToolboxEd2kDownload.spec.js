import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxEd2kDownload.vue', import.meta.url), 'utf8')

describe('ToolboxEd2kDownload', () => {
  it('keeps the download workbench as a local task manager with hash dedupe and history lockout', () => {
    expect(source).toContain('ED2K 下载工作台')
    expect(source).toContain('任务标题默认取链接里的文件名')
    expect(source).toContain('当前资源已命中历史任务，不允许重建下载任务')
    expect(source).toContain('永久删除')
    expect(source).toContain('resourceHash')
    expect(source).toContain('titleInput')
    expect(source).toContain('normalizeTaskFromLink')
    expect(source).toContain('findTaskByHash')
    expect(source).toContain('markTaskDeleted')
    expect(source).toContain('历史任务存在就直接命中')
    expect(source).toContain('localStorage')
    expect(source).toContain('parseEd2kLinks')
    expect(source).toContain('getEd2kLinkLabel')
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
    expect(source).toContain('重置')
    expect(source).toContain('切到下载中')
    expect(source).toContain('刷新本地记录')
  })
})
