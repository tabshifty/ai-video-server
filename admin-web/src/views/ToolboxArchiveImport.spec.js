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

  it('uses batch title as the default video title and keeps batch title overrides available', () => {
    expect(source).toContain('视频默认标题')
    expect(source).toContain('可不填，默认取压缩包文件名')
    expect(source).toContain('title_enabled')
    expect(source).toContain('batchEditForm.title')
    expect(source).toContain('统一覆盖为同一个标题；留空回到视频默认标题')
    expect(source).not.toContain('批次标题')
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

  it('keeps single-file editing as an exception path without a process button', () => {
    expect(source).toContain('这里只处理例外项修正；真正的处理动作统一留在下方主动作区。')
    expect(source).toContain('单文件精修')
    expect(source).toContain('saveSelectedFile')
    expect(source).not.toContain('processSelectedFile')
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
})
