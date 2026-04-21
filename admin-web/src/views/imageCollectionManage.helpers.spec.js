import { describe, expect, it, vi } from 'vitest'

import { buildImageCollectionPayload, revokePreviewURLs } from './imageCollectionManage.helpers'

describe('buildImageCollectionPayload', () => {
  it('includes cover_image_id while keeping manual cover url fallback', () => {
    const payload = buildImageCollectionPayload({
      name: '  图集甲  ',
      description: '  简介  ',
      cover_url: ' https://legacy.example/cover.jpg ',
      cover_image_id: 'image-1',
      sort_order: '12.8',
      active: 1
    })

    expect(payload).toEqual({
      name: '图集甲',
      description: '简介',
      cover_url: 'https://legacy.example/cover.jpg',
      cover_image_id: 'image-1',
      sort_order: 12,
      active: true
    })
  })
})

describe('revokePreviewURLs', () => {
  it('releases every created object url', () => {
    const revoke = vi.fn()

    revokePreviewURLs(
      {
        first: 'blob:first',
        second: 'blob:second',
        empty: ''
      },
      revoke
    )

    expect(revoke).toHaveBeenCalledTimes(2)
    expect(revoke).toHaveBeenCalledWith('blob:first')
    expect(revoke).toHaveBeenCalledWith('blob:second')
  })
})
