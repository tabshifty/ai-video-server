import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const source = readFileSync(new URL('./ToolboxArchiveImport.vue', import.meta.url), 'utf8')
const adminDrawerHeaderSource = readFileSync(new URL('../components/base/AdminDrawerHeader.vue', import.meta.url), 'utf8')

function extractTemplate(sfc) {
  const opening = sfc.match(/<template[^>]*>/)
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const end = sfc.lastIndexOf('</template>')

  return end > start ? sfc.slice(start, end) : ''
}

function extractStyle(sfc) {
  const match = sfc.match(/<style scoped>([\s\S]*?)<\/style>/)
  return match?.[1] || ''
}

function cssDeclarationPattern(property, value) {
  const escapeRegExp = (input) => input.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  return new RegExp(`(?:^|[;{])\\s*${escapeRegExp(property)}\\s*:\\s*${escapeRegExp(value)}\\s*;`)
}

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

  it('keeps the archive file picker bound to an explicit file list before upload', () => {
    expect(source).toContain('v-model:file-list="uploadFiles"')
    expect(source).toContain(':on-change="onUploadChange"')
    expect(source).toContain(':on-remove="onUploadRemove"')
    expect(source).toContain('const uploadFiles = ref([])')
    expect(source).toContain('const input = uploadFiles.value[0]?.raw')
    expect(source).not.toContain('uploadRef.value?.files?.length')
    expect(source).not.toContain('uploadRef.value?.uploadFiles?.length')
  })

  it('supports creating video or image collections in place and filling selectors back', () => {
    expect(source).toContain('createAdminCollection')
    expect(source).toContain('createAdminImageCollection')
    expect(source).toContain('openCreateVideoCollection')
    expect(source).toContain('openCreateImageCollection')
    expect(source).toContain('saveQuickCollection')
    expect(source).toContain('新建视频合集')
    expect(source).toContain('新建图片合集')
    expect(source).toContain('pushSelectedCollectionValue(target, created.id)')
  })

  it('shows localized archive file type and skipped reason in the file list', () => {
    expect(source).toContain('formatArchiveFileType')
    expect(source).toContain('archiveMediaKindLabel')
    expect(source).toContain('formatArchiveReason')
    expect(source).toContain('class="archive-file-item__meta"')
    expect(source).toContain('v-if="file.reason"')
    expect(source).not.toContain('{{ file.media_kind }} · {{ formatFileSize(file.file_size) }}')
  })

  it('lets the archive file list toggle between original order and type sorting', () => {
    expect(source).toContain("const fileSortMode = ref('original')")
    expect(source).toContain('const filteredBatchFiles = computed(() => filterArchiveFilesByGroup(selectedBatchFiles.value, activeGroupFilter.value))')
    expect(source).toContain('const displayedBatchFiles = computed(() => sortArchiveFiles(filteredBatchFiles.value))')
    expect(source).toContain('function sortArchiveFiles(files)')
    expect(source).toContain("'按原始顺序'")
    expect(source).toContain("'按类型排序'")
    expect(source).toContain('<el-segmented v-model="fileSortMode"')
    expect(source).toContain('v-for="file in displayedBatchFiles"')
  })

  it('supports selecting archive files, fast type selection and batch-processing the selected items', () => {
    expect(source).toContain('const selectedFileIDs = ref([])')
    expect(source).toContain('const bulkActions = computed(() => [')
    expect(source).toContain('processSelectionActionLabel')
    expect(source).toContain('function onArchiveFileSelectToggle(row, event)')
    expect(source).toContain('const shouldSelectRange = !isSelected')
    expect(source).toContain('if (shouldSelectRange) {')
    expect(source).toContain('function processSelectedArchiveFiles()')
    expect(source).toContain("selectArchiveFilesByKind('video')")
    expect(source).toContain("selectArchiveFilesByKind('image')")
    expect(source).toContain('BulkActionBar')
    expect(source).toContain('selectedBatchFileCount')
    expect(source).toContain('shiftKey')
    expect(source).toContain(":class=\"{ 'is-selected': selectedFileIDs.includes(String(file.id)) }\"")
    expect(source).toContain('border: 1px solid var(--line-strong);')
    expect(source).toContain('左侧勾选位支持多选和 Shift 连选')
  })

  it('adds a persistent group workspace with ungrouped cards and batch move actions', () => {
    expect(source).toContain('const selectedBatchGroups = ref([])')
    expect(source).toContain("const ARCHIVE_GROUP_FILTER_UNGROUPED = '__ungrouped__'")
    expect(source).toContain('archiveGroupCards')
    expect(source).toContain('未分组固定置顶')
    expect(source).toContain('创建分组')
    expect(source).toContain('加入分组')
    expect(source).toContain('移出分组')
    expect(source).toContain('成员')
    expect(source).toContain('未入库')
    expect(source).toContain('已入库')
    expect(source).toContain("file.group_name || '未分组'")
  })

  it('lets archive video files link to one image collection while images can join image collections', () => {
    expect(source).toContain('selectedVideoImageCollectionID')
    expect(source).toContain('batchEditVideoImageCollectionID')
    expect(source).toContain('仅可关联一个图片图集')
    expect(source).toContain('视频关联的图片合集')
    expect(source).toContain('图片入库后加入的合集')
  })

  it('uses batch title as the default video title and keeps explicit batch title modes available', () => {
    expect(source).toContain('视频默认标题')
    expect(source).toContain('可不填，默认取压缩包文件名')
    expect(source).toContain("title_mode: 'none'")
    expect(source).toContain('batchEditForm.title')
    expect(source).toContain('统一覆盖为同一个标题；留空回到视频默认标题')
    expect(source).not.toContain('title_enabled')
    expect(source).not.toContain('批次标题')
  })

  it('adds filename title drafts without coupling them to process actions', () => {
    expect(source).toContain("{ label: '不修改', value: 'none' }")
    expect(source).toContain("{ label: '统一标题', value: 'uniform' }")
    expect(source).toContain("{ label: '按各自文件名替换', value: 'filename' }")
    expect(source).toContain('DocumentCopy')
    expect(source).toContain('canReplaceArchiveFilenameTitle(selectedFile)')
    expect(source).toContain('applySelectedFilenameTitleDraft')
    expect(source).toContain('使用文件名')
    expect(source).toContain('@input="deactivateSelectedFilenameMode"')
    expect(source).toContain('标题已与文件名一致')
    expect(source).toContain('const persistedFile = findArchiveFileByID(selectedFile.value.id)')
    expect(source).toContain('archiveFilenameEditorMatchesPersistedText(selectedFile.value, persistedFile)')
    expect(source).toContain('标题或说明有未保存修改，请先保存或撤销后再使用文件名')
    expect(source).toContain('buildArchiveFilenameTitleDraft(persistedFile)')
    expect(source).not.toContain('processSelectedArchiveFiles()\n  applySelectedFilenameTitleDraft')
    expect(source).not.toContain('processAdminArchiveImportFile(file.id)\n        applySelectedFilenameTitleDraft')
  })

  it('routes an active single-file filename draft through one semantic request and preserves manual saves', () => {
    expect(source).toContain('const selectedFilenameDraftSnapshot = ref(null)')
    expect(source).toContain('title: String(selectedFile.value.title || \'\')')
    expect(source).toContain('description: String(selectedFile.value.description || \'\')')
    expect(source).toContain('isArchiveFilenameDraftSnapshotCurrent(selectedFilenameDraftSnapshot.value, selectedFile.value)')
    expect(source).toContain('if (selectedFilenameDraftSnapshot.value && !filenameModeActive)')
    expect(source).toContain('if (filenameModeActive)')
    expect(source).toContain('buildArchiveFilenameBatchPayload([selectedFile.value], {')
    expect(source).toContain('update_tags: true')
    expect(source).toContain('update_video_type: true')
    expect(source).toContain('update_video_collection_ids: true')
    expect(source).toContain('update_image_collection_ids: true')
    expect(source).toContain('await batchUpdateAdminArchiveImportFiles(filenamePayload)')
    expect(source).toContain('await updateAdminArchiveImportFile(selectedFile.value.id, payload)')
    expect(source).toContain("error?.data?.reason === 'stale_target'")
    expect(source).toContain("error?.data?.reason === 'ineligible_target'")
  })

  it('uses filename as a mutually exclusive batch title mode with a bounded issue-aware preview', () => {
    expect(source).toContain('<el-segmented')
    expect(source).toContain('v-model="batchEditForm.title_mode"')
    expect(source).toContain('@change="onBatchTitleModeChange"')
    expect(source).toContain("batchEditForm.title_mode === 'filename'")
    expect(source).toContain(':disabled="batchEditForm.title_mode === \'filename\' || !batchEditForm.description_enabled"')
    expect(source).toContain('buildArchiveFilenameBatchPreview(selectedBatchFilesForActions.value, 5)')
    expect(source).toContain('batchFilenamePreview.total')
    expect(source).toContain('batchFilenamePreview.remaining')
    expect(source).toContain('batchFilenameIssuePreview.items')
    expect(source).toContain('batchFilenameIssuePreview.remaining')
    expect(source).toContain('batchFilenamePreview.value.issues.slice(0, 5)')
    expect(source).toContain('issue.relative_path')
    expect(source).toContain('原标题')
    expect(source).toContain('新标题')
  })

  it('saves filename batches transactionally while leaving uniform mode on the existing loop', () => {
    expect(source).toContain('saveArchiveFilenameBatchUpdate')
    expect(source).toContain("if (batchEditForm.title_mode === 'filename')")
    expect(source).toContain('await batchUpdateAdminArchiveImportFiles(payload)')
    expect(source).toContain('buildArchiveFilenameBatchOptionalPatch({')
    expect(source).toContain('for (const file of targets) {')
    expect(source).toContain("batchEditForm.title_mode === 'uniform'")
    expect(source).toContain('已更新 ${updatedCount} 个视频')
    expect(source).toContain(':disabled="batchEditForm.title_mode === \'filename\' && !batchFilenamePreview.ok"')
  })

  it('keeps dialogs and drafts open when committed updates cannot refresh authority', () => {
    expect(source).toContain('executeArchiveFilenameUpdate')
    expect(source.match(/await executeArchiveFilenameUpdate\(\{/g)).toHaveLength(2)
    expect(source.match(/outcome\.status === 'refresh_failed'/g)).toHaveLength(2)
    expect(source.match(/outcome\.status === 'success'/g)).toHaveLength(2)
    expect(source).toContain('文件信息已保存，但刷新失败，请手动刷新')
    expect(source).toContain('更新已提交，但刷新失败，请手动刷新')
    expect(source).toContain('const result = await batchUpdateAdminArchiveImportFiles(payload)')
    expect(source).toContain('outcome.data?.updated_count')
    expect(source).toContain('Number.isInteger(apiUpdatedCount)')
    expect(source).toContain('已更新 ${updatedCount} 个视频')
  })

  it('keeps the title mode and preview stable in the dense responsive dialog', () => {
    expect(source).toContain('class="archive-title-mode"')
    expect(source).toContain('<el-form-item label="标题处理" class="archive-title-mode-item">')
    expect(source).toContain('class="archive-title-preview"')
    expect(source).toContain('class="archive-title-preview__row"')
    expect(source).toContain('class="archive-title-preview__issues"')
    expect(source).toContain('grid-template-columns: repeat(2, minmax(0, 1fr));')
    expect(source).toContain('@media (max-width: 40rem)')
    expect(source).toMatch(/@media \(max-width: 40rem\)[\s\S]*?\.archive-title-preview__row[\s\S]*?grid-template-columns: 1fr;/)
    expect(source).toMatch(/@media \(max-width: 40rem\)[\s\S]*?\.archive-title-mode-item \{[\s\S]*?display: block;[\s\S]*?\.archive-title-mode-item :deep\(\.el-form-item__label\)[\s\S]*?width: auto !important;[\s\S]*?\.archive-title-mode-item :deep\(\.el-form-item__content\)[\s\S]*?margin-left: 0 !important;/)
  })

  it('keeps the page batch-first by moving upload into a dialog and batch detail into a drawer', () => {
    expect(source).toContain('uploadDialogVisible')
    expect(source).toContain('batchDrawerVisible')
    expect(source).toContain('title="上传压缩包"')
    expect(source).toContain('title="批次详情"')
    expect(source).toContain('batchEditDialogVisible')
    expect(source).toContain('批量编辑')
    expect(source).toContain('上传成功后会自动打开新批次详情')
  })

  it('moves single-file editing into a row-level dialog without mixing in process actions', () => {
    expect(source).toContain('const selectedFileDialogVisible = ref(false)')
    expect(source).toContain('async function openArchiveFileEditor(row)')
    expect(source).toContain('requestSelectedFileDialogClose')
    expect(source).toContain('handleSelectedFileDialogBeforeClose')
    expect(source).toContain('handleSelectedFileDialogClosed')
    expect(source).toContain('v-model="selectedFileDialogVisible"')
    expect(source).toContain('@click.stop="openArchiveFileEditor(file)"')
    expect(source).toContain('class="archive-file-item__edit"')
    expect(source).toContain('canEditArchiveFile(file)')
    expect(source).toContain('直接点文件行右侧“编辑”')
    expect(source).toContain('处理动作仍走下方主动作区')
    expect(source).toContain('saveSelectedFile')
    expect(source).not.toContain('<SectionCard v-if="shouldShowSingleFileEditor"')
    expect(source).not.toContain('单文件精修')
    expect(source).not.toContain('处理文件</el-button>')
  })

  it('allows deleting archive batches from the batch list with confirmation', () => {
    expect(source).toContain('deleteAdminArchiveImportBatch')
    expect(source).toContain('function canDeleteArchiveBatch(batch)')
    expect(source).toContain('async function removeArchiveBatch(batch)')
    expect(source).toContain("'删除批次'")
    expect(source).toContain('删除后会清空该批次的压缩包记录、文件清单和解包目录')
    expect(source).toContain('archive-batch-card__actions')
  })

  it('uses a standalone form workspace and a non-card metric strip without scope hints', () => {
    const template = extractTemplate(source)
    const overviewBlock = source.match(/const overviewCards = computed\(\(\) => \{[\s\S]*?\n\}\)\nconst selectionAlert/)?.[0] || ''
    const metricStrip = template.match(/<MetricStrip\b[\s\S]*?\/>/)?.[0] || ''

    expect(source).toContain("import MetricStrip from '../components/base/MetricStrip.vue'")
    expect(template.trimStart()).toMatch(/^<main class="tool-workspace archive-import-tool" data-density="form">/)
    expect(metricStrip).toContain(':items="overviewCards"')
    expect(metricStrip).toContain('aria-label="压缩包批次摘要"')
    expect(metricStrip).not.toContain('scope')
    expect(template).toContain('class="archive-overview-note"')
    expect(template).toContain('按上传时间倒序；待继续处理包含待处理、失败或待纠偏批次；待纠偏可补密码或确认编码后继续解包；处理中表示后台仍在解包或入库。')
    expect(template).not.toContain('archive-overview-grid')
    expect(template).not.toContain('archive-overview-card')
    expect(overviewBlock).toContain('const needingAction = batches.value.filter((batch) => batchNeedsAction(batch)).length')
    expect(overviewBlock).toContain("const needExtractRetry = batches.value.filter((batch) => batch.status === 'needs_password' || batch.status === 'needs_encoding').length")
    expect(overviewBlock).toContain("const processing = batches.value.filter((batch) => batch.status === 'processing').length")
    expect(overviewBlock).toContain("{ key: '批次总数', label: '批次总数', value: batches.value.length }")
    expect(overviewBlock).toContain("{ key: '待继续处理', label: '待继续处理', value: needingAction }")
    expect(overviewBlock).toContain("{ key: '待纠偏', label: '待纠偏', value: needExtractRetry }")
    expect(overviewBlock).toContain("{ key: '处理中', label: '处理中', value: processing }")
    expect(overviewBlock).not.toContain('hint:')
  })

  it('keeps compact collections and all existing Teleport surfaces at form density', () => {
    const template = extractTemplate(source)
    const dialogs = template.match(/<el-dialog\b[\s\S]*?>/g) || []
    const drawers = template.match(/<el-drawer\b[\s\S]*?>/g) || []

    expect(template).toContain('<SectionCard class="archive-batch-panel" data-density="compact">')
    expect(template).toContain('<SectionCard class="archive-file-panel" data-density="compact">')
    expect(dialogs).toHaveLength(5)
    expect(drawers).toHaveLength(1)
    dialogs.forEach((dialog) => expect(dialog).toContain('data-density="form"'))
    drawers.forEach((drawer) => expect(drawer).toContain('data-density="form"'))

    expect(drawers[0]).toContain('v-model="batchDrawerVisible"')
    expect(drawers[0]).toContain(':before-close="handleBatchDrawerBeforeClose"')
    expect(drawers[0]).toContain('@closed="handleBatchDrawerClosed"')
    expect(template).toContain('v-model="selectedFileDialogVisible"')
    expect(template).toContain(':before-close="handleSelectedFileDialogBeforeClose"')
    expect(template).toContain('@closed="handleSelectedFileDialogClosed"')
    expect(template).toContain('v-model="uploadDialogVisible"')
    expect(template).toContain('v-model="batchEditDialogVisible"')
    expect(template).toContain(':before-close="handleBatchEditBeforeClose"')
    expect(template).toContain('v-model="archiveGroupDialogVisible"')
    expect(template).toContain(':before-close="handleArchiveGroupDialogBeforeClose"')
    expect(template).toContain('v-model="quickCollectionDialogVisible"')
    expect(template).toContain('<BulkActionBar :count="selectedFileIDs.length" :actions="bulkActions" />')
  })

  it('uses a 52px minimum media row and exposes the full skipped reason through a focusable tooltip', () => {
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const fileItemRule = style.match(/\.archive-file-item\s*\{[^}]*\}/s)?.[0] || ''
    const reasonRule = style.match(/\.archive-file-item__reason\s*\{[^}]*\}/s)?.[0] || ''
    const narrowReasonRule = style.match(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-file-item__reason\s*\{[^}]*\}/s)?.[0] || ''

    expect(fileItemRule).toContain('min-height: var(--media-row-height);')
    expect(fileItemRule).not.toMatch(/(?:^|[;{]\s*)height:/)
    expect(template).toMatch(/<el-tooltip\s+v-if="file\.reason"\s+:content="formatArchiveReason\(file\.reason\)"[\s\S]*?<span class="archive-file-item__reason" tabindex="0">\{\{ formatArchiveReason\(file\.reason\) \}\}<\/span>[\s\S]*?<\/el-tooltip>/)
    expect(reasonRule).toContain('display: block;')
    expect(reasonRule).toContain('max-width: 100%;')
    expect(reasonRule).toContain('overflow: hidden;')
    expect(reasonRule).toContain('text-overflow: ellipsis;')
    expect(reasonRule).toContain('white-space: nowrap;')
    expect(reasonRule).toMatch(cssDeclarationPattern('min-height', 'var(--control-height)'))
    expect(narrowReasonRule).toMatch(cssDeclarationPattern('min-height', '44px'))
  })

  it('sizes the file selection target from compact density and raises every dimension to 44px below desktop', () => {
    const style = extractStyle(source)
    const selectionRule = style.match(/\.archive-file-item__selection\s*\{[^}]*\}/s)?.[0] || ''
    const narrowSelectionRule = style.match(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-file-item__selection\s*\{[^}]*\}/s)?.[0] || ''

    for (const property of ['width', 'height', 'min-width', 'min-height', 'flex-basis']) {
      expect(selectionRule, property).toMatch(cssDeclarationPattern(property, 'var(--control-height)'))
      expect(narrowSelectionRule, property).toMatch(cssDeclarationPattern(property, '44px'))
    }
  })

  it('opens the skipped reason tooltip on keyboard focus and provides a visible focus ring', () => {
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const reasonTooltip = template.match(/<el-tooltip\s+v-if="file\.reason"[\s\S]*?<\/el-tooltip>/)?.[0] || ''
    const reasonFocusRule = style.match(/\.archive-file-item__reason:focus-visible\s*\{[^}]*\}/s)?.[0] || ''

    expect(reasonTooltip).toContain(':content="formatArchiveReason(file.reason)"')
    expect(reasonTooltip).toContain('trigger="focus"')
    expect(reasonTooltip).toContain('<span class="archive-file-item__reason" tabindex="0">{{ formatArchiveReason(file.reason) }}</span>')
    expect(reasonFocusRule).toContain('outline: 2px solid var(--line-focus);')
    expect(reasonFocusRule).toContain('outline-offset: 2px;')
  })

  it('gives only the batch Drawer close button a 36px desktop and 44px narrow target', () => {
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const drawer = template.match(/<el-drawer\b[\s\S]*?>/)?.[0] || ''
    const closeRule = style.match(/:global\(\.archive-batch-drawer \.el-drawer__close-btn\)\s*\{[^}]*\}/s)?.[0] || ''
    const narrowCloseRule = style.match(/@media \(max-width: 63\.9375rem\)[\s\S]*?:global\(\.archive-batch-drawer \.el-drawer__close-btn\)\s*\{[^}]*\}/s)?.[0] || ''

    expect(drawer).toContain('class="archive-batch-drawer"')
    expect(drawer).toContain('v-model="batchDrawerVisible"')
    expect(drawer).toContain('data-density="form"')
    for (const property of ['width', 'height', 'min-width', 'min-height']) {
      expect(closeRule, property).toMatch(cssDeclarationPattern(property, '36px'))
      expect(narrowCloseRule, property).toMatch(cssDeclarationPattern(property, '44px'))
    }
  })

  it('uses the shared labeled Drawer header without rendering the framework close button', () => {
    const template = extractTemplate(source)
    const drawer = template.match(/<el-drawer\b[\s\S]*?>/)?.[0] || ''
    const drawerHeader = template.match(/<el-drawer\b[\s\S]*?>\s*<template #header="\{ close, titleId, titleClass \}">[\s\S]*?<\/template>/)?.[0] || ''

    expect(source).toContain("import AdminDrawerHeader from '../components/base/AdminDrawerHeader.vue'")
    for (const contract of [
      'v-model="batchDrawerVisible"',
      'class="archive-batch-drawer"',
      'title="批次详情"',
      'direction="rtl"',
      ':size="batchDrawerSize"',
      'destroy-on-close',
      ':show-close="false"',
      ':before-close="handleBatchDrawerBeforeClose"',
      '@closed="handleBatchDrawerClosed"',
      'data-density="form"'
    ]) {
      expect(drawer, contract).toContain(contract)
    }
    expect(drawerHeader).toContain('<template #header="{ close, titleId, titleClass }">')
    expect(drawerHeader).toContain('<AdminDrawerHeader title="批次详情" :title-id="titleId" :title-class="titleClass" :close="close" />')
    expect(template.match(/<AdminDrawerHeader\b/g)).toHaveLength(1)
    expect(template.match(/<el-dialog\b/g)).toHaveLength(5)
    expect(adminDrawerHeaderSource).toContain('class="el-drawer__close-btn"')
    expect(adminDrawerHeaderSource).toContain('aria-label="关闭此对话框"')
    expect(adminDrawerHeaderSource).toContain('title="关闭此对话框"')
    expect(adminDrawerHeaderSource).toContain('@click="close"')
  })

  it('keeps file sort segmented items at compact desktop height and 44px below desktop', () => {
    const style = extractStyle(source)
    const desktopStyle = style.slice(0, style.indexOf('@media'))
    const desktopItemRule = desktopStyle.match(/\.archive-file-sort :deep\(\.el-segmented__item\)\s*\{[^}]*\}/s)?.[0] || ''
    const narrowItemRule = style.match(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-file-sort :deep\(\.el-segmented__item\)\s*\{[^}]*\}/s)?.[0] || ''

    expect(desktopItemRule).toContain('min-height: var(--control-height);')
    expect(narrowItemRule).toContain('min-height: 44px;')
  })

  it('removes decorative gradients and contains dynamic content at 768px and 375px', () => {
    const style = extractStyle(source)
    const rootRule = style.match(/\.tool-workspace\s*\{[^}]*\}/s)?.[0] || ''
    const innerRule = style.match(/\.tool-workspace__inner\s*\{[^}]*\}/s)?.[0] || ''

    expect(source).not.toMatch(/(?:linear|radial)-gradient\(/)
    expect(rootRule).toContain('min-width: 0;')
    expect(rootRule).toContain('overflow-x: clip;')
    expect(innerRule).toContain('min-width: 0;')
    expect(style).toMatch(/\.archive-drawer,[\s\S]*?\.quick-collection-form\s*\{[^}]*max-width:\s*100%;/s)
    expect(style).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-file-item__selection\s*\{[^}]*min-width:\s*44px;[^}]*min-height:\s*44px;/s)
    expect(style).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-group-panel__actions,[\s\S]*?\.archive-drawer__hero-actions\s*\{[^}]*width:\s*100%;[^}]*flex-wrap:\s*wrap;/s)
    expect(style).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-import-tool :deep\(\.page-header-shell__actions\),[\s\S]*?\.archive-file-panel :deep\(\.section-card__actions\)\s*\{[^}]*width:\s*100%;[^}]*margin-left:\s*0;[^}]*flex-wrap:\s*wrap;/s)
    expect(style).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-batch-panel :deep\(\.section-card__header\),[\s\S]*?\.archive-file-panel :deep\(\.section-card__header\)\s*\{[^}]*flex-wrap:\s*wrap;/s)
    expect(style).toMatch(/@media \(max-width: 63\.9375rem\)[\s\S]*?\.archive-drawer :deep\(\.bulk-action-bar\)\s*\{[^}]*flex-direction:\s*column;[\s\S]*?\.archive-drawer :deep\(\.bulk-action-bar__actions\)\s*\{[^}]*width:\s*100%;[^}]*flex-wrap:\s*wrap;/s)
    expect(style).toMatch(/@media \(max-width: 48rem\)[\s\S]*?\.archive-group-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\);/s)
    expect(style).toMatch(/@media \(max-width: 48rem\)[\s\S]*?\.archive-file-sort\s*\{[^}]*min-width:\s*0;[^}]*width:\s*100%;/s)
    expect(style).toMatch(/@media \(max-width: 48rem\)[\s\S]*?\.archive-file-editor__form :deep\(\.el-form-item__content\),[\s\S]*?\.quick-collection-form :deep\(\.el-form-item__content\)\s*\{[^}]*margin-left:\s*0 !important;[^}]*min-width:\s*0;/s)
  })

  it('keeps the critical archive commands and dirty guards on their original handlers', () => {
    const template = extractTemplate(source)

    for (const command of [
      '@click="uploadArchive"',
      '@click="runRetryExtract"',
      '@click="saveSelectedFile"',
      '@click="saveBatchEdit"',
      '@click="saveArchiveGroup"',
      '@click.stop="onArchiveFileSelectToggle(file, $event)"'
    ]) {
      expect(template, command).toContain(command)
    }
    expect(source).toContain('onClick: processSelectedArchiveFiles')
    expect(source).toContain('confirmDiscardUnsavedFileChanges')
    expect(source).toContain('requestSelectedFileDialogClose')
    expect(source).toContain('requestBatchEditClose')
    expect(source).toContain('requestArchiveGroupDialogClose')
  })
})
