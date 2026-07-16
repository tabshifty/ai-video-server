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

  it('注册待删除短视频路由并隐藏壳层重复标题', () => {
    const source = readFileSync(new URL('./index.js', import.meta.url), 'utf8')
    const page = readFileSync(new URL('../views/PendingDeleteShorts.vue', import.meta.url), 'utf8')

    expect(source).toContain("import PendingDeleteShorts from '../views/PendingDeleteShorts.vue'")
    expect(source).toContain("{ path: '/short-pending-delete', component: PendingDeleteShorts, meta: { hideShellPageHeader: true } }")
    expect(page).toContain("import Layout from '../components/Layout.vue'")
    expect(page).toMatch(/<Layout>\s*<div class="page-shell pending-delete-page">/)
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
})
