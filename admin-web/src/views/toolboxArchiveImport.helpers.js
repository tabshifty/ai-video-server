const archiveFilenameAdPattern = /www\.98T\.la@/gi
const archiveFilenameTitleMaxCodePoints = 200
// Mirrors Go unicode.IsSpace (Unicode White_Space), which excludes U+FEFF.
const goWhitespaceEdgesPattern = /^[\u0009-\u000D\u0020\u0085\u00A0\u1680\u2000-\u200A\u2028-\u2029\u202F\u205F\u3000]+|[\u0009-\u000D\u0020\u0085\u00A0\u1680\u2000-\u200A\u2028-\u2029\u202F\u205F\u3000]+$/gu
const goWhitespaceRunPattern = /[\u0009-\u000D\u0020\u0085\u00A0\u1680\u2000-\u200A\u2028-\u2029\u202F\u205F\u3000]+/gu

function archiveFileText(value) {
  return String(value ?? '')
}

function trimGoWhitespace(value) {
  return value.replace(goWhitespaceEdgesPattern, '')
}

function archiveFilenameIssue(file, message) {
  return {
    id: archiveFileText(file?.id),
    relative_path: archiveFileText(file?.relative_path),
    message
  }
}

export function deriveArchiveFilenameTitle(relativePath) {
  const normalized = trimGoWhitespace(archiveFileText(relativePath)).replace(/\\/g, '/')
  const pathWithoutTrailingSlashes = normalized.replace(/\/+$/, '')
  if (!pathWithoutTrailingSlashes) return ''

  const filename = pathWithoutTrailingSlashes.split('/').pop() || ''
  const dotIndex = filename.lastIndexOf('.')
  const basename = dotIndex >= 0 ? filename.slice(0, dotIndex) : filename

  return trimGoWhitespace(basename.replace(archiveFilenameAdPattern, ''))
    .replace(goWhitespaceRunPattern, ' ')
}

export function canReplaceArchiveFilenameTitle(file) {
  return file?.entry_type === 'file'
    && file?.media_kind === 'video'
    && (file?.status === 'pending' || file?.status === 'failed')
}

export function archiveFilenameEditorMatchesPersistedText(editorFile, persistedFile) {
  if (!editorFile || !persistedFile) return false
  return archiveFileText(editorFile.title) === archiveFileText(persistedFile.title)
    && archiveFileText(editorFile.description) === archiveFileText(persistedFile.description)
}

export function isArchiveFilenameDraftSnapshotCurrent(snapshot, file) {
  if (!snapshot || !file) return false
  return ['id', 'updated_at', 'relative_path', 'title', 'description'].every((field) =>
    Object.prototype.hasOwnProperty.call(snapshot, field)
      && archiveFileText(snapshot[field]) === archiveFileText(file[field])
  )
}

export function buildArchiveFilenameTitleDraft(file) {
  const currentTitle = archiveFileText(file?.title)
  const currentDescription = archiveFileText(file?.description)
  if (!canReplaceArchiveFilenameTitle(file)) {
    return {
      ok: false,
      title: currentTitle,
      description: currentDescription,
      changed: false,
      issue: archiveFilenameIssue(file, '仅待处理或失败的视频可以替换标题')
    }
  }

  const title = deriveArchiveFilenameTitle(file?.relative_path)
  if (!title || Array.from(title).length > archiveFilenameTitleMaxCodePoints) {
    return {
      ok: false,
      title: currentTitle,
      description: currentDescription,
      changed: false,
      issue: archiveFilenameIssue(file, '文件名无法生成有效标题')
    }
  }

  const originalTitle = trimGoWhitespace(currentTitle)
  const description = originalTitle && originalTitle !== title
    ? currentDescription === '' ? originalTitle : `${originalTitle}\n${currentDescription}`
    : currentDescription

  return {
    ok: true,
    title,
    description,
    changed: title !== originalTitle || description !== currentDescription,
    issue: null
  }
}

export function buildArchiveFilenameBatchPreview(files, limit = 5) {
  const selectedFiles = Array.isArray(files) ? files : []
  const issues = []
  const validItems = []

  for (const file of selectedFiles) {
    const draft = buildArchiveFilenameTitleDraft(file)
    if (!draft.ok) {
      issues.push(draft.issue)
    } else {
      validItems.push({
        id: archiveFileText(file?.id),
        relative_path: archiveFileText(file?.relative_path),
        old_title: trimGoWhitespace(archiveFileText(file?.title)),
        new_title: draft.title
      })
    }

    if (!trimGoWhitespace(archiveFileText(file?.id))) {
      issues.push(archiveFilenameIssue(file, '文件 ID 不能为空'))
    }
    if (!trimGoWhitespace(archiveFileText(file?.updated_at))) {
      issues.push(archiveFilenameIssue(file, '文件更新时间不能为空'))
    }
  }

  const firstVisibleBatchID = selectedFiles
    .map((file) => trimGoWhitespace(archiveFileText(file?.batch_id)))
    .find(Boolean)
  if (firstVisibleBatchID) {
    for (const file of selectedFiles) {
      const batchID = trimGoWhitespace(archiveFileText(file?.batch_id))
      if (batchID && batchID !== firstVisibleBatchID) {
        issues.push(archiveFilenameIssue(file, '所选文件不属于同一批次'))
      }
    }
  }

  const previewLimit = Number.isFinite(Number(limit))
    ? Math.max(Math.trunc(Number(limit)), 0)
    : 5

  return {
    ok: selectedFiles.length > 0 && issues.length === 0,
    total: selectedFiles.length,
    items: validItems.slice(0, previewLimit),
    remaining: Math.max(validItems.length - previewLimit, 0),
    issues
  }
}

export function buildArchiveFilenameBatchTargets(files) {
  return (Array.isArray(files) ? files : []).map((file) => ({
    id: archiveFileText(file?.id),
    updated_at: archiveFileText(file?.updated_at)
  }))
}

export function buildArchiveFilenameBatchOptionalPatch(options = {}) {
  const patch = {}
  if (options.tags_enabled) {
    patch.update_tags = true
    patch.tags = Array.isArray(options.tags) ? [...options.tags] : []
  }
  if (options.video_type_enabled) {
    patch.update_video_type = true
    patch.video_type = archiveFileText(options.video_type)
  }
  if (options.video_collection_enabled) {
    patch.update_video_collection_ids = true
    patch.video_collection_ids = Array.isArray(options.video_collection_ids) ? [...options.video_collection_ids] : []
  }
  if (options.video_image_collection_enabled) {
    patch.update_image_collection_ids = true
    patch.image_collection_ids = Array.isArray(options.image_collection_ids) ? [...options.image_collection_ids] : []
  }
  return patch
}

export function buildArchiveFilenameBatchPayload(files, patch = {}) {
  return {
    ...patch,
    targets: buildArchiveFilenameBatchTargets(files),
    title_mode: 'filename'
  }
}

export async function executeArchiveFilenameUpdate({ submit, refresh, isConflict, recoverConflict }) {
  let data
  try {
    data = await submit()
  } catch (error) {
    if (isConflict(error)) {
      await recoverConflict(error)
      return { status: 'conflict', data: null, error }
    }
    return { status: 'error', data: null, error }
  }

  try {
    const refreshed = await refresh()
    return {
      status: refreshed ? 'success' : 'refresh_failed',
      data,
      error: null
    }
  } catch (error) {
    return { status: 'refresh_failed', data, error }
  }
}
