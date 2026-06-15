import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const layout = readFileSync(new URL('./Layout.vue', import.meta.url), 'utf8')
const commandPaletteHelper = readFileSync(new URL('./base/commandPalette.helpers.js', import.meta.url), 'utf8')

describe('Layout shell', () => {
  it('uses grouped navigation and the command palette shell', () => {
    expect(layout).toContain('分组')
    expect(commandPaletteHelper).toContain('媒体库')
    expect(commandPaletteHelper).toContain('录入处理')
    expect(commandPaletteHelper).toContain('工具箱')
    expect(commandPaletteHelper).toContain('系统')
    expect(layout).toMatch(/CommandPalette|command-palette/i)
  })

  it('persists the sidebar collapse preference and profile chip', () => {
    expect(layout).toContain('admin-sidebar-collapsed')
    expect(layout).toMatch(/profile/i)
  })

  it('keeps a visible expand affordance in the collapsed sidebar', () => {
    expect(layout).toContain('.admin-shell.is-collapsed .brand-block {')
    expect(layout).toContain('flex-direction: column;')
    expect(layout).toContain('.admin-shell.is-collapsed .collapse-button {')
    expect(layout).toContain('display: inline-flex;')
    expect(layout).toContain('width: var(--space-8);')
  })

  it('removes the legacy rose gradient and admin subtitle copy', () => {
    expect(layout).not.toContain('#881337')
    expect(layout).not.toContain('#7f1d1d')
    expect(layout).not.toContain('linear-gradient(180deg, #881337')
    expect(layout).not.toContain("'管理员工作台'")
  })
})
