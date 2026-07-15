import { describe, expect, it } from 'vitest'
import { buildDashboardMetricGroups, dashboardQuickActions } from './dashboard.helpers'

describe('dashboard helpers', () => {
  it('只使用当前 stats API 已提供的字段', () => {
    const groups = buildDashboardMetricGroups({
      short_videos: 12,
      movie_videos: 3,
      episode_videos: 8,
      av_videos: 4,
      total_users: 5,
      today_uploads: 6,
      queue_length: 2,
      disk_free_bytes: 10737418240
    })

    expect(groups.runtime.map((item) => [item.label, item.value])).toEqual([
      ['今日上传', 6],
      ['转码队列', 2],
      ['磁盘剩余', '10.00 GB'],
      ['总用户数', 5]
    ])
    expect(groups.inventory.map((item) => item.value)).toEqual([12, 3, 8, 4])
    expect(JSON.stringify(groups)).not.toMatch(/失败率|健康|告警/)
  })

  it('将非有限正数指标稳定回落为零', () => {
    const groups = buildDashboardMetricGroups({
      today_uploads: -1,
      queue_length: Number.NaN,
      disk_free_bytes: undefined,
      total_users: '7'
    })

    expect(groups.runtime.map((item) => item.value)).toEqual([0, 0, '0.00 GB', 7])
    expect(groups.inventory.map((item) => item.value)).toEqual([0, 0, 0, 0])
  })

  it('只链接到四个已有管理端目标', () => {
    expect(dashboardQuickActions.map((item) => item.path)).toEqual([
      '/upload',
      '/tasks',
      '/videos',
      '/images'
    ])
    expect(dashboardQuickActions.map((item) => item.label)).toEqual([
      '上传视频',
      '查看任务',
      '管理视频',
      '管理图片'
    ])
  })
})
