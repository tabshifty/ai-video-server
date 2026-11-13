import { beforeEach, describe, expect, it, vi } from 'vitest'

const mocks = vi.hoisted(() => ({
  preview: vi.fn(), cleanup: vi.fn(), confirm: vi.fn(), info: vi.fn(), success: vi.fn(), error: vi.fn()
}))
vi.mock('../api/admin', () => ({ previewAdminEmptyCollections: mocks.preview, cleanupAdminEmptyCollections: mocks.cleanup }))
vi.mock('element-plus', () => ({ ElMessage: { info: mocks.info, success: mocks.success, error: mocks.error }, ElMessageBox: { confirm: mocks.confirm } }))
import { useEmptyCollectionCleanup } from './emptyCollectionCleanup'

beforeEach(() => {
  vi.resetAllMocks()
  mocks.preview.mockResolvedValue({ candidate_count: 3, preview_token: 'signed-preview' })
  mocks.confirm.mockResolvedValue('confirm')
  mocks.cleanup.mockResolvedValue({ deleted_count: 2, skipped_count: 1 })
})

describe('空合集清理交互', () => {
  it.each(['image', 'video'])('仅执行 %s 本次签名预览并报告实际结果', async (kind) => {
    const completed = vi.fn()
    const action = useEmptyCollectionCleanup(kind, completed)
    await action.run()
    expect(mocks.preview).toHaveBeenCalledWith(kind)
    const message = mocks.confirm.mock.calls[0][0]
    expect(message).toContain('3 个空')
    expect(message).toContain('不受当前搜索、筛选和分页限制')
    expect(message).toContain('不会删除图片、视频或媒体文件')
    expect(mocks.cleanup).toHaveBeenCalledWith(kind, 'signed-preview')
    expect(mocks.success).toHaveBeenCalledWith('已清理 2 个空合集，跳过 1 个合集')
    expect(completed).toHaveBeenCalledTimes(1)
    expect(action.busy.value).toBe(false)
  })

  it('没有候选时不弹出确认、不执行删除', async () => {
    mocks.preview.mockResolvedValue({ candidate_count: 0, preview_token: '' })
    await useEmptyCollectionCleanup('image', vi.fn()).run()
    expect(mocks.info).toHaveBeenCalledOnce()
    expect(mocks.confirm).not.toHaveBeenCalled()
    expect(mocks.cleanup).not.toHaveBeenCalled()
  })

  it.each(['cancel', 'close'])('取消确认 %s 不删除、不报错', async (reason) => {
    mocks.confirm.mockRejectedValue(reason)
    const action = useEmptyCollectionCleanup('video', vi.fn())
    await action.run()
    expect(mocks.cleanup).not.toHaveBeenCalled()
    expect(mocks.error).not.toHaveBeenCalled()
    expect(action.busy.value).toBe(false)
  })

  it('确认等待期间阻止重复预览和提交', async () => {
    let resolve
    mocks.confirm.mockImplementation(() => new Promise((done) => { resolve = done }))
    const action = useEmptyCollectionCleanup('image', vi.fn())
    const first = action.run()
    await vi.waitFor(() => expect(mocks.confirm).toHaveBeenCalledOnce())
    await action.run()
    expect(mocks.preview).toHaveBeenCalledOnce()
    resolve('confirm')
    await first
    expect(mocks.cleanup).toHaveBeenCalledOnce()
  })

  it.each(['preview', 'cleanup'])('%s 失败时不报成功并解除忙碌状态', async (step) => {
    mocks[step].mockRejectedValue(new Error('模拟网络异常'))
    const completed = vi.fn()
    const action = useEmptyCollectionCleanup('video', completed)
    await action.run()
    expect(mocks.error).toHaveBeenCalledOnce()
    expect(mocks.success).not.toHaveBeenCalled()
    expect(completed).not.toHaveBeenCalled()
    expect(action.busy.value).toBe(false)
  })
})
