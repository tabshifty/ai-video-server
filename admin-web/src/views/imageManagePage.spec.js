import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import ImageManage from './ImageManage.vue'

const source = readFileSync(new URL('./ImageManage.vue', import.meta.url), 'utf8')

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

function tokensAppearInOrder(sourceText, tokens) {
  const indexes = tokens.map((token) => sourceText.indexOf(token))
  return indexes.every((index) => index >= 0)
    && indexes.every((index, position) => position === 0 || indexes[position - 1] < index)
}

function ownsSnapshotApplyOrder(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /function applyImageViewSnapshot\(snapshot\)\s*\{/)
  if (!block) return false

  return tokensAppearInOrder(block.body, [
    'const next = normalizeImageViewSnapshot(snapshot)',
    'query.q = next.q',
    'query.status = next.status',
    'query.active = next.active',
    'query.actor_id = next.actor_id',
    'query.collection_id = next.collection_id',
    'query.page = 1',
    'setViewMode(next.viewMode)',
    'quickSearch.value = query.q',
    'syncFilterDraftFromQuery()',
    'resetQueryIdentity()'
  ])
}

function resetsBeforeLoad(sourceText, openingPattern) {
  const block = extractBalancedBraceBlock(sourceText, openingPattern)
  if (!block) return false

  const resetIndex = block.body.indexOf('resetQueryIdentity()')
  const loadIndex = block.body.indexOf('load()')
  return resetIndex >= 0 && loadIndex > resetIndex
}

function resetOwnsDisplayIdentity(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /function resetQueryIdentity\(\)\s*\{/)
  if (!block) return false

  return tokensAppearInOrder(block.body, [
    'list.value = []',
    'total.value = 0',
    'clearImageSelection()',
    "listError.value = ''"
  ])
}

function viewModePreservesSelection(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /function setViewMode\(mode\)\s*\{/)
  if (!block) return false

  return block.body.includes("viewMode.value = mode === 'list' ? 'list' : 'grid'")
    && block.body.includes('window.localStorage.setItem(IMAGEMANAGE_VIEW_KEY, viewMode.value)')
    && !block.body.includes('clearImageSelection()')
}

function latestRequestOwnsState(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /async function load\(\)\s*\{/)
  if (!block) return false
  const body = block.body
  const requestIndex = body.indexOf('const data = await getAdminImages(buildListParams())')
  const catchMatch = body.match(/catch\s*\(error\)\s*\{/)
  if (!body.includes('const seq = ++loadSeq') || requestIndex < 0 || !catchMatch || catchMatch.index === undefined) return false

  const successSource = body.slice(requestIndex, catchMatch.index)
  const successGuard = extractBalancedBraceBlock(successSource, /if\s*\(seq !== loadSeq\)\s*\{/)
  if (!successGuard || successGuard.body.trim() !== 'return null') return false
  const successTail = successSource.slice(successGuard.closeBraceIndex + 1)
  if (![
    'list.value = data.items || []',
    'total.value = data.total_count || 0',
    "listError.value = ''",
    'clearImageSelection()',
    'return true'
  ].every((token) => successTail.includes(token))) return false

  const catchBlock = extractBalancedBraceBlock(body, /catch\s*\(error\)\s*\{/)
  const catchGuard = catchBlock
    ? extractBalancedBraceBlock(catchBlock.body, /if\s*\(seq !== loadSeq\)\s*\{/)
    : null
  if (!catchBlock || !catchGuard || catchGuard.body.trim() !== 'return null') return false
  const catchTail = catchBlock.body.slice(catchGuard.closeBraceIndex + 1)
  if (!catchTail.includes("listError.value = extractErrorMessage(error, '加载图片列表失败')")) return false
  if (!catchTail.includes('return false') || /list\.value\s*=|total\.value\s*=|clearImageSelection\(\)/.test(catchTail)) return false

  const finallyBlock = extractBalancedBraceBlock(body, /finally\s*\{/)
  const ownerGuard = finallyBlock
    ? extractBalancedBraceBlock(finallyBlock.body, /if\s*\(seq === loadSeq\)\s*\{/)
    : null
  if (!finallyBlock || !ownerGuard || !ownerGuard.body.includes('loading.value = false')) return false
  const outsideOwner = finallyBlock.body.slice(0, ownerGuard.openingIndex)
    + finallyBlock.body.slice(ownerGuard.closeBraceIndex + 1)
  return outsideOwner.trim() === '' && (body.match(/loading\.value\s*=\s*false/g) || []).length === 1
}

function handlesImageActionFailClosed(sourceText) {
  const block = extractBalancedBraceBlock(sourceText, /function handleImageRowAction\(command, row\)\s*\{/)
  if (!block) return false

  const toggleBlock = extractBalancedBraceBlock(block.body, /if\s*\(command === 'toggle'\)\s*\{/)
  const deleteBlock = extractBalancedBraceBlock(block.body, /if\s*\(command === 'delete'\)\s*\{/)
  if (!toggleBlock || !deleteBlock) return false
  if (!toggleBlock.body.includes('toggleActive(row).catch(') || !toggleBlock.body.includes('return true')) return false
  if (!deleteBlock.body.includes('doDelete(row).catch(') || !deleteBlock.body.includes('return true')) return false

  const tail = block.body.slice(deleteBlock.closeBraceIndex + 1)
  return tail.includes('return false')
    && (block.body.match(/toggleActive\(row\)/g) || []).length === 1
    && (block.body.match(/doDelete\(row\)/g) || []).length === 1
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
const listPreviewBlock = extractBalancedBraceBlock(script, /async function loadListPreviews\(items, listRequestSeq\)\s*\{/)
const clearListPreviewsBlock = extractBalancedBraceBlock(script, /function clearListPreviews\(\)\s*\{/)
const deleteBlock = extractBalancedBraceBlock(script, /async function doDelete\(row\)\s*\{/)
const filterDrawer = extractElement(template, /<el-drawer\b(?=[^>]*v-model="filterDrawerVisible")[^>]*>/, '</el-drawer>')
const paginationTag = template.match(/<AdminTablePagination\b[\s\S]*?\/>/)?.[0] || ''
const gridCard = extractElement(template, /<article\b(?=[^>]*class="image-grid-card")[^>]*>/, '</article>')
const mainTable = extractElement(template, /<el-table\b(?=[^>]*row-key="id")[^>]*>/, '</el-table>')
const operationsColumn = extractElement(template, /<el-table-column\b(?=[^>]*label="操作")[^>]*>/, '</el-table-column>')
const headerActions = extractElement(template, /<template #header-actions>/, '</template>')
const detailDrawer = extractElement(template, /<el-drawer\b(?=[^>]*v-model="detailVisible")[^>]*>/, '</el-drawer>')

describe('图片资产集合页', () => {
  it('通过真实 SFC 编译并只接入共享保存视图职责', () => {
    expect(ImageManage).toBeTruthy()
    expect(script).toContain("import SavedViewTabs from '../components/base/SavedViewTabs.vue'")
    expect(script).toContain("import { useSavedViews } from '../components/base/useSavedViews'")
    expect(script).toContain("const SAVED_VIEWS_KEY = 'admin-imagemanage-saved-views-v1'")
    expect(savedViewsOptions).not.toBeNull()
    expect(savedViewsOptions?.body).toContain('storageKey: SAVED_VIEWS_KEY')
    expect(savedViewsOptions?.body).toContain('builtInViews')
    expect(savedViewsOptions?.body).toContain('normalizeSnapshot: normalizeImageViewSnapshot')
    expect(savedViewsOptions?.body).toContain('applySnapshot: applyImageViewSnapshot')
    expect(savedViewsOptions?.body).toContain('refresh: refreshSavedView')
    expect(script).not.toMatch(/function\s+(?:persistUserViews|readUserViews)\b/)
    expect(script).not.toContain('selectedViewID')
    expect(script).not.toContain("from '../components/base/savedView.helpers'")
    expect(script).not.toContain("ElMessageBox.prompt('请输入视图名称'")
  })

  it('快照只包含六个业务字段并按约定顺序应用新查询身份', () => {
    const validFixture = `function applyImageViewSnapshot(snapshot) {
  const next = normalizeImageViewSnapshot(snapshot)
  query.q = next.q
  query.status = next.status
  query.active = next.active
  query.actor_id = next.actor_id
  query.collection_id = next.collection_id
  query.page = 1
  setViewMode(next.viewMode)
  quickSearch.value = query.q
  syncFilterDraftFromQuery()
  resetQueryIdentity()
}`
    const unsafeFixture = validFixture.replace('setViewMode(next.viewMode)\n  quickSearch.value', 'quickSearch.value')

    expect(ownsSnapshotApplyOrder(validFixture)).toBe(true)
    expect(ownsSnapshotApplyOrder(unsafeFixture)).toBe(false)
    expect(ownsSnapshotApplyOrder(script)).toBe(true)
    expect(currentSnapshotBlock).not.toBeNull()
    expect(currentSnapshotBlock?.body.match(/^\s*(q|status|active|actor_id|collection_id|viewMode):/gm)?.map((line) => line.trim().split(':')[0])).toEqual([
      'q',
      'status',
      'active',
      'actor_id',
      'collection_id',
      'viewMode'
    ])
    expect(currentSnapshotBlock?.body).not.toMatch(/\bpage\b|selectedImageRows|detailVisible|uploadDialogVisible/)
    expect(script).toContain('const builtInViews = createImageBuiltInViews(query.active, viewMode.value)')
    expect(script).toContain('active: DEFAULT_IMAGE_ACTIVE')
    expect(script).toContain("window.localStorage.setItem(IMAGEMANAGE_VIEW_KEY, viewMode.value)")
  })

  it('所有服务端查询入口先重置展示身份再通过唯一入口加载', () => {
    const validFixture = `function setPage(page) {
  query.page = page
  resetQueryIdentity()
  load()
}`
    const unsafeFixture = validFixture.replace('resetQueryIdentity()\n  load()', 'load()\n  resetQueryIdentity()')

    expect(resetsBeforeLoad(validFixture, /function setPage\(page\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(unsafeFixture, /function setPage\(page\)\s*\{/)).toBe(false)
    expect(resetOwnsDisplayIdentity(script)).toBe(true)
    expect(ownsSnapshotApplyOrder(script)).toBe(true)
    expect(resetsBeforeLoad(script, /function applyQuickSearch\(\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function applyFilterDrawer\(\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function removeFilter\(key\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function resetFilters\(\)\s*\{/)).toBe(true)
    expect(resetsBeforeLoad(script, /function setPage\(page\)\s*\{/)).toBe(true)
    expect(script).toContain('const quickSearch = ref(query.q)')
    expect(script).toContain('const filterDraft = reactive({')
    expect(template).toContain('v-model="quickSearch"')
    expect(template).toContain('@keyup.enter="applyQuickSearch"')
    expect(template).toContain('@clear="applyQuickSearch"')
    expect(template).toContain('@click="openFilterDrawer"')
    for (const key of ['q', 'status', 'active', 'actor_id', 'collection_id']) {
      expect(filterDrawer).toContain(`v-model="filterDraft.${key}"`)
    }
    expect(filterDrawer).toContain('@closed="syncFilterDraftFromQuery"')
    expect(filterDrawer).toContain('@click="applyFilterDrawer"')
    expect(filterDrawer).not.toMatch(/v-model="query\.(?:q|status|active|actor_id|collection_id)"/)
    expect(paginationTag).toContain(':current-page="query.page"')
    expect(paginationTag).toContain(':page-size="query.page_size"')
    expect(paginationTag).toContain('@current-change="setPage"')
    expect(paginationTag).not.toContain('v-model:current-page')
    expect(paginationTag).not.toContain('v-model:page-size')
  })

  it('默认 active 保持结果集基线且筛选移除和重置不会扩大结果集', () => {
    const removeBlock = extractBalancedBraceBlock(script, /function removeFilter\(key\)\s*\{/)
    const resetBlock = extractBalancedBraceBlock(script, /function resetFilters\(\)\s*\{/)

    expect(script).toContain('DEFAULT_IMAGE_ACTIVE,')
    expect(script).toContain("from './imageManage.helpers'")
    expect(removeBlock?.body).toContain('query.active = DEFAULT_IMAGE_ACTIVE')
    expect(resetBlock?.body).toContain('query.active = DEFAULT_IMAGE_ACTIVE')
    expect(script).toContain('const hasActiveFilters = computed(() => hasImageActiveFilters(query))')
  })

  it('只有最新请求可以写列表、错误、选择和加载完成状态', () => {
    const protectedLoad = `let loadSeq = 0
async function load() {
  const seq = ++loadSeq
  loading.value = true
  try {
    const data = await getAdminImages(buildListParams())
    if (seq !== loadSeq) {
      return null
    }
    list.value = data.items || []
    total.value = data.total_count || 0
    listError.value = ''
    clearImageSelection()
    return true
  } catch (error) {
    if (seq !== loadSeq) {
      return null
    }
    listError.value = extractErrorMessage(error, '加载图片列表失败')
    return false
  } finally {
    if (seq === loadSeq) {
      loading.value = false
    }
  }
}`
    const unsafeLoad = protectedLoad
      .replace(/\s*if \(seq !== loadSeq\) \{\s*return null\s*\}/g, '')
      .replace('if (seq === loadSeq) {\n      loading.value = false\n    }', 'loading.value = false')

    expect(latestRequestOwnsState(protectedLoad)).toBe(true)
    expect(latestRequestOwnsState(unsafeLoad)).toBe(false)
    expect(latestRequestOwnsState(script)).toBe(true)
    const requestIndex = loadBlock?.body.indexOf('const data = await getAdminImages(buildListParams())') || -1
    expect(requestIndex).toBeGreaterThan(-1)
    expect(loadBlock?.body.slice(0, requestIndex)).not.toContain("listError.value = ''")
    expect(catchClearsRows(loadBlock?.body || '')).toBe(false)
  })

  it('保存视图刷新失败继续拒绝并由页面事件边界安全消费', () => {
    const refreshBlock = extractBalancedBraceBlock(script, /async function refreshSavedView\(\)\s*\{/)
    const selectBlock = extractBalancedBraceBlock(script, /async function selectImageView\(id\)\s*\{/)
    const removeBlock = extractBalancedBraceBlock(script, /async function removeImageView\(id\)\s*\{/)

    expect(refreshBlock).not.toBeNull()
    expect(refreshBlock?.body).toContain('const loaded = await load()')
    expect(refreshBlock?.body).toContain('if (loaded === false)')
    expect(refreshBlock?.body).toContain("throw new Error(listError.value || '加载图片列表失败')")
    expect(selectBlock?.body).toContain('await selectView(id)')
    expect(selectBlock?.body).toMatch(/catch \(error\) \{[\s\S]*?listError\.value/)
    expect(removeBlock?.body).toContain('await removeView(id)')
    expect(removeBlock?.body).toMatch(/catch \(error\) \{[\s\S]*?listError\.value/)
    expect(template).toContain('@select="selectImageView"')
    expect(template).toContain('@remove="removeImageView"')
  })

  it('按固定顺序使用壳层操作区、保存视图、指标和紧凑工作区', () => {
    expect(script).not.toContain('PageHeader')
    expect(template).not.toContain('<PageHeader')
    expect(headerActions).toContain('@click="openUploadDialog"')
    expect(headerActions).toContain('上传图片')
    expect(template).toContain('class="page-shell image-page" data-density="compact"')
    expect(template).toContain('<SavedViewTabs')
    expect(template).toContain('<MetricStrip :items="summaryMetrics" aria-label="图片摘要" />')
    expect(template).toContain('<Toolbar dense>')
    expect(tokensAppearInOrder(template, [
      '<SavedViewTabs',
      '<MetricStrip',
      '<Toolbar dense>',
      '<el-alert v-if="listError"',
      '<el-skeleton v-if="loading && list.length === 0"',
      '<SectionCard v-else-if="!listError || list.length > 0" dense>',
      '<AdminTablePagination',
      '<BulkActionBar'
    ])).toBe(true)
    expect(script).toContain("{ key: 'total', label: '结果总数', value: total.value, scope: '当前条件·全部页' }")
    expect(script).toContain("{ key: 'ready', label: '可用', value: readyCount.value, scope: '本页', tone: 'success' }")
    expect(script).toContain("{ key: 'failed', label: '失败', value: failedCount.value, scope: '本页', tone: 'danger' }")
    expect(script).toContain("{ key: 'inactive', label: '停用', value: inactiveCount.value, scope: '本页', tone: 'warning' }")
  })

  it('同查询刷新保留旧资产并以持久错误、首次骨架和诚实空态反馈', () => {
    const violatingLoad = `async function load() {
  try {
    list.value = await getAdminImages()
  } catch (error) {
    listError.value = error.message
    list.value = []
  }
}`

    expect(catchClearsRows(violatingLoad)).toBe(true)
    expect(catchClearsRows(loadBlock?.body || '')).toBe(false)
    expect(loadBlock?.body).toContain("listError.value = ''")
    expect(loadBlock?.body).toContain("listError.value = extractErrorMessage(error, '加载图片列表失败')")
    expect(template).toContain('<el-alert v-if="listError"')
    expect(template).toContain('@click="load">重试</el-button>')
    expect(template).toContain('<el-skeleton v-if="loading && list.length === 0"')
    expect(template).toContain('<SectionCard v-else-if="!listError || list.length > 0" dense>')
    expect(template).toContain("hasActiveFilters ? '当前筛选无结果' : '暂无图片'")
    expect(template).toContain("hasActiveFilters ? '清除筛选后查看全部图片' : '上传图片后会显示在这里'")
    expect(template).toContain('v-if="hasActiveFilters" @click="resetFilters"')
  })

  it('列表成功后非阻塞加载缺失的认证 blob 预览并保留单项失败占位', () => {
    expect(script).toContain('const listPreviewUrls = ref({})')
    expect(script).toContain('const listPreviewErrors = ref({})')
    expect(script).toContain('const IMAGE_LIST_PREVIEW_PARAMS = Object.freeze({')
    expect(listPreviewBlock).not.toBeNull()
    expect(listPreviewBlock?.body).toContain('await Promise.all(')
    expect(listPreviewBlock?.body).toContain('if (!item?.id || resolveImagePreviewUrl(item)) return')
    expect(listPreviewBlock?.body).toContain('getAdminImageViewBlob(item.id, IMAGE_LIST_PREVIEW_PARAMS)')
    expect(listPreviewBlock?.body).toContain("readImagePreviewBlob(blob, '加载图片缩略图失败')")
    expect(listPreviewBlock?.body).toContain('const ownedUrl = URL.createObjectURL(imageBlob)')
    expect(listPreviewBlock?.body).toContain("nextErrors[item.id] = extractErrorMessage(error, '加载图片缩略图失败')")

    expect(tokensAppearInOrder(loadBlock?.body || '', [
      'list.value = data.items || []',
      'void loadListPreviews(list.value, seq)',
      'return true'
    ])).toBe(true)
    expect(loadBlock?.body).not.toContain('await loadListPreviews(list.value, seq)')

    expect(gridCard).toContain('v-if="imagePreviewUrl(item)"')
    expect(gridCard).toContain(':src="imagePreviewUrl(item)"')
    expect(gridCard).toContain('v-else-if="listPreviewErrors[item.id]"')
    expect(gridCard).toContain('预览加载失败')
    expect(findRule(style, '.image-grid-card__preview img')).toContain('object-fit: contain')
  })

  it('图片 blob 预览只接受最新列表并在替换、stale、查询重置和卸载时回收', () => {
    const resetBlock = extractBalancedBraceBlock(script, /function resetQueryIdentity\(\)\s*\{/)
    const unmountBlock = extractBalancedBraceBlock(script, /onBeforeUnmount\(\(\)\s*=>\s*\{/)

    expect(script).toContain('let listPreviewSeq = 0')
    expect(clearListPreviewsBlock).not.toBeNull()
    expect(tokensAppearInOrder(clearListPreviewsBlock?.body || '', [
      'listPreviewSeq += 1',
      'listPreviewUrls.value = revokeImagePreviewUrls(listPreviewUrls.value)',
      'listPreviewErrors.value = {}'
    ])).toBe(true)
    expect(listPreviewBlock?.body).toContain('const previewRequestSeq = ++listPreviewSeq')
    expect(listPreviewBlock?.body).toContain('listPreviewUrls.value = revokeImagePreviewUrls(listPreviewUrls.value)')
    expect(listPreviewBlock?.body).toContain('if (listRequestSeq !== loadSeq || previewRequestSeq !== listPreviewSeq)')
    expect(listPreviewBlock?.body).toContain('URL.revokeObjectURL(ownedUrl)')
    expect((listPreviewBlock?.body || '').lastIndexOf('if (listRequestSeq !== loadSeq || previewRequestSeq !== listPreviewSeq)'))
      .toBeLessThan((listPreviewBlock?.body || '').lastIndexOf('listPreviewUrls.value = nextUrls'))
    expect((listPreviewBlock?.body || '').lastIndexOf('listPreviewUrls.value = nextUrls'))
      .toBeLessThan((listPreviewBlock?.body || '').lastIndexOf('listPreviewErrors.value = nextErrors'))
    expect(resetBlock?.body).toContain('clearListPreviews()')
    expect(unmountBlock?.body).toContain('clearListPreviews()')
  })

  it('每张认证缩略图完成后立即发布并在发布前复核双代次', () => {
    const body = listPreviewBlock?.body || ''
    const ownedIndex = body.indexOf('const ownedUrl = URL.createObjectURL(imageBlob)')
    const staleIndex = body.indexOf('if (listRequestSeq !== loadSeq || previewRequestSeq !== listPreviewSeq)', ownedIndex)
    const revokeIndex = body.indexOf('URL.revokeObjectURL(ownedUrl)', staleIndex)
    const publishIndex = body.indexOf('nextUrls[item.id] = ownedUrl', revokeIndex)
    const reactivePublishIndex = body.indexOf(
      'listPreviewUrls.value = { ...listPreviewUrls.value, [item.id]: ownedUrl }',
      publishIndex
    )

    expect(ownedIndex).toBeGreaterThan(-1)
    expect(staleIndex).toBeGreaterThan(ownedIndex)
    expect(revokeIndex).toBeGreaterThan(staleIndex)
    expect(publishIndex).toBeGreaterThan(revokeIndex)
    expect(reactivePublishIndex).toBeGreaterThan(publishIndex)
    expect(listPreviewBlock?.body).not.toContain('nextUrls[item.id] = URL.createObjectURL(imageBlob)')
  })

  it('网格和列表动作共用显式白名单并安全消费异步拒绝', () => {
    const protectedHandler = `function handleImageRowAction(command, row) {
  if (command === 'toggle') {
    toggleActive(row).catch(() => {})
    return true
  }
  if (command === 'delete') {
    doDelete(row).catch(() => {})
    return true
  }
  return false
}`
    const unsafeHandler = `function handleImageRowAction(command, row) {
  return command === 'toggle' ? toggleActive(row) : doDelete(row)
}`

    expect(handlesImageActionFailClosed(protectedHandler)).toBe(true)
    expect(handlesImageActionFailClosed(unsafeHandler)).toBe(false)
    expect(handlesImageActionFailClosed(script)).toBe(true)
    expect(gridCard).toContain('@command="(command) => handleImageRowAction(command, item)"')
    expect(operationsColumn).toContain('@command="(command) => handleImageRowAction(command, row)"')
    for (const element of [gridCard, operationsColumn]) {
      expect(element).toContain('command="toggle"')
      expect(element).toContain('command="delete"')
      expect(element).toContain('popper-class="image-row-actions-popper"')
      expect(element).toContain('content="图片操作"')
      expect(element).toContain('aria-label="图片操作"')
      expect(element).toMatch(/<el-tooltip\b[^>]*content="图片操作"[^>]*>\s*<el-dropdown\b[\s\S]*?>\s*<el-button\b/)
      expect(element).not.toMatch(/<el-dropdown\b[^>]*>\s*<el-tooltip\b/)
    }
    expect(deleteBlock?.body.indexOf('ElMessageBox.confirm')).toBeGreaterThan(-1)
    expect(deleteBlock?.body.indexOf('deleteAdminImage(row.id)')).toBeGreaterThan(deleteBlock?.body.indexOf('ElMessageBox.confirm') || -1)
  })

  it('高密度网格完整展示图片并让选择和操作始终可访问', () => {
    expect(gridCard).toContain(':aria-label="`选择图片：${item.title || item.id}`"')
    expect(gridCard).toContain(':alt="item.title || \'图片预览\'"')
    expect(gridCard).toContain('<StatusIndicator')
    expect(operationsColumn).toContain('width="108"')
    expect(operationsColumn).toContain('fixed="right"')
    expect(template).toContain('prop="title" label="标题" min-width="200" show-overflow-tooltip')
    expect(template).toContain('class="table-wrap has-media-rows"')
    expect(findRule(style, '.image-grid')).toContain('grid-template-columns: repeat(auto-fill, minmax(184px, 1fr))')
    expect(findRule(style, '.image-grid-card')).toContain('gap: var(--space-1)')
    expect(findRule(style, '.image-grid-card__preview')).toContain('aspect-ratio: 16 / 9')
    expect(findRule(style, '.image-grid-card__preview img')).toContain('object-fit: contain')
    expect(findRule(style, '.image-grid-card__actions')).not.toMatch(/opacity:\s*0/)
  })

  it('手动切换网格与列表模式时保留当前选择', () => {
    const validFixture = `function setViewMode(mode) {
  viewMode.value = mode === 'list' ? 'list' : 'grid'
  window.localStorage.setItem(IMAGEMANAGE_VIEW_KEY, viewMode.value)
}`
    const unsafeFixture = validFixture.replace(
      'window.localStorage.setItem',
      'clearImageSelection()\n  window.localStorage.setItem'
    )

    expect(viewModePreservesSelection(validFixture)).toBe(true)
    expect(viewModePreservesSelection(unsafeFixture)).toBe(false)
    expect(viewModePreservesSelection(script)).toBe(true)
  })

  it('列表选择提供逐行中文名称和本页全选完整状态', () => {
    const togglePageBlock = extractBalancedBraceBlock(script, /function toggleCurrentPageSelection\(checked\)\s*\{/)
    const toggleRowBlock = extractBalancedBraceBlock(script, /function toggleGridSelection\(row, checked\)\s*\{/)

    expect(mainTable).not.toContain('type="selection"')
    expect(mainTable).toContain('aria-label="选择本页全部图片"')
    expect(mainTable).toContain(':aria-label="`选择图片：${row.title || row.id}`"')
    expect(mainTable).toContain(':model-value="isGridSelected(row)"')
    expect(mainTable).toContain('@update:model-value="(checked) => toggleGridSelection(row, checked)"')
    expect(script).toContain('const allPageSelected = computed(() => list.value.length > 0 && list.value.every((item) => isGridSelected(item)))')
    expect(script).toContain('const somePageSelected = computed(() => !allPageSelected.value && list.value.some((item) => isGridSelected(item)))')
    expect(tokensAppearInOrder(togglePageBlock?.body || '', [
      'const currentPageIDs = new Set(list.value.map((item) => item.id))',
      'const next = selectedImageRows.value.filter((item) => !currentPageIDs.has(item.id))',
      'if (checked)',
      'next.push(...list.value)',
      'onImageSelectionChange(next)'
    ])).toBe(true)
    expect(toggleRowBlock?.body).toContain('onImageSelectionChange(next)')
    expect(template).toContain('<BulkActionBar :count="selectedImageRows.length" :actions="bulkActions" />')
  })

  it('详情预览使用标题或 ID 形成动态中文替代文本', () => {
    expect(detailDrawer).toContain(':alt="`图片预览：${detail.title || detail.id}`"')
  })

  it('棋盘背景只用于图片预览且样式保持语义化和克制', () => {
    const previewRule = findRule(style, '.image-grid-card__preview')

    expect(previewRule.match(/linear-gradient/g)).toHaveLength(4)
    expect(style.match(/linear-gradient/g)).toHaveLength(4)
    expect(style).not.toMatch(/#[0-9a-f]{3,8}\b/i)
    expect(style).not.toMatch(/rgba?\(\s*\d/i)
    expect(style).not.toContain('box-shadow')
    expect(style).not.toContain('letter-spacing')
    expect(style).not.toContain('var(--radius-lg)')
  })

  it('窄屏壳层、分段控件、菜单按钮和 Teleport 目标至少 44px', () => {
    const mediaStart = style.indexOf('@media (max-width: 63.9375rem)')

    expect(mediaStart).toBeGreaterThan(-1)
    const mobileStyle = style.slice(mediaStart)
    expect(findRule(mobileStyle, '.image-header-actions :deep(.el-button)')).toContain('min-height: 44px')
    expect(findRule(mobileStyle, '.image-view-switch :deep(.el-segmented__item)')).toContain('min-height: 44px')
    expect(findRule(mobileStyle, '.image-row-actions :deep(.el-button)')).toContain('min-height: 44px')
    expect(findRule(mobileStyle, ':global(.image-row-actions-popper .el-dropdown-menu__item)')).toContain('min-height: 44px')
    expect(style).toContain('.image-row-actions :deep(.el-button:focus-visible)')
    expect(style).toContain('outline: 2px solid var(--line-focus)')
  })

  it('保留上传、选择、批量、详情守卫、路由详情、预览和保存 payload', () => {
    expect(script).toContain('const uploadFileList = ref([])')
    expect(script).toContain('const selectedImageRows = ref([])')
    expect(script).toContain('function toggleGridSelection(row, checked)')
    expect(script).toContain('function onImageSelectionChange(rows)')
    expect(script).toContain('sha256File(file)')
    expect(script).toContain('checkAdminImageUpload({ hash, file_size: file.size })')
    expect(script).toContain('uploadAdminImages(buildFormData())')
    expect(script).toContain("const drawerSize = computed(() => (viewportWidth.value < 1024 ? '100%' : '560px'))")
    expect(script).toContain('const uploadDrawerDirty = computed(')
    expect(script).toContain('const detailDrawerDirty = computed(')
    expect(script).toContain("const imageID = String(route.query.image_id || '').trim()")
    expect(script).toContain('await openDetailFromRouteQuery()')
    expect(script).toContain("fit: 'inside'")
    expect(script).toContain('zoom: 100')
    expect(script).toContain('actor_ids: actorIDs')
    expect(script).toContain('collection_ids: normalizeCollectionSelection(detail.value.collection_ids)')
    expect(template).toContain(':before-close="handleUploadDrawerBeforeClose"')
    expect(template).toContain(':before-close="handleDetailDrawerBeforeClose"')
    expect(template).toContain('<BulkActionBar :count="selectedImageRows.length" :actions="bulkActions" />')
  })
})
