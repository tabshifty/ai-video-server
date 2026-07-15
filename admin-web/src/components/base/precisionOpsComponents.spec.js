import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import MetricStrip from './MetricStrip.vue'
import StatusIndicator from './StatusIndicator.vue'

const metricStripSource = readFileSync(new URL('./MetricStrip.vue', import.meta.url), 'utf8')
const statusIndicatorSource = readFileSync(new URL('./StatusIndicator.vue', import.meta.url), 'utf8')
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

const metricTemplate = extractBlock(metricStripSource, 'template')
const metricStyle = extractBlock(metricStripSource, 'style')
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
})
