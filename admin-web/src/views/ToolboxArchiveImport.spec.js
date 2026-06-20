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
    expect(source).toContain('{{ formatArchiveFileType(file) }}')
    expect(source).toContain('v-if="file.reason"')
    expect(source).not.toContain('{{ file.media_kind }} · {{ formatFileSize(file.file_size) }}')
  })
})
