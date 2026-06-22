import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const displayFiles = [
  './ActorManage.vue',
  './CollectionManage.vue',
  './ImageCollectionManage.vue',
  './ImageManage.vue',
  './IPTVManage.vue',
  './ShortReview.vue',
  './TaskMonitor.vue',
  './ToolboxArchiveImport.vue',
  './ToolboxImageWorkbench.vue',
  './ToolboxOrphanFiles.vue',
  './ToolboxPasswordVault.vue',
  './TvAppManage.vue',
  './UserManage.vue',
  './VideoList.vue',
  './toolboxOrphanFiles.helpers.js'
]

function readDisplaySources() {
  return displayFiles.map((file) => ({
    file,
    source: readFileSync(new URL(file, import.meta.url), 'utf8')
  }))
}

describe('admin date time display', () => {
  it('does not use locale-dependent formatting in visible time displays', () => {
    for (const { file, source } of readDisplaySources()) {
      expect(source, file).not.toContain('toLocaleString(')
    }
  })

  it('keeps visible time formatting routed through the shared formatter', () => {
    const sources = readDisplaySources()
    const combined = sources.map(({ source }) => source).join('\n')

    expect(combined).toContain('formatAdminDateTime')
    expect(combined).toContain('formatOrphanScanTime')
  })

  it('does not render absolute timestamp table columns through raw props', () => {
    const rawTimestampProp = /prop="(?:created_at|updated_at|started_at|progress_updated_at|original_uploaded_at|published_at|last_status_changed_at|mod_time)"/

    for (const { file, source } of readDisplaySources()) {
      expect(source, file).not.toMatch(rawTimestampProp)
    }
  })
})
