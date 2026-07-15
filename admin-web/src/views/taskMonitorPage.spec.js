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

const script = extractBlock('script')
const template = extractBlock('template')
const style = extractBlock('style')
const loadBlock = script.match(/async function load\(options = \{\}\) \{[\s\S]*?\n\}(?=\n\nfunction toNumber)/)?.[0] || ''
const statusOptionsBlock = script.match(/const statusOptions = \[[\s\S]*?\n\]/)?.[0] || ''
const summaryMetricsBlock = script.match(/const summaryMetrics = computed\(\(\) => \[[\s\S]*?\n\]\)/)?.[0] || ''
const latestLoadedInFinallyPattern = /finally\s*\{\s*if\s*\(seq === loadSeq\)\s*\{(?=[\s\S]*?loaded\.value\s*=\s*true)(?=[\s\S]*?loading\.value\s*=\s*false)[\s\S]*?\}\s*\}\s*$/
const unguardedLoadedInFinallyPattern = /finally\s*\{\s*loaded\.value\s*=\s*true/
const rowsResetInCatchPattern = /catch\s*\(error\)\s*\{(?:(?!\}\s*finally)[\s\S])*?list\.value\s*=\s*\[\]/

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
  }`
    const unsafeFinally = `finally {
    loaded.value = true
    if (seq === loadSeq) loading.value = false
  }`

    expect(protectedFinally).toMatch(latestLoadedInFinallyPattern)
    expect(unsafeFinally).toMatch(unguardedLoadedInFinallyPattern)
    expect(unsafeFinally).not.toMatch(latestLoadedInFinallyPattern)
  })

  it('只允许最新请求结束首次加载并保留原请求参数与轮询约定', () => {
    expect(loadBlock).toContain('const { skipIfLoading = false } = options')
    expect(loadBlock).toContain('if (skipIfLoading && loading.value)')
    expect(loadBlock).toContain('const seq = ++loadSeq')
    expect(loadBlock).toContain('page: query.page')
    expect(loadBlock).toContain('page_size: query.page_size')
    expect(loadBlock).toContain('if (query.status)')
    expect(loadBlock).toContain('params.status = query.status')
    expect(loadBlock).toMatch(latestLoadedInFinallyPattern)
    expect(loadBlock).not.toMatch(unguardedLoadedInFinallyPattern)
    expect(loadBlock.match(/loaded\.value\s*=\s*true/g)).toHaveLength(1)
    expect(script).toContain('setInterval(() => load({ skipIfLoading: true }), 5000)')
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

  it('保留完整任务列并用文字和语义色共同表达状态', () => {
    for (const label of ['任务', '状态', '进度', '剩余时间', '已耗时', '重试', '错误', '开始时间', '进度更新时间']) {
      expect(template).toContain(`label="${label}"`)
    }
    expect(template).toMatch(/<div class="task-cell">\s*<strong>\{\{ taskTitle\(row\) \}\}<\/strong>\s*<span>任务 ID：\{\{ row\.id \}\} · 视频 ID：\{\{ row\.video_id \|\| '--' \}\}<\/span>\s*<\/div>/)
    expect(template).toContain('<StatusIndicator :label="statusLabel(row.status)" :tone="taskStatusTone(row.status)" />')
    expect(script).toContain("if (status === 'success') return 'success'")
    expect(script).toContain("if (status === 'failed') return 'danger'")
    expect(script).toContain("if (status === 'running') return 'warning'")
    expect(script).toContain("if (status === 'pending') return 'info'")
    expect(template).toContain(':stroke-width="6"')
    expect(template).toContain('<AdminTablePagination')
    expect(template).toContain('v-model:current-page="query.page"')
    expect(template).toContain('v-model:page-size="query.page_size"')
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
    expect(style).not.toMatch(/#[0-9a-f]{3,8}\b/i)
    expect(style).not.toMatch(/rgba?\(/i)
    expect(style).not.toContain('linear-gradient')
    expect(style).not.toContain('letter-spacing')
    expect(style).not.toContain('box-shadow')
  })
})
