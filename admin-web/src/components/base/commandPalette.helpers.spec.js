import { describe, expect, it } from 'vitest'
import {
  adminShellNavItems,
  matchMenuItem,
  searchMenuItems
} from './commandPalette.helpers'

describe('command palette helpers', () => {
  it('matches menu items by label, path, alias and initials', () => {
    expect(searchMenuItems('上传').map((item) => item.label).slice(0, 1)).toEqual(['上传视频'])
    expect(searchMenuItems('sj').map((item) => item.label).slice(0, 1)).toEqual(['视频管理'])
    expect(searchMenuItems('dssh').map((item) => item.label).slice(0, 1)).toEqual(['短视频审核'])
    expect(searchMenuItems('/short-review').map((item) => item.label).slice(0, 1)).toEqual(['短视频审核'])
    expect(searchMenuItems('/videos').map((item) => item.label).slice(0, 1)).toEqual(['视频管理'])
    expect(searchMenuItems('upload').map((item) => item.label).slice(0, 1)).toEqual(['上传视频'])
    expect(searchMenuItems('gjx').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('ed2k').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('孤儿文件扫描').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('orphan scan').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('压缩包导入').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('archive import').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
  })

  it('keeps every item when the query is empty', () => {
    const result = searchMenuItems('')
    expect(result.map((item) => item.score)).toEqual(adminShellNavItems.map(() => 0))
    expect(result.map((item) => item.label)).toEqual(adminShellNavItems.map((item) => item.label))
  })

  it('exposes the raw match function for direct scoring checks', () => {
    expect(matchMenuItem('upload', adminShellNavItems.find((item) => item.path === '/upload'))).toMatchObject({
      score: 60,
      matched: true
    })
  })

  it('registers short video review under media navigation', () => {
    const item = adminShellNavItems.find((entry) => entry.path === '/short-review')
    expect(item).toMatchObject({
      label: '短视频审核',
      groupLabel: '媒体库',
      icon: 'VideoCamera'
    })
  })

  it('keeps individual toolbox tools out of direct shell navigation', () => {
    expect(adminShellNavItems.some((entry) => entry.path === '/toolbox/orphan-files')).toBe(false)
    expect(adminShellNavItems.some((entry) => entry.path === '/toolbox/archive-import')).toBe(false)
    expect(searchMenuItems('/toolbox/orphan-files')).toHaveLength(0)
    expect(searchMenuItems('/toolbox/archive-import')).toHaveLength(0)
  })
})
