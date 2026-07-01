import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxEd2kDownload.vue', import.meta.url), 'utf8')

describe('ToolboxEd2kDownload', () => {
  it('keeps the download workbench backed by the admin API with hash dedupe and history lockout', () => {
    expect(source).toContain('ED2K 下载工作台')
    expect(source).toContain('任务标题默认取链接里的文件名')
    expect(source).toContain('当前资源已命中历史任务，不允许重建下载任务')
    expect(source).toContain('永久删除')
    expect(source).toContain('新建任务')
    expect(source).toContain('新建下载任务')
    expect(source).toContain('createDialogVisible')
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

  it('keeps the workbench focused on dialog-based task submission, task list, detail, and history sections', () => {
    expect(source).toContain('下载任务')
    expect(source).toContain('任务详情')
    expect(source).toContain('状态反馈')
    expect(source).toContain('任务操作')
    expect(source).toContain('文件区')
    expect(source).toContain('历史记录')
    expect(source).toContain('重试任务')
    expect(source).toContain('刷新任务')
    expect(source).toContain('后端工作台里管理下载任务')
    expect(source).toContain('<el-dialog')
    expect(source).not.toContain('<template #title>提交链接</template>')
  })

  it('keeps the task list as a full-width workspace row instead of a narrow side rail', () => {
    expect(source).toContain('class="task-workspace"')
    expect(source).toContain('class="task-list-card"')
    expect(source).toMatch(/\.task-workspace\s*\{[^}]*display:\s*grid;[^}]*gap:\s*var\(--space-4\);/s)
    expect(source).not.toMatch(/\.task-workspace\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*24rem\)/s)
    expect(source).toMatch(/\.task-row\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+auto;/s)
  })
})
