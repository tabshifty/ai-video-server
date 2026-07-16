import { describe, expect, it } from 'vitest'
import {
  DEFAULT_IMAGE_ACTIVE,
  createImageBuiltInViews,
  hasImageActiveFilters,
  normalizeImageViewSnapshot
} from './imageManage.helpers'

describe('图片管理视图 helper', () => {
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
      active: '',
      actor_id: '',
      collection_id: '',
      viewMode: 'grid'
    })
    expect(DEFAULT_IMAGE_ACTIVE).toBe('1')
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
