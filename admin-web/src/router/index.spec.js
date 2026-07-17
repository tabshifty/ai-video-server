import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { resolveRouterHistoryBase } from './historyBase'

describe('resolveRouterHistoryBase', () => {
  it('保留 Vite 生产构建的 /admin/ 基路径', () => {
    expect(resolveRouterHistoryBase('/admin/')).toBe('/admin/')
  })

  it('缺省时回落到根路径', () => {
    expect(resolveRouterHistoryBase('')).toBe('/')
    expect(resolveRouterHistoryBase()).toBe('/')
  })

  it('媒体集合页使用壳层标题且保持原有路由目标', () => {
    const routerSource = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const pendingSource = readFileSync(new URL('../views/PendingDeleteShorts.vue', import.meta.url), 'utf8')
    const collectionSource = readFileSync(new URL('../views/ImageCollectionManage.vue', import.meta.url), 'utf8')

    expect(routerSource).toContain("{ path: '/short-pending-delete', component: PendingDeleteShorts },")
    expect(routerSource).toContain("{ path: '/image-collections', component: ImageCollectionManage },")
    expect(pendingSource).not.toContain('<PageHeader')
    expect(collectionSource).not.toContain('<PageHeader')
  })

  it('仪表盘使用壳层标题且保持原有路由目标', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const dashboardRoute = source.split('\n').find((line) => line.includes("path: '/dashboard'"))

    expect(dashboardRoute).toContain("{ path: '/dashboard', component: Dashboard }")
    expect(dashboardRoute).not.toContain('hideShellPageHeader')
  })

  it('任务监控使用壳层标题且保持原有路由目标', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const taskRoute = source.split('\n').find((line) => line.includes("path: '/tasks'"))

    expect(taskRoute).toContain("{ path: '/tasks', component: TaskMonitor }")
    expect(taskRoute).not.toContain('hideShellPageHeader')
  })

  it('视频资源页使用壳层标题且保持原有路由目标', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const videoRoute = source.split('\n').find((line) => line.includes("path: '/videos'"))

    expect(videoRoute).toContain("{ path: '/videos', component: VideoList }")
    expect(videoRoute).not.toContain('hideShellPageHeader')
  })

  it('视频上传页使用壳层标题且只移除对应兼容 meta', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const uploadRoute = source.split('\n').find((line) => line.includes("path: '/upload'")) || ''

    expect(uploadRoute.trim()).toBe("{ path: '/upload', component: VideoUpload },")
    expect(uploadRoute).not.toContain('hideShellPageHeader')
  })

  it('媒体刮削工作台使用壳层标题且只移除对应兼容 meta', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const routes = [
      ["path: '/scrape'", "{ path: '/scrape', component: ScrapePreview },"],
      ["path: '/av-scrape'", "{ path: '/av-scrape', component: AVManualScrape },"]
    ]

    routes.forEach(([path, expected]) => {
      const route = source.split('\n').find((line) => line.includes(path)) || ''

      expect(route.trim()).toBe(expected)
      expect(route).not.toContain('hideShellPageHeader')
    })
  })

  it('电视剧、工具箱与系统设置使用壳层标题且只移除对应兼容 meta', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const routes = [
      ["path: '/tv-series'", "{ path: '/tv-series', component: TvSeriesManage },"],
      ["path: '/toolbox'", "{ path: '/toolbox', component: Toolbox },"],
      ["path: '/settings'", "{ path: '/settings', component: SystemSettings },"]
    ]

    routes.forEach(([path, expected]) => {
      const route = source.split('\n').find((line) => line.includes(path)) || ''

      expect(route.trim()).toBe(expected)
      expect(route).not.toContain('hideShellPageHeader')
    })
  })

  it('图片资产页使用壳层标题且保持原有路由目标', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const imageRoute = source.split('\n').find((line) => line.includes("path: '/images'"))

    expect(imageRoute).toContain("{ path: '/images', component: ImageManage }")
    expect(imageRoute).not.toContain('hideShellPageHeader')
  })

  it('基础 CRUD 集合页使用壳层标题且只移除对应兼容 meta', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const routes = [
      ["path: '/actors'", "{ path: '/actors', component: ActorManage }"],
      ["path: '/collections'", "{ path: '/collections', component: CollectionManage }"],
      ["path: '/users'", "{ path: '/users', component: UserManage }"]
    ]

    routes.forEach(([path, expected]) => {
      const route = source.split('\n').find((line) => line.includes(path)) || ''

      expect(route).toContain(expected)
      expect(route).not.toContain('hideShellPageHeader')
    })
  })

  it('服务资源集合页使用壳层标题且只移除对应兼容 meta', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const routes = [
      ["path: '/iptv'", "{ path: '/iptv', component: IPTVManage }"],
      ["path: '/tv-app'", "{ path: '/tv-app', component: TvAppManage }"]
    ]

    routes.forEach(([path, expected]) => {
      const route = source.split('\n').find((line) => line.includes(path)) || ''

      expect(route).toContain(expected)
      expect(route).not.toContain('hideShellPageHeader')
    })
  })
})
