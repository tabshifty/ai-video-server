import { describe, expect, it } from 'vitest'
import {
  dedupeEd2kLinksByInfoHash,
  extractApiErrorMessage,
  getEd2kInfoHash,
  getEd2kLinkLabel,
  indexCloudBatchItems,
  parseEd2kLinks,
  summarizeEd2kBatchSelection
} from './toolbox.helpers'

describe('toolbox ed2k helpers', () => {
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

  it('treats non-file ed2k links as invalid input for the link generator', () => {
    const result = parseEd2kLinks(`
      ed2k://|server|127.0.0.1|4661|/
      ed2k://|file|ok.mkv|123|0123456789ABCDEF0123456789ABCDEF|/
    `)

    expect(result.invalidCount).toBe(1)
    expect(result.links).toHaveLength(1)
    expect(result.links[0].label).toBe('ok.mkv')
  })

  it('extracts normalized info hashes and rejects non-file or malformed links', () => {
    expect(getEd2kInfoHash('ed2k://|file|a.mkv|1|0123456789ABCDEF0123456789ABCDEF|/'))
      .toBe('0123456789abcdef0123456789abcdef')
    expect(getEd2kInfoHash('ED2K://|file|a.mkv|1|0123456789abcdef0123456789abcdef|/'))
      .toBe('0123456789abcdef0123456789abcdef')
    expect(getEd2kInfoHash('ed2k://|server|127.0.0.1|4661|/')).toBe('')
    expect(getEd2kInfoHash('ed2k://|file|a.mkv|1|not-a-hash|/')).toBe('')
  })

  it('deduplicates selected links by info hash and keeps the first occurrence', () => {
    const first = {
      id: '1:a',
      href: 'ed2k://|file|a.mkv|1|0123456789ABCDEF0123456789ABCDEF|/',
      infoHash: '0123456789abcdef0123456789abcdef'
    }
    const duplicate = {
      id: '2:b',
      href: 'ed2k://|file|a-copy.mkv|1|0123456789abcdef0123456789abcdef|/',
      infoHash: '0123456789abcdef0123456789abcdef'
    }
    const other = {
      id: '3:c',
      href: 'ed2k://|file|c.mkv|1|FEDCBA9876543210FEDCBA9876543210|/',
      infoHash: 'fedcba9876543210fedcba9876543210'
    }

    const result = dedupeEd2kLinksByInfoHash([first, duplicate, other])

    expect(result.links).toEqual([first, other])
    expect(result.duplicates).toEqual([duplicate])
  })

  it('extracts backend error messages from the response envelope', () => {
    expect(extractApiErrorMessage({ response: { data: { msg: '云下载功能未启用' } } }, 'fallback'))
      .toBe('云下载功能未启用')
    expect(extractApiErrorMessage({ response: { status: 404 } }, 'fallback')).toBe('云下载功能未启用')
    expect(extractApiErrorMessage({ response: { status: 503 } }, 'fallback')).toBe('云下载功能未启用')
    expect(extractApiErrorMessage({ code: 404, message: 'cloud: not found' }, 'fallback'))
      .toBe('云下载功能未启用')
    expect(extractApiErrorMessage({ code: 503, message: '云下载导入功能未启用' }, 'fallback'))
      .toBe('云下载功能未启用')
    expect(extractApiErrorMessage({ message: 'network down' }, 'fallback')).toBe('network down')
    expect(extractApiErrorMessage({}, 'fallback')).toBe('fallback')
  })

  it('indexes an invalid batch result by both its raw link and parsed hash', () => {
    const link = 'ed2k://|file|bad-size.mkv|not-a-number|0123456789ABCDEF0123456789ABCDEF|/'
    const item = { link, info_hash: '', status: 'invalid' }

    const index = indexCloudBatchItems([item])

    expect(index[link]).toBe(item)
    expect(index['0123456789abcdef0123456789abcdef']).toBe(item)
  })

  it('separates invalid-hash selections and counts existing task states', () => {
    const hashA = 'a'.repeat(32)
    const hashB = 'b'.repeat(32)
    const hashC = 'c'.repeat(32)
    const links = [
      { id: '1', href: 'ed2k://|file|a|1|' + hashA + '|/', infoHash: hashA },
      { id: '2', href: 'ed2k://|file|b|1|' + hashB + '|/', infoHash: hashB },
      { id: '3', href: 'ed2k://|file|c|1|' + hashC + '|/', infoHash: hashC },
      { id: '4', href: 'ed2k://|file|bad|1|not-a-hash|/', infoHash: '' }
    ]
    const summary = summarizeEd2kBatchSelection(links, [{ id: '5' }], {
      [hashA]: { status: 'completed' },
      [hashB]: { status: 'queued' },
      [hashC]: { status: 'failed' }
    })

    expect(summary.invalidLinks).toEqual([links[3]])
    expect(summary.submittableLinks).toEqual([links[0], links[1], links[2]])
    expect(summary.hashes).toEqual([hashA, hashB, hashC])
    expect(summary.duplicateCount).toBe(1)
    expect(summary.completed).toBe(1)
    expect(summary.active).toBe(1)
    expect(summary.retry).toBe(1)
    expect(summary.fresh).toBe(0)
  })

})
