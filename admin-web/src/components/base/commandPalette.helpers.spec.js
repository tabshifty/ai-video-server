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
    expect(searchMenuItems('/videos').map((item) => item.label).slice(0, 1)).toEqual(['视频管理'])
    expect(searchMenuItems('upload').map((item) => item.label).slice(0, 1)).toEqual(['上传视频'])
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
})
