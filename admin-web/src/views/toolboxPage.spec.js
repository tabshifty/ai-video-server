import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const toolbox = readFileSync(new URL('./Toolbox.vue', import.meta.url), 'utf8')
const ed2kTool = readFileSync(new URL('./ToolboxEd2k.vue', import.meta.url), 'utf8')
const imageWorkbench = readFileSync(new URL('./ToolboxImageWorkbench.vue', import.meta.url), 'utf8')
const router = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
const ed2kRoute = router.match(/\{ path: '\/toolbox\/ed2k'[^}]+\}/)?.[0] || ''
const imageWorkbenchRoute = router.match(/\{ path: '\/toolbox\/image-workbench'[^}]+\}/)?.[0] || ''

describe('toolbox pages', () => {
  it('keeps the toolbox page as a shell menu of tool entry buttons', () => {
    expect(toolbox).toContain('工具箱')
    expect(toolbox).toContain('ED2K 链接生成器')
    expect(toolbox).toContain('图像生成工作台')
    expect(toolbox).toContain('target="_blank"')
    expect(toolbox).toContain('rel="noopener noreferrer"')
    expect(toolbox).toContain('/toolbox/ed2k')
    expect(toolbox).toContain('/toolbox/image-workbench')
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

  it('does not track or display clicked state in the ED2K generated links', () => {
    expect(ed2kTool).not.toContain('ed2kClickedLinks')
    expect(ed2kTool).not.toContain('markEd2kLinkClicked')
    expect(ed2kTool).not.toContain('isEd2kLinkClicked')
    expect(ed2kTool).not.toContain('is-clicked')
    expect(ed2kTool).not.toContain('已点击')
    expect(ed2kTool).not.toContain('未点击')
  })

  it('registers the ED2K tool page as an authenticated route without shell header metadata', () => {
    expect(ed2kRoute).toContain("path: '/toolbox/ed2k'")
    expect(ed2kRoute).toContain('ToolboxEd2k')
    expect(ed2kRoute).not.toContain('public: true')
    expect(ed2kRoute).not.toContain('hideShellPageHeader')
  })

  it('keeps the image workbench as a no-shell authenticated tool page', () => {
    expect(imageWorkbench).toContain('图像生成工作台')
    expect(imageWorkbench).toContain('返回工具箱')
    expect(imageWorkbench).toContain('打开媒体库')
    expect(imageWorkbench).toContain('导入媒体库')
    expect(imageWorkbench).toContain('复用为参考图')
    expect(imageWorkbench).toContain('本地历史')
    expect(imageWorkbench).not.toContain('components/Layout.vue')
    expect(imageWorkbench).not.toMatch(/<Layout[>\s]/)
    expect(imageWorkbenchRoute).toContain("path: '/toolbox/image-workbench'")
    expect(imageWorkbenchRoute).toContain('ToolboxImageWorkbench')
    expect(imageWorkbenchRoute).not.toContain('public: true')
    expect(imageWorkbenchRoute).not.toContain('hideShellPageHeader')
  })

  it('surfaces failed workbench generation reasons in local history and the selected-task empty state', () => {
    expect(imageWorkbench).toContain('WarningFilled')
    expect(imageWorkbench).toContain('selectedTaskError')
    expect(imageWorkbench).toContain('task.error')
    expect(imageWorkbench).toContain('生成失败')
    expect(imageWorkbench).toContain('未返回错误原因')
    expect(imageWorkbench).toContain('selectTask(fresh)')
  })
})
