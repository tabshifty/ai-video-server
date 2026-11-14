import { describe, expect, it } from 'vitest'

import { buildExistingUploadMessage, hasExistingAttachmentSelection } from './videoUpload.existing'

describe('buildExistingUploadMessage', () => {
  it('待删除视频明确说明未关联合集', () => {
    expect(buildExistingUploadMessage({ pending_delete: true, added_collections: 0 }))
      .toBe('已存在（该视频待删除），未关联合集')
  })

  it('有新增时显示实际新增数量', () => {
    expect(buildExistingUploadMessage({ added_collections: 2 })).toBe('已存在，新增 2 个合集')
  })

  it('补齐图片合集时附加说明', () => {
    expect(buildExistingUploadMessage({ added_collections: 1, added_image_collection: true }))
      .toBe('已存在，新增 1 个合集，并补齐图片合集')
  })

  it('只补齐图片合集时不能报“合集无变化”', () => {
    expect(buildExistingUploadMessage({ added_collections: 0, added_image_collection: true }))
      .toBe('已存在，已补齐图片合集')
  })

  it('非法新增计数按 0 处理', () => {
    expect(buildExistingUploadMessage({ added_collections: 'oops' })).toBe('已存在，合集无变化')
  })

  it('选中的关系都已存在时不算新增', () => {
    expect(buildExistingUploadMessage({ added_collections: 0 })).toBe('已存在，合集无变化')
    expect(buildExistingUploadMessage({})).toBe('已存在，合集无变化')
    expect(buildExistingUploadMessage(null)).toBe('已存在，合集无变化')
  })
})

describe('hasExistingAttachmentSelection', () => {
  it('没有选择合集和图片合集时不需要调用追加接口', () => {
    expect(hasExistingAttachmentSelection([], '')).toBe(false)
    expect(hasExistingAttachmentSelection(null, '  ')).toBe(false)
    expect(hasExistingAttachmentSelection([''], '')).toBe(false)
  })

  it('任一项有选择就要追加', () => {
    expect(hasExistingAttachmentSelection(['a'], '')).toBe(true)
    expect(hasExistingAttachmentSelection([], 'img-1')).toBe(true)
  })
})
