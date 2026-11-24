import { describe, expect, it } from 'vitest'
import { parseEd2kLinks } from './toolbox.helpers'
import {
  ED2K_INGEST_NOT_VIDEO,
  ED2K_INGEST_UNKNOWN,
  ED2K_INGEST_VIDEO,
  buildEd2kIngestLinks,
  buildEd2kIngestPayload,
  buildEd2kIngestPlan,
  classifyEd2kIngestLink,
  cleanEd2kIngestTitle,
  getEd2kFilenameExtension,
  resolveIngestStatusMeta,
  ingestBatchLabel,
  ingestBatchSelectedLabel,
  normalizedIngestBatchID,
  truncateTitle,
  validateEd2kIngestForm
} from './ed2kIngest.helpers'

describe('getEd2kFilenameExtension', () => {
  it('取最后一段扩展名并小写', () => {
    expect(getEd2kFilenameExtension('Movie.MKV')).toBe('mkv')
    expect(getEd2kFilenameExtension('Show.tar.gz')).toBe('gz')
  })

  it('没有扩展名或以点开头时返回空', () => {
    expect(getEd2kFilenameExtension('Movie')).toBe('')
    expect(getEd2kFilenameExtension('.mp4')).toBe('')
    expect(getEd2kFilenameExtension('trailing.')).toBe('')
  })
})

describe('classifyEd2kIngestLink', () => {
  it('白名单内的扩展名判为视频', () => {
    expect(classifyEd2kIngestLink('a.mp4').kind).toBe(ED2K_INGEST_VIDEO)
    expect(classifyEd2kIngestLink('b.M2TS').kind).toBe(ED2K_INGEST_VIDEO)
  })

  it('压缩包与图片是"明确不是"，并给出确定原因', () => {
    const archive = classifyEd2kIngestLink('a.zip')
    expect(archive.kind).toBe(ED2K_INGEST_NOT_VIDEO)
    expect(archive.reason).toContain('压缩包')
    expect(classifyEd2kIngestLink('a.tar.gz').kind).toBe(ED2K_INGEST_NOT_VIDEO)
    expect(classifyEd2kIngestLink('a.jpg').kind).toBe(ED2K_INGEST_NOT_VIDEO)
  })

  it('其它已识别媒体不进入裁决区', () => {
    expect(classifyEd2kIngestLink('a.srt').kind).toBe(ED2K_INGEST_NOT_VIDEO)
    expect(classifyEd2kIngestLink('a.mp3').kind).toBe(ED2K_INGEST_NOT_VIDEO)
  })

  it('无扩展名与未知扩展名才需要管理员裁决', () => {
    expect(classifyEd2kIngestLink('Movie.2024.1080p.BluRay.x264-GROUP').kind).toBe(ED2K_INGEST_UNKNOWN)
    expect(classifyEd2kIngestLink('Movie').kind).toBe(ED2K_INGEST_UNKNOWN)
    expect(classifyEd2kIngestLink('a.part').kind).toBe(ED2K_INGEST_UNKNOWN)
  })
})

describe('cleanEd2kIngestTitle', () => {
  const cases = [
    ['Movie.2024.1080p.BluRay.x264-GROUP.mkv', 'Movie 2024 1080p BluRay x264-GROUP'],
    ['www.98t.la@Movie.2024.mkv', 'Movie 2024'],
    ['hhd800.com@ABC_1080p.mp4', 'ABC 1080p'],
    ['【中文字幕】Movie.2024.mkv', 'Movie 2024'],
    ['【1080p】Movie.mkv', '1080p Movie'],
    ['Movie(98t.la).mkv', 'Movie'],
    ['Movie (2019).mkv', 'Movie (2019)'],
    ['Movie - 副本.mkv', 'Movie'],
    ['Movie(1).mkv', 'Movie'],
    ['Movie 副本 3.mkv', 'Movie'],
    ['Show_01.mkv', 'Show 01']
  ]

  for (const [input, expected] of cases) {
    it(`${input} -> ${expected}`, () => {
      expect(cleanEd2kIngestTitle(input)).toBe(expected)
    })
  }

  it('只剥白名单视频扩展名', () => {
    expect(cleanEd2kIngestTitle('Show.S01.tar.gz')).toBe('Show S01 tar gz')
  })

  it('空输入与仅扩展名返回空串', () => {
    expect(cleanEd2kIngestTitle('')).toBe('')
    expect(cleanEd2kIngestTitle('   ')).toBe('')
  })

  it('保留中文与 emoji，不劈代理对', () => {
    expect(cleanEd2kIngestTitle('【中字】夏日限定🌊.2024.mkv')).toBe('夏日限定🌊 2024')
  })
})

describe('buildEd2kIngestPlan', () => {
  const links = [
    { href: 'ed2k://a', label: 'a.mp4', infoHash: 'a'.repeat(32) },
    { href: 'ed2k://b', label: 'b.zip', infoHash: 'b'.repeat(32) },
    { href: 'ed2k://c', label: 'c.unknown', infoHash: 'c'.repeat(32) }
  ]

  it('三分法分组并算好标题', () => {
    const plan = buildEd2kIngestPlan(links)
    expect(plan.videos).toHaveLength(1)
    expect(plan.notVideos).toHaveLength(1)
    expect(plan.unknown).toHaveLength(1)
    expect(plan.videos[0].title).toBe('a')
    expect(plan.unknown[0].reason).toContain('unknown')
  })

  it('只有被强制放行的无法识别项会进入提交列表', () => {
    const plan = buildEd2kIngestPlan(links)
    expect(buildEd2kIngestLinks(plan)).toHaveLength(1)
    const forced = new Set([plan.unknown[0].infoHash])
    const items = buildEd2kIngestLinks(plan, { forcedKeys: forced })
    expect(items).toHaveLength(2)
    expect(items[1]).toMatchObject({ link: 'ed2k://c', forced: true })
  })
})

describe('buildEd2kIngestPayload', () => {
  const links = [{ link: 'ed2k://a', title: '标题 A', forced: false }]

  it('按文件名模式带每条标题', () => {
    const payload = buildEd2kIngestPayload({
      links,
      form: { titleMode: 'filename', videoType: 'short', collectionIDs: ['c1'] }
    })
    expect(payload.title_mode).toBe('filename')
    expect(payload.uniform_title).toBe('')
    expect(payload.links[0].title).toBe('标题 A')
    expect(payload.collection_ids).toEqual(['c1'])
    expect(payload.site_category).toBe('')
  })

  it('统一标题模式清掉每条标题', () => {
    const payload = buildEd2kIngestPayload({
      links,
      form: { titleMode: 'uniform', uniformTitle: '合辑', videoType: 'movie' }
    })
    expect(payload.uniform_title).toBe('合辑')
    expect(payload.links[0].title).toBe('')
    expect(payload.collection_ids).toEqual([])
    expect(payload.site_category).toBe('')
  })

  it('AV 默认带日本地区', () => {
    const payload = buildEd2kIngestPayload({ links, form: { videoType: 'av' } })
    expect(payload.site_category).toBe('japanese')
  })
})

describe('validateEd2kIngestForm', () => {
  it('统一标题必填且不超过 200 字', () => {
    expect(validateEd2kIngestForm({ titleMode: 'uniform', uniformTitle: '' })).toContain('不能为空')
    expect(validateEd2kIngestForm({ titleMode: 'uniform', uniformTitle: 'あ'.repeat(201) })).toContain('200')
    expect(validateEd2kIngestForm({ titleMode: 'uniform', uniformTitle: 'ok' })).toBe('')
  })

  it('非短视频不能带视频合集', () => {
    expect(
      validateEd2kIngestForm({ titleMode: 'filename', videoType: 'movie', collectionIDs: ['c1'] })
    ).toContain('仅短视频')
  })
})

describe('resolveIngestStatusMeta', () => {
  it('非批成员返回 null', () => {
    expect(resolveIngestStatusMeta({ status: 'completed' })).toBeNull()
  })

  it('按任务与导入状态派生文案', () => {
    expect(resolveIngestStatusMeta({ ingest_batch_id: 'b', status: 'downloading' })).toMatchObject({
      label: '等待落盘'
    })
    expect(resolveIngestStatusMeta({ ingest_batch_id: 'b', status: 'completed' })).toMatchObject({
      label: '待自动入库'
    })
    expect(
      resolveIngestStatusMeta({ ingest_batch_id: 'b', status: 'completed', import: { status: 'imported' } })
    ).toMatchObject({ label: '已入库' })
    expect(
      resolveIngestStatusMeta({
        ingest_batch_id: 'b',
        status: 'completed',
        import: { status: 'skipped_existing' }
      })
    ).toMatchObject({ label: '已在库' })
    expect(
      resolveIngestStatusMeta({ ingest_batch_id: 'b', status: 'completed', import: { status: 'failed' } })
    ).toMatchObject({ label: '入库失败' })
  })

  it('混合括号里画质标记必须保留，语言词不能连坐', () => {
    // 去括号、保留内容；分隔符归一后 `98t.la` 里的点也会变成空格。
    expect(cleanEd2kIngestTitle('Movie[中文字幕 1080p].mkv')).toBe('Movie 中文字幕 1080p')
    expect(cleanEd2kIngestTitle('Movie(98t.la 720p).mkv')).toBe('Movie 98t la 720p')
  })

  it('剥离 `_2` 这类重复后缀但不动集数', () => {
    expect(cleanEd2kIngestTitle('Movie_2.mkv')).toBe('Movie')
    expect(cleanEd2kIngestTitle('Movie_9.mkv')).toBe('Movie')
    expect(cleanEd2kIngestTitle('Show_01.mkv')).toBe('Show 01')
  })
})

describe('buildEd2kIngestLinks 的形状', () => {
  it('保留 label 与 infoHash，预览与云端核对都依赖它们', () => {
    const plan = buildEd2kIngestPlan([{ href: 'ed2k://a', label: 'a.mp4', infoHash: 'a'.repeat(32) }])
    const [item] = buildEd2kIngestLinks(plan)
    expect(item).toMatchObject({ link: 'ed2k://a', label: 'a.mp4', infoHash: 'a'.repeat(32), forced: false })
  })
})

describe('toolbox.helpers.parseEd2kLinks 与服务端一致', () => {
  it('接受缺尾斜杠的四段链接', () => {
    const { links, invalidCount } = parseEd2kLinks('ed2k://|file|a.mkv|1|0123456789abcdef0123456789abcdef')
    expect(invalidCount).toBe(0)
    expect(links).toHaveLength(1)
    expect(links[0].label).toBe('a.mkv')
  })

  it('仍然拒绝非 file 段', () => {
    expect(parseEd2kLinks('ed2k://|server|1.2.3.4|4661|/').invalidCount).toBe(1)
  })
})

describe('truncateTitle', () => {
  it('按码点截断并补省略号', () => {
    expect(truncateTitle('あ'.repeat(250))).toHaveLength(200)
    expect(truncateTitle('短标题')).toBe('短标题')
  })
})

describe('入库批筛选标签', () => {
  const batch = {
    id: '43ec49eb-383f-4912-a777-6eb0929a95f8',
    video_type: 'short',
    counts: { members: 14 },
    created_at: '2026-11-24T01:06:07Z'
  }

  it('选项文案带类型、条数与时间，且不含 UUID', () => {
    const label = ingestBatchLabel(batch)
    expect(label).toContain('入库批')
    expect(label).toContain('短视频')
    expect(label).toContain('14 条')
    expect(label).toMatch(/\d\d-\d\d \d\d:\d\d/)
    expect(label).not.toContain('43ec49eb')
  })

  it('选中项在任何分支都不会显示 id', () => {
    // 未选择
    expect(ingestBatchSelectedLabel('')).toBe('全部入库批')
    // 选项还没到
    expect(ingestBatchSelectedLabel(batch.id, [], true)).toBe('入库批（加载中…）')
    // 选项已到但该批不在列表里（深链、超过条数、接口失败）
    expect(ingestBatchSelectedLabel(batch.id, [], false)).toBe('入库批（当前批）')
    // 命中选项
    const hit = ingestBatchSelectedLabel(batch.id, [batch], false)
    expect(hit).toContain('短视频')
    expect(hit).toContain('14 条')

    for (const value of ['', batch.id, 'unknown-id']) {
      for (const loading of [true, false]) {
        const text = ingestBatchSelectedLabel(value, [], loading)
        expect(text).not.toContain('43ec49eb')
        expect(text).not.toContain('unknown-id')
        expect(text).toMatch(/[\u4e00-\u9fa5]/)
      }
    }
  })

  it('清空值统一成空串，避免 undefined 让筛选状态漂移', () => {
    expect(normalizedIngestBatchID(undefined)).toBe('')
    expect(normalizedIngestBatchID(null)).toBe('')
    expect(normalizedIngestBatchID(batch.id)).toBe(batch.id)
  })

  it('占位选项也只用中文', () => {
    expect(ingestBatchLabel({ id: batch.id, placeholder: true })).toBe('入库批（当前批）')
  })
})
