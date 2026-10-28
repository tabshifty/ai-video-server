import { readFileSync } from 'node:fs'
import { computed, reactive, ref } from 'vue'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

// Run the page's actual setup/lifecycle with controlled API responses.
const setup = readFileSync(new URL('./CloudDownloadList.vue', import.meta.url), 'utf8')
  .match(/<script setup>([\s\S]*?)<\/script>/)[1]
  .replace(/^import\s[\s\S]*?from ['"][^'"]+['"]\n/gm, '')
function createPage() {
  const mounted = [], unmounted = []
  const api = {
    getCloudBrowserState: vi.fn().mockResolvedValue({ up: true, logged_in: false }),
    getCloudLoginQR: vi.fn().mockResolvedValue({ image: 'data:image/png;base64,qr', expires_in_seconds: 120 }),
    listCloudDownloads: vi.fn().mockResolvedValue({ items: [], total_count: 0 })
  }
  const bindings = {
    computed, reactive, ref, useRouter: () => ({}),
    onMounted: fn => mounted.push(fn), onUnmounted: fn => unmounted.push(fn),
    hasActiveImportTasks: () => false, hasActiveCloudFileDeletions: () => false, ...api
  }
  const page = new Function(...Object.keys(bindings), `${setup}\nreturn { loadBrowserState, qrDialogVisible, qrImage, browserStateError, openQrDialog };`)(...Object.values(bindings))
  return { page, api, mount: () => mounted.forEach(fn => fn()), unmount: () => unmounted.forEach(fn => fn()) }
}
const flush = async () => { for (let i = 0; i < 8; i++) await Promise.resolve() }

describe('115 会话过期扫码提示', () => {
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())
  it('未登录自动弹码，手动关闭不重复弹，仍可主动取码', async () => {
    const { page, api, mount, unmount } = createPage()
    mount(); await flush()
    expect(page.qrDialogVisible.value).toBe(true)
    expect(page.qrImage.value).toContain('data:image/')
    page.qrDialogVisible.value = false
    await vi.advanceTimersByTimeAsync(20000)
    expect(page.qrDialogVisible.value).toBe(false)
    expect(api.getCloudLoginQR).toHaveBeenCalledTimes(1)
    page.openQrDialog(); await flush()
    expect(api.getCloudLoginQR).toHaveBeenCalledTimes(2)
    unmount(); expect(vi.getTimerCount()).toBe(0)
  })
  it('无活动任务仍检查，登录成功关窗清码，下次失效再提示', async () => {
    const { page, api, mount, unmount } = createPage()
    api.getCloudBrowserState.mockResolvedValue({ up: true, logged_in: true })
    mount(); await flush()
    expect(page.qrDialogVisible.value).toBe(false)
    api.getCloudBrowserState.mockResolvedValue({ up: true, logged_in: false })
    await vi.advanceTimersByTimeAsync(10000)
    expect(page.qrDialogVisible.value).toBe(true)
    api.getCloudBrowserState.mockResolvedValue({ up: true, logged_in: true })
    await vi.advanceTimersByTimeAsync(10000)
    expect(page.qrDialogVisible.value).toBe(false)
    expect(page.qrImage.value).toBe('')
    api.getCloudBrowserState.mockResolvedValue({ up: true, logged_in: false })
    await vi.advanceTimersByTimeAsync(10000)
    expect(page.qrDialogVisible.value).toBe(true)
    expect(api.getCloudLoginQR).toHaveBeenCalledTimes(2)
    unmount()
  })
  it('浏览器离线、状态验证异常和网络失败不会自动要求扫码', async () => {
    const { page, api, unmount } = createPage()
    for (const data of [{ up: false, logged_in: false }, { up: true, logged_in: false, session_error: '115 登录状态验证失败' }]) {
      api.getCloudBrowserState.mockResolvedValue(data)
      await page.loadBrowserState()
      expect(page.qrDialogVisible.value).toBe(false)
    }
    expect(page.browserStateError.value).toBe('115 登录状态验证失败')
    api.getCloudBrowserState.mockRejectedValue(new Error('网络超时'))
    await page.loadBrowserState()
    expect(page.browserStateError.value).toBe('网络超时')
    expect(api.getCloudLoginQR).not.toHaveBeenCalled()
    unmount()
  })
  it('慢请求不会并发，卸载后的状态响应不会弹码', async () => {
    const { page, api, mount, unmount } = createPage()
    let resolve
    api.getCloudBrowserState.mockReturnValue(new Promise(done => { resolve = done }))
    mount(); await vi.advanceTimersByTimeAsync(20000)
    expect(api.getCloudBrowserState).toHaveBeenCalledTimes(1)
    unmount(); resolve({ up: true, logged_in: false }); await flush()
    expect(page.qrDialogVisible.value).toBe(false)
    expect(api.getCloudLoginQR).not.toHaveBeenCalled()
    expect(vi.getTimerCount()).toBe(0)
  })
  it('卸载后的二维码响应不会恢复图像或倒计时', async () => {
    const { page, api, mount, unmount } = createPage()
    let resolve
    api.getCloudLoginQR.mockReturnValue(new Promise(done => { resolve = done }))
    mount(); await flush()
    expect(api.getCloudLoginQR).toHaveBeenCalledTimes(1)
    unmount(); resolve({ image: 'data:image/png;base64,late', expires_in_seconds: 120 }); await flush()
    expect(page.qrImage.value).toBe('')
    expect(vi.getTimerCount()).toBe(0)
  })
})
