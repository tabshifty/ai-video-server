import { existsSync, readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import MetricStrip from './MetricStrip.vue'
import SectionCard from './SectionCard.vue'
import SavedViewTabs from './SavedViewTabs.vue'
import StatusIndicator from './StatusIndicator.vue'

const metricStripSource = readFileSync(new URL('./MetricStrip.vue', import.meta.url), 'utf8')
const sectionCardSource = readFileSync(new URL('./SectionCard.vue', import.meta.url), 'utf8')
const savedViewTabsSource = readFileSync(new URL('./SavedViewTabs.vue', import.meta.url), 'utf8')
const statusIndicatorSource = readFileSync(new URL('./StatusIndicator.vue', import.meta.url), 'utf8')
const drawerHeaderURL = new URL('./AdminDrawerHeader.vue', import.meta.url)
const semanticTones = ['neutral', 'success', 'warning', 'danger', 'info']

function extractBlock(source, tag) {
  const match = source.match(new RegExp(`<${tag}[^>]*>([\\s\\S]*?)</${tag}>`))

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

function getValidator(component, propName) {
  const validator = component.props?.[propName]?.validator

  expect(validator).toBeTypeOf('function')
  return validator
}

function findRule(style, selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = style.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`))

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

function extractBraceBlock(source, marker) {
  const start = source.indexOf(marker)
  const openingBrace = source.indexOf('{', start)

  expect(start, marker).toBeGreaterThanOrEqual(0)
  expect(openingBrace, marker).toBeGreaterThan(start)
  let depth = 0
  for (let index = openingBrace; index < source.length; index += 1) {
    if (source[index] === '{') depth += 1
    if (source[index] === '}') depth -= 1
    if (depth === 0) return source.slice(start, index + 1)
  }
  throw new Error(`未找到完整代码块：${marker}`)
}

const metricTemplate = extractBlock(metricStripSource, 'template')
const metricStyle = extractBlock(metricStripSource, 'style')
const sectionCardScript = extractBlock(sectionCardSource, 'script')
const sectionCardTemplate = extractBlock(sectionCardSource, 'template')
const sectionCardStyle = extractBlock(sectionCardSource, 'style')
const savedViewScript = extractBlock(savedViewTabsSource, 'script')
const savedViewTemplate = extractBlock(savedViewTabsSource, 'template')
const savedViewStyle = extractBlock(savedViewTabsSource, 'style')
const statusTemplate = extractBlock(statusIndicatorSource, 'template')
const statusStyle = extractBlock(statusIndicatorSource, 'style')

describe('Precision Ops base components', () => {
  it('compiles both SFCs and renders labeled tabular metrics without cards', () => {
    expect(MetricStrip).toBeTruthy()
    expect(StatusIndicator).toBeTruthy()
    expect(metricTemplate).toContain('class="metric-strip"')
    expect(metricTemplate).toContain('tabular-num')
    expect(metricTemplate).toContain('v-if="item.scope"')
    expect(metricStyle).not.toContain('box-shadow')
  })

  it('validates metric item shape and semantic tone', () => {
    const validateItems = getValidator(MetricStrip, 'items')

    expect(validateItems([])).toBe(true)
    expect(validateItems([
      { key: 'all', label: '任务总数', value: 12, scope: '全局', tone: 'neutral' },
      { label: '失败', value: '2', tone: 'danger' }
    ])).toBe(true)
    expect(validateItems([{ label: ' ', value: 1 }])).toBe(false)
    expect(validateItems([{ label: '总数', value: true }])).toBe(false)
    expect(validateItems([{ label: '总数', value: 1, scope: 24 }])).toBe(false)
    expect(validateItems([{ label: '总数', value: 1, tone: 'critical' }])).toBe(false)
    expect(validateItems([{ key: 1, label: '总数', value: 1 }])).toBe(false)
  })

  it('requires unique metric identities and explicit keys for duplicate labels', () => {
    const validateItems = getValidator(MetricStrip, 'items')

    expect(validateItems([
      { label: '重复指标', value: 1 },
      { label: '重复指标', value: 2 }
    ])).toBe(false)
    expect(validateItems([
      { key: 'first', label: '重复指标', value: 1 },
      { key: 'second', label: '重复指标', value: 2 }
    ])).toBe(true)
    expect(validateItems([
      { key: 'same', label: '指标一', value: 1 },
      { key: 'same', label: '指标二', value: 2 }
    ])).toBe(false)
  })

  it('normalizes invalid metric tone classes to neutral', () => {
    expect(metricTemplate).toContain("METRIC_TONES.has(item.tone) ? item.tone : 'neutral'")
  })

  it('closes every desktop grid row and preserves responsive overrides', () => {
    expect(findRule(metricStyle, '.metric-strip__item:nth-child(4n)')).toContain('border-right: 0')
    expect(metricStyle).toContain('.metric-strip__item:nth-child(2n)')
    expect(metricStyle).toMatch(/@media \(max-width: 36rem\)[\s\S]*?\.metric-strip__item\s*\{[^}]*border-right:\s*0/)
  })

  it('requires a non-empty status label', () => {
    const validateLabel = getValidator(StatusIndicator, 'label')

    expect(validateLabel('运行中')).toBe(true)
    expect(validateLabel('  ')).toBe(false)
  })

  it('validates status tone while normalizing invalid classes', () => {
    const validateTone = getValidator(StatusIndicator, 'tone')

    semanticTones.forEach((tone) => expect(validateTone(tone)).toBe(true))
    expect(validateTone('critical')).toBe(false)
    expect(statusTemplate).toContain("STATUS_TONES.has(tone) ? tone : 'neutral'")
  })

  it('combines accessible text, shape and semantic tone', () => {
    expect(statusTemplate).toContain('status-indicator__dot')
    expect(statusTemplate).toContain('{{ label }}')
    expect(statusTemplate).toContain('aria-label')
    expect(statusStyle).toContain('.status-indicator--danger')
  })

  it('wraps long visible status labels without overflowing', () => {
    const rootRule = findRule(statusStyle, '.status-indicator')
    const labelRule = findRule(statusStyle, '.status-indicator__label')

    expect(statusTemplate).toContain('class="status-indicator__label"')
    expect(rootRule).toContain('max-width: 100%')
    expect(rootRule).toContain('min-width: 0')
    expect(statusStyle).not.toContain('white-space: nowrap')
    expect(labelRule).toContain('min-width: 0')
    expect(labelRule).toContain('overflow-wrap: anywhere')
  })

  it('compiles SavedViewTabs and exposes the complete shared command surface', () => {
    expect(SavedViewTabs).toBeTruthy()
    expect(Object.keys(SavedViewTabs.props)).toEqual(['items', 'activeId', 'editableSourceId'])
    expect(SavedViewTabs.props.items.required).toBe(true)
    expect(SavedViewTabs.props.activeId.required).toBe(true)
    expect(SavedViewTabs.emits).toEqual(['select', 'save', 'update', 'rename', 'remove'])
  })

  it('owns confirmed saved-view naming and deletion commands', () => {
    expect.soft(savedViewScript).toContain("import { ElMessage, ElMessageBox } from 'element-plus'")
    expect(savedViewScript).toContain("ElMessageBox.prompt('请输入视图名称'")
    expect(savedViewScript).toContain("ElMessageBox.prompt('请输入新的视图名称'")
    expect(savedViewScript).toContain("ElMessageBox.confirm('确认删除这个保存视图？'")
    expect(savedViewScript).toContain("emit('save', String(value).trim())")
    expect(savedViewScript).toContain("emit('rename', { id: activeItem.value.id, label: String(value).trim() })")
    expect(savedViewScript).toContain("emit('remove', activeItem.value.id)")
    expect(savedViewScript).toContain("return error === 'cancel' || error === 'close'")
    expect.soft(savedViewScript).toContain("reportDialogError(error, '保存视图失败，请重试')")
    expect.soft(savedViewScript).toContain("reportDialogError(error, '重命名视图失败，请重试')")
    expect.soft(savedViewScript).toContain("reportDialogError(error, '删除视图失败，请重试')")
    expect.soft(savedViewScript).toContain('ElMessage.error(message)')
    expect.soft(savedViewScript).not.toContain('throw error')
    expect(savedViewScript).not.toContain('localStorage')
  })

  it('shows transient custom commands while protecting built-in views', () => {
    expect(savedViewScript).toContain("{ id: CUSTOM_VIEW_ID, label: '自定义', builtIn: true, transient: true }")
    expect(savedViewScript).toContain('props.editableSourceId && !item.builtIn')
    expect(savedViewScript).toContain('if (!activeItem.value || activeItem.value.builtIn) return')
    expect(savedViewTemplate).toContain('v-if="activeId === CUSTOM_VIEW_ID"')
    expect(savedViewTemplate).toContain('v-if="activeId === CUSTOM_VIEW_ID && editableSource"')
    expect(savedViewTemplate).toContain("emit('update', editableSource.id)")
    expect(savedViewTemplate).toContain('v-if="activeItem && !activeItem.builtIn"')
  })

  it('keeps saved-view commands compact, responsive and accessible without cards', () => {
    const rootRule = findRule(savedViewStyle, '.saved-view-tabs')
    const actionButtonRule = findRule(savedViewStyle, '.saved-view-tabs__actions :deep(.el-button)')
    const tabItemRule = findRule(savedViewStyle, ':deep(.el-tabs__item)')

    expect(savedViewTemplate).toContain('class="saved-view-tabs" aria-label="保存视图"')
    expect(savedViewTemplate).toContain('<el-tooltip content="视图操作"')
    expect(savedViewTemplate).toContain('aria-label="视图操作"')
    expect(rootRule).toContain('border-bottom: 1px solid var(--line-soft)')
    expect(rootRule).toContain('letter-spacing: 0')
    expect(actionButtonRule).toContain('height: 32px')
    expect(tabItemRule).toContain('overflow: hidden')
    expect(tabItemRule).toContain('text-overflow: ellipsis')
    expect(savedViewStyle).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.saved-view-tabs__tabs\s*\{[^}]*overflow-x:\s*auto/)
    expect(savedViewStyle).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.saved-view-tabs__actions :deep\(\.el-button\)\s*\{[^}]*min-height:\s*44px/)
    expect(savedViewStyle).not.toContain('box-shadow')
    expect(savedViewStyle).not.toMatch(/\bbackground(?:-color)?:/)
  })

  it('gives every saved-view tab a 44px mobile click target', () => {
    const mediaStart = savedViewStyle.indexOf('@media (max-width: 63.9375rem)')

    expect(mediaStart).toBeGreaterThan(-1)
    const mobileStyle = savedViewStyle.slice(mediaStart)
    expect(mobileStyle).toMatch(/:deep\(\.el-tabs__item\)\s*\{[^}]*height:\s*44px/)
    expect(mobileStyle).toMatch(/:deep\(\.el-tabs__item\)\s*\{[^}]*line-height:\s*44px/)
  })

  it('共享 Drawer 标题使用 Element Plus close 回调并提供中文名称与提示', async () => {
    expect(existsSync(drawerHeaderURL)).toBe(true)
    if (!existsSync(drawerHeaderURL)) return

    const drawerHeaderSource = readFileSync(drawerHeaderURL, 'utf8')
    const AdminDrawerHeader = (await import('./AdminDrawerHeader.vue')).default

    expect(AdminDrawerHeader).toBeTruthy()
    expect(Object.keys(AdminDrawerHeader.props)).toEqual(['title', 'titleId', 'titleClass', 'close'])
    expect(AdminDrawerHeader.props.close.required).toBe(true)
    expect(drawerHeaderSource).toContain('aria-label="关闭此对话框"')
    expect(drawerHeaderSource).toContain('title="关闭此对话框"')
    expect(drawerHeaderSource).toContain('class="el-drawer__close-btn"')
    expect(drawerHeaderSource).toContain('@click="close"')
    expect(drawerHeaderSource).not.toContain('<el-tooltip')
    expect(drawerHeaderSource).not.toContain("emit('update:modelValue'")
  })

  it('共享折叠区块按钮公开当前状态的中文名称且保留原折叠语义', () => {
    const toggle = sectionCardTemplate.match(/<button\b(?=[^>]*class="section-card__toggle")[\s\S]*?>/)?.[0] || ''

    expect(SectionCard).toBeTruthy()
    expect(sectionCardScript).toContain('const expanded = ref(props.defaultExpanded)')
    expect(SectionCard.props.defaultExpanded.default).toBe(true)
    expect(toggle).toContain(':aria-label="expanded ? \'收起区块\' : \'展开区块\'"')
    expect(toggle).toContain(':title="expanded ? \'收起区块\' : \'展开区块\'"')
    expect(toggle).toContain(':aria-expanded="expanded"')
    expect(toggle).toContain('@click="expanded = !expanded"')
    expect(sectionCardTemplate).toContain('<el-icon aria-hidden="true">')
    expect(sectionCardTemplate).toContain('<div v-show="expanded" class="section-card__body">')
  })

  it('共享折叠区块按钮消费桌面密度并在窄屏提升为 44px', () => {
    const desktopRule = findRule(sectionCardStyle, '.section-card__toggle')
    const mobileStyle = extractBraceBlock(sectionCardStyle, '@media (max-width: 63.9375rem)')
    const mobileRule = findRule(mobileStyle, '.section-card__toggle')

    for (const property of ['width', 'height', 'min-width', 'min-height']) {
      expect(desktopRule).toContain(`${property}: var(--control-height)`)
      expect(mobileRule).toContain(`${property}: 44px`)
    }
  })
})
