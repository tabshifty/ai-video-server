import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import Dashboard from './Dashboard.vue'
import ImageManage from './ImageManage.vue'
import TaskMonitor from './TaskMonitor.vue'
import VideoList from './VideoList.vue'

const readView = (file) => readFileSync(new URL(`./${file}`, import.meta.url), 'utf8')
const router = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
const migratedViews = [
  { file: 'Dashboard.vue', component: 'Dashboard', density: 'compact', compiled: Dashboard },
  { file: 'TaskMonitor.vue', component: 'TaskMonitor', density: 'monitor', compiled: TaskMonitor },
  { file: 'VideoList.vue', component: 'VideoList', density: 'compact', compiled: VideoList },
  { file: 'ImageManage.vue', component: 'ImageManage', density: 'compact', compiled: ImageManage }
]
const pendingShellViews = [
  'AVManualScrape.vue',
  'ActorManage.vue',
  'CollectionManage.vue',
  'IPTVManage.vue',
  'ImageCollectionManage.vue',
  'PendingDeleteShorts.vue',
  'ScrapePreview.vue',
  'SystemSettings.vue',
  'Toolbox.vue',
  'TvAppManage.vue',
  'TvSeriesManage.vue',
  'UserManage.vue',
  'VideoUpload.vue'
]

function extractTemplate(source) {
  const opening = source.match(/<template[^>]*>/)

  expect(opening).not.toBeNull()
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const end = source.lastIndexOf('</template>')

  expect(end).toBeGreaterThan(start)
  return source.slice(start, end)
}

function routeLine(component) {
  const componentPattern = new RegExp(`\\bcomponent:\\s*${component}(?=\\s*[,}])`)
  return router.split('\n').find((line) => componentPattern.test(line)) || ''
}

function exactRoutePattern(component, withCompatibilityMeta) {
  const meta = withCompatibilityMeta
    ? ',\\s*meta:\\s*\\{\\s*hideShellPageHeader:\\s*true\\s*\\}'
    : ''

  return new RegExp(
    `^\\s*\\{\\s*path:\\s*'[^']+',\\s*component:\\s*${component}${meta}\\s*\\},?\\s*$`
  )
}

describe('Precision Ops 第一阶段 rollout', () => {
  it('固定四个已迁移页面与 13 个兼容页面，且集合互不重叠', () => {
    const migratedFiles = migratedViews.map(({ file }) => file)

    expect(migratedViews).toHaveLength(4)
    expect(pendingShellViews).toHaveLength(13)
    expect(new Set(migratedFiles).size).toBe(migratedFiles.length)
    expect(new Set(pendingShellViews).size).toBe(pendingShellViews.length)
    expect(migratedFiles.filter((file) => pendingShellViews.includes(file))).toEqual([])
  })

  migratedViews.forEach(({ file, component, density, compiled }) => {
    it(`${file} 通过真实 SFC 编译并使用合并工作区页头`, () => {
      const template = extractTemplate(readView(file))
      const line = routeLine(component)

      expect(compiled).toBeTruthy()
      expect(template).toMatch(/<Layout(?:\\s|>)/)
      expect(template).toContain(`data-density="${density}"`)
      expect(template).not.toContain('<PageHeader')
      expect(line).toMatch(exactRoutePattern(component, false))
    })
  })

  it('13 个待迁移 shell 页面保持 PageHeader 与精确兼容 meta', () => {
    pendingShellViews.forEach((file) => {
      const component = file.replace('.vue', '')
      const template = extractTemplate(readView(file))
      const line = routeLine(component)

      expect(template, file).toContain('<PageHeader')
      expect(line, component).toMatch(exactRoutePattern(component, true))
    })
  })
})
