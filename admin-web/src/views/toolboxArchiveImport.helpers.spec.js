import { describe, expect, it } from 'vitest'
import {
  buildArchiveFilenameBatchPayload,
  buildArchiveFilenameBatchPreview,
  buildArchiveFilenameBatchTargets,
  buildArchiveFilenameTitleDraft,
  canReplaceArchiveFilenameTitle,
  deriveArchiveFilenameTitle
} from './toolboxArchiveImport.helpers'

const updatedAt = '2026-07-12T05:00:00Z'

function pendingVideo(relativePath, overrides = {}) {
  return {
    id: 'file-1',
    batch_id: 'batch-1',
    updated_at: updatedAt,
    relative_path: relativePath,
    title: '旧标题',
    description: '原说明',
    entry_type: 'file',
    media_kind: 'video',
    status: 'pending',
    ...overrides
  }
}

describe('archive filename title derivation', () => {
  it.each([
    ['removes ad text', '目录/www.98T.la@ABC-123.mp4', 'ABC-123'],
    ['matches case insensitively and globally', 'WWW.98t.LA@ A www.98T.la@ B.mkv', 'A B'],
    ['keeps punctuation', '目录/www.98T.la@-ABC_[01].mp4', '-ABC_[01]'],
    ['keeps near match', '目录/98T.la@ABC.mp4', '98T.la@ABC'],
    ['collapses whitespace', '目录/www.98T.la@  A  B .mp4', 'A B'],
    ['returns empty after cleaning', '目录/www.98T.la@.mp4', '']
  ])('%s', (_, relativePath, expected) => {
    expect(deriveArchiveFilenameTitle(relativePath)).toBe(expected)
  })

  it('uses only the filename and removes only its last extension', () => {
    expect(deriveArchiveFilenameTitle('目录名.mp4/子目录\\www.98T.la@archive.part.name.tar.mp4')).toBe('archive.part.name.tar')
  })

  it('trims and collapses U+0085 like Go unicode.IsSpace', () => {
    const nextLine = '\u0085'

    expect(deriveArchiveFilenameTitle(`目录/${nextLine}www.98T.la@A${nextLine}${nextLine}B${nextLine}.mp4`)).toBe('A B')
  })

  it('preserves U+FEFF because Go unicode.IsSpace excludes it', () => {
    const byteOrderMark = '\uFEFF'

    expect(deriveArchiveFilenameTitle(`目录/${byteOrderMark}www.98T.la@A${byteOrderMark}${byteOrderMark}B${byteOrderMark}.mp4`))
      .toBe(`${byteOrderMark}A${byteOrderMark}${byteOrderMark}B${byteOrderMark}`)
  })

  it('matches Go path.Base for trailing separators and empty roots', () => {
    expect(deriveArchiveFilenameTitle('目录/片名.mp4/')).toBe('片名')
    expect(deriveArchiveFilenameTitle('目录/片名.mp4///')).toBe('片名')
    expect(deriveArchiveFilenameTitle('')).toBe('')
    expect(deriveArchiveFilenameTitle('/')).toBe('')
    expect(deriveArchiveFilenameTitle('///')).toBe('')
  })
})

describe('archive filename title eligibility', () => {
  it('accepts pending and failed videos', () => {
    expect(canReplaceArchiveFilenameTitle(pendingVideo('pending.mp4'))).toBe(true)
    expect(canReplaceArchiveFilenameTitle(pendingVideo('failed.mp4', { status: 'failed' }))).toBe(true)
  })

  it.each([
    ['processing video', { status: 'processing' }],
    ['ready video', { status: 'ready' }],
    ['existing video', { status: 'existing' }],
    ['skipped video', { status: 'skipped' }],
    ['pending image', { media_kind: 'image' }],
    ['pending directory', { entry_type: 'directory' }]
  ])('rejects %s', (_, overrides) => {
    expect(canReplaceArchiveFilenameTitle(pendingVideo('video.mp4', overrides))).toBe(false)
  })
})

describe('archive filename title drafts', () => {
  it('derives cleaned titles and preserves old titles in descriptions', () => {
    expect(buildArchiveFilenameTitleDraft(pendingVideo('目录/WWW.98t.LA@  ABC  123.mp4'))).toEqual({
      ok: true,
      title: 'ABC 123',
      description: '旧标题\n原说明',
      changed: true,
      issue: null
    })
  })

  it('trims old titles for case-sensitive comparison and transfer', () => {
    expect(buildArchiveFilenameTitleDraft(pendingVideo('title.mp4', {
      title: '  Title  ',
      description: '原说明'
    }))).toMatchObject({
      ok: true,
      title: 'title',
      description: 'Title\n原说明',
      changed: true
    })
  })

  it('keeps descriptions exact when the trimmed title is unchanged', () => {
    expect(buildArchiveFilenameTitleDraft(pendingVideo('同名.mp4', {
      title: '  同名  ',
      description: '  原说明\n'
    }))).toEqual({
      ok: true,
      title: '同名',
      description: '  原说明\n',
      changed: false,
      issue: null
    })
  })

  it.each([
    ['empty', '', '旧标题'],
    ['whitespace-only', '  ', '旧标题\n  '],
    ['leading and trailing whitespace', '  原说明\n\n', '旧标题\n  原说明\n\n']
  ])('preserves %s descriptions while transferring the old title', (_, description, expected) => {
    expect(buildArchiveFilenameTitleDraft(pendingVideo('新标题.mp4', { description })).description).toBe(expected)
  })

  it('accepts 200 Unicode code points and rejects 201 without changing the draft', () => {
    const validTitle = '😀'.repeat(200)
    const invalidTitle = '😀'.repeat(201)
    const original = {
      title: '  原标题  ',
      description: '  原说明\n'
    }

    expect(buildArchiveFilenameTitleDraft(pendingVideo(`${validTitle}.mp4`, original))).toMatchObject({
      ok: true,
      title: validTitle
    })
    expect(buildArchiveFilenameTitleDraft(pendingVideo(`${invalidTitle}.mp4`, original))).toEqual({
      ok: false,
      title: original.title,
      description: original.description,
      changed: false,
      issue: {
        id: 'file-1',
        relative_path: `${invalidTitle}.mp4`,
        message: '文件名无法生成有效标题'
      }
    })
  })

  it('returns the exact current draft when the file is ineligible or the derived title is empty', () => {
    const original = {
      title: '  当前标题  ',
      description: '  当前说明\n'
    }

    expect(buildArchiveFilenameTitleDraft(pendingVideo('www.98T.la@.mp4', original))).toMatchObject({
      ok: false,
      title: original.title,
      description: original.description,
      changed: false
    })
    expect(buildArchiveFilenameTitleDraft(pendingVideo('有效标题.mp4', {
      ...original,
      status: 'processing'
    }))).toMatchObject({
      ok: false,
      title: original.title,
      description: original.description,
      changed: false
    })
  })
})

describe('archive filename batch helpers', () => {
  it('previews only the first five valid items and reports the remainder', () => {
    const files = Array.from({ length: 7 }, (_, index) => pendingVideo(`标题 ${index + 1}.mp4`, {
      id: `file-${index + 1}`,
      title: `原标题 ${index + 1}`
    }))

    const preview = buildArchiveFilenameBatchPreview(files, 5)

    expect(preview).toMatchObject({ ok: true, total: 7, remaining: 2, issues: [] })
    expect(preview.items).toHaveLength(5)
    expect(preview.items[0]).toEqual({
      id: 'file-1',
      relative_path: '标题 1.mp4',
      old_title: '原标题 1',
      new_title: '标题 1'
    })
  })

  it('blocks the whole preview when any derived title is invalid', () => {
    const preview = buildArchiveFilenameBatchPreview([
      pendingVideo('www.98T.la@ABC.mp4', { id: 'file-1' }),
      pendingVideo('www.98T.la@.mp4', { id: 'file-2' })
    ], 5)

    expect(preview.ok).toBe(false)
    expect(preview.issues).toEqual([{
      id: 'file-2',
      relative_path: 'www.98T.la@.mp4',
      message: '文件名无法生成有效标题'
    }])
    expect(preview.items).toHaveLength(1)
  })

  it('rejects files missing ids or update timestamps with file-specific issues', () => {
    const preview = buildArchiveFilenameBatchPreview([
      pendingVideo('缺少 ID.mp4', { id: '' }),
      pendingVideo('缺少时间.mp4', { id: 'file-2', updated_at: '' })
    ])

    expect(preview.ok).toBe(false)
    expect(preview.issues).toEqual([
      { id: '', relative_path: '缺少 ID.mp4', message: '文件 ID 不能为空' },
      { id: 'file-2', relative_path: '缺少时间.mp4', message: '文件更新时间不能为空' }
    ])
  })

  it('rejects visible batch ids that differ and identifies the conflicting file', () => {
    const preview = buildArchiveFilenameBatchPreview([
      pendingVideo('第一批.mp4', { id: 'file-1', batch_id: 'batch-1' }),
      pendingVideo('第二批.mp4', { id: 'file-2', batch_id: 'batch-2' })
    ])

    expect(preview.ok).toBe(false)
    expect(preview.issues).toContainEqual({
      id: 'file-2',
      relative_path: '第二批.mp4',
      message: '所选文件不属于同一批次'
    })
  })

  it('does not treat an unavailable batch id as a cross-batch conflict', () => {
    const preview = buildArchiveFilenameBatchPreview([
      pendingVideo('未提供批次.mp4', { id: 'file-1', batch_id: undefined }),
      pendingVideo('已提供批次.mp4', { id: 'file-2', batch_id: 'batch-1' })
    ])

    expect(preview.ok).toBe(true)
  })

  it('returns a blocked empty preview for an empty selection', () => {
    expect(buildArchiveFilenameBatchPreview([], 5)).toEqual({
      ok: false,
      total: 0,
      items: [],
      remaining: 0,
      issues: []
    })
  })

  it('builds exact targets and prevents patch fields from overriding fixed semantics', () => {
    const files = [
      pendingVideo('第一集.mp4', { id: 'file-1' }),
      pendingVideo('第二集.mp4', { id: 'file-2', updated_at: '2026-07-12T05:01:00Z' })
    ]
    const targets = [
      { id: 'file-1', updated_at: updatedAt },
      { id: 'file-2', updated_at: '2026-07-12T05:01:00Z' }
    ]

    expect(buildArchiveFilenameBatchTargets(files)).toEqual(targets)
    expect(buildArchiveFilenameBatchPayload(files, {
      targets: [{ id: 'other', updated_at: 'other' }],
      title_mode: 'uniform',
      update_tags: true,
      tags: ['标签']
    })).toEqual({
      update_tags: true,
      tags: ['标签'],
      targets,
      title_mode: 'filename'
    })
  })
})
