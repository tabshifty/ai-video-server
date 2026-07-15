import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const metricStrip = readFileSync(new URL('./MetricStrip.vue', import.meta.url), 'utf8')
const statusIndicator = readFileSync(new URL('./StatusIndicator.vue', import.meta.url), 'utf8')

describe('Precision Ops base components', () => {
  it('renders labeled tabular metrics without cards', () => {
    expect(metricStrip).toContain('metric-strip')
    expect(metricStrip).toContain('tabular-num')
    expect(metricStrip).toContain('item.scope')
    expect(metricStrip).not.toContain('box-shadow')
  })

  it('combines text, shape and semantic tone', () => {
    expect(statusIndicator).toContain('status-indicator__dot')
    expect(statusIndicator).toContain('{{ label }}')
    expect(statusIndicator).toContain('aria-label')
    expect(statusIndicator).toContain('status-indicator--danger')
  })
})
