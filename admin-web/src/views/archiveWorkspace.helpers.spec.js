import { describe, expect, it } from 'vitest'
import { descendantIDs, selectedWorkspaceFiles, importableWorkspaceFiles, resolveWorkspaceValues, summarizeWorkspaceImport, newWorkspaceID } from './archiveWorkspace.helpers'

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
