import { describe, expect, it } from 'vitest'
import {
  SHELL_PREFERENCE_VERSION,
  ensureActiveGroup,
  parseExpandedGroupKeys,
  parseRecentRoutes,
  pushRecentRoute,
  serializeExpandedGroupKeys,
  serializeRecentRoutes
} from './adminShellPreferences'
import { adminShellNavGroups } from './base/commandPalette.helpers'

const paths = ['/dashboard', '/videos', '/tasks', '/toolbox']
const groups = ['workspace', 'content', 'ingest', 'operations', 'system', 'service-tools']

describe('admin shell preferences', () => {
  it('keeps three unique recent known routes with newest first', () => {
    expect(pushRecentRoute(['/videos', '/dashboard'], '/tasks', paths)).toEqual(['/tasks', '/videos', '/dashboard'])
    expect(pushRecentRoute(['/tasks', '/videos', '/dashboard'], '/videos', paths)).toEqual(['/videos', '/tasks', '/dashboard'])
    expect(pushRecentRoute(['/tasks'], '/unknown', paths)).toEqual(['/tasks'])
    expect(pushRecentRoute(['/unknown', '/tasks', '/tasks', '/videos'], '/dashboard', paths, 2)).toEqual(['/dashboard', '/tasks'])
  })

  it('caps the recent route limit at three', () => {
    const existing = ['/videos', '/tasks', '/toolbox']

    expect(pushRecentRoute(existing, '/dashboard', paths, 4)).toEqual(['/dashboard', '/videos', '/tasks'])
  })

  it('returns no recent routes for zero or negative limits', () => {
    const existing = ['/videos', '/tasks', '/toolbox']

    expect(pushRecentRoute(existing, '/dashboard', paths, 0)).toEqual([])
    expect(pushRecentRoute(existing, '/dashboard', paths, -1)).toEqual([])
  })

  it('truncates fractional limits and defaults non-finite limits to three', () => {
    const existing = ['/videos', '/tasks', '/toolbox']

    expect(pushRecentRoute(existing, '/dashboard', paths, 2.9)).toEqual(['/dashboard', '/videos'])
    expect(pushRecentRoute(existing, '/dashboard', paths, Number.POSITIVE_INFINITY)).toEqual(['/dashboard', '/videos', '/tasks'])
    expect(pushRecentRoute(existing, '/dashboard', paths, Number.NaN)).toEqual(['/dashboard', '/videos', '/tasks'])
  })

  it('filters unknown and duplicate recent routes and caps parsed history at three', () => {
    const raw = JSON.stringify({
      version: SHELL_PREFERENCE_VERSION,
      paths: ['/videos', '/unknown', '/videos', '/tasks', '/dashboard', '/toolbox']
    })

    expect(parseRecentRoutes(raw, paths)).toEqual(['/videos', '/tasks', '/dashboard'])
  })

  it('falls back safely for corrupt or incompatible documents', () => {
    expect(parseRecentRoutes('{bad', paths)).toEqual([])
    expect(parseRecentRoutes('{"version":2,"paths":["/videos"]}', paths)).toEqual([])
    expect(parseRecentRoutes(null, paths)).toEqual([])
    expect(parseExpandedGroupKeys('{bad', groups)).toEqual(groups)
    expect(parseExpandedGroupKeys('{"version":2,"keys":["overview"]}', groups)).toEqual(groups)
  })

  it.each([
    ['missing keys', '{"version":1}'],
    ['null keys', '{"version":1,"keys":null}'],
    ['non-array keys', '{"version":1,"keys":"overview"}']
  ])('falls back for a semantically corrupt v1 document with %s', (_, raw) => {
    expect(parseExpandedGroupKeys(raw, groups)).toEqual(groups)
  })

  it('preserves a valid empty expanded group list', () => {
    expect(parseExpandedGroupKeys('{"version":1,"keys":[]}', groups)).toEqual([])
  })

  it('round trips versioned documents with the documented structure', () => {
    const recentRaw = serializeRecentRoutes(['/videos'])
    const expandedRaw = serializeExpandedGroupKeys(['workspace'])

    expect(SHELL_PREFERENCE_VERSION).toBe(1)
    expect(JSON.parse(recentRaw)).toEqual({ version: 1, paths: ['/videos'] })
    expect(JSON.parse(expandedRaw)).toEqual({ version: 1, keys: ['workspace'] })
    expect(parseRecentRoutes(recentRaw, paths)).toEqual(['/videos'])
    expect(parseExpandedGroupKeys(expandedRaw, groups)).toEqual(['workspace'])
  })

  it('filters duplicate expanded groups and forces a known active group open once', () => {
    const raw = JSON.stringify({
      version: SHELL_PREFERENCE_VERSION,
      keys: ['workspace', 'workspace', 'system', 'system']
    })

    expect(parseExpandedGroupKeys(raw, groups)).toEqual(['workspace', 'system'])
    expect(ensureActiveGroup(['workspace', 'content', 'workspace'], 'content', groups)).toEqual(['workspace', 'content'])
    expect(ensureActiveGroup(['workspace'], 'service-tools', groups)).toEqual(['workspace', 'service-tools'])
    expect(ensureActiveGroup(['workspace'], 'unknown', groups)).toEqual(['workspace'])
  })

  it('re-expands every group when stored keys belong to an older nav layout', () => {
    const legacyRaw = JSON.stringify({
      version: SHELL_PREFERENCE_VERSION,
      keys: ['overview', 'media']
    })

    expect(parseExpandedGroupKeys(legacyRaw, groups)).toEqual(groups)
  })

  it('exposes exactly six purpose-based navigation groups', () => {
    expect(adminShellNavGroups.map((group) => [group.key, group.label])).toEqual([
      ['workspace', '工作台'],
      ['content', '内容管理'],
      ['ingest', '采集与导入'],
      ['operations', '任务与运维'],
      ['system', '应用与系统'],
      ['service-tools', '服务与工具']
    ])
  })
})
