import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxEd2k.vue', import.meta.url), 'utf8')

describe('ToolboxEd2k', () => {
  it('renders a fixed ED2K download detail skeleton with stable section order', () => {
    expect(source).toContain('ED2K 下载工作台')
    expect(source).toContain('来源标识')
    expect(source).toContain('预期文件信息')
    expect(source).toContain('状态反馈')
    expect(source).toContain('进行中落盘快照')
    expect(source).toContain('失败残留快照')
    expect(source).toContain('最终结果清单')
    expect(source).toContain('source_link')
    expect(source).toContain('resource_hash')
    expect(source).toContain('expected_files')
    expect(source).toContain('running_files')
    expect(source).toContain('failed_files')
    expect(source).toContain('completed_files')
    expect(source).toContain('showExpectedFiles')
    expect(source).toContain('showFileSection')
    expect(source).toContain('fileSectionTitle')
    expect(source).toContain('currentStatusDescription')
    expect(source).toContain('formatFileSize(task.declared_size)')
    expect(source).not.toContain('parseEd2kLinks')
    expect(source).not.toContain('ed2kClickedLinks')
    expect(source).not.toContain('已点击')
  })
})
