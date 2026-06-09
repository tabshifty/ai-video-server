import { describe, expect, it } from 'vitest'
import {
  IMAGE_WORKBENCH_LIMITS,
  buildImageGenerationPayload,
  normalizeImageWorkbenchParams,
  validateReferenceImageFiles
} from './imageWorkbench.helpers'

describe('image workbench helpers', () => {
  it('normalizes common generation params', () => {
    expect(
      normalizeImageWorkbenchParams({
        size: 'bad',
        quality: 'high',
        output_format: 'jpeg',
        output_compression: 200,
        n: 99
      })
    ).toEqual({
      size: 'auto',
      quality: 'high',
      output_format: 'jpeg',
      output_compression: 100,
      n: 4
    })
  })

  it('clears compression for png output', () => {
    expect(normalizeImageWorkbenchParams({ output_format: 'png', output_compression: 80 }).output_compression).toBeNull()
    expect(normalizeImageWorkbenchParams({ output_format: 'webp' }).output_compression).toBe(82)
  })

  it('validates reference image count, type and size', () => {
    const goodFile = new File(['x'], 'ref.png', { type: 'image/png' })
    expect(validateReferenceImageFiles([goodFile]).ok).toBe(true)

    const badType = new File(['x'], 'ref.gif', { type: 'image/gif' })
    expect(validateReferenceImageFiles([badType])).toEqual({ ok: false, message: '仅支持 PNG、JPEG、WebP 参考图' })

    const tooMany = Array.from({ length: IMAGE_WORKBENCH_LIMITS.maxReferenceImages + 1 }, (_, index) =>
      new File(['x'], `ref-${index}.png`, { type: 'image/png' })
    )
    expect(validateReferenceImageFiles(tooMany).message).toContain('参考图最多')
  })

  it('builds backend payload without leaking local-only fields', () => {
    const payload = buildImageGenerationPayload(
      '  生成海报  ',
      { output_format: 'webp', output_compression: 76, n: 2 },
      [{ id: 'local-1', name: 'ref.png', mime: 'image/png', dataUrl: 'data:image/png;base64,abc' }]
    )

    expect(payload).toMatchObject({
      prompt: '生成海报',
      size: 'auto',
      quality: 'auto',
      output_format: 'webp',
      output_compression: 76,
      n: 2,
      reference_images: [{ name: 'ref.png', mime: 'image/png', data_url: 'data:image/png;base64,abc' }]
    })
    expect(payload.reference_images[0]).not.toHaveProperty('id')
  })
})
