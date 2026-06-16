import { describe, expect, it } from 'vitest'
import {
  buildOrphanScanPromptKey,
  formatOrphanScanBytes,
  formatOrphanScanTime,
  getOrphanScanStatusLabel,
  getOrphanScanStatusType,
  shouldPromptDeleteOrphanScan
} from './toolboxOrphanFiles.helpers'

describe('toolbox orphan file scan helpers', () => {
  it('formats orphan scan state labels and tag types', () => {
    expect(getOrphanScanStatusLabel('completed')).toBe('已完成')
    expect(getOrphanScanStatusLabel('unknown')).toBe('未知')
    expect(getOrphanScanStatusType('failed')).toBe('danger')
    expect(getOrphanScanStatusType('unknown')).toBe('info')
  })

  it('formats orphan scan bytes and time values', () => {
    const date = new Date('2026-06-05T01:02:03.000Z')
    const wantTime = `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')} ${String(date.getHours()).padStart(2, '0')}:${String(date.getMinutes()).padStart(2, '0')}:${String(date.getSeconds()).padStart(2, '0')}`

    expect(formatOrphanScanBytes(0)).toBe('0 B')
    expect(formatOrphanScanBytes(1536)).toBe('1.5 KB')
    expect(formatOrphanScanBytes(1024 * 1024)).toBe('1.0 MB')
    expect(formatOrphanScanTime(date.toISOString())).toBe(wantTime)
    expect(formatOrphanScanTime('bad-value')).toBe('--')
  })

  it('builds a stable prompt key for completed scans and suppresses repeated prompts', () => {
    const scan = {
      id: 1,
      status: 'completed',
      finished_at: '2026-06-05T09:00:00.000Z',
      updated_at: '2026-06-05T09:00:01.000Z',
      orphan_files: 3,
      deleted_files: 0
    }

    const promptKey = buildOrphanScanPromptKey(scan)
    expect(promptKey).toContain('1:')
    expect(promptKey).toContain(':3:0')
    expect(shouldPromptDeleteOrphanScan(scan, '')).toBe(true)
    expect(shouldPromptDeleteOrphanScan(scan, promptKey)).toBe(false)
    expect(shouldPromptDeleteOrphanScan({ ...scan, orphan_files: 0 }, '')).toBe(false)
    expect(shouldPromptDeleteOrphanScan({ ...scan, status: 'deleted' }, '')).toBe(false)
  })
})
