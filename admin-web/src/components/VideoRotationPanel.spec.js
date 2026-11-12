import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { parse, compileScript } from '@vue/compiler-sfc'
import { transformSync } from 'esbuild'
import * as Vue from 'vue'

const confirm = vi.fn()
const api = { getVideoRotation: vi.fn(), previewVideoRotation: vi.fn(), submitVideoRotation: vi.fn() }
const messages = { success: vi.fn() }
const renderer = Vue.createRenderer({
  createElement: () => ({}), createText: () => ({}), createComment: () => ({}),
  setText() {}, setElementText() {}, parentNode() {}, nextSibling() {}, insert() {}, remove() {}, patchProp() {}
})
let app
async function flush() { for (let n = 0; n < 8; n++) await Promise.resolve(); await Vue.nextTick() }
function mount(video = { id: 'video-1', status: 'ready', metadata: {} }) {
  const { descriptor } = parse(readFileSync(new URL('./VideoRotationPanel.vue', import.meta.url), 'utf8'))
  const script = compileScript(descriptor, { id: 'rotation-test' })
  const code = transformSync(script.content, { loader: 'ts', format: 'cjs' }).code
  const module = { exports: {} }
  new Function('require', 'module', 'exports', code)((name) => {
    if (name === 'vue') return Vue
    if (name === 'element-plus') return { ElMessage: messages, ElMessageBox: { confirm } }
    if (name === '../api/admin') return api
    throw new Error(name)
  }, module, module.exports)
  const component = module.exports.default
  component.render = () => null
  const completed = vi.fn()
  app = renderer.createApp(component, { video, onCompleted: completed })
  app.mount({})
  return { state: app._instance.setupState, completed }
}
beforeEach(() => {
  vi.useFakeTimers()
  api.getVideoRotation.mockResolvedValue({ job: null })
  api.previewVideoRotation.mockResolvedValue({ mode: 'lossless', version: 'v1', temporary_bytes: 100 })
  api.submitVideoRotation.mockResolvedValue({ id: 'job-1', status: 'queued', progress: 0 })
  confirm.mockResolvedValue('confirm')
})
afterEach(() => { app?.unmount(); app = null; vi.clearAllMocks(); vi.useRealTimers() })

describe('视频方向修正交互', () => {
  it('预检后提交固定的无损方式与当前版本', async () => {
    const { state } = mount(); await flush()
    state.open(); state.angle = 270
    await state.inspect()
    await state.submit()
    expect(api.submitVideoRotation).toHaveBeenCalledWith('video-1', expect.objectContaining({ angle: 270, mode: 'lossless', version: 'v1', reencode_confirmed: false }))
    expect(confirm).toHaveBeenCalledWith(expect.stringContaining('不重新编码'), '确认修正方向', expect.anything())
  })
  it('重新编码预检须经过提示确认，取消时不提交', async () => {
    api.previewVideoRotation.mockResolvedValue({ mode: 'reencode', version: 'v2' })
    const { state } = mount(); await flush(); state.open(); await state.inspect()
    confirm.mockRejectedValue('cancel')
    await state.submit()
    expect(confirm).toHaveBeenCalledWith(expect.stringContaining('可能损失画质'), '确认修正方向', expect.anything())
    expect(api.submitVideoRotation).not.toHaveBeenCalled()
  })
  it('修改处理方式后必须重新预检', async () => {
    const { state } = mount(); await flush(); state.open(); await state.inspect()
    state.mode = 'reencode'; await Vue.nextTick()
    expect(state.plan).toBeNull()
    await state.submit()
    expect(api.submitVideoRotation).not.toHaveBeenCalled()
  })
  it('清理失败仍为活动任务，不重复受理并显示持久化错误', async () => {
    api.getVideoRotation.mockResolvedValue({ job: { id: 'job-1', status: 'cleanup', error_message: '清理待重试' } })
    const { state } = mount(); await flush()
    expect(state.active).toBe(true)
    expect(state.job.error_message).toBe('清理待重试')
    expect(state.jobLabel).toContain('清理')
  })
  it('处理完成仅通知刷新详情，不控制现有播放器', async () => {
    api.getVideoRotation.mockResolvedValueOnce({ job: { id: 'job-1', status: 'running', progress: 50 } }).mockResolvedValue({ job: { id: 'job-1', status: 'succeeded', progress: 100 } })
    const { completed } = mount(); await flush()
    await vi.advanceTimersByTimeAsync(3000); await flush()
    expect(completed).toHaveBeenCalledTimes(1)
  })
})
