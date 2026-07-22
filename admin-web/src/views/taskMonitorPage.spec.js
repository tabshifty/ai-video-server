import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import TaskMonitor from './TaskMonitor.vue'

const source = readFileSync(new URL('./TaskMonitor.vue', import.meta.url), 'utf8')

function extractBlock(tag) {
  const opening = source.match(new RegExp(`<${tag}[^>]*>`))

  expect(opening).not.toBeNull()
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const closing = `</${tag}>`
  const end = tag === 'template' ? source.lastIndexOf(closing) : source.indexOf(closing, start)

  expect(end).toBeGreaterThan(start)
  return source.slice(start, end)
}

function findRule(styleSource, selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = styleSource.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`))

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

function extractBalancedBraceBlock(sourceText, openingPattern) {
  const opening = sourceText.match(openingPattern)
  if (!opening || opening.index === undefined) return null

  const openBraceIndex = opening.index + opening[0].lastIndexOf('{')
  let depth = 0

  for (let index = openBraceIndex; index < sourceText.length; index += 1) {
    if (sourceText[index] === '{') depth += 1
    if (sourceText[index] !== '}') continue

    depth -= 1
    if (depth === 0) {
      return {
        body: sourceText.slice(openBraceIndex + 1, index),
        closeBraceIndex: index,
        openingIndex: opening.index
      }
    }
  }

  return null
}

function extractFunctionStatements(sourceText, openingPattern) {
  const block = extractBalancedBraceBlock(sourceText, openingPattern)
  if (!block) return null

  return block.body
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
}

function ownsQueryIdentityReset(sourceText) {
  const resetStatements = extractFunctionStatements(sourceText, /function resetQueryIdentity\(\)\s*\{/)
  const setPageStatements = extractFunctionStatements(sourceText, /function setPage\(page\)\s*\{/)

  return JSON.stringify(resetStatements) === JSON.stringify([
    'list.value = []',
    'total.value = 0',
    'loaded.value = false',
    "loadError.value = ''"
  ]) && JSON.stringify(setPageStatements) === JSON.stringify([
    'query.page = page',
    'resetQueryIdentity()',
    'load()'
  ])
}

function latestFinallyOwnsCompletion(sourceText) {
  const finallyBlock = extractBalancedBraceBlock(sourceText, /finally\s*\{/)
  if (!finallyBlock) return false
  if (!/^\s*\}\s*$/.test(sourceText.slice(finallyBlock.closeBraceIndex + 1))) return false

  const guardBlock = extractBalancedBraceBlock(finallyBlock.body, /if\s*\(seq === loadSeq\)\s*\{/)
  if (!guardBlock) return false

  const outsideGuard = finallyBlock.body.slice(0, guardBlock.openingIndex)
    + finallyBlock.body.slice(guardBlock.closeBraceIndex + 1)

  return outsideGuard.trim() === ''
    && /loaded\.value\s*=\s*true/.test(guardBlock.body)
    && /loading\.value\s*=\s*false/.test(guardBlock.body)
}

const script = extractBlock('script')
const template = extractBlock('template')
const style = extractBlock('style')
const loadBlock = script.match(/async function load\(options = \{\}\) \{[\s\S]*?\n\}(?=\n\nfunction toNumber)/)?.[0] || ''
const statusOptionsBlock = script.match(/const statusOptions = \[[\s\S]*?\n\]/)?.[0] || ''
const summaryMetricsBlock = script.match(/const summaryMetrics = computed\(\(\) => \[[\s\S]*?\n\]\)/)?.[0] || ''
const rowsResetInCatchPattern = /catch\s*\(error\)\s*\{(?:(?!\}\s*finally)[\s\S])*?list\.value\s*=\s*\[\]/
const paginationTag = template.match(/<AdminTablePagination\b[\s\S]*?\/>/)?.[0] || ''

describe('任务监控页', () => {
  it('通过真实 SFC 编译并使用壳层标题操作区', () => {
    expect(TaskMonitor).toBeTruthy()
    expect(script).toContain("import MetricStrip from '../components/base/MetricStrip.vue'")
    expect(script).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(script).not.toContain('PageHeader')
    expect(script).not.toContain('StatCard')
    expect(template).toContain('<template #header-actions>')
    expect(template).toContain('<MetricStrip :items="summaryMetrics" aria-label="任务摘要" />')
    expect(template).not.toContain('<PageHeader')
    expect(template).not.toContain('<StatCard')
  })

  it('移除混合口径成功率并逐项标明摘要范围', () => {
    expect(script).not.toContain('successRate')
    expect(script).not.toContain('successCount')
    expect([...summaryMetricsBlock.matchAll(/label: '([^']+)'/g)].map((match) => match[1])).toEqual([
      '任务总量',
      '排队',
      '处理中',
      '失败'
    ])
    expect(summaryMetricsBlock).toContain("scope: query.status ? '当前筛选·全部页' : '全局'")
    expect(summaryMetricsBlock).toContain("{ key: 'queued', label: '排队', value: queuedCount.value, scope: '本页', tone: 'info' }")
    expect(summaryMetricsBlock).toContain("{ key: 'running', label: '处理中', value: runningCount.value, scope: '本页', tone: 'warning' }")
    expect(summaryMetricsBlock).toContain("{ key: 'failed', label: '失败', value: failedCount.value, scope: '本页', tone: 'danger' }")
  })

  it('状态机门禁能区分最新响应与未受保护的 loaded 更新', () => {
    const protectedFinally = `finally {
    if (seq === loadSeq) {
      loaded.value = true
      loading.value = false
    }
  }
}`
    const unsafeFinally = `finally {
    loaded.value = true
    if (seq === loadSeq) {
      loading.value = false
    }
  }
}`
    const escapedGuardFinally = `finally {
    if (seq === loadSeq) {
      loading.value = false
    }
    loaded.value = true
  }
}`

    expect(latestFinallyOwnsCompletion(protectedFinally)).toBe(true)
    expect(latestFinallyOwnsCompletion(unsafeFinally)).toBe(false)
    expect(latestFinallyOwnsCompletion(escapedGuardFinally)).toBe(false)
  })

  it('只允许最新请求结束首次加载并保留原请求参数与轮询约定', () => {
    expect(loadBlock).toContain('const { skipIfLoading = false } = options')
    expect(loadBlock).toContain('if (skipIfLoading && loading.value)')
    expect(loadBlock).toContain('const seq = ++loadSeq')
    expect(loadBlock).toContain('page: query.page')
    expect(loadBlock).toContain('page_size: query.page_size')
    expect(loadBlock).toContain('if (query.status)')
    expect(loadBlock).toContain('params.status = query.status')
    expect(latestFinallyOwnsCompletion(loadBlock)).toBe(true)
    expect(loadBlock.match(/loaded\.value\s*=\s*true/g)).toHaveLength(1)
    expect(script).toContain('setInterval(() => load({ skipIfLoading: true }), 5000)')
  })

  it('只在最新请求成功后清除持久错误', () => {
    const requestIndex = loadBlock.indexOf('const data = await getAdminTasks(params)')
    const staleGuard = extractBalancedBraceBlock(loadBlock.slice(requestIndex), /if\s*\(seq !== loadSeq\)\s*\{/)
    const clearErrorIndex = loadBlock.indexOf("loadError.value = ''")
    const rowsIndex = loadBlock.indexOf('list.value = data.items || []')

    expect(requestIndex).toBeGreaterThan(-1)
    expect(staleGuard).not.toBeNull()
    expect(clearErrorIndex).toBeGreaterThan(requestIndex + (staleGuard?.closeBraceIndex || 0))
    expect(clearErrorIndex).toBeLessThan(rowsIndex)
    expect(loadBlock.match(/loadError\.value\s*=\s*''/g)).toHaveLength(1)
  })

  it('失败分支门禁能识别清空已有任务的违规实现', () => {
    const violatingLoad = `async function load() {
  try {
    list.value = await getAdminTasks()
  } catch (error) {
    loadError.value = error.message
    list.value = []
  } finally {
    loading.value = false
  }
}`

    expect(violatingLoad).toMatch(rowsResetInCatchPattern)
    expect(loadBlock).not.toMatch(rowsResetInCatchPattern)
  })

  it('后台刷新保留已有行并让首次失败保持为错误态', () => {
    expect(script).toContain('const loaded = ref(false)')
    expect(script).toContain("const loadError = ref('')")
    expect(script).toContain('const initialLoading = computed(() => loading.value && !loaded.value)')
    expect(script).toContain('const backgroundRefreshing = computed(() => loading.value && loaded.value)')
    expect(loadBlock).toContain("loadError.value = ''")
    expect(loadBlock).toContain("loadError.value = error?.message || '加载任务失败'")
    expect(script).not.toContain('ElMessage')
    expect(template).not.toContain('<el-table v-loading="loading"')
    expect(template).toContain('<el-alert v-if="loadError"')
    expect(template).toContain('<el-skeleton v-if="initialLoading"')
    expect(template).toContain('<SectionCard v-else-if="!loadError || list.length > 0" dense>')
    expect(template).toContain('@click="load">重试</el-button>')
  })

  it('提供全部五项状态筛选和诚实的空状态', () => {
    expect([...statusOptionsBlock.matchAll(/\{ label: '([^']+)', value: '([^']*)' \}/g)].map((match) => match.slice(1))).toEqual([
      ['全部', ''],
      ['排队', 'pending'],
      ['处理中', 'running'],
      ['已完成', 'success'],
      ['失败', 'failed']
    ])
    expect(script).toContain("const hasStatusFilter = computed(() => query.status !== '')")
    expect(template).toContain('<el-segmented')
    expect(template).toContain(':options="statusOptions"')
    expect(template).toContain("hasStatusFilter ? '当前筛选无结果' : '暂无任务'")
    expect(template).toContain("hasStatusFilter ? '清除状态筛选后查看全部任务' : '任务创建后会显示在这里'")
    expect(template).toContain("@click=\"setStatus('')\"")
  })

  it('切换筛选时清空旧查询身份并重新进入首次加载', () => {
    expect(extractFunctionStatements(script, /function setStatus\(status\)\s*\{/)).toEqual([
      'query.status = status',
      'query.page = 1',
      'resetQueryIdentity()',
      'load()'
    ])
  })

  it('分页查询身份门禁能拒绝直接加载与错误重置顺序', () => {
    const validFixture = `function resetQueryIdentity() {
  list.value = []
  total.value = 0
  loaded.value = false
  loadError.value = ''
}

function setPage(page) {
  query.page = page
  resetQueryIdentity()
  load()
}`
    const directLoadFixture = `function setPage(page) {
  query.page = page
  load()
}`
    const lateResetFixture = `${validFixture.replace('resetQueryIdentity()\n  load()', 'load()\n  resetQueryIdentity()')}`

    expect(ownsQueryIdentityReset(validFixture)).toBe(true)
    expect(ownsQueryIdentityReset(directLoadFixture)).toBe(false)
    expect(ownsQueryIdentityReset(lateResetFixture)).toBe(false)
  })

  it('分页切换使用只读参数并在请求前接管新查询身份', () => {
    expect(paginationTag).not.toBe('')
    expect(paginationTag).toContain(':current-page="query.page"')
    expect(paginationTag).toContain(':page-size="query.page_size"')
    expect(paginationTag).toContain('@current-change="setPage"')
    expect(paginationTag).not.toContain('v-model:current-page')
    expect(paginationTag).not.toContain('v-model:page-size')
    expect(ownsQueryIdentityReset(script)).toBe(true)
  })

  it('保留完整任务列并用文字和语义色共同表达状态', () => {
    for (const label of ['任务', '状态', '进度', '剩余时间', '已耗时', '重试', '错误', '开始时间', '进度更新时间']) {
      expect(template).toContain(`label="${label}"`)
    }
    expect(template).toContain('label="任务" min-width="210"')
    expect(template).toContain('label="状态" width="96"')
    expect(template).toContain('label="进度" min-width="150"')
    expect(template).toContain('label="剩余时间" width="88"')
    expect(template).toContain('label="已耗时" width="88"')
    expect(template).toContain('prop="retry_count" label="重试" width="60"')
    expect(template).toContain('label="错误" min-width="160"')
    expect(template).toContain('label="开始时间" width="145"')
    expect(template).toContain('label="进度更新时间" width="145"')
    expect(template).toMatch(/<div class="task-cell">\s*<strong>\{\{ taskTitle\(row\) \}\}<\/strong>\s*<span>任务 ID：\{\{ row\.id \}\} · 视频 ID：\{\{ row\.video_id \|\| '--' \}\}<\/span>\s*<\/div>/)
    expect(template).toContain('<StatusIndicator :label="statusLabel(row.status)" :tone="taskStatusTone(row.status)" />')
    expect(script).toContain("if (status === 'success') return 'success'")
    expect(script).toContain("if (status === 'failed') return 'danger'")
    expect(script).toContain("if (status === 'running') return 'warning'")
    expect(script).toContain("if (status === 'pending') return 'info'")
    expect(template).toContain(':stroke-width="6"')
    expect(template).toContain('<AdminTablePagination')
  })

  it('单列轨道允许窄屏分段筛选在自身横滚而不撑宽页面', () => {
    expect(findRule(style, '.task-monitor-page')).toContain('grid-template-columns: minmax(0, 1fr)')
  })

  it('错误全文可访问且监控密度和窄屏点击目标稳定', () => {
    expect(template).toContain('data-density="monitor"')
    expect(template).toContain('<el-tooltip :content="row.error || \'无错误\'"')
    expect(template).toContain('class="task-error"')
    expect(template).toContain('tabindex="0"')
    expect(template).toContain(':aria-label="row.error || \'无错误\'"')
    expect(findRule(style, '.task-error')).toMatch(/display:\s*block/)
    expect(findRule(style, '.task-error')).toMatch(/overflow:\s*hidden/)
    expect(findRule(style, '.task-error')).toMatch(/text-overflow:\s*ellipsis/)
    expect(findRule(style, '.task-error')).toMatch(/white-space:\s*nowrap/)
    expect(findRule(style, '.task-error:focus-visible')).toMatch(/outline:\s*2px solid var\(--line-focus\)/)

    const mobileStart = style.indexOf('@media (max-width: 63.9375rem)')
    expect(mobileStart).toBeGreaterThan(-1)
    const mobileStyle = style.slice(mobileStart)
    expect(mobileStyle).toMatch(/\.task-refresh\s*\{[^}]*min-height:\s*44px/)
    expect(mobileStyle).toMatch(/:deep\(\.el-segmented__item\)\s*\{[^}]*min-height:\s*44px/)
    expect(findRule(mobileStyle, '.status-filter')).toContain('overflow-x: auto')
    expect(findRule(mobileStyle, '.status-filter')).toContain('max-width: 100%')
    expect(findRule(mobileStyle, '.status-filter :deep(.el-segmented__group)')).toContain('width: max-content')
    expect(findRule(mobileStyle, '.status-filter :deep(.el-segmented__item-label)')).toContain('white-space: nowrap')
    expect(findRule(mobileStyle, '.status-filter :deep(.el-segmented__item-label)')).toContain('text-overflow: clip')
    expect(style).not.toMatch(/#[0-9a-f]{3,8}\b/i)
    expect(style).not.toMatch(/rgba?\(/i)
    expect(style).not.toContain('linear-gradient')
    expect(style).not.toContain('letter-spacing')
    expect(style).not.toContain('box-shadow')
  })

  it('错误 Tooltip 触发器在窄屏保持 44px 高度并垂直对齐单行文字', () => {
    const errorTrigger = template.match(/<span\b(?=[^>]*class="task-error")[^>]*>/)?.[0] || ''
    const mobileStart = style.indexOf('@media (max-width: 63.9375rem)')

    expect(errorTrigger).toContain('tabindex="0"')
    expect(errorTrigger).toContain(':aria-label="row.error || \'无错误\'"')
    expect(findRule(style, '.task-error')).toMatch(/display:\s*block/)
    expect(findRule(style, '.task-error')).toMatch(/text-overflow:\s*ellipsis/)
    expect(findRule(style, '.task-error')).toMatch(/white-space:\s*nowrap/)
    expect(mobileStart).toBeGreaterThan(-1)
    expect(findRule(style.slice(mobileStart), '.task-error')).toMatch(/min-height:\s*44px/)
    expect(findRule(style.slice(mobileStart), '.task-error')).toMatch(/align-content:\s*center/)
  })
})
