import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import VideoList from './VideoList.vue'

const source = readFileSync(new URL('./VideoList.vue', import.meta.url), 'utf8')

function extractBlock(tag) {
  const opening = source.match(new RegExp(`<${tag}[^>]*>`))

  expect(opening).not.toBeNull()
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const closing = `</${tag}>`
  const end = tag === 'template' ? source.lastIndexOf(closing) : source.indexOf(closing, start)

  expect(end).toBeGreaterThan(start)
  return source.slice(start, end)
}

function extractBalancedBraceBlock(sourceText, openingPattern) {
  const opening = sourceText.match(openingPattern)
  if (!opening || opening.index === undefined) return null

  const openBraceIndex = opening.index + opening[0].lastIndexOf('{')
  let depth = 0

  for (let index = openBraceIndex; index < sourceText.length; index += 1) {
    if (sourceText[index] === '{') depth += 1
    if (sourceText[index] !== '}') continue

    depth -= 1
    if (depth === 0) {
      return {
        body: sourceText.slice(openBraceIndex + 1, index),
        closeBraceIndex: index,
        openingIndex: opening.index
      }
    }
  }

  return null
}

function extractElement(sourceText, openingPattern, closingTag) {
  const opening = sourceText.match(openingPattern)
  if (!opening || opening.index === undefined) return ''

  const start = opening.index
  const end = sourceText.indexOf(closingTag, start)
  if (end < 0) return ''
  return sourceText.slice(start, end + closingTag.length)
}

function findRule(styleSource, selector) {
  const escapedSelector = selector.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const match = styleSource.match(new RegExp(`${escapedSelector}\\s*\\{([^}]*)\\}`))

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

function ownsSnapshotApplyOrder(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /applySnapshot:\s*\(snapshot\)\s*=>\s*\{/)
  if (!block) return false

  const tokens = [
    'const next = normalizeVideoViewSnapshot(',
    'query.q = next.q',
    'query.type = next.type',
    'query.status = next.status',
    'query.page = 1',
    'columnVisibility.value = [...next.columns]',
    'persistColumns()',
    'clearSelection()'
  ]
  const indexes = tokens.map((token) => block.body.indexOf(token))

  return indexes.every((index) => index >= 0)
    && indexes.every((index, position) => position === 0 || indexes[position - 1] < index)
}

function catchClearsRows(sourceText) {
  const catchBlock = extractBalancedBraceBlock(sourceText, /catch\s*\(error\)\s*\{/)
  return catchBlock ? /list\.value\s*=\s*\[\]/.test(catchBlock.body) : false
}

const script = extractBlock('script')
const template = extractBlock('template')
const style = extractBlock('style')
const savedViewsOptions = extractBalancedBraceBlock(script, /useSavedViews\(\s*\{/)
const currentSnapshotBlock = savedViewsOptions
  ? extractBalancedBraceBlock(savedViewsOptions.body, /getCurrentSnapshot:\s*\(\)\s*=>\s*\(\s*\{/)
  : null
const loadBlock = extractBalancedBraceBlock(script, /async function load\(\)\s*\{/)
const deleteBlock = extractBalancedBraceBlock(script, /async function doDelete\(row\)\s*\{/)
const operationsColumn = extractElement(
  template,
  /<el-table-column\b(?=[^>]*isColumnVisible\('operations'\))[^>]*>/,
  '</el-table-column>'
)

describe('视频资源集合页', () => {
  it('通过真实 SFC 编译并只接入共享保存视图职责', () => {
    expect(VideoList).toBeTruthy()
    expect(script).toContain("import SavedViewTabs from '../components/base/SavedViewTabs.vue'")
    expect(script).toContain("import { useSavedViews } from '../components/base/useSavedViews'")
    expect(script).toContain("const SAVED_VIEWS_KEY = 'admin-videolist-saved-views-v1'")
    expect(savedViewsOptions).not.toBeNull()
    expect(savedViewsOptions?.body).toContain('storageKey: SAVED_VIEWS_KEY')
    expect(savedViewsOptions?.body).toContain('builtInViews')
    expect(savedViewsOptions?.body).toContain('normalizeSnapshot:')
    expect(savedViewsOptions?.body).toContain('refresh: load')
    expect(script).not.toMatch(/function\s+(?:persistUserViews|readUserViews)\b/)
    expect(script).not.toContain('selectedViewID')
    expect(script).not.toContain("from '../components/base/savedView.helpers'")
    expect(script).not.toContain("ElMessageBox.prompt('请输入视图名称'")
    expect(script).toContain("ElMessageBox.prompt('请输入弃刮原因'")
  })

  it('快照只包含查询和列，并按约定顺序应用后交给共享 refresh', () => {
    const validFixture = `applySnapshot: (snapshot) => {
  const next = normalizeVideoViewSnapshot(snapshot)
  query.q = next.q
  query.type = next.type
  query.status = next.status
  query.page = 1
  columnVisibility.value = [...next.columns]
  persistColumns()
  clearSelection()
}`
    const invalidFixture = validFixture.replace('persistColumns()\n  clearSelection()', 'clearSelection()\n  persistColumns()')

    expect(ownsSnapshotApplyOrder(validFixture)).toBe(true)
    expect(ownsSnapshotApplyOrder(invalidFixture)).toBe(false)
    expect(ownsSnapshotApplyOrder(savedViewsOptions?.body || '')).toBe(true)
    expect(currentSnapshotBlock).not.toBeNull()
    expect(currentSnapshotBlock?.body.match(/^\s*(q|type|status|columns):/gm)?.map((line) => line.trim().split(':')[0])).toEqual([
      'q',
      'type',
      'status',
      'columns'
    ])
    expect(currentSnapshotBlock?.body).not.toMatch(/\bpage\b|selectedRows|detailVisible|filterDrawerVisible/)
  })

  it('使用壳层操作区、保存视图和紧凑筛选工具条', () => {
    expect(script).not.toContain('PageHeader')
    expect(template).not.toContain('<PageHeader')
    expect(template).toContain('<template #header-actions>')
    expect(template).toContain("router.push('/upload')")
    expect(template).toContain('<SavedViewTabs')
    expect(template).toContain(':items="availableViews"')
    expect(template).toContain(':active-id="activeViewId"')
    expect(template).toContain(':editable-source-id="editableSourceId"')
    expect(template).toContain('@select="selectView"')
    expect(template).toContain('@save="saveView"')
    expect(template).toContain('@update="updateView"')
    expect(template).toContain('@rename="renameView"')
    expect(template).toContain('@remove="removeView"')
    expect(template).toContain('<Toolbar dense>')
    expect(template).toContain('class="page-shell video-list-page" data-density="compact"')
  })

  it('保留已有行并以行内错误、首次骨架和诚实空态反馈请求结果', () => {
    const violatingLoad = `async function load() {
  try {
    list.value = await getAdminVideos(query)
  } catch (error) {
    listError.value = error.message
    list.value = []
  }
}`

    expect(catchClearsRows(violatingLoad)).toBe(true)
    expect(loadBlock).not.toBeNull()
    expect(catchClearsRows(loadBlock?.body || '')).toBe(false)
    expect(loadBlock?.body).toContain("listError.value = ''")
    expect(loadBlock?.body).toContain("listError.value = error?.message || '加载视频列表失败'")
    expect(loadBlock?.body.indexOf('clearSelection()')).toBeGreaterThan(loadBlock?.body.indexOf('list.value = data.items || []'))
    expect(template).toContain('<el-alert v-if="listError"')
    expect(template).toContain('<el-skeleton v-if="listLoading && list.length === 0"')
    expect(template).toContain('<SectionCard v-else-if="!listError || list.length > 0" dense>')
    expect(template).toContain("hasActiveFilters ? '当前筛选无结果' : '暂无视频'")
    expect(template).toContain("hasActiveFilters ? '清除筛选后查看全部视频' : '上传视频后会显示在这里'")
    expect(template).toContain('@click="resetFilters">清除筛选</el-button>')
    expect(template).toContain('共 {{ total }} 条')
  })

  it('详情常驻且低频行操作进入可访问菜单并保留删除确认', () => {
    expect(operationsColumn).toContain('width="108"')
    expect(operationsColumn).toContain('fixed="right"')
    expect(operationsColumn).toMatch(/<el-button\b[^>]*link[^>]*@click="showDetail\(row\)"[^>]*>详情<\/el-button>/)
    expect(operationsColumn).toContain('<el-dropdown')
    expect(operationsColumn).toContain('content="更多视频操作"')
    expect(operationsColumn).toContain('aria-label="更多视频操作"')
    expect(operationsColumn).toContain('command="retranscode"')
    expect(operationsColumn).toContain('command="delete"')
    expect(operationsColumn).not.toContain('@click="doRetranscode(row)"')
    expect(operationsColumn).not.toContain('@click="doDelete(row)"')
    expect(deleteBlock).not.toBeNull()
    expect(deleteBlock?.body.indexOf('ElMessageBox.confirm')).toBeGreaterThan(-1)
    expect(deleteBlock?.body.indexOf('deleteAdminVideo(row.id)')).toBeGreaterThan(deleteBlock?.body.indexOf('ElMessageBox.confirm'))
  })

  it('使用固定媒体几何、状态组件和防溢出表格', () => {
    expect(template).toContain('class="table-wrap has-media-rows"')
    expect(template).toContain('<StatusIndicator')
    expect(template).toContain('width="44"')
    expect(template).toContain('show-overflow-tooltip')
    expect(findRule(style, '.video-cover-cell')).toContain('width: 72px')
    expect(findRule(style, '.video-cover-cell')).toContain('height: 40px')
    expect(findRule(style, '.video-cover-image')).toContain('width: 72px')
    expect(findRule(style, '.video-cover-image')).toContain('height: 40px')
    expect(findRule(style, '.video-cover-placeholder')).toContain('width: 72px')
    expect(findRule(style, '.video-cover-placeholder')).toContain('height: 40px')
    expect(style).toContain('color-mix(')
    expect(style).not.toMatch(/#[0-9a-f]{3,8}\b/i)
    expect(style).not.toMatch(/rgba?\(\s*\d/i)
    expect(style).not.toContain('linear-gradient')
    expect(style).not.toContain('box-shadow')
    expect(style).not.toContain('letter-spacing')
  })

  it('窄屏壳层操作和 Teleport 菜单保持至少 44px 点击目标', () => {
    const mediaStart = style.indexOf('@media (max-width: 63.9375rem)')

    expect(template).toContain('popper-class="video-column-settings-popper"')
    expect(operationsColumn).toContain('popper-class="video-row-actions-popper"')
    expect(mediaStart).toBeGreaterThan(-1)
    const mobileStyle = style.slice(mediaStart)
    expect(findRule(mobileStyle, '.video-header-actions :deep(.el-button)')).toContain('min-height: 44px')
    expect(findRule(mobileStyle, ':global(.video-column-settings-popper .el-checkbox)')).toContain('min-height: 44px')
    expect(findRule(mobileStyle, ':global(.video-row-actions-popper .el-dropdown-menu__item)')).toContain('min-height: 44px')
  })

  it('保留选择、批量操作、响应式列、Drawer、字幕和编辑 payload', () => {
    expect(script).toContain('viewportWidth.value < 1280')
    expect(script).toContain('function onRowSelectionSelect(selection, row)')
    expect(script).toContain('applySelectionByIDs(Array.from(selectedSet), selectionAnchorIndex.value)')
    expect(script).toContain('function clearSelection()')
    expect(script).toContain('const batchActionBusy = computed(() => deletingBatch.value || updatingBatch.value)')
    expect(script).toContain("const detailDrawerSize = computed(() => (viewportWidth.value < 1024 ? '100%' : '560px'))")
    expect(script).toContain("formData.append('file', file)")
    expect(script).toContain('actor_ids: actorIDs')
    expect(template).toContain(':before-close="handleDetailBeforeClose"')
    expect(template).toContain('<BulkActionBar :count="selectedRows.length" :actions="bulkActions" />')
    expect(template).toContain('<AdminTablePagination')
    expect(template).toContain('@current-change="load"')
  })
})
