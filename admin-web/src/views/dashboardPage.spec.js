import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import Dashboard from './Dashboard.vue'

const source = readFileSync(new URL('./Dashboard.vue', import.meta.url), 'utf8')

function extractBlock(tag) {
  const opening = source.match(new RegExp(`<${tag}[^>]*>`))

  expect(opening).not.toBeNull()
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const closing = `</${tag}>`
  const end = tag === 'template' ? source.lastIndexOf(closing) : source.indexOf(closing, start)

  expect(end).toBeGreaterThan(start)
  return source.slice(start, end)
}

function findRule(style, selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = style.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`))

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

const script = extractBlock('script')
const template = extractBlock('template')
const style = extractBlock('style')

describe('Precision Ops dashboard', () => {
  it('通过真实 SFC 编译并使用壳层标题操作区', () => {
    expect(Dashboard).toBeTruthy()
    expect(template).toContain('<template #header-actions>')
    expect(template).toContain('<MetricStrip')
    expect(script).toContain("import MetricStrip from '../components/base/MetricStrip.vue'")
    expect(script).not.toContain('PageHeader')
    expect(script).not.toContain('StatCard')
    expect(template).not.toContain('<PageHeader')
    expect(template).not.toContain('<StatCard')
  })

  it('以语义导航呈现严格的快捷入口', () => {
    expect(template).toContain('<nav class="dashboard-quick-actions" aria-label="常用入口">')
    expect(template).toContain('<RouterLink')
    expect(template).toContain('v-for="item in dashboardQuickActions"')
    expect(template).toContain(':to="item.path"')
    expect(template).toContain('{{ item.label }}')
  })

  it('保留加载骨架、行内错误重试与空趋势状态', () => {
    expect(template).toContain('v-if="errorMessage"')
    expect(template).toContain('@click="load">重试</el-button>')
    expect(template).toContain('v-if="loading && !stats"')
    expect(template).toContain('v-else-if="stats"')
    expect(template).toContain('v-if="trendPoints.length"')
    expect(template).toContain('title="暂无趋势数据"')
    expect(template).toContain('description="后端暂未返回最近 7 天上传趋势"')
  })

  it('刷新失败时保留已有 stats 并保留图表生命周期', () => {
    expect(script).toContain('const nextStats = await getAdminStats()')
    expect(script).toContain('stats.value = nextStats')
    expect(script).not.toMatch(/catch \(error\) \{[\\s\\S]*?stats\.value\s*=\s*null/)
    expect(script).toContain("window.addEventListener('resize', handleResize)")
    expect(script).toContain("window.removeEventListener('resize', handleResize)")
    expect(script).toContain('chart?.dispose()')
  })

  it('使用主列与 240px 快捷入口，并在小于 1024px 时改为单列', () => {
    expect(findRule(style, '.dashboard-workspace')).toContain(
      'grid-template-columns: minmax(0, 1fr) 240px'
    )
    expect(findRule(style, '.trend-chart')).toContain('height: 240px')
    expect(style).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.dashboard-workspace\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/)
    expect(style).not.toMatch(/#[0-9a-f]{3,8}\b/i)
    expect(style).not.toMatch(/rgba?\(/i)
    expect(style).not.toContain('linear-gradient')
    expect(style).not.toContain('letter-spacing')
    expect(style).not.toContain('box-shadow')
  })

  it('保持窄屏快捷入口可点击且长文本不溢出', () => {
    const mediaStart = style.indexOf('@media (max-width: 63.9375rem)')

    expect(mediaStart).toBeGreaterThan(-1)
    const mobileStyle = style.slice(mediaStart)
    expect(mobileStyle).toMatch(/\.dashboard-quick-actions a\s*\{[^}]*min-height:\s*44px/)
    expect(findRule(style, '.dashboard-quick-actions a')).toContain('overflow-wrap: anywhere')
  })
})
