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

function extractSfcBlock(source, tag) {
  const opening = source.match(new RegExp(`<${tag}[^>]*>`))
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const end = source.lastIndexOf(`</${tag}>`)

  expect(opening).not.toBeNull()
  expect(end).toBeGreaterThan(start)
  return source.slice(start, end)
}

function extractBraceBlock(source, marker) {
  const start = source.indexOf(marker)
  const openingBrace = source.indexOf('{', start)

  expect(start, marker).toBeGreaterThanOrEqual(0)
  expect(openingBrace, marker).toBeGreaterThan(start)
  let depth = 0
  for (let index = openingBrace; index < source.length; index += 1) {
    if (source[index] === '{') depth += 1
    if (source[index] === '}') depth -= 1
    if (depth === 0) return source.slice(start, index + 1)
  }
  throw new Error(`未找到完整代码块：${marker}`)
}

function findStyleRule(style, selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = style.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`))

  expect(match, selector).not.toBeNull()
  return match?.[1] || ''
}

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
  it('为主提示词输入提供独立中文程序化名称', () => {
    const source = readView('ToolboxImageWorkbench.vue')
    const template = extractSfcBlock(source, 'template')
    const promptInput = template.match(
      /<el-input\b(?=[^>]*v-model="prompt")(?=[^>]*type="textarea")[^>]*>/
    )?.[0] || ''

    expect(promptInput).toContain('aria-label="图像生成提示词"')
  })

  it('keeps the mask editor component-only boundary and separates data colors from UI colors', () => {
    const source = readView('ImageWorkbenchMaskEditor.vue')
    const script = extractSfcBlock(source, 'script')
    const drawPreview = extractBraceBlock(script, 'function drawPreview()')
    const resolveCanvasColor = extractBraceBlock(script, 'function resolveCanvasColor(token)')
    const sourceWithoutMaskDataColor = source.replace("const MASK_OPAQUE_COLOR = '#ffffff'", '')

    expect(source).toContain('data-density="form"')
    expect(source).toContain('width="min(96vw, 1080px)"')
    expect(source).not.toContain('.mask-editor :deep(.el-dialog)')
    expect(source).toContain('mask-editor__toolbar')
    expect(source).toContain("const MASK_OPAQUE_COLOR = '#ffffff'")
    expect(source.match(/= MASK_OPAQUE_COLOR/g)).toHaveLength(5)
    expect(resolveCanvasColor).toContain("if (typeof window === 'undefined' || typeof document === 'undefined') return ''")
    expect(resolveCanvasColor).toContain('return window.getComputedStyle(document.documentElement).getPropertyValue(token).trim()')
    expect(resolveCanvasColor).not.toMatch(/getPropertyValue\(token\).*?(?:\|\||\?\?|fallback)/i)
    expect(drawPreview.match(/^\s*overlayCtx\.fillStyle = resolveCanvasColor\('--primary'\)\s*$/gm)).toHaveLength(1)
    expect(drawPreview).not.toMatch(/overlayCtx\.fillStyle\s*=.*(?:\|\||\?\?|fallback)/i)
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
    const style = extractSfcBlock(source, 'style')
    const mobileStyle = extractBraceBlock(style, '@media (max-width: 63.9375rem)')
    const mobileModeRule = findStyleRule(mobileStyle, '.mask-editor__mode :deep(.el-radio-button__inner)')

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
    expect(mobileModeRule).toContain('min-height: 44px;')
    expect(source).toContain("emit('save', maskCanvas.toDataURL('image/png'))")
  })

  it('keeps the mask dialog close control named and its footer inside the viewport', () => {
    const source = readView('ImageWorkbenchMaskEditor.vue')
    const template = extractSfcBlock(source, 'template')
    const style = extractSfcBlock(source, 'style')
    const dialogOpening = template.match(/<el-dialog\b[\s\S]*?>/)?.[0] || ''
    const header = template.match(/<template #header="\{ close, titleId, titleClass \}">[\s\S]*?<\/template>/)?.[0] || ''
    const rootRule = findStyleRule(style, ':global(.mask-editor)')
    const chromeRule = style.match(/:global\(\.mask-editor \.el-dialog__header\),\s*:global\(\.mask-editor \.el-dialog__footer\)\s*\{([^}]*)\}/s)?.[1] || ''
    const bodyRule = findStyleRule(style, ':global(.mask-editor .el-dialog__body)')

    expect(source).toMatch(/import \{(?=[^}]*\bClose\b)(?=[^}]*\bDelete\b)[^}]*\} from '@element-plus\/icons-vue'/)
    expect(dialogOpening).toContain('v-model="visible"')
    expect(dialogOpening).toContain('width="min(96vw, 1080px)"')
    expect(dialogOpening).toContain(':show-close="false"')
    expect(dialogOpening).toContain(':close-on-click-modal="false"')
    expect(header).toContain(':id="titleId"')
    expect(header).toContain(':class="titleClass"')
    expect(header).toContain('class="el-dialog__headerbtn"')
    expect(header).toContain('aria-label="关闭此对话框"')
    expect(header).toContain('title="关闭此对话框"')
    expect(header).toContain('@click="close"')
    expect(header).toContain('aria-hidden="true"')
    expect(header).toContain('<Close />')
    expect(template.match(/class="el-dialog__headerbtn"/g)).toHaveLength(1)
    expect(rootRule).toContain('max-height: 92vh;')
    expect(rootRule).toContain('display: flex;')
    expect(rootRule).toContain('flex-direction: column;')
    expect(style).toContain(':global(.mask-editor .el-dialog__header),')
    expect(style).toContain(':global(.mask-editor .el-dialog__footer) {')
    expect(chromeRule).toContain('flex: 0 0 auto;')
    expect(bodyRule).toContain('min-height: 0;')
    expect(bodyRule).toContain('overflow: auto;')
    expect(bodyRule).toContain('overscroll-behavior: contain;')
    expect(style).not.toMatch(/(^|\n)\s*\.mask-editor\s*\{/)
    expect(style).not.toContain('.mask-editor :deep(.el-dialog__body)')
  })

  it('uses the shared single close entry for the library drawer', () => {
    const source = readView('ToolboxImageWorkbench.vue')
    const template = extractSfcBlock(source, 'template')
    const style = extractSfcBlock(source, 'style')
    const mobileStyle = extractBraceBlock(style, '@media (max-width: 63.9375rem)')
    const drawer = template.match(/<el-drawer\b[\s\S]*?<\/el-drawer>/)?.[0] || ''
    const drawerOpening = drawer.match(/<el-drawer\b[\s\S]*?>/)?.[0] || ''
    const header = drawer.match(/<template #header="\{ close, titleId, titleClass \}">[\s\S]*?<\/template>/)?.[0] || ''
    const closeSelector = ':global(.image-workbench-library-drawer .el-drawer__close-btn)'
    const desktopCloseRule = findStyleRule(style, closeSelector)
    const mobileCloseRule = findStyleRule(mobileStyle, closeSelector)

    expect(source).toContain("import AdminDrawerHeader from '../components/base/AdminDrawerHeader.vue'")
    expect(drawerOpening).toContain('v-model="libraryPickerVisible"')
    expect(drawerOpening).toContain('class="image-workbench-library-drawer"')
    expect(drawerOpening).toContain('title="选择图库参考图"')
    expect(drawerOpening).toContain('size="min(100%, 760px)"')
    expect(drawerOpening).toContain(':show-close="false"')
    expect(drawerOpening).toContain(':close-on-click-modal="!libraryAdding"')
    expect(drawerOpening).toContain(':close-on-press-escape="!libraryAdding"')
    expect(drawerOpening).toContain('@closed="onLibraryPickerClosed"')
    expect(header).toContain('<AdminDrawerHeader')
    expect(header).toContain('title="选择图库参考图"')
    expect(header).toContain(':title-id="titleId"')
    expect(header).toContain(':title-class="titleClass"')
    expect(header).toContain(':close="close"')
    expect(template.match(/<AdminDrawerHeader\b/g)).toHaveLength(1)
    for (const [rule, size] of [[desktopCloseRule, 36], [mobileCloseRule, 44]]) {
      for (const property of ['width', 'height', 'min-width', 'min-height']) {
        expect(rule).toMatch(new RegExp(`(?:^|\\n)\\s*${property}:\\s*${size}px;`))
      }
    }
  })

  it('uses stable compact media-grid geometry in the image workbench', () => {
    const source = readView('ToolboxImageWorkbench.vue')
    const resultGridRule = source.match(/\.result-grid\s*\{[^}]*\}/s)?.[0] || ''
    const libraryGridRule = source.match(/\.library-grid\s*\{[^}]*\}/s)?.[0] || ''
    const template = extractSfcBlock(source, 'template')
    const inputOpening = template.match(/<aside class="workbench-panel workbench-panel--input"[^>]*>/)?.[0] || ''

    expect(source).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(source).toContain('<StatusIndicator :label="statusLabel" :tone="statusType" />')
    expect(source).toContain('<section class="workbench-panel workbench-panel--preview" data-density="compact">')
    expect(source).toContain('<div class="library-picker" data-density="compact">')
    expect(source.match(/data-density="compact"/g)).toHaveLength(2)
    expect(inputOpening).not.toContain('data-density')
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
