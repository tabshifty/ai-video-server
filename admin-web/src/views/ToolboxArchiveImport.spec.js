import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxArchiveImport.vue', import.meta.url), 'utf8')

describe('ToolboxArchiveImport', () => {
  it('keeps tags and collections as selector-based inputs instead of JSON or raw ID fields', () => {
    expect(source).toContain('默认标签')
    expect(source).toContain('可选择或输入标签')
    expect(source).toContain('默认视频合集')
    expect(source).toContain('默认图片合集')
    expect(source).toContain('可选，可多选')
    expect(source).not.toContain('JSON 数组')
    expect(source).not.toContain('合集 ID')
  })
})
