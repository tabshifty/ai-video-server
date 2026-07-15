import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const css = readFileSync(new URL('./theme.css', import.meta.url), 'utf8')
const overrides = readFileSync(new URL('./element-overrides.css', import.meta.url), 'utf8')
const mainSource = readFileSync(new URL('../main.js', import.meta.url), 'utf8')
const sectionCard = readFileSync(new URL('../components/base/SectionCard.vue', import.meta.url), 'utf8')
const emptyState = readFileSync(new URL('../components/base/EmptyState.vue', import.meta.url), 'utf8')
const bulkActionBar = readFileSync(new URL('../components/base/BulkActionBar.vue', import.meta.url), 'utf8')
const VIEW_HEX_AUDIT_TARGETS = [
  '../views/Dashboard.vue',
  '../views/SystemSettings.vue',
  '../views/Toolbox.vue',
  '../views/ToolboxArchiveImport.vue',
  '../views/ToolboxOrphanFiles.vue',
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
const densityTokenContracts = [
  {
    density: 'compact',
    tokens: [
      ['--control-height', '32px'],
      ['--table-head-height', '36px'],
      ['--table-row-height', '40px'],
      ['--media-row-height', '52px'],
      ['--section-padding', '12px']
    ]
  },
  {
    density: 'monitor',
    tokens: [
      ['--control-height', '32px'],
      ['--table-head-height', '36px'],
      ['--table-row-height', '44px'],
      ['--media-row-height', '44px'],
      ['--section-padding', '12px']
    ]
  },
  {
    density: 'form',
    tokens: [
      ['--control-height', '36px'],
      ['--section-padding', '16px']
    ]
  }
]

describe('theme tokens', () => {
  it('exports the approved Precision Ops shell and semantic tokens', () => {
    expect(css).toContain('--admin-sidebar-width: 224px')
    expect(css).toContain('--admin-sidebar-collapsed-width: 56px')
    expect(css).toContain('--admin-header-height: 52px')
    expect(css).toContain('--bg-canvas: #f7f8fa')
    expect(css).toContain('--text-primary: #172033')
    expect(css).toContain('--text-muted: #607085')
    expect(css).toContain('--success-600: #047857')
    expect(css).toContain('--warning-600: #b45309')
    expect(css).toContain('--danger-600: #c81e1e')
    expect(css).toContain('--info-600: #0369a1')
  })

  it('scopes compact sizing instead of applying it to every form', () => {
    expect(css).toContain('[data-density="compact"]')
    expect(css).toContain('[data-density="monitor"]')
    expect(css).toContain('[data-density="form"]')
    expect(overrides).toContain('min-width: 44px')
    expect(overrides).toContain('.el-button.is-circle')
    expect(overrides).toContain('.el-checkbox')
    expect(overrides).not.toMatch(/^:where\(\.el-button\)\s*\{[^}]*min-height:\s*32px/m)
  })

  it.each(densityTokenContracts)('exports every $density density token', ({ density, tokens }) => {
    const densityRule = css.match(new RegExp(`\\[data-density="${density}"\\]\\s*\\{([^}]*)\\}`))

    expect(densityRule).not.toBeNull()
    tokens.forEach(([token, value]) => {
      expect(densityRule?.[1]).toContain(`${token}: ${value}`)
    })
  })

  it('keeps component specificity in density overrides', () => {
    expect(overrides).not.toMatch(/:where\(\[data-density[^\n]*\)\s+:where\(/)

    const selectors = [
      ':where([data-density="compact"], [data-density="monitor"]) .el-select__wrapper',
      ':where([data-density="compact"], [data-density="monitor"]) .el-table th.el-table__cell',
      ':where([data-density="compact"], [data-density="monitor"]) .el-table td.el-table__cell',
      ':where([data-density="compact"]) .has-media-rows .el-table td.el-table__cell',
      ':where([data-density="form"]) .el-select__wrapper',
      ':where([data-density]) .el-pagination button',
      ':where([data-density]) .el-button.is-circle'
    ]

    selectors.forEach((selector) => {
      expect(overrides).toContain(selector)
    })
  })

  it('loads density overrides after Element Plus defaults and theme tokens', () => {
    const elementPlusStyles = mainSource.indexOf("import 'element-plus/dist/index.css'")
    const themeStyles = mainSource.indexOf("import './assets/theme.css'")
    const densityOverrides = mainSource.indexOf("import './assets/element-overrides.css'")

    expect(elementPlusStyles).toBeGreaterThan(-1)
    expect(themeStyles).toBeGreaterThan(elementPlusStyles)
    expect(densityOverrides).toBeGreaterThan(themeStyles)
  })

  it('keeps narrow-screen touch targets at least 44px in both dimensions', () => {
    const touchTargetRule = overrides.match(
      /@media \(max-width: 63\.9375rem\) \{[\s\S]*?:where\(\[data-density\]\) \.el-button\.is-circle,[^{]*\{([^}]*)\}/
    )

    expect(touchTargetRule).not.toBeNull()
    expect(touchTargetRule?.[1]).toMatch(/min-width:\s*44px/)
    expect(touchTargetRule?.[1]).toMatch(/min-height:\s*44px/)
  })

  it('exports the approved typography scale', () => {
    expect(css).toMatch(/--text-h1:\s*20px/)
    expect(css).toMatch(/--text-h2:\s*14px/)
    expect(css).toMatch(/--text-body:\s*14px/)
    expect(css).toMatch(/--text-small:\s*13px/)
    expect(css).toMatch(/--text-caption:\s*12px/)
    expect(css).toMatch(/--text-kpi:\s*24px/)
  })

  it('uses the approved base section geometry', () => {
    expect(sectionCard).not.toMatch(/\.section-card\s*\{[^}]*box-shadow:/)
    expect(emptyState).toMatch(/\.empty-state\s*\{[^}]*min-height:\s*160px/)
    expect(bulkActionBar).toMatch(/\.bulk-action-bar\s*\{[^}]*border-radius:\s*var\(--radius-md\)/)
    expect(bulkActionBar).toMatch(/\.bulk-action-bar\s*\{[^}]*box-shadow:\s*var\(--shadow-lg\)/)
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
