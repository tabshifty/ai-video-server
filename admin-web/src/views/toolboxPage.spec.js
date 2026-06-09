import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const toolbox = readFileSync(new URL('./Toolbox.vue', import.meta.url), 'utf8')
const ed2kTool = readFileSync(new URL('./ToolboxEd2k.vue', import.meta.url), 'utf8')
const router = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
const ed2kRoute = router.match(/\{ path: '\/toolbox\/ed2k'[^}]+\}/)?.[0] || ''

describe('toolbox pages', () => {
  it('keeps the toolbox page as a shell menu of tool entry buttons', () => {
    expect(toolbox).toContain('工具箱')
    expect(toolbox).toContain('ED2K 链接生成器')
    expect(toolbox).toContain('target="_blank"')
    expect(toolbox).toContain('rel="noopener noreferrer"')
    expect(toolbox).toContain('/toolbox/ed2k')
    expect(toolbox).not.toContain('parseEd2kLinks')
    expect(toolbox).not.toContain('ed2kInput')
  })

  it('keeps the ED2K tool page outside the admin shell while preserving the tool workflow', () => {
    expect(ed2kTool).toContain('ED2K 链接生成器')
    expect(ed2kTool).toContain('parseEd2kLinks')
    expect(ed2kTool).toContain('返回工具箱')
    expect(ed2kTool).toContain('/toolbox')
    expect(ed2kTool).not.toContain('components/Layout.vue')
    expect(ed2kTool).not.toMatch(/<Layout[>\s]/)
  })

  it('registers the ED2K tool page as an authenticated route without shell header metadata', () => {
    expect(ed2kRoute).toContain("path: '/toolbox/ed2k'")
    expect(ed2kRoute).toContain('ToolboxEd2k')
    expect(ed2kRoute).not.toContain('public: true')
    expect(ed2kRoute).not.toContain('hideShellPageHeader')
  })
})
