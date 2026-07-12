const archiveFilenameAdPattern = /www\.98T\.la@/gi
const archiveFilenameTitleMaxCodePoints = 200

function archiveFileText(value) {
  return String(value ?? '')
}

function archiveFilenameIssue(file, message) {
  return {
    id: archiveFileText(file?.id),
    relative_path: archiveFileText(file?.relative_path),
    message
  }
}

export function deriveArchiveFilenameTitle(relativePath) {
  const normalized = archiveFileText(relativePath).trim().replace(/\\/g, '/')
  const filename = normalized.split('/').pop() || ''
  const dotIndex = filename.lastIndexOf('.')
  const basename = dotIndex >= 0 ? filename.slice(0, dotIndex) : filename

  return basename.replace(archiveFilenameAdPattern, '').trim().replace(/\s+/g, ' ')
}

export function canReplaceArchiveFilenameTitle(file) {
  return file?.entry_type === 'file'
    && file?.media_kind === 'video'
    && (file?.status === 'pending' || file?.status === 'failed')
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

  const originalTitle = currentTitle.trim()
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
        old_title: archiveFileText(file?.title).trim(),
        new_title: draft.title
      })
    }

    if (!archiveFileText(file?.id).trim()) {
      issues.push(archiveFilenameIssue(file, '文件 ID 不能为空'))
    }
    if (!archiveFileText(file?.updated_at).trim()) {
      issues.push(archiveFilenameIssue(file, '文件更新时间不能为空'))
    }
  }

  const firstVisibleBatchID = selectedFiles
    .map((file) => archiveFileText(file?.batch_id).trim())
    .find(Boolean)
  if (firstVisibleBatchID) {
    for (const file of selectedFiles) {
      const batchID = archiveFileText(file?.batch_id).trim()
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

export function buildArchiveFilenameBatchPayload(files, patch = {}) {
  return {
    ...patch,
    targets: buildArchiveFilenameBatchTargets(files),
    title_mode: 'filename'
  }
}
