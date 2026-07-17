import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import {
  IMAGE_WORKBENCH_LIMITS,
  buildImageGenerationPayload,
  buildReferenceImageSnapshots,
  createImageWorkbenchTask,
  estimateDataUrlBytes,
  hydrateReferenceImageFromSnapshot,
  normalizeImageWorkbenchParams,
  validateReferenceImageFiles
} from './imageWorkbench.helpers'

const readView = (file) => readFileSync(new URL(`./${file}`, import.meta.url), 'utf8')

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

  it('estimates data URL payload bytes for restored local images', () => {
    expect(estimateDataUrlBytes('data:image/png;base64,eA==')).toBe(1)
    expect(estimateDataUrlBytes('data:image/png;base64,YWJj')).toBe(3)
    expect(estimateDataUrlBytes('bad')).toBe(0)
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

  it('hydrates frozen library asset snapshots without refetching source assets', () => {
    const reference = hydrateReferenceImageFromSnapshot(
      {
        image_id: 'local-copy',
        name: 'frozen-title.png',
        mime: 'image/png',
        source_kind: 'library_asset',
        source_image_id: 'asset-1',
        source_title: '加入时标题',
        source_status: 'ready',
        source_active: true,
        source_view_url: '/api/v1/admin/images/asset-1/view',
        source_frozen_at: 1781343900000
      },
      {
        id: 'local-copy',
        dataUrl: 'data:image/png;base64,frozen',
        mime: 'image/png'
      }
    )

    expect(reference).toEqual({
      id: 'local-copy',
      file: null,
      name: 'frozen-title.png',
      mime: 'image/png',
      size: 4,
      dataUrl: 'data:image/png;base64,frozen',
      sourceKind: 'library_asset',
      sourceTaskId: '',
      sourceResultId: '',
      sourceImageId: 'asset-1',
      sourceTitle: '加入时标题',
      sourceStatus: 'ready',
      sourceActive: true,
      sourceViewUrl: '/api/v1/admin/images/asset-1/view',
      sourceFrozenAt: 1781343900000
    })
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

describe('image workbench Precision Ops contracts', () => {
  it('keeps the mask editor component-only boundary and separates data colors from UI colors', () => {
    const source = readView('ImageWorkbenchMaskEditor.vue')
    const sourceWithoutMaskDataColor = source.replace("const MASK_OPAQUE_COLOR = '#ffffff'", '')

    expect(source).toContain('data-density="form"')
    expect(source).toContain('width="min(96vw, 1080px)"')
    expect(source).not.toContain('.mask-editor :deep(.el-dialog)')
    expect(source).toContain('mask-editor__toolbar')
    expect(source).toContain("const MASK_OPAQUE_COLOR = '#ffffff'")
    expect(source.match(/= MASK_OPAQUE_COLOR/g)).toHaveLength(5)
    expect(source).toContain("function resolveCanvasColor(token) {\n  if (typeof window === 'undefined' || typeof document === 'undefined') return ''\n  return window.getComputedStyle(document.documentElement).getPropertyValue(token).trim()\n}")
    expect(source).toContain("overlayCtx.fillStyle = resolveCanvasColor('--primary')")
    expect(source).toContain('color-mix(in srgb, var(--text-primary) 4%, transparent)')
    expect(source.match(/linear-gradient\(/g)).toHaveLength(2)
    expect(sourceWithoutMaskDataColor).not.toMatch(/#[0-9a-fA-F]{3,8}\b/)
    expect(source).not.toMatch(/rgba?\(\s*\d/)
    expect(source).not.toMatch(/\.mask-editor__canvas\s*\{[^}]*box-shadow:/s)
    expect(source).not.toContain('<Layout')
    expect(source).not.toContain('<PageHeader')
  })

  it('uses fixed-format mask controls without changing pointer or save contracts', () => {
    const source = readView('ImageWorkbenchMaskEditor.vue')

    expect(source).toContain('<el-radio-group v-model="tool"')
    expect(source).toContain('<el-slider v-model="brushSize"')
    expect(source).toContain('class="mask-editor__icon-button"')
    expect(source).toContain('aria-label="清空蒙版"')
    expect(source).toContain('title="清空蒙版"')
    expect(source).toMatch(/\.mask-editor__icon-button\s*\{[^}]*width:\s*36px;[^}]*height:\s*36px;/s)
    expect(source).toContain('grid-template-columns: minmax(12rem, 1fr) auto minmax(14rem, 20rem) auto;')
    expect(source).toMatch(/@media \(max-width: 47\.9375rem\)[\s\S]*?\.mask-editor__toolbar\s*\{[^}]*flex-wrap:\s*wrap;/s)
    expect(source).toContain("const emit = defineEmits(['update:modelValue', 'save'])")
    for (const binding of [
      '@pointerdown.prevent="handlePointerDown"',
      '@pointermove.prevent="handlePointerMove"',
      '@pointerup.prevent="stopDrawing"',
      '@pointerleave.prevent="stopDrawing"',
      '@pointercancel.prevent="stopDrawing"'
    ]) {
      expect(source).toContain(binding)
    }
    expect(source).toContain('event.target?.setPointerCapture?.(event.pointerId)')
    expect(source).toContain('event.target.releasePointerCapture(event.pointerId)')
    expect(source).toContain('x: ((event.clientX - rect.left) / rect.width) * canvas.width')
    expect(source).toContain('y: ((event.clientY - rect.top) / rect.height) * canvas.height')
    expect(source).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.mask-editor__icon-button\s*\{[^}]*width:\s*44px;[^}]*height:\s*44px;/s)
    expect(source).toContain("emit('save', maskCanvas.toDataURL('image/png'))")
  })

  it('uses stable compact media-grid geometry in the image workbench', () => {
    const source = readView('ToolboxImageWorkbench.vue')
    const resultGridRule = source.match(/\.result-grid\s*\{[^}]*\}/s)?.[0] || ''
    const libraryGridRule = source.match(/\.library-grid\s*\{[^}]*\}/s)?.[0] || ''

    expect(source).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(source).toContain('<StatusIndicator :label="statusLabel" :tone="statusType" />')
    expect(source).toContain('<section class="workbench-panel workbench-panel--preview" data-density="compact">')
    expect(source).toContain('<div class="library-picker" data-density="compact">')
    for (const rule of [resultGridRule, libraryGridRule]) {
      expect(rule).toContain('grid-template-columns: repeat(auto-fill, minmax(184px, 1fr));')
      expect(rule).toContain('gap: 12px;')
    }
    expect(source).toMatch(/\.result-card img\s*\{[^}]*aspect-ratio:[^}]*object-fit:\s*contain;/s)
    expect(source).toMatch(/\.library-card__preview\s*\{[^}]*aspect-ratio:/s)
    expect(source).toMatch(/\.library-card__preview img\s*\{[^}]*object-fit:\s*contain;/s)
    expect(source).not.toMatch(/\bbox-shadow\s*:/)
    expect(source).toMatch(/\.result-card__actions\s*\{[^}]*display:\s*flex;[^}]*flex-wrap:\s*wrap;/s)
    expect(source).toContain('@click="downloadResult(item, index)"')
    expect(source).toContain('@click="reuseAsReference(item)"')
    expect(source).toContain('@click="importToMediaLibrary(item, index)"')
    expect(source).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.library-card__select\s*\{[^}]*width:\s*44px;[^}]*height:\s*44px;/s)
    expect(source).toContain('<PageHeader')
  })
})
