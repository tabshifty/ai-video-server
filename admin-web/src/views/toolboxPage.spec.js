import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const toolbox = readFileSync(new URL('./Toolbox.vue', import.meta.url), 'utf8')
const ed2kTool = readFileSync(new URL('./ToolboxEd2k.vue', import.meta.url), 'utf8')
const archiveImportTool = readFileSync(new URL('./ToolboxArchiveImport.vue', import.meta.url), 'utf8')
const imageWorkbench = readFileSync(new URL('./ToolboxImageWorkbench.vue', import.meta.url), 'utf8')
const orphanFilesTool = readFileSync(new URL('./ToolboxOrphanFiles.vue', import.meta.url), 'utf8')
const passwordVaultTool = readFileSync(new URL('./ToolboxPasswordVault.vue', import.meta.url), 'utf8')
const systemSettings = readFileSync(new URL('./SystemSettings.vue', import.meta.url), 'utf8')
const router = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
const ed2kRoute = router.match(/\{ path: '\/toolbox\/ed2k'[^}]+\}/)?.[0] || ''
const archiveImportRoute = router.match(/\{ path: '\/toolbox\/archive-import'[^}]+\}/)?.[0] || ''
const imageWorkbenchRoute = router.match(/\{ path: '\/toolbox\/image-workbench'[^}]+\}/)?.[0] || ''
const orphanFilesRoute = router.match(/\{ path: '\/toolbox\/orphan-files'[^}]+\}/)?.[0] || ''
const passwordVaultRoute = router.match(/\{ path: '\/toolbox\/password-vault'[^}]+\}/)?.[0] || ''

describe('toolbox pages', () => {
  it('keeps the toolbox page as a shell menu of tool entry buttons', () => {
    expect(toolbox).toContain('工具箱')
    expect(toolbox).toContain('ED2K 链接生成器')
    expect(toolbox).toContain('压缩包导入')
    expect(toolbox).toContain('图像生成工作台')
    expect(toolbox).toContain('孤儿文件扫描')
    expect(toolbox).toContain('密码管理')
    expect(toolbox).toContain('target="_blank"')
    expect(toolbox).toContain('rel="noopener noreferrer"')
    expect(toolbox).toContain('/toolbox/ed2k')
    expect(toolbox).toContain('/toolbox/archive-import')
    expect(toolbox).toContain('/toolbox/image-workbench')
    expect(toolbox).toContain('/toolbox/orphan-files')
    expect(toolbox).toContain('/toolbox/password-vault')
    expect(toolbox).not.toContain('parseEd2kLinks')
    expect(toolbox).not.toContain('ed2kInput')
  })

  it('keeps the toolbox menu responsive with at most four items per row', () => {
    expect(toolbox).toContain('grid-template-columns: repeat(4, minmax(0, 1fr))')
    expect(toolbox).toContain('grid-template-columns: repeat(3, minmax(0, 1fr))')
    expect(toolbox).toContain('grid-template-columns: repeat(2, minmax(0, 1fr))')
    expect(toolbox).toContain('grid-template-columns: 1fr')
    expect(toolbox).not.toContain('grid-template-columns: repeat(auto-fit')
  })

  it('keeps the ED2K tool page outside the admin shell while preserving the tool workflow', () => {
    expect(ed2kTool).toContain('ED2K 链接生成器')
    expect(ed2kTool).toContain('parseEd2kLinks')
    expect(ed2kTool).toContain('返回工具箱')
    expect(ed2kTool).toContain('/toolbox')
    expect(ed2kTool).not.toContain('components/Layout.vue')
    expect(ed2kTool).not.toMatch(/<Layout[>\s]/)
  })

  it('only displays clicked state after an ED2K link is clicked', () => {
    expect(ed2kTool).toContain('ed2kClickedLinks')
    expect(ed2kTool).toContain('markEd2kLinkClicked')
    expect(ed2kTool).toContain('isEd2kLinkClicked')
    expect(ed2kTool).toContain('@click="markEd2kLinkClicked(link)"')
    expect(ed2kTool).toContain('v-if="isEd2kLinkClicked(link)"')
    expect(ed2kTool).toContain('已点击')
    expect(ed2kTool).not.toContain('未点击')
    expect(ed2kTool).toContain('ed2k-link__status')
  })

  it('registers the ED2K tool page as an authenticated route without shell header metadata', () => {
    expect(ed2kRoute).toContain("path: '/toolbox/ed2k'")
    expect(ed2kRoute).toContain('ToolboxEd2k')
    expect(ed2kRoute).not.toContain('public: true')
    expect(ed2kRoute).not.toContain('hideShellPageHeader')
  })

  it('keeps the archive import tool page outside the admin shell while preserving batch-first archive workflows', () => {
    expect(archiveImportTool).toContain('压缩包导入')
    expect(archiveImportTool).toContain('zip、rar、7z')
    expect(archiveImportTool).toContain('支持密码包')
    expect(archiveImportTool).toContain('压缩包内不允许嵌套压缩包')
    expect(archiveImportTool).toContain('批次列表')
    expect(archiveImportTool).toContain('title="上传压缩包"')
    expect(archiveImportTool).toContain('title="批次详情"')
    expect(archiveImportTool).toContain('批量编辑')
    expect(archiveImportTool).toContain('formatFileSize(file.file_size)')
    expect(archiveImportTool).not.toContain('file.file_size | formatFileSize')
    expect(archiveImportTool).not.toContain('components/Layout.vue')
    expect(archiveImportTool).not.toMatch(/<Layout[>\s]/)
    expect(archiveImportRoute).toContain("path: '/toolbox/archive-import'")
    expect(archiveImportRoute).toContain('ToolboxArchiveImport')
    expect(archiveImportRoute).not.toContain('public: true')
    expect(archiveImportRoute).not.toContain('hideShellPageHeader')
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

  it('moves orphan file scanning into a no-shell authenticated toolbox page', () => {
    expect(orphanFilesTool).toContain('孤儿文件扫描')
    expect(orphanFilesTool).toContain('返回工具箱')
    expect(orphanFilesTool).toContain('startOrphanFileScan')
    expect(orphanFilesTool).toContain('getLatestOrphanFileScan')
    expect(orphanFilesTool).toContain('deleteLatestOrphanFileScan')
    expect(orphanFilesTool).toContain('shouldPromptDeleteOrphanScan')
    expect(orphanFilesTool).not.toContain('components/Layout.vue')
    expect(orphanFilesTool).not.toMatch(/<Layout[>\s]/)
    expect(orphanFilesRoute).toContain("path: '/toolbox/orphan-files'")
    expect(orphanFilesRoute).toContain('ToolboxOrphanFiles')
    expect(orphanFilesRoute).not.toContain('public: true')
    expect(orphanFilesRoute).not.toContain('hideShellPageHeader')
  })

  it('adds password management as a no-shell authenticated toolbox page', () => {
    expect(passwordVaultTool).toContain('密码管理')
    expect(passwordVaultTool).toContain('返回工具箱')
    expect(passwordVaultTool).toContain('getAdminPasswordVaultEntries')
    expect(passwordVaultTool).toContain('createAdminPasswordVaultEntry')
    expect(passwordVaultTool).toContain('updateAdminPasswordVaultEntry')
    expect(passwordVaultTool).toContain('deleteAdminPasswordVaultEntry')
    expect(passwordVaultTool).toContain('getAdminPasswordVaultPassword')
    expect(passwordVaultTool).toContain('显示密码')
    expect(passwordVaultTool).toContain('复制密码')
    expect(passwordVaultTool).not.toContain('components/Layout.vue')
    expect(passwordVaultTool).not.toMatch(/<Layout[>\s]/)
    expect(passwordVaultRoute).toContain("path: '/toolbox/password-vault'")
    expect(passwordVaultRoute).toContain('ToolboxPasswordVault')
    expect(passwordVaultRoute).not.toContain('public: true')
    expect(passwordVaultRoute).not.toContain('hideShellPageHeader')
  })

  it('removes orphan scanning from system settings after the toolbox migration', () => {
    expect(systemSettings).toContain('系统设置')
    expect(systemSettings).toContain('临时文件清理')
    expect(systemSettings).toContain('系统日志')
    expect(systemSettings).not.toContain('孤儿文件扫描')
    expect(systemSettings).not.toContain('startOrphanFileScan')
    expect(systemSettings).not.toContain('getLatestOrphanFileScan')
    expect(systemSettings).not.toContain('shouldPromptDeleteOrphanScan')
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
