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
    'resetQueryIdentity()'
  ]
  const indexes = tokens.map((token) => block.body.indexOf(token))

  return indexes.every((index) => index >= 0)
    && indexes.every((index, position) => position === 0 || indexes[position - 1] < index)
}

function resetsBeforeLoad(sourceText, openingPattern) {
  const block = extractBalancedBraceBlock(sourceText, openingPattern)
  if (!block) return false

  const resetIndex = block.body.indexOf('resetQueryIdentity()')
  const loadIndex = block.body.indexOf('load()')
  return resetIndex >= 0 && loadIndex > resetIndex
}

function latestRequestOwnsState(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /async function load\(\)\s*\{/)
  if (!block) return false
  const body = block.body
  const requestIndex = body.indexOf('const data = await getAdminVideos(query)')
  const catchMatch = body.match(/catch\s*\(error\)\s*\{/)
  if (!body.includes('const seq = ++loadSeq') || requestIndex < 0 || !catchMatch || catchMatch.index === undefined) return false

  const successSource = body.slice(requestIndex, catchMatch.index)
  const successGuard = extractBalancedBraceBlock(successSource, /if\s*\(seq !== loadSeq\)\s*\{/)
  if (!successGuard || successGuard.body.trim() !== 'return null') return false
  const successTail = successSource.slice(successGuard.closeBraceIndex + 1)
  if (!['list.value = data.items || []', 'total.value = data.total_count || 0', "listError.value = ''", 'clearSelection()', 'return true']
    .every((token) => successTail.includes(token))) return false

  const catchBlock = extractBalancedBraceBlock(body, /catch\s*\(error\)\s*\{/)
  const catchGuard = catchBlock
    ? extractBalancedBraceBlock(catchBlock.body, /if\s*\(seq !== loadSeq\)\s*\{/)
    : null
  if (!catchBlock || !catchGuard || catchGuard.body.trim() !== 'return null') return false
  const catchTail = catchBlock.body.slice(catchGuard.closeBraceIndex + 1)
  if (!catchTail.includes("listError.value = error?.message || '加载视频列表失败'") || !catchTail.includes('return false')) return false
  if (/list\.value\s*=|total\.value\s*=|clearSelection\(\)/.test(catchTail)) return false

  const finallyBlock = extractBalancedBraceBlock(body, /finally\s*\{/)
  const ownerGuard = finallyBlock
    ? extractBalancedBraceBlock(finallyBlock.body, /if\s*\(seq === loadSeq\)\s*\{/)
    : null
  if (!finallyBlock || !ownerGuard || !ownerGuard.body.includes('listLoading.value = false')) return false
  const outsideOwner = finallyBlock.body.slice(0, ownerGuard.openingIndex)
    + finallyBlock.body.slice(ownerGuard.closeBraceIndex + 1)
  return outsideOwner.trim() === '' && (body.match(/listLoading\.value\s*=\s*false/g) || []).length === 1
}

function handlesRowActionFailClosed(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /function handleVideoRowAction\(command, row\)\s*\{/)
  if (!block) return false

  const retranscodeBlock = extractBalancedBraceBlock(block.body, /if\s*\(command === 'retranscode'\)\s*\{/)
  const deleteBlock = extractBalancedBraceBlock(block.body, /if\s*\(command === 'delete'\)\s*\{/)
  if (!retranscodeBlock || !deleteBlock) return false
  if (!retranscodeBlock.body.includes('doRetranscode(row)') || !retranscodeBlock.body.includes('return true')) return false
  if (!deleteBlock.body.includes('doDelete(row)') || !deleteBlock.body.includes('return true')) return false

  const tail = block.body.slice(deleteBlock.closeBraceIndex + 1)
  return tail.includes('return false')
    && (block.body.match(/doRetranscode\(row\)/g) || []).length === 1
    && (block.body.match(/doDelete\(row\)/g) || []).length === 1
}

function consumesRowActionRejections(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /function handleVideoRowAction\(command, row\)\s*\{/)
  if (!block) return false

  const retranscodeBlock = extractBalancedBraceBlock(block.body, /if\s*\(command === 'retranscode'\)\s*\{/)
  const deleteBlock = extractBalancedBraceBlock(block.body, /if\s*\(command === 'delete'\)\s*\{/)
  return Boolean(
    retranscodeBlock
    && deleteBlock
    && /doRetranscode\(row\)\.catch\(/.test(retranscodeBlock.body)
    && /doDelete\(row\)\.catch\(/.test(deleteBlock.body)
  )
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
const filterDrawer = extractElement(template, /<el-drawer\b(?=[^>]*v-model="filterDrawerVisible")[^>]*>/, '</el-drawer>')
const paginationTag = template.match(/<AdminTablePagination\b[\s\S]*?\/>/)?.[0] || ''
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
    expect(savedViewsOptions?.body).toContain('refresh: refreshSavedView')
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
  resetQueryIdentity()
}`
    const invalidFixture = validFixture.replace('persistColumns()\n  resetQueryIdentity()', 'resetQueryIdentity()\n  persistColumns()')

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

  it('所有新查询身份先清旧数据并通过唯一受控入口加载', () => {
    const validFixture = `function setPage(page) {
  query.page = page
  resetQueryIdentity()
  load()
}`
    const unsafeFixture = validFixture.replace('resetQueryIdentity()\n  load()', 'load()\n  resetQueryIdentity()')
    const resetBlock = extractBalancedBraceBlock(script, /function resetQueryIdentity\(\)\s*\{/)

    expect(resetsBeforeLoad(validFixture, /function setPage\(page\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(unsafeFixture, /function setPage\(page\)\s*\{/)).toBe(false)
    expect(resetBlock).not.toBeNull()
    expect(resetBlock?.body).toContain('list.value = []')
    expect(resetBlock?.body).toContain('total.value = 0')
    expect(resetBlock?.body).toContain("listError.value = ''")
    expect(resetBlock?.body).toContain('clearSelection()')
    expect(ownsSnapshotApplyOrder(savedViewsOptions?.body || '')).toBe(true)
    expect(resetsBeforeLoad(script, /function applyFilters\(\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function applyFilterDrawer\(\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function removeFilter\(key\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function resetFilters\(\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function setPage\(page\)\s*\{/)).toBe(true)
    expect(script).toContain("const quickSearch = ref(query.q)")
    expect(script).toContain('const filterDraft = reactive({ q: query.q, type: query.type, status: query.status })')
    expect(template).toContain('v-model="quickSearch"')
    expect(template).toContain('@click="openFilterDrawer"')
    expect(filterDrawer).toContain('v-model="filterDraft.q"')
    expect(filterDrawer).toContain('v-model="filterDraft.type"')
    expect(filterDrawer).toContain('v-model="filterDraft.status"')
    expect(filterDrawer).toContain('@closed="syncFilterDraftFromQuery"')
    expect(filterDrawer).toContain('@click="applyFilterDrawer"')
    expect(filterDrawer).not.toMatch(/v-model="query\.(?:q|type|status)"/)
    expect(paginationTag).toContain(':current-page="query.page"')
    expect(paginationTag).toContain(':page-size="query.page_size"')
    expect(paginationTag).toContain('@current-change="setPage"')
    expect(paginationTag).not.toContain('v-model:current-page')
    expect(paginationTag).not.toContain('v-model:page-size')
  })

  it('只有最新请求可以写入列表、错误、选择和加载完成状态', () => {
    const protectedLoad = `let loadSeq = 0
async function load() {
  const seq = ++loadSeq
  listLoading.value = true
  try {
    const data = await getAdminVideos(query)
    if (seq !== loadSeq) {
      return null
    }
    list.value = data.items || []
    total.value = data.total_count || 0
    listError.value = ''
    clearSelection()
    return true
  } catch (error) {
    if (seq !== loadSeq) {
      return null
    }
    listError.value = error?.message || '加载视频列表失败'
    return false
  } finally {
    if (seq === loadSeq) {
      listLoading.value = false
    }
  }
}`
    const unsafeLoad = protectedLoad
      .replace(/\s*if \(seq !== loadSeq\) \{\s*return null\s*\}/g, '')
      .replace('if (seq === loadSeq) {\n      listLoading.value = false\n    }', 'listLoading.value = false')

    expect(latestRequestOwnsState(protectedLoad)).toBe(true)
    expect(latestRequestOwnsState(unsafeLoad)).toBe(false)
    expect(latestRequestOwnsState(script)).toBe(true)
    const requestIndex = loadBlock?.body.indexOf('const data = await getAdminVideos(query)') || -1
    expect(requestIndex).toBeGreaterThan(-1)
    expect(loadBlock?.body.slice(0, requestIndex)).not.toContain("listError.value = ''")
    expect(catchClearsRows(loadBlock?.body || '')).toBe(false)
  })

  it('保存视图刷新失败会拒绝且页面事件安全消费 rejection', () => {
    const refreshBlock = extractBalancedBraceBlock(script, /async function refreshSavedView\(\)\s*\{/)
    const selectBlock = extractBalancedBraceBlock(script, /async function selectVideoView\(id\)\s*\{/)
    const removeBlock = extractBalancedBraceBlock(script, /async function removeVideoView\(id\)\s*\{/)

    expect(refreshBlock).not.toBeNull()
    expect(refreshBlock?.body).toContain('const loaded = await load()')
    expect(refreshBlock?.body).toContain('if (loaded === false)')
    expect(refreshBlock?.body).toContain("throw new Error(listError.value || '加载视频列表失败')")
    expect(selectBlock).not.toBeNull()
    expect(selectBlock?.body).toContain('await selectView(id)')
    expect(selectBlock?.body).toMatch(/catch \(error\) \{[\s\S]*?listError\.value/)
    expect(removeBlock).not.toBeNull()
    expect(removeBlock?.body).toContain('await removeView(id)')
    expect(removeBlock?.body).toMatch(/catch \(error\) \{[\s\S]*?listError\.value/)
    expect(template).toContain('@select="selectVideoView"')
    expect(template).toContain('@remove="removeVideoView"')
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
    expect(template).toContain('@select="selectVideoView"')
    expect(template).toContain('@save="saveView"')
    expect(template).toContain('@update="updateView"')
    expect(template).toContain('@rename="renameView"')
    expect(template).toContain('@remove="removeVideoView"')
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

  it('未知行操作命令 fail-closed 且已知命令安全消费异步拒绝', () => {
    const protectedHandler = `function handleVideoRowAction(command, row) {
  if (command === 'retranscode') {
    doRetranscode(row).catch(() => {})
    return true
  }
  if (command === 'delete') {
    doDelete(row).catch(() => {})
    return true
  }
  return false
}`
    const unsafeHandler = `function handleVideoRowAction(command, row) {
  return command === 'retranscode' ? doRetranscode(row) : doDelete(row)
}`

    expect(handlesRowActionFailClosed(protectedHandler)).toBe(true)
    expect(handlesRowActionFailClosed(unsafeHandler)).toBe(false)
    expect(consumesRowActionRejections(protectedHandler)).toBe(true)
    expect(consumesRowActionRejections(unsafeHandler)).toBe(false)
    expect(handlesRowActionFailClosed(script)).toBe(true)
    expect(consumesRowActionRejections(script)).toBe(true)
    expect(operationsColumn).toContain('@command="(command) => handleVideoRowAction(command, row)"')
    expect(operationsColumn).not.toMatch(/command === 'retranscode' \? doRetranscode\(row\) : doDelete\(row\)/)
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
    expect(template).toContain('@current-change="setPage"')
  })
})
