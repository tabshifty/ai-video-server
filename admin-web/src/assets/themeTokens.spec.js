import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const css = readFileSync(new URL('./theme.css', import.meta.url), 'utf8')
const VIEW_HEX_AUDIT_TARGETS = [
  '../views/Dashboard.vue',
  '../views/SystemSettings.vue',
  '../views/UserManage.vue',
  '../views/TaskMonitor.vue',
  '../views/IPTVManage.vue',
  '../views/CollectionManage.vue',
  '../views/ActorManage.vue',
  '../components/UploadProgress.vue',
  '../views/ScrapePreview.vue',
  '../views/AVManualScrape.vue',
  '../views/TvSeriesManage.vue',
  '../views/ImageCollectionManage.vue',
  '../views/VideoList.vue',
  '../views/ImageManage.vue',
  '../views/VideoUpload.vue'
]
const roseHexPatterns = [/#881337/i, /#be123c/i, /#7f1d1d/i]
const dashboardLegacyPatterns = [/#2563eb/i, /#eff6ff/i, /#64748b/i, /#e2e8f0/i, /#cad8f5/i, /#e11d48/i, /#fda4af/i]

describe('theme tokens', () => {
  it('exports the core admin design tokens', () => {
    expect(css).toContain('--primary: var(--blue-600)')
    expect(css).toContain('--text-primary: var(--slate-900)')
    expect(css).toContain('--bg-canvas: var(--slate-50)')
    expect(css).toContain('--bg-sidebar: var(--slate-100)')
  })

  it('exports six typography tiers', () => {
    expect(css).toMatch(/--text-display:\s*28px/)
    expect(css).toMatch(/--text-h1:\s*20px/)
    expect(css).toMatch(/--text-h2:\s*15px/)
    expect(css).toMatch(/--text-body:\s*14px/)
    expect(css).toMatch(/--text-small:\s*13px/)
    expect(css).toMatch(/--text-caption:\s*11px/)
  })

  it('removes the old rose palette and Fira stack', () => {
    expect(css).not.toMatch(/#881337/i)
    expect(css).not.toMatch(/#be123c/i)
    expect(css).not.toMatch(/#7f1d1d/i)
    expect(css).not.toContain('Fira Code')
    expect(css).not.toContain('Fira Sans')
    expect(css).not.toContain('--font-code')
  })
})

describe('view hex audit', () => {
  VIEW_HEX_AUDIT_TARGETS.forEach((relativePath) => {
    it(`${relativePath} 不含旧玫红 hex`, () => {
      const source = readFileSync(new URL(relativePath, import.meta.url), 'utf8')
      const patterns = relativePath.includes('Dashboard.vue')
        ? [...roseHexPatterns, ...dashboardLegacyPatterns]
        : roseHexPatterns
      patterns.forEach((pattern) => {
        expect(source).not.toMatch(pattern)
      })
    })
  })
})
