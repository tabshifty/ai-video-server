import { describe, expect, it, vi } from 'vitest'
import * as imageManageHelpers from './imageManage.helpers'
import {
  DEFAULT_IMAGE_ACTIVE,
  createImageBuiltInViews,
  hasImageActiveFilters,
  normalizeImageViewSnapshot
} from './imageManage.helpers'

describe('图片管理视图 helper', () => {
  it('图片预览地址优先使用服务端直连字段并回退到当前页 owned URL', () => {
    const resolveImagePreviewUrl = imageManageHelpers.resolveImagePreviewUrl

    expect(resolveImagePreviewUrl).toBeTypeOf('function')
    expect(resolveImagePreviewUrl(
      { id: 'image-1', view_url: '/direct/view', url: '/direct/url', thumbnail_url: '/direct/thumb' },
      { 'image-1': 'blob:owned' }
    )).toBe('/direct/view')
    expect(resolveImagePreviewUrl(
      { id: 'image-1', url: '/direct/url', thumbnail_url: '/direct/thumb' },
      { 'image-1': 'blob:owned' }
    )).toBe('/direct/url')
    expect(resolveImagePreviewUrl(
      { id: 'image-1', thumbnail_url: '/direct/thumb' },
      { 'image-1': 'blob:owned' }
    )).toBe('/direct/thumb')
    expect(resolveImagePreviewUrl({ id: 'image-1' }, { 'image-1': 'blob:owned' })).toBe('blob:owned')
    expect(resolveImagePreviewUrl({ id: 'image-2' }, { 'image-1': 'blob:owned' })).toBe('')
  })

  it('图片列表预览替换时回收全部 owned URL 并返回空映射', () => {
    const revokeImagePreviewUrls = imageManageHelpers.revokeImagePreviewUrls
    const revoke = vi.fn()

    expect(revokeImagePreviewUrls).toBeTypeOf('function')
    expect(revokeImagePreviewUrls({
      first: 'blob:first',
      second: 'blob:second',
      empty: ''
    }, revoke)).toEqual({})
    expect(revoke).toHaveBeenCalledTimes(2)
    expect(revoke).toHaveBeenCalledWith('blob:first')
    expect(revoke).toHaveBeenCalledWith('blob:second')
  })

  it('只创建受支持的内置视图并保持快照引用隔离', () => {
    const views = createImageBuiltInViews(DEFAULT_IMAGE_ACTIVE, 'list')

    expect(views.map((item) => [
      item.id,
      item.label,
      item.snapshot.status,
      item.snapshot.active,
      item.snapshot.viewMode
    ])).toEqual([
      ['builtin-all', '全部图片', '', '1', 'list'],
      ['builtin-ready', '可用', 'ready', '1', 'list'],
      ['builtin-failed', '失败', 'failed', '1', 'list']
    ])
    expect(views.every((item) => item.builtIn === true)).toBe(true)
    expect(views[0].snapshot).not.toBe(views[1].snapshot)

    views[0].snapshot.status = 'changed'
    expect(views[1].snapshot.status).toBe('ready')
  })

  it('规范图片视图快照并排除分页、选择和 Drawer 状态', () => {
    const snapshot = normalizeImageViewSnapshot({
      q: 'A',
      status: 'ready',
      active: '0',
      actor_id: 'actor-1',
      collection_id: 'collection-1',
      viewMode: 'list',
      page: 4,
      selectedImageRows: ['image-1'],
      detailVisible: true,
      uploadDialogVisible: true
    })

    expect(snapshot).toEqual({
      q: 'A',
      status: 'ready',
      active: '0',
      actor_id: 'actor-1',
      collection_id: 'collection-1',
      viewMode: 'list'
    })
    expect(snapshot).not.toHaveProperty('page')
    expect(snapshot).not.toHaveProperty('selectedImageRows')
    expect(snapshot).not.toHaveProperty('detailVisible')
    expect(snapshot).not.toHaveProperty('uploadDialogVisible')
  })

  it('规范 active 与视图模式的非法值而不扩大字段集合', () => {
    expect(normalizeImageViewSnapshot({ active: true, viewMode: 'table' })).toEqual({
      q: '',
      status: '',
      active: DEFAULT_IMAGE_ACTIVE,
      actor_id: '',
      collection_id: '',
      viewMode: 'grid'
    })
    expect(DEFAULT_IMAGE_ACTIVE).toBe('1')
  })

  it('图片 active 缺失或非法时回退默认启用并保留显式全部语义', () => {
    expect.soft(normalizeImageViewSnapshot({}).active).toBe(DEFAULT_IMAGE_ACTIVE)
    expect.soft(normalizeImageViewSnapshot({ active: undefined }).active).toBe(DEFAULT_IMAGE_ACTIVE)
    expect.soft(normalizeImageViewSnapshot({ active: null }).active).toBe(DEFAULT_IMAGE_ACTIVE)
    expect.soft(normalizeImageViewSnapshot({ active: 'invalid' }).active).toBe(DEFAULT_IMAGE_ACTIVE)
    expect.soft(normalizeImageViewSnapshot({ active: [] }).active).toBe(DEFAULT_IMAGE_ACTIVE)
    expect.soft(normalizeImageViewSnapshot({ active: '' }).active).toBe('')
    expect.soft(normalizeImageViewSnapshot({ active: '0' }).active).toBe('0')
    expect.soft(normalizeImageViewSnapshot({ active: '1' }).active).toBe('1')
  })

  it('把默认启用状态视为结果集基线而不是用户筛选', () => {
    const baseline = {
      q: '',
      status: '',
      active: DEFAULT_IMAGE_ACTIVE,
      actor_id: '',
      collection_id: ''
    }

    expect(hasImageActiveFilters(baseline)).toBe(false)
    expect(hasImageActiveFilters({ ...baseline, q: ' 封面 ' })).toBe(true)
    expect(hasImageActiveFilters({ ...baseline, status: 'failed' })).toBe(true)
    expect(hasImageActiveFilters({ ...baseline, active: '0' })).toBe(true)
    expect(hasImageActiveFilters({ ...baseline, active: '' })).toBe(true)
    expect(hasImageActiveFilters({ ...baseline, actor_id: 'actor-1' })).toBe(true)
    expect(hasImageActiveFilters({ ...baseline, collection_id: 'collection-1' })).toBe(true)
  })
})
