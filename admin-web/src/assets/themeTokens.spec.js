import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const css = readFileSync(new URL('./theme.css', import.meta.url), 'utf8')

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
