import { describe, expect, it, vi, beforeEach } from 'vitest'

const { post, get, put, remove } = vi.hoisted(() => ({
  post: vi.fn(),
  get: vi.fn(),
  put: vi.fn(),
  remove: vi.fn()
}))

vi.mock('./request', () => ({
  default: {
    post,
    get,
    put,
    delete: remove
  }
}))

import { uploadAdminImages } from './admin'

describe('uploadAdminImages', () => {
  beforeEach(() => {
    post.mockReset()
    post.mockResolvedValue({ ok: true })
  })

  it('disables request timeout for long-running image uploads', async () => {
    const formData = new FormData()

    await uploadAdminImages(formData)

    expect(post).toHaveBeenCalledWith('/admin/images/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 0
    })
  })
})
