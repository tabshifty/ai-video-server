import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxEd2k.vue', import.meta.url), 'utf8')

describe('ToolboxEd2k', () => {
  it('renders the legacy ED2K link generator workflow', () => {
    expect(source).toContain('ED2K 链接生成器')
    expect(source).toContain('parseEd2kLinks')
    expect(source).toContain('ed2kInput')
    expect(source).toContain('ed2kClickedLinks')
    expect(source).toContain('markEd2kLinkClicked')
    expect(source).toContain('isEd2kLinkClicked')
    expect(source).toContain('已点击')
    expect(source).not.toContain('ED2K 下载工作台')
    expect(source).not.toContain('source_link')
    expect(source).not.toContain('expected_files')
  })

  it('adds explicit checkbox batch submission with a 50-link confirmation', () => {
    expect(source).toContain('全选')
    expect(source).toContain('发送选中到115')
    expect(source).toContain('selectedLinkIds')
    expect(source).toContain('duplicateLinkIds')
    expect(source).toContain('createCloudDownloadBatch')
    expect(source).toContain('cloudDownloadStatuses')
    expect(source).toContain('ElMessageBox.confirm')
    expect(source).toContain('BATCH_LIMIT = 50')
    expect(source).toContain('单次最多')
    expect(source).toContain('批量提交完成')
    expect(source).toContain('cloudStatusFor')
  })
})
