import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxEd2kDownload.vue', import.meta.url), 'utf8')

function extractTemplate(sfcSource) {
  const opening = sfcSource.match(/<template[^>]*>/)

  expect(opening).not.toBeNull()
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const end = sfcSource.lastIndexOf('</template>')

  expect(end).toBeGreaterThan(start)
  return sfcSource.slice(start, end)
}

function extractStyle(sfcSource) {
  const match = sfcSource.match(/<style scoped>([\s\S]*?)<\/style>/)

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

function directPixelRadiusValues(style) {
  return [...style.matchAll(/(?:^|[;{])\s*border-radius\s*:\s*([^;}]+)/gim)]
    .flatMap((declaration) => [...declaration[1].matchAll(/([+-]?(?:\d+(?:\.\d+)?|\.\d+))px\b/gi)])
    .map((match) => Number(match[1]))
}

function elementBlock(content, tagName, openingPattern) {
  const opening = content.match(openingPattern)

  expect(opening, openingPattern.toString()).not.toBeNull()
  const start = opening?.index || 0
  const tags = new RegExp(`<\\/?${tagName}\\b[^>]*>`, 'g')
  tags.lastIndex = start
  let depth = 0
  let match

  while ((match = tags.exec(content))) {
    if (match[0].startsWith('</')) {
      depth -= 1
      if (depth === 0) return content.slice(start, tags.lastIndex)
    } else if (!match[0].endsWith('/>')) {
      depth += 1
    }
  }

  throw new Error(`未找到 ${tagName} 的闭合标签`)
}

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
    expect(source).toContain('createSessionResults')
    expect(source).toContain('getAdminEd2kDownloadTasks')
    expect(source).toContain('getAdminEd2kDownloadTask')
    expect(source).toContain('getAdminEd2kDownloadStatus')
    expect(source).toContain('createAdminEd2kDownloadTasks')
    expect(source).toContain('deleteAdminEd2kDownloadTask')
    expect(source).toContain('loadingTasks')
    expect(source).toContain('loadingDownloadStatus')
    expect(source).toContain('submitting')
    expect(source).toContain('历史任务存在就直接命中')
    expect(source).toContain('parseEd2kLinks')
    expect(source).toContain('buildPendingEd2kInput')
    expect(source).toContain('parseEd2kCreateEntries')
    expect(source).toContain('mergeEd2kDraftEntries')
    expect(source).toContain('getEd2kCreateResultMeta')
    expect(source).toContain('handleCreateDialogBeforeClose')
    expect(source).toContain('scheduleNextPoll')
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
    expect(source).toContain('刷新任务')
    expect(source).toContain('后端工作台里管理下载任务')
    expect(source).toContain('删除暂存文件')
    expect(source).toContain('重新下载')
    expect(source).toContain('重试清理')
    expect(source).toContain('取消任务')
    expect(source).toContain('继续取消')
    expect(source).toContain("['queued', 'running', 'canceling', 'cancelled', 'failed']")
    expect(source).toContain('删除已取消任务')
    expect(source).toContain('预期文件信息')
    expect(source).toContain('文件清理时间')
    expect(source).toContain('残留清理完成时间')
    expect(source).toContain('<el-dialog')
    expect(source).toContain('结果区')
    expect(source).toContain('定位任务')
    expect(source).toContain('focusTask(item.task)')
    expect(source).toContain('下载引擎正常')
    expect(source).toContain('自动刷新：')
    expect(source).toContain('列表更新于：')
    expect(source).toContain('引擎检查于：')
    expect(source).toContain('详情已保留')
    expect(source).toContain('最近更新时间')
    expect(source).toContain('开始执行时间')
    expect(source).toContain('最近事件时间')
    expect(source).toContain('getEd2kCreateResultMeta(item.status).label')
    expect(source).not.toContain("{ value: 'deleted', label: '已删除' }")
    expect(source).not.toContain('<template #title>提交链接</template>')
  })

  it('keeps the primary task commands and history in their existing work areas', () => {
    const template = extractTemplate(source)
    const pageHeader = elementBlock(template, 'PageHeader', /<PageHeader\b[^>]*>/)
    const taskWorkspace = elementBlock(template, 'section', /<section class="task-workspace">/)
    const detailCard = elementBlock(taskWorkspace, 'SectionCard', /<SectionCard v-if="selectedTask"[^>]*>/)
    const dialog = elementBlock(template, 'el-dialog', /<el-dialog\b[^>]*v-model="createDialogVisible"[^>]*>/)

    expect(pageHeader).toContain('@click="openCreateDialog"')
    expect(pageHeader).toContain('@click="manualRefresh"')
    expect(detailCard).toContain('@click="deleteTask(selectedTask)"')
    expect(detailCard).toContain('@click="cleanTaskFiles(selectedTask)"')
    expect(detailCard).toContain('@click="retryTask(selectedTask)"')
    expect(detailCard).toContain('@click="retryCleanup(selectedTask)"')
    expect(detailCard).toContain('<h3 id="task-history-title" class="detail-section__title">历史记录</h3>')
    expect(dialog).toContain('@click="submitLinks"')
    expect(dialog).toContain('@click="focusTask(item.task)"')
    expect(dialog).toContain('@click="clearComposer"')
    expect(template).toContain('@click="returnToToolbox"')
  })

  it('uses semantic status indicators in every existing task status surface', () => {
    const template = extractTemplate(source)
    const pageHeader = elementBlock(template, 'PageHeader', /<PageHeader\b[^>]*>/)
    const engineStatus = elementBlock(template, 'section', /<section class="workbench-status"[^>]*>/)
    const taskButton = elementBlock(template, 'button', /<button\s+v-for="task in visibleTasks"[\s\S]*?>/)
    const detailCard = elementBlock(template, 'SectionCard', /<SectionCard v-if="selectedTask"[^>]*>/)
    const resultCard = elementBlock(template, 'SectionCard', /<SectionCard v-if="createSessionResults\.length > 0"[^>]*>/)

    expect(source).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(pageHeader).toContain('<StatusIndicator :label="selectedTaskLabel" :tone="selectedTaskTone" />')
    expect(engineStatus).toContain('<StatusIndicator :label="engineLevelLabel" :tone="engineTone" />')
    expect(taskButton).toContain(':label="taskStatusLabelMap[task.status] || task.status"')
    expect(taskButton).toContain(":tone=\"taskStatusToneMap[task.status] || 'info'\"")
    expect(detailCard).toContain('<StatusIndicator :label="selectedTaskLabel" :tone="selectedTaskTone" />')
    expect(resultCard).toContain(':label="getEd2kCreateResultMeta(item.status).label"')
    expect(resultCard).toContain(':tone="getEd2kCreateResultMeta(item.status).tone"')
  })

  it('exposes the active status filter with pressed state and a non-color check marker', () => {
    const template = extractTemplate(source)
    const taskWorkspace = elementBlock(template, 'section', /<section class="task-workspace">/)
    const listCard = elementBlock(taskWorkspace, 'SectionCard', /<SectionCard class="task-list-card"[^>]*>/)
    const statusFilters = elementBlock(listCard, 'div', /<div class="status-filters">/)
    const filterButton = elementBlock(statusFilters, 'el-button', /<el-button\s+v-for="option in statusOptions"[\s\S]*?>/)

    expect.soft(source).toContain("import { Back, Check, Delete, Download, Plus, RefreshRight } from '@element-plus/icons-vue'")
    expect.soft(filterButton).toContain(":type=\"currentFilter === option.value ? 'primary' : ''\"")
    expect.soft(filterButton).toContain(':aria-pressed="currentFilter === option.value"')
    expect.soft(filterButton).toContain('@click="setFilter(option.value)"')
    expect.soft(filterButton).toContain('<el-icon v-if="currentFilter === option.value" aria-hidden="true">')
    expect.soft(filterButton.match(/<Check\s*\/>/g) || []).toHaveLength(1)
  })

  it('keeps native task selection accurate, keyboard-visible, and fixed to compact media rows', () => {
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const taskButton = elementBlock(template, 'button', /<button\s+v-for="task in visibleTasks"[\s\S]*?>/)
    const rowRule = style.match(/\.task-row\s*\{[^}]*\}/s)?.[0] || ''
    const headRule = style.match(/\.task-row__head\s*\{[^}]*\}/s)?.[0] || ''
    const textRule = style.match(/\.task-row__head strong,\s*\.task-row__sub\s*\{[^}]*\}/s)?.[0] || ''
    const focusRule = style.match(/\.task-row:focus-visible\s*\{[^}]*\}/s)?.[0] || ''

    expect.soft(taskButton).toMatch(/^<button\b/)
    expect.soft(taskButton).toContain('type="button"')
    expect.soft(taskButton).toContain(':aria-pressed="selectedTask?.id === task.id"')
    expect.soft(taskButton).toContain('@click="selectTask(task)"')
    expect.soft(rowRule).toContain('height: var(--media-row-height);')
    expect.soft(rowRule).toContain('padding: var(--space-1) var(--space-3);')
    expect.soft(rowRule).toContain('font: inherit;')
    expect.soft(headRule).toContain('display: grid;')
    expect.soft(textRule).toContain('overflow: hidden;')
    expect.soft(textRule).toContain('text-overflow: ellipsis;')
    expect.soft(textRule).toContain('white-space: nowrap;')
    expect.soft(textRule).toContain('line-height: var(--leading-small);')
    expect.soft(focusRule).toContain('outline: 2px solid var(--line-focus);')
    expect.soft(focusRule).toContain('outline-offset: -2px;')
  })

  it('keeps the task list as a full-width workspace row instead of a narrow side rail', () => {
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const taskWorkspace = elementBlock(template, 'section', /<section class="task-workspace">/)
    const listCard = elementBlock(taskWorkspace, 'SectionCard', /<SectionCard class="task-list-card"[^>]*>/)
    const detailCard = elementBlock(taskWorkspace, 'SectionCard', /<SectionCard v-if="selectedTask"[^>]*>/)
    const taskWorkspaceRule = style.match(/\.task-workspace\s*\{[^}]*\}/s)?.[0] || ''

    expect(listCard).toMatch(/^<SectionCard class="task-list-card" data-density="compact">/)
    expect(taskWorkspace.indexOf(listCard)).toBeLessThan(taskWorkspace.indexOf(detailCard))
    expect(taskWorkspace.match(/<SectionCard\b/g)).toHaveLength(2)
    expect(detailCard.match(/<SectionCard\b/g)).toHaveLength(1)
    expect(taskWorkspaceRule).toContain('display: grid;')
    expect(taskWorkspaceRule).toContain('grid-template-columns: minmax(0, 1fr);')
    expect(taskWorkspaceRule).toContain('gap: var(--space-4);')
    expect(style).toMatch(/\.task-row\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+auto;/s)
    for (const [id, title] of [
      ['expected-file-title', '预期文件信息'],
      ['status-feedback-title', '状态反馈'],
      ['task-actions-title', '任务操作'],
      ['task-files-title', '文件区'],
      ['task-history-title', '历史记录']
    ]) {
      expect(detailCard).toContain(`<section class="detail-section`)
      expect(detailCard).toContain(`aria-labelledby="${id}"`)
      expect(detailCard).toContain(`<h3 id="${id}" class="detail-section__title">${title}</h3>`)
    }
    expect(detailCard).toContain('<section class="detail-section file-section" data-density="compact" aria-labelledby="task-files-title">')
  })

  it('gives the teleported create dialog form density and an accurate textarea name', () => {
    const template = extractTemplate(source)
    const dialogOpening = template.match(/<el-dialog\b[^>]*v-model="createDialogVisible"[\s\S]*?>/)?.[0] || ''
    const textarea = template.match(/<el-input\b[^>]*v-model="ed2kInput"[\s\S]*?\/>/)?.[0] || ''

    expect(dialogOpening).toContain('data-density="form"')
    expect(textarea).toContain('type="textarea"')
    expect(textarea).toContain('aria-label="ED2K 下载链接"')
  })

  it('contains long content locally and keeps narrow-screen actions touchable and wrapping', () => {
    const style = extractStyle(source)
    const rootRule = style.match(/\.tool-workspace\s*\{[^}]*\}/s)?.[0] || ''
    const innerRule = style.match(/\.tool-workspace__inner\s*\{[^}]*\}/s)?.[0] || ''

    expect.soft(rootRule).toContain('min-width: 0;')
    expect.soft(rootRule).toContain('overflow-x: clip;')
    expect.soft(innerRule).toContain('min-width: 0;')
    expect.soft(style).toMatch(
      /@media \(max-width: 63\.9375rem\)[\s\S]*?\.tool-workspace :deep\(\.page-header-shell__actions\),\s*\.task-list-card :deep\(\.section-card__actions\)\s*\{[^}]*width:\s*100%;[^}]*flex-wrap:\s*wrap;/s
    )
    expect.soft(style).toMatch(
      /@media \(max-width: 63\.9375rem\)[\s\S]*?\.tool-workspace :deep\(\.el-button\),\s*\.crud-dialog :deep\(\.el-button\)\s*\{[^}]*min-height:\s*44px;/s
    )
    expect.soft(style).toMatch(
      /@media \(max-width: 63\.9375rem\)[\s\S]*?\.status-filters\s*\{[^}]*width:\s*100%;[^}]*flex-wrap:\s*wrap;/s
    )
    expect(style).not.toMatch(/#[0-9a-fA-F]{3,8}\b/)
    expect(style).not.toMatch(/rgba?\(\s*\d/)
    expect(style).not.toMatch(/(?:linear|radial)-gradient\(/)
    expect(directPixelRadiusValues('.ok { border-radius: 8px; } .nine { border-radius: 9px; } .multi { border-radius: 8px 9.25px / 4px; } .decimal { border-radius: .5px; }')).toEqual([8, 9, 8, 9.25, 4, 0.5])
    expect(directPixelRadiusValues(style).filter((value) => value > 8)).toEqual([])
    expect(style).not.toContain('box-shadow:')
  })
})
