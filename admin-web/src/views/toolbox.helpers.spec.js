import { describe, expect, it } from 'vitest'
import {
  getEd2kLinkLabel,
  parseEd2kLinks
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
})
