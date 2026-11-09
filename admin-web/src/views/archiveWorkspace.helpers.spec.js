import { describe, expect, it } from 'vitest'
import { descendantIDs, selectedWorkspaceFiles, importableWorkspaceFiles, resolveWorkspaceValues, summarizeWorkspaceImport, newWorkspaceID, directoryBreadcrumbs } from './archiveWorkspace.helpers'

const draft = {
  directories: [{ id: 'a', parent_id: '', name: '甲' }, { id: 'b', parent_id: 'a', name: '乙' }],
  groups: [{ id: 'main', parent_id: '', name: '主分组', config: { title: '父标题', tags: ['A'], image_collection_ids: ['图集1', '图集2'] } }, { id: 'g', parent_id: 'main', name: '子组', config: { tags: [] } }],
  files: { v: { name: 'WWW.98t.LA@新名称.mp4', directory_id: 'b', group_id: 'g', config: {} }, i: { name: '图片.jpg', directory_id: 'a', group_id: 'g', config: {} } }
}
const files = [{ id: 'v', media_kind: 'video', status: 'pending' }, { id: 'i', media_kind: 'image', status: 'pending' }]
describe('压缩包全屏工作台规则', () => {
  it('递归选择并去重文件夹与单文件重叠选择', () => {
    expect([...descendantIDs(draft.directories, 'a')]).toEqual(['a', 'b'])
    expect(selectedWorkspaceFiles(draft, files, ['v'], ['a'])).toEqual(['v', 'i'])
  })
  it('保留逐字段继承、显式清空及多图集顺序', () => {
    expect(resolveWorkspaceValues(draft, 'v', 'video')).toMatchObject({ title: '父标题', tags: [], image_collection_ids: ['图集1', '图集2'] })
  })
  it('文件名模式使用虚拟名称并清理指定广告', () => {
    const d = structuredClone(draft); d.groups[1].config.title_mode = 'filename'
    expect(resolveWorkspaceValues(d, 'v', 'video').title).toBe('新名称')
    expect(resolveWorkspaceValues(d, 'i', 'image').title).toBe('图片')
  })
  it('单文件标题优先于文件名模式', () => {
    const d = structuredClone(draft); d.groups[1].config.title_mode = 'filename'; d.files.v.config.title = '指定标题'
    expect(resolveWorkspaceValues(d, 'v', 'video').title).toBe('指定标题')
  })
  it('排除、成功和活动文件不能重复提交', () => {
    const d = structuredClone(draft); d.files.i.excluded = true
    expect(importableWorkspaceFiles({ draft: d, files, jobs: [{ file_id: 'v', state: 'queued' }] }, ['v', 'v', 'i'])).toEqual([])
    expect(importableWorkspaceFiles({ draft, files: [{ ...files[0], status: 'ready' }], jobs: [] }, ['v'])).toEqual([])
  })
  it('预览分别提示未选视频合集和图片合集', () => {
    const data = { draft, files, effective: { v: { title: '视频' }, i: { title: '图片' } }, jobs: [] }
    expect(summarizeWorkspaceImport(data, ['v', 'i'])).toMatchObject({ videos: 1, images: 1, missingVideoCollections: 1, missingImageCollections: 1, missingVideoImageCollections: 1, emptyTitles: [] })
  })
})

it('家庭局域网 HTTP 没有 randomUUID 时仍可新建目录和分组', () => {
  const id = newWorkspaceID({ getRandomValues: (bytes) => bytes.fill(17) })
  expect(id).toMatch(/^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$/)
})


describe('工作台目录路径导航', () => {
  it('根目录、嵌套路径与不存在的目录均可安全导航', () => {
    expect(directoryBreadcrumbs(draft.directories, '')).toEqual([{ id: '', name: '根目录' }])
    expect(directoryBreadcrumbs(draft.directories, 'b')).toEqual([{ id: '', name: '根目录' }, { id: 'a', name: '甲' }, { id: 'b', name: '乙' }])
    expect(directoryBreadcrumbs(draft.directories, 'missing')).toEqual([{ id: '', name: '根目录' }])
  })
  it('目录重命名及时反映在路径中，异常循环不会无限遍历', () => {
    expect(directoryBreadcrumbs([{ id: 'a', parent_id: '', name: '新名称' }], 'a')[1].name).toBe('新名称')
    expect(directoryBreadcrumbs([{ id: 'a', parent_id: 'b', name: '甲' }, { id: 'b', parent_id: 'a', name: '乙' }], 'b')).toHaveLength(3)
  })
})

const desktopDraft = () => ({
  layout_version: 2, directories: [],
  groups: [{ id: 'a', parent_id: '', name: '甲组', config: { title_mode: 'filename', title: '组标题', tags: ['甲'], video_collection_ids: ['v1'], image_collection_ids: ['i1'] } }, { id: 'b', parent_id: '', name: '乙组', config: { title_mode: 'uniform', title: '乙标题', video_collection_ids: ['v2'], image_collection_ids: [] } }],
  files: { v: { group_id: 'a', name: '新视频.mp4', config: {} }, i: { group_id: 'a', name: '图片.jpg', config: {} } }
})
describe('平级桌面分组', () => {
  it('单组移动采用目标配置，单文件内容覆盖保留，合集不能覆盖分组', () => {
    const d = desktopDraft(); d.files.v.config = { title: '单独标题', video_collection_ids: ['绕过组'] }
    expect(resolveWorkspaceValues(d, 'v', 'video')).toMatchObject({ title: '单独标题', video_collection_ids: ['v1'] })
    d.files.v.group_id = 'b'
    expect(resolveWorkspaceValues(d, 'v', 'video')).toMatchObject({ title: '单独标题', video_collection_ids: ['v2'], tags: [] })
    expect(resolveWorkspaceValues(d, 'i', 'image').title).toBe('图片')
  })
  it('成功媒体新增目标可补关系，完成目标不重复提交，移出后不能直接入库', () => {
    const d = desktopDraft(); const data = { draft: d, files: [{ ...files[0], status: 'ready' }], jobs: [] }
    expect(importableWorkspaceFiles(data, ['v'])).toEqual(['v'])
    data.jobs = [{ file_id: 'v', state: 'completed', snapshot: { group_id: 'a', group_relations: true, video_collection_ids: ['v1'], image_collection_ids: ['i1'] } }]
    expect(importableWorkspaceFiles(data, ['v'])).toEqual([])
    d.groups[0].config.image_collection_ids.push('i2')
    expect(importableWorkspaceFiles(data, ['v'])).toEqual(['v'])
    d.files.v.group_id = ''
    expect(importableWorkspaceFiles(data, ['v'])).toEqual([])
  })
  it('组内存在活动任务时禁止对同组其他文件重复提交', () => {
    const d = desktopDraft()
    expect(importableWorkspaceFiles({ draft: d, files, jobs: [{ file_id: 'v', state: 'running', snapshot: { group_id: 'a' } }] }, ['i'])).toEqual([])
  })
})

it('混合组只选择视频也必须补齐组内图片目标', () => {
 const d = desktopDraft(); d.groups[0].config.image_collection_ids = []
 const result = summarizeWorkspaceImport({ draft: d, files, jobs: [], effective: { v: resolveWorkspaceValues(d, 'v', 'video') } }, ['v'])
 expect(result.missingImageCollections).toBe(1)
})
