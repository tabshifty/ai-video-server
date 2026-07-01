import { describe, expect, it } from 'vitest'
import {
  buildEd2kTaskFocusState,
  filterEd2kTasks,
  getEd2kLinkLabel,
  parseEd2kLinks,
  parseEd2kCreateEntries,
  normalizeEd2kCreateResults,
  mergeEd2kCreateSession,
  mergeEd2kDraftEntries,
  buildPendingEd2kInput,
  shouldKeepEd2kCreateDialogOpen
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

  it('treats non-file ed2k links as invalid input for the download workbench', () => {
    const result = parseEd2kLinks(`
      ed2k://|server|127.0.0.1|4661|/
      ed2k://|file|ok.mkv|123|0123456789ABCDEF0123456789ABCDEF|/
    `)

    expect(result.invalidCount).toBe(1)
    expect(result.links).toHaveLength(1)
    expect(result.links[0].label).toBe('ok.mkv')
  })

  it('merges create-session results by original line number and keeps only unresolved input lines', () => {
    const entries = parseEd2kCreateEntries(`
      ed2k://|file|first.mkv|123|0123456789ABCDEF0123456789ABCDEF|/
      ed2k://|file|second.mkv|456|FEDCBA9876543210FEDCBA9876543210|/
      ed2k://|file|third.mkv|789|11111111111111111111111111111111|/
    `)
    const session = mergeEd2kCreateSession([], normalizeEd2kCreateResults([
      { line_number: entries[0].lineNumber, source_link: entries[0].sourceLink, status: 'created', message: 'ok' },
      { line_number: entries[1].lineNumber, source_link: entries[1].sourceLink, status: 'enqueue_failed', message: 'boom' },
      { line_number: entries[2].lineNumber, source_link: entries[2].sourceLink, status: 'reused', message: 'history' }
    ]))

    expect(session.map((item) => `${item.lineNumber}:${item.status}`)).toEqual([
      '1:created',
      '2:enqueue_failed',
      '3:reused'
    ])
    expect(buildPendingEd2kInput(entries, session)).toBe(entries[1].sourceLink)
    expect(shouldKeepEd2kCreateDialogOpen(session)).toBe(true)
  })

  it('keeps first-assigned line numbers for retried input and gives edited links new line numbers', () => {
    const first = parseEd2kCreateEntries(`
      bad-link
      ed2k://|file|ok.mkv|123|0123456789ABCDEF0123456789ABCDEF|/
    `)
    const retried = parseEd2kCreateEntries(`
      bad-link
      ed2k://|file|fixed.mkv|456|FEDCBA9876543210FEDCBA9876543210|/
    `, first)

    expect(retried[0]).toMatchObject({ lineNumber: 1, sourceLink: 'bad-link' })
    expect(retried[1]).toMatchObject({ lineNumber: 3, sourceLink: 'ed2k://|file|fixed.mkv|456|FEDCBA9876543210FEDCBA9876543210|/' })
  })

  it('drops settled entries from the draft while keeping unresolved ones by stable line number', () => {
    const previousEntries = [
      { lineNumber: 1, sourceLink: 'bad-link' },
      { lineNumber: 2, sourceLink: 'ed2k://|file|old.mkv|123|0123456789ABCDEF0123456789ABCDEF|/' }
    ]
    const latestEntries = [
      { lineNumber: 1, sourceLink: 'ed2k://|file|fixed.mkv|456|FEDCBA9876543210FEDCBA9876543210|/' }
    ]
    const results = normalizeEd2kCreateResults([
      { line_number: 2, source_link: previousEntries[1].sourceLink, status: 'created', message: 'ok' },
      { line_number: 1, source_link: latestEntries[0].sourceLink, status: 'invalid', message: 'bad' }
    ])

    expect(mergeEd2kDraftEntries(previousEntries, latestEntries, results)).toEqual([
      { lineNumber: 1, sourceLink: 'ed2k://|file|fixed.mkv|456|FEDCBA9876543210FEDCBA9876543210|/' }
    ])
  })

  it('hides legacy deleted tasks from the visible task list and keeps status filtering client-side', () => {
    const tasks = [
      { id: 'task-1', status: 'queued' },
      { id: 'task-2', status: 'deleted' },
      { id: 'task-3', status: 'cancelled' }
    ]

    expect(filterEd2kTasks(tasks, 'all').map((item) => item.id)).toEqual(['task-1', 'task-3'])
    expect(filterEd2kTasks(tasks, 'cancelled').map((item) => item.id)).toEqual(['task-3'])
  })

  it('switches filter to the target task status when the current filtered view cannot show that task', () => {
    expect(buildEd2kTaskFocusState('running', { id: 'task-1', status: 'cancelled' })).toEqual({
      filter: 'cancelled',
      selectedTaskID: 'task-1'
    })
    expect(buildEd2kTaskFocusState('all', { id: 'task-2', status: 'files_cleaned' })).toEqual({
      filter: 'all',
      selectedTaskID: 'task-2'
    })
  })
})
