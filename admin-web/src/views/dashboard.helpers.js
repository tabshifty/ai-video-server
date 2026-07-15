function count(value) {
  const number = Number(value)
  return Number.isFinite(number) && number > 0 ? number : 0
}

function formatDiskFree(value) {
  return `${(count(value) / 1024 / 1024 / 1024).toFixed(2)} GB`
}

export function buildDashboardMetricGroups(stats = {}) {
  return {
    runtime: [
      { key: 'today', label: '今日上传', value: count(stats.today_uploads), tone: 'info' },
      { key: 'queue', label: '转码队列', value: count(stats.queue_length) },
      { key: 'disk', label: '磁盘剩余', value: formatDiskFree(stats.disk_free_bytes) },
      { key: 'users', label: '总用户数', value: count(stats.total_users) }
    ],
    inventory: [
      { key: 'short', label: '短视频', value: count(stats.short_videos) },
      { key: 'movie', label: '电影', value: count(stats.movie_videos) },
      { key: 'episode', label: '电视剧集', value: count(stats.episode_videos) },
      { key: 'av', label: 'AV', value: count(stats.av_videos) }
    ]
  }
}

export const dashboardQuickActions = [
  { path: '/upload', label: '上传视频' },
  { path: '/tasks', label: '查看任务' },
  { path: '/videos', label: '管理视频' },
  { path: '/images', label: '管理图片' }
]
