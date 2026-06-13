import { describe, expect, it } from 'vitest'
import {
  IMAGE_WORKBENCH_LIMITS,
  buildImageGenerationPayload,
  buildReferenceImageSnapshots,
  createImageWorkbenchTask,
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

  it('builds backend payload without leaking local-only fields', async () => {
    const payload = await buildImageGenerationPayload(
      '  生成海报  ',
      { output_format: 'webp', output_compression: 76, n: 2 },
      [{
        id: 'local-1',
        name: 'ref.png',
        mime: 'image/png',
        dataUrl: 'data:image/png;base64,abc',
        sourceKind: 'library_asset',
        sourceImageId: 'asset-1',
        sourceTitle: '图库原图'
      }]
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
    expect(payload.reference_images[0]).not.toHaveProperty('source_image_id')
    expect(payload.reference_images[0]).not.toHaveProperty('sourceTitle')
  })

  it('stores library asset metadata only in local reference snapshots', () => {
    expect(buildReferenceImageSnapshots([
      {
        id: 'local-asset-copy',
        name: 'frozen.png',
        mime: 'image/png',
        sourceKind: 'library_asset',
        sourceImageId: 'asset-1',
        sourceTitle: '原始图库标题',
        sourceStatus: 'ready',
        sourceActive: true,
        sourceViewUrl: '/api/v1/admin/images/asset-1/view',
        sourceFrozenAt: 1781343900000
      },
      {
        id: 'local-upload',
        name: 'upload.webp',
        mime: 'image/webp',
        sourceKind: 'browser_input'
      }
    ])).toEqual([
      {
        image_id: 'local-asset-copy',
        name: 'frozen.png',
        mime: 'image/png',
        slot_index: 0,
        source_kind: 'library_asset',
        source_task_id: '',
        source_result_id: '',
        source_image_id: 'asset-1',
        source_title: '原始图库标题',
        source_status: 'ready',
        source_active: true,
        source_view_url: '/api/v1/admin/images/asset-1/view',
        source_frozen_at: 1781343900000
      },
      {
        image_id: 'local-upload',
        name: 'upload.webp',
        mime: 'image/webp',
        slot_index: 1,
        source_kind: 'browser_input',
        source_task_id: '',
        source_result_id: ''
      }
    ])
  })

  it('builds mask payload against the original reference slot', async () => {
    const payload = await buildImageGenerationPayload(
      '编辑这张图',
      { output_format: 'png', n: 1 },
      [
        { id: 'ref-a', name: 'first.jpg', mime: 'image/jpeg', dataUrl: 'data:image/png;base64,aaa' },
        { id: 'ref-b', name: 'second.png', mime: 'image/png', dataUrl: 'data:image/png;base64,bbb' }
      ],
      { targetImageId: 'ref-b', maskDataUrl: 'data:image/png;base64,mask' }
    )

    expect(payload.mask).toEqual({
      name: 'mask.png',
      mime: 'image/png',
      data_url: 'data:image/png;base64,mask',
      target_index: 1
    })
    expect(payload.reference_images).toHaveLength(2)
    expect(payload.reference_images[1]).toMatchObject({
      name: 'second.png',
      mime: 'image/png',
      data_url: 'data:image/png;base64,bbb'
    })
  })

  it('stores optional structured snapshots in local task records', () => {
    const task = createImageWorkbenchTask({
      prompt: '继续改图',
      params: { n: 1 },
      referenceImageIds: ['ref-a'],
      referenceSnapshots: [{ image_id: 'ref-a', source_kind: 'browser_input', slot_index: 0 }],
      outputImageIds: ['out-a'],
      mask: { image_id: 'mask-a', target_reference_index: 0 }
    })

    expect(task.referenceSnapshots).toEqual([{ image_id: 'ref-a', source_kind: 'browser_input', slot_index: 0 }])
    expect(task.mask).toEqual({ image_id: 'mask-a', target_reference_index: 0 })
  })
})
