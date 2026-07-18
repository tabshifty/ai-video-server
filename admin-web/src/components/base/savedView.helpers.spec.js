import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import {
  CUSTOM_VIEW_ID,
  SAVED_VIEW_SCHEMA_VERSION,
  createSavedViewId,
  parseSavedViewDocument,
  removeSavedView,
  serializeSavedViews,
  snapshotKey,
  upsertSavedView
} from './savedView.helpers'

const normalize = (snapshot) => ({
  q: String(snapshot?.q || ''),
  status: String(snapshot?.status || ''),
  columns: Array.isArray(snapshot?.columns) ? [...snapshot.columns] : []
})

describe('saved view helpers', () => {
  it('exposes the version 1 schema and custom view identity', () => {
    expect(SAVED_VIEW_SCHEMA_VERSION).toBe(1)
    expect(CUSTOM_VIEW_ID).toBe('custom')
  })

  it('compares nested snapshots independent of object key order', () => {
    const first = { status: 'failed', filters: { owner: 'me', range: { to: 2, from: 1 } }, columns: ['title'] }
    const second = { columns: ['title'], filters: { range: { from: 1, to: 2 }, owner: 'me' }, status: 'failed' }

    expect(snapshotKey(first)).toBe(snapshotKey(second))
    expect(snapshotKey(null)).toBe('{}')
  })

  it('round trips versioned user views', () => {
    const items = [{ id: 'user-1', label: '失败处理', snapshot: normalize({ status: 'failed' }) }]

    expect(parseSavedViewDocument(serializeSavedViews(items), normalize)).toEqual(items)
  })

  it('drops corrupt, incompatible, duplicate and invalid records', () => {
    expect(parseSavedViewDocument('{bad', normalize)).toEqual([])
    expect(parseSavedViewDocument('{"version":2,"items":[]}', normalize)).toEqual([])
    expect(parseSavedViewDocument('{"version":1,"items":{}}', normalize)).toEqual([])

    const raw = JSON.stringify({
      version: 1,
      items: [
        { id: 'user-1', label: ' A ', snapshot: { status: 'failed' } },
        { id: 'user-1', label: 'B', snapshot: {} },
        { id: '', label: 'C', snapshot: {} },
        { id: 'builtin-all', label: 'D', snapshot: {} },
        { id: 'user-', label: 'E', snapshot: {} },
        { id: 'user-has space', label: 'F', snapshot: {} },
        { id: 'user-2', label: '   ', snapshot: {} }
      ]
    })

    expect(parseSavedViewDocument(raw, normalize)).toEqual([
      { id: 'user-1', label: 'A', snapshot: normalize({ status: 'failed' }) }
    ])
  })

  it('drops a record whose snapshot cannot be normalized without losing valid records', () => {
    const raw = JSON.stringify({
      version: 1,
      items: [
        { id: 'user-bad', label: '损坏', snapshot: { invalid: true } },
        { id: 'user-good', label: '可用', snapshot: { q: '保留' } }
      ]
    })
    const guardedNormalize = (snapshot) => {
      if (snapshot?.invalid) throw new Error('invalid snapshot')
      return normalize(snapshot)
    }

    expect(parseSavedViewDocument(raw, guardedNormalize)).toEqual([
      { id: 'user-good', label: '可用', snapshot: normalize({ q: '保留' }) }
    ])
  })

  it('drops missing, null and array snapshots before normalization without losing valid records', () => {
    const raw = JSON.stringify({
      version: 1,
      items: [
        { id: 'user-missing', label: '缺失' },
        { id: 'user-null', label: '空值', snapshot: null },
        { id: 'user-array', label: '数组', snapshot: [{ q: '错误结构' }] },
        { id: 'user-good', label: '可用', snapshot: { q: '保留' } }
      ]
    })

    expect(parseSavedViewDocument(raw, normalize)).toEqual([
      { id: 'user-good', label: '可用', snapshot: normalize({ q: '保留' }) }
    ])
  })

  it('upserts and removes immutably while creating deterministic ids', () => {
    const original = [{ id: 'user-1', label: 'A', snapshot: {} }]
    const inserted = upsertSavedView(original, { id: 'user-2', label: 'B', snapshot: {} })
    const updated = upsertSavedView(inserted, { id: 'user-1', label: 'C', snapshot: {} })

    expect(inserted).toHaveLength(2)
    expect(updated.find((item) => item.id === 'user-1')?.label).toBe('C')
    expect(removeSavedView(updated, 'user-1')).toEqual([
      { id: 'user-2', label: 'B', snapshot: {} }
    ])
    expect(original).toEqual([{ id: 'user-1', label: 'A', snapshot: {} }])
    expect(createSavedViewId(123)).toBe('user-123')
  })

  it('remains independent of browser storage', () => {
    const source = readFileSync(new URL('./savedView.helpers.js', import.meta.url), 'utf8')

    expect(source).not.toMatch(/localStorage|sessionStorage/)
  })
})
