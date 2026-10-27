import { describe, expect, it } from 'vitest'
import {
  ACTIVE_IMPORT_STATUSES,
  VIDEO_IMPORT_TYPES,
  basenameWithoutExtension,
  buildArchiveImportPayload,
  buildVideoImportPayload,
  formatImportFileSize,
  hasActiveImportTasks,
  importDestination,
  cloudArchiveImportPath,
  cloudArchiveImportState,
  isImportActive,
  isStaleActiveImport,
  normalizeImportTags,
  normalizeImportUUIDs,
  resolveCloudImportAction,
  resolveCloudImportDefaultTitle,
  splitImportActors,
  validateArchiveImportForm,
  validateVideoImportForm
} from './cloudDownloadImport.helpers'

const ACTOR_ID = 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa'
const VIDEO_COLLECTION_ID = 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb'
const IMAGE_COLLECTION_ID = 'cccccccc-cccc-4ccc-8ccc-cccccccccccc'

describe('115 云下载导入表单 helpers', () => {
  it('标签统一小写、合并空白并稳定去重', () => {
    expect(
      normalizeImportTags([' 动作 ', '动作', 'Sci-Fi', ' sci-fi ', '  ', null, 'A  B'])
    ).toEqual(['动作', 'sci-fi', 'a b'])
    expect(normalizeImportTags('标签一,标签二，标签一\n标签三')).toEqual(['标签一', '标签二', '标签三'])
    expect(normalizeImportTags(null)).toEqual([])
  })

  it('只保留合法 UUID 并按大小写不敏感去重', () => {
    expect(
      normalizeImportUUIDs([
        VIDEO_COLLECTION_ID,
        VIDEO_COLLECTION_ID.toUpperCase(),
        'not-a-uuid',
        IMAGE_COLLECTION_ID,
        ''
      ])
    ).toEqual([VIDEO_COLLECTION_ID, IMAGE_COLLECTION_ID])
  })

  it('把演员选择拆成 UUID 与新建姓名，并去重', () => {
    expect(splitImportActors([ACTOR_ID, ACTOR_ID.toUpperCase(), 'Alice', 'alice', '  Bob   Smith ']))
      .toEqual({
        actorIDs: [ACTOR_ID],
        actorNames: ['Alice', 'Bob Smith']
      })
    expect(splitImportActors(null)).toEqual({ actorIDs: [], actorNames: [] })
  })

  it('从 Unix/Windows 路径取 basename 并去掉扩展名', () => {
    expect(basenameWithoutExtension('/mnt/115/downloads/Movie.2024.mkv')).toBe('Movie.2024')
    expect(basenameWithoutExtension('D:\\115\\Show.S01E01.mp4')).toBe('Show.S01E01')
    expect(basenameWithoutExtension('archive.zip')).toBe('archive')
    expect(basenameWithoutExtension('noext')).toBe('noext')
    expect(basenameWithoutExtension('.hidden')).toBe('.hidden')
    expect(basenameWithoutExtension('')).toBe('')
  })

  it('默认标题优先取任务 local_path，其次 preview.name', () => {
    expect(
      resolveCloudImportDefaultTitle(
        { local_path: '/downloads/本地文件.mkv', declared_name: 'fallback.mkv' },
        { name: 'preview.mkv' }
      )
    ).toBe('本地文件')
    expect(
      resolveCloudImportDefaultTitle({ local_path: '' }, { name: '/downloads/预览文件.mp4' })
    ).toBe('预览文件')
    expect(resolveCloudImportDefaultTitle(null, null)).toBe('')
  })

  it('格式化文件大小', () => {
    expect(formatImportFileSize(0)).toBe('--')
    expect(formatImportFileSize(512)).toBe('512 B')
    expect(formatImportFileSize(1024)).toBe('1.0 KB')
    expect(formatImportFileSize(1536)).toBe('1.5 KB')
    expect(formatImportFileSize(1024 * 1024)).toBe('1.0 MB')
  })
})

describe('115 云下载导入请求体', () => {
  it('构造视频导入体：短标题、标签、演员、合集和图集', () => {
    expect(
      buildVideoImportPayload({
        type: 'short',
        site_category: 'western',
        title: '  标题  ',
        description: ' 描述 ',
        tags: ['Tag', 'tag', ' '],
        actors: [ACTOR_ID, '新演员', '新演员'],
        collections: [VIDEO_COLLECTION_ID, 'bad-id', VIDEO_COLLECTION_ID],
        image_collection_id: IMAGE_COLLECTION_ID
      })
    ).toEqual({
      kind: 'video',
      type: 'short',
      site_category: '',
      title: '标题',
      description: '描述',
      tags: ['tag'],
      actor_ids: [ACTOR_ID],
      actor_names: ['新演员'],
      collection_ids: [VIDEO_COLLECTION_ID],
      image_collection_id: IMAGE_COLLECTION_ID
    })
  })

  it('构造 AV 导入体时保留地区，非 short 类型丢弃合集', () => {
    const avPayload = buildVideoImportPayload({
      type: 'av',
      site_category: 'fc2',
      collections: [VIDEO_COLLECTION_ID]
    })
    expect(avPayload.site_category).toBe('fc2')
    expect(avPayload.collection_ids).toEqual([])
    expect(buildVideoImportPayload({ type: 'av', site_category: '  ' }).site_category).toBe('japanese')

    const moviePayload = buildVideoImportPayload({
      type: 'movie',
      site_category: 'western',
      collections: [VIDEO_COLLECTION_ID]
    })
    expect(moviePayload.site_category).toBe('')
    expect(moviePayload.collection_ids).toEqual([])
  })

  it('缺少或非法类型时回退 short，并过滤非法图集 ID', () => {
    const payload = buildVideoImportPayload({ type: 'unknown', image_collection_id: 'bad-id' })
    expect(payload.type).toBe('short')
    expect(payload.image_collection_id).toBeNull()
    expect(buildVideoImportPayload({ type: 'short', image_collection_id: '' }).image_collection_id).toBeNull()
    expect(validateVideoImportForm({ type: 'short' })).toBe('')
    expect(validateVideoImportForm({ type: 'unknown' })).toBe('请选择视频类型')
  })

  it('构造压缩包导入体：标签、合集、密码', () => {
    expect(
      buildArchiveImportPayload({
        title: ' 包名 ',
        default_description: ' 说明 ',
        default_tags: ['A', 'a', 'B'],
        default_video_collection_ids: [VIDEO_COLLECTION_ID, 'bad-id', VIDEO_COLLECTION_ID],
        default_image_collection_ids: [IMAGE_COLLECTION_ID],
        has_password: true,
        password: ' secret '
      })
    ).toEqual({
      kind: 'archive',
      title: '包名',
      default_description: '说明',
      default_tags: ['a', 'b'],
      default_video_collection_ids: [VIDEO_COLLECTION_ID],
      default_image_collection_ids: [IMAGE_COLLECTION_ID],
      has_password: true,
      password: 'secret'
    })
  })

  it('关闭密码时清空密码字段，开启密码时必须填写', () => {
    expect(
      buildArchiveImportPayload({
        has_password: false,
        password: 'should-not-send'
      }).password
    ).toBe('')
    expect(validateArchiveImportForm({ has_password: true, password: '   ' })).toBe('请输入压缩包密码')
    expect(validateArchiveImportForm({ has_password: true, password: 'x' })).toBe('')
    expect(validateArchiveImportForm({ has_password: false, password: '' })).toBe('')
  })
})

describe('115 云下载导入状态', () => {
  it('识别所有活动导入状态', () => {
    for (const status of ACTIVE_IMPORT_STATUSES) {
      expect(isImportActive({ import: { status } })).toBe(true)
      expect(resolveCloudImportAction({ status: 'completed', import: { status } })).toBe('importing')
    }
    expect(hasActiveImportTasks([
      { status: 'completed', import: { status: 'hashing' } },
      { status: 'downloading' }
    ])).toBe(true)
    expect(hasActiveImportTasks([])).toBe(false)
    expect(hasActiveImportTasks(null)).toBe(false)
  })

  it('completed 只在有本地路径时提供入媒体库', () => {
    expect(resolveCloudImportAction({ status: 'completed', local_path: '/downloads/a.mkv' })).toBe('import')
    expect(resolveCloudImportAction({ status: 'completed', local_path: '' })).toBe('missing')
    expect(resolveCloudImportAction({ status: 'completed', import: null, local_path: '/downloads/a.mkv' }))
      .toBe('import')
  })

  it('活动导入超过宽限期后才允许重新提交', () => {
    const now = Date.parse('2026-10-18T01:00:00Z')
    const activeTask = (createdAt) => ({
      status: 'completed',
      import: { status: 'copying', created_at: createdAt }
    })
    expect(isStaleActiveImport(activeTask('2026-10-18T00:50:00Z'), now)).toBe(true)
    expect(isStaleActiveImport(activeTask('2026-10-18T00:59:00Z'), now)).toBe(false)
    expect(isStaleActiveImport(activeTask(''), now)).toBe(false)
    expect(resolveCloudImportAction(activeTask('2026-10-18T00:50:00Z'), now)).toBe('retry-import')
    expect(resolveCloudImportAction(activeTask('2026-10-18T00:59:00Z'), now)).toBe('importing')
  })

  it('按下载与导入状态决定操作列动作', () => {
    expect(resolveCloudImportAction({ status: 'completed', import: { status: 'imported' } }))
      .toBe('imported')
    expect(resolveCloudImportAction({ status: 'completed', import: { status: 'skipped_existing' } }))
      .toBe('existing')
    expect(resolveCloudImportAction({ status: 'completed', import: { status: 'failed' } }))
      .toBe('retry-import')
    expect(resolveCloudImportAction({ status: 'failed' })).toBe('retry-download')
    expect(resolveCloudImportAction({ status: 'blocked' })).toBe('retry-download')
    expect(resolveCloudImportAction({ status: 'downloading' })).toBe('none')
    expect(resolveCloudImportAction({ status: 'completed', import: { status: 'unknown' } }))
      .toBe('none')
  })

  it('压缩包导入结果直达批次工作台，视频保留列表入口', () => {
    expect(importDestination({
      import: { kind: 'video', status: 'imported', video_id: 'video-uuid' }
    })).toBe('/videos')
    expect(importDestination({
      import: { kind: 'archive', status: 'imported', archive_batch_id: '11111111-1111-4111-8111-111111111111' }
    })).toBe('/toolbox/archive-import/11111111-1111-4111-8111-111111111111')
    expect(importDestination({
      import: { kind: 'archive', status: 'skipped_existing', video_id: 'video-uuid' }
    })).toBe('/videos')
    expect(importDestination({ import: { kind: 'archive', status: 'imported' } })).toBe('/toolbox/archive-import')
    expect(importDestination({ import: { kind: 'archive', status: 'failed', archive_batch_id: '11111111-1111-4111-8111-111111111111' } })).toBe('/toolbox/archive-import/11111111-1111-4111-8111-111111111111')
    expect(importDestination({ status: 'completed' })).toBe('')
  })

  it('活动状态集合与视频类型集合保持契约', () => {
    expect(ACTIVE_IMPORT_STATUSES).toEqual(['queued', 'hashing', 'copying', 'creating'])
    expect(VIDEO_IMPORT_TYPES).toEqual(['short', 'movie', 'episode', 'av'])
  })
})


describe('115 压缩包全屏导入恢复', () => {
  it('任务入口保留大整数 ID，不经过 Number 转换或拼接任意路径', () => {
    expect(cloudArchiveImportPath('9223372036854775807')).toBe('/cloud-downloads/9223372036854775807/archive-import')
    for (const id of ['', 0, -1, '1/other', '1?redirect=x']) expect(cloudArchiveImportPath(id)).toBe('')
  })
  it('批次已创建时，源是否存在和导入成功/失败都不阻止恢复工作台', () => {
    for (const status of ['queued', 'creating', 'imported', 'failed']) {
      const task = { import: { kind: 'archive', status, archive_batch_id: '11111111-1111-4111-8111-111111111111' } }
      expect(cloudArchiveImportState(task)).toBe('workspace')
      expect(importDestination(task)).toBe('/toolbox/archive-import/11111111-1111-4111-8111-111111111111')
    }
  })
  it('刷新活动任务只等待，未提交或失败才显示可提交表单', () => {
    for (const status of ACTIVE_IMPORT_STATUSES) expect(cloudArchiveImportState({ import: { kind: 'archive', status } })).toBe('waiting')
    expect(cloudArchiveImportState({ status: 'completed' })).toBe('form')
    expect(cloudArchiveImportState({ import: { kind: 'archive', status: 'failed' } })).toBe('form')
    expect(cloudArchiveImportState({ import: { kind: 'video', status: 'queued' } })).toBe('unsupported')
    expect(cloudArchiveImportState({ import: { kind: 'archive', status: 'imported' } })).toBe('missing-batch')
  })
  it('失效批次 ID 安全回退列表，不能注入路径', () => {
    expect(importDestination({ import: { kind: 'archive', status: 'imported', archive_batch_id: '../../videos' } })).toBe('/toolbox/archive-import')
  })
})
