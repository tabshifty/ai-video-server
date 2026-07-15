import { readFileSync } from 'node:fs'
import { reactive } from 'vue'
import { describe, expect, it, vi } from 'vitest'
import { CUSTOM_VIEW_ID } from './savedView.helpers'
import { useSavedViews } from './useSavedViews'

const normalize = (snapshot) => ({ q: String(snapshot?.q || '') })
const builtInViews = [
  { id: 'builtin-all', label: '全部', builtIn: true, snapshot: normalize({}) }
]

function createStorage(raw = '') {
  let value = raw
  return {
    getItem: vi.fn(() => value),
    setItem: vi.fn((_, next) => { value = next })
  }
}

function createController(overrides = {}) {
  const page = reactive({ q: '', ignored: 'page-only' })
  const storage = createStorage()
  const refresh = vi.fn()
  const controller = useSavedViews({
    storageKey: 'test-saved-views-v1',
    builtInViews,
    normalizeSnapshot: normalize,
    getCurrentSnapshot: () => page,
    applySnapshot: (snapshot) => { page.q = snapshot.q },
    refresh,
    storage,
    now: () => 123,
    ...overrides
  })

  return { controller, page, refresh, storage }
}

describe('useSavedViews', () => {
  it('owns custom state and the complete saved-view lifecycle', async () => {
    const { controller, page, refresh, storage } = createController()

    expect(controller.availableViews.value).toEqual(builtInViews)
    expect(controller.activeViewId.value).toBe('builtin-all')
    expect(controller.editableSourceId.value).toBe('')

    page.q = '失败'
    expect(controller.activeViewId.value).toBe(CUSTOM_VIEW_ID)
    expect(controller.saveView('   ')).toBeNull()

    const id = controller.saveView('  失败处理  ')
    expect(id).toBe('user-123')
    expect(controller.userViews.value).toEqual([
      { id, label: '失败处理', snapshot: { q: '失败' } }
    ])
    expect(controller.activeViewId.value).toBe(id)
    expect(controller.editableSourceId.value).toBe(id)

    expect(controller.renameView({ id: 'builtin-all', label: '覆盖内置' })).toBe(false)
    expect(controller.renameView({ id, label: '   ' })).toBe(false)
    expect(controller.renameView()).toBe(false)
    expect(controller.renameView({ id, label: '  待处理  ' })).toBe(true)
    expect(controller.userViews.value[0].label).toBe('待处理')

    page.q = ''
    expect(await controller.selectView('missing')).toBe(false)
    expect(refresh).not.toHaveBeenCalled()
    expect(await controller.selectView(id)).toBe(true)
    expect(page.q).toBe('失败')
    expect(refresh).toHaveBeenCalledTimes(1)

    page.q = '重试'
    expect(controller.activeViewId.value).toBe(CUSTOM_VIEW_ID)
    expect(controller.editableSourceId.value).toBe(id)
    expect(controller.updateView('builtin-all')).toBe(false)
    expect(controller.updateView('missing')).toBe(false)
    expect(controller.updateView(id)).toBe(true)
    expect(controller.userViews.value[0].snapshot).toEqual({ q: '重试' })
    expect(controller.activeViewId.value).toBe(id)

    expect(await controller.removeView('builtin-all')).toBe(false)
    expect(await controller.removeView('missing')).toBe(false)
    expect(await controller.removeView(id)).toBe(true)
    expect(controller.userViews.value).toEqual([])
    expect(page.q).toBe('')
    expect(controller.activeViewId.value).toBe('builtin-all')
    expect(controller.editableSourceId.value).toBe('')
    expect(refresh).toHaveBeenCalledTimes(2)
    expect(storage.setItem).toHaveBeenCalled()
  })

  it('loads compatible user views from the injected storage key', async () => {
    const storage = createStorage(JSON.stringify({
      version: 1,
      items: [{ id: 'user-loaded', label: '已保存', snapshot: { q: '历史' } }]
    }))
    const { controller, page, refresh } = createController({ storage })

    expect(storage.getItem).toHaveBeenCalledWith('test-saved-views-v1')
    expect(controller.availableViews.value.map((item) => item.id)).toEqual(['builtin-all', 'user-loaded'])
    expect(await controller.selectView('user-loaded')).toBe(true)
    expect(page.q).toBe('历史')
    expect(refresh).toHaveBeenCalledTimes(1)
  })

  it('keeps the session usable when storage reads and writes fail', () => {
    const storage = {
      getItem: () => { throw new Error('denied') },
      setItem: () => { throw new Error('denied') }
    }
    const { controller } = createController({ storage, now: () => 456 })

    expect(controller.saveView('会话视图')).toBe('user-456')
    expect(controller.userViews.value).toHaveLength(1)
    expect(controller.renameView({ id: 'user-456', label: '会话内可用' })).toBe(true)
    expect(controller.userViews.value[0].label).toBe('会话内可用')
  })

  it('does not overwrite a view when the injected clock repeats', () => {
    const { controller, page } = createController()

    expect(controller.saveView('第一个')).toBe('user-123')
    page.q = '不同快照'
    expect(controller.saveView('第二个')).toBe('user-123-2')
    expect(controller.userViews.value.map((item) => item.id)).toEqual(['user-123', 'user-123-2'])
  })

  it('stays independent of page business fields and confirmation UI', () => {
    const source = readFileSync(new URL('./useSavedViews.js', import.meta.url), 'utf8')

    expect(source).not.toMatch(/\bq\b|\bstatus\b|\bcolumns\b/)
    expect(source).not.toContain('ElMessageBox')
  })
})
