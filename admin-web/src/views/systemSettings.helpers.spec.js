import { describe, expect, it } from 'vitest'
import {
  buildOrphanScanPromptKey,
  formatOrphanScanBytes,
  formatOrphanScanTime,
  getEd2kLinkLabel,
  getOrphanScanStatusLabel,
  getOrphanScanStatusType,
  parseEd2kLinks,
  shouldPromptDeleteOrphanScan
} from './systemSettings.helpers'

describe('system settings orphan scan helpers', () => {
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

describe('system settings ed2k helpers', () => {
  it('parses multi-line ed2k text into clickable link records and ignores blank lines', () => {
    const result = parseEd2kLinks(`
      ed2k://|file|first.mkv|123|0123456789ABCDEF0123456789ABCDEF|/

      https://example.test/not-ed2k
      ED2K://|file|second.mkv|456|FEDCBA9876543210FEDCBA9876543210|/
    `)

    expect(result.invalidCount).toBe(1)
    expect(result.links).toHaveLength(2)
    expect(result.links[0]).toMatchObject({
      lineNumber: 2,
      href: 'ed2k://|file|first.mkv|123|0123456789ABCDEF0123456789ABCDEF|/',
      label: 'first.mkv'
    })
    expect(result.links[1].href).toBe('ED2K://|file|second.mkv|456|FEDCBA9876543210FEDCBA9876543210|/')
    expect(result.links[1].id).toContain('5:')
  })

  it('uses decoded file names as ed2k labels and falls back to href for non-file links', () => {
    expect(getEd2kLinkLabel('ed2k://|file|%E4%B8%BB%E8%A7%92.mkv|123|HASH|/')).toBe('主角.mkv')
    expect(getEd2kLinkLabel('ed2k://|server|127.0.0.1|4661|/')).toBe('ed2k://|server|127.0.0.1|4661|/')
  })
})
