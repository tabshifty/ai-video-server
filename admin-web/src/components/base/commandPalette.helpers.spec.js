import { describe, expect, it } from 'vitest'
import {
  adminShellNavGroups,
  adminShellNavItems,
  matchMenuItem,
  searchMenuItems
} from './commandPalette.helpers'

describe('command palette helpers', () => {
  it('matches menu items by label, path, alias and initials', () => {
    expect(searchMenuItems('上传').map((item) => item.label).slice(0, 1)).toEqual(['上传视频'])
    expect(searchMenuItems('sj').map((item) => item.label).slice(0, 1)).toEqual(['视频管理'])
    expect(searchMenuItems('dscdsp').map((item) => item.label).slice(0, 1)).toEqual(['待删除短视频'])
    expect(searchMenuItems('/short-pending-delete').map((item) => item.label).slice(0, 1)).toEqual(['待删除短视频'])
    expect(searchMenuItems('/videos').map((item) => item.label).slice(0, 1)).toEqual(['视频管理'])
    expect(searchMenuItems('upload').map((item) => item.label).slice(0, 1)).toEqual(['上传视频'])
    expect(searchMenuItems('gjx').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('ed2k').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('孤儿文件扫描').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('orphan scan').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('压缩包导入').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('archive import').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('密码管理').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('password vault').map((item) => item.label).slice(0, 1)).toEqual(['工具箱'])
    expect(searchMenuItems('论坛资源').map((item) => item.label).slice(0, 1)).toEqual(['论坛资源'])
    expect(searchMenuItems('hermes forum').map((item) => item.label).slice(0, 1)).toEqual(['论坛资源'])
    expect(searchMenuItems('telegram').map((item) => item.label).slice(0, 1)).toEqual(['Telegram 管理'])
    expect(searchMenuItems('tg').map((item) => item.label).slice(0, 1)).toEqual(['Telegram 管理'])
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

  it('registers pending delete shorts under media navigation', () => {
    const item = adminShellNavItems.find((entry) => entry.path === '/short-pending-delete')
    expect(item).toMatchObject({
      label: '待删除短视频',
      groupLabel: '媒体库',
      icon: 'Delete'
    })
  })

  it('keeps service and toolbox navigation capabilities in one authoritative group', () => {
    const group = adminShellNavGroups.find((entry) => entry.key === 'service-tools')

    expect(group).toMatchObject({ label: '服务与工具' })
    expect(group.items.map(({ path, label, title, icon, alias }) => ({ path, label, title, icon, alias }))).toEqual([
      { path: '/iptv', label: 'IPTV 管理', title: 'IPTV 管理', icon: 'Monitor', alias: 'iptv live' },
      { path: '/telegram', label: 'Telegram 管理', title: 'Telegram 管理', icon: 'Connection', alias: 'telegram tg 电报 telegram管理' },
      { path: '/tasks', label: '任务监控', title: '任务监控', icon: 'List', alias: 'task tasks jobs rw' },
      { path: '/forum-posts', label: '论坛资源', title: '论坛资源', icon: 'Link', alias: 'forum forums hermes forum-posts hermes-forum luntan ziyuan ltz y' },
      {
        path: '/toolbox',
        label: '工具箱',
        title: '工具箱',
        icon: 'Tools',
        alias: 'toolbox tools ed2k orphan scan orphan-files 孤儿文件扫描 archive archive-import archive import zip rar 7z 压缩包导入 压缩包 password vault credentials 密码 密码库 密码管理 gjx'
      }
    ])
    expect(searchMenuItems('live')[0]).toMatchObject({ path: '/iptv', groupKey: 'service-tools' })
    expect(searchMenuItems('telegram')[0]).toMatchObject({ path: '/telegram', groupKey: 'service-tools' })
    expect(searchMenuItems('jobs')[0]).toMatchObject({ path: '/tasks', groupKey: 'service-tools' })
    expect(searchMenuItems('hermes')[0]).toMatchObject({ path: '/forum-posts', groupKey: 'service-tools' })
    expect(searchMenuItems('gjx')[0]).toMatchObject({ path: '/toolbox', groupKey: 'service-tools' })
  })

  it('keeps individual toolbox tools out of direct shell navigation', () => {
    expect(adminShellNavItems.some((entry) => entry.path === '/toolbox/orphan-files')).toBe(false)
    expect(adminShellNavItems.some((entry) => entry.path === '/toolbox/archive-import')).toBe(false)
    expect(adminShellNavItems.some((entry) => entry.path === '/toolbox/password-vault')).toBe(false)
    expect(searchMenuItems('/toolbox/orphan-files')).toHaveLength(0)
    expect(searchMenuItems('/toolbox/archive-import')).toHaveLength(0)
    expect(searchMenuItems('/toolbox/password-vault')).toHaveLength(0)
  })
})
