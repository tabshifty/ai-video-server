import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import ActorManage from './ActorManage.vue'
import CollectionManage from './CollectionManage.vue'
import Dashboard from './Dashboard.vue'
import ImageCollectionManage from './ImageCollectionManage.vue'
import ImageManage from './ImageManage.vue'
import PendingDeleteShorts from './PendingDeleteShorts.vue'
import TaskMonitor from './TaskMonitor.vue'
import UserManage from './UserManage.vue'
import VideoList from './VideoList.vue'

const readView = (file) => readFileSync(new URL(`./${file}`, import.meta.url), 'utf8')
const router = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
const migratedViews = [
  { file: 'Dashboard.vue', component: 'Dashboard', density: 'compact', compiled: Dashboard },
  { file: 'TaskMonitor.vue', component: 'TaskMonitor', density: 'monitor', compiled: TaskMonitor },
  { file: 'VideoList.vue', component: 'VideoList', density: 'compact', compiled: VideoList },
  { file: 'ImageManage.vue', component: 'ImageManage', density: 'compact', compiled: ImageManage },
  { file: 'PendingDeleteShorts.vue', component: 'PendingDeleteShorts', density: 'compact', compiled: PendingDeleteShorts },
  { file: 'ImageCollectionManage.vue', component: 'ImageCollectionManage', density: 'compact', compiled: ImageCollectionManage },
  { file: 'ActorManage.vue', component: 'ActorManage', density: 'compact', compiled: ActorManage },
  { file: 'CollectionManage.vue', component: 'CollectionManage', density: 'compact', compiled: CollectionManage },
  { file: 'UserManage.vue', component: 'UserManage', density: 'compact', compiled: UserManage }
]
const pendingShellViews = [
  'AVManualScrape.vue',
  'IPTVManage.vue',
  'ScrapePreview.vue',
  'SystemSettings.vue',
  'Toolbox.vue',
  'TvAppManage.vue',
  'TvSeriesManage.vue',
  'VideoUpload.vue'
]

const crudViews = [
  {
    file: 'ActorManage.vue',
    error: '加载演员列表失败',
    createHandler: 'openCreate',
    saveHandler: 'save',
    titleAttribute: `:title="editingID ? '编辑演员' : '创建演员'"`
  },
  {
    file: 'CollectionManage.vue',
    error: '加载合集列表失败',
    createHandler: 'openCreate',
    saveHandler: 'save',
    titleAttribute: `:title="editingID ? '编辑合集' : '新增合集'"`
  },
  {
    file: 'UserManage.vue',
    error: '加载用户列表失败',
    createHandler: 'openCreateDialog',
    saveHandler: 'saveUser',
    titleAttribute: 'title="添加用户"'
  }
]

function extractTemplate(source) {
  const opening = source.match(/<template[^>]*>/)

  expect(opening).not.toBeNull()
  const start = (opening?.index || 0) + (opening?.[0].length || 0)
  const end = source.lastIndexOf('</template>')

  expect(end).toBeGreaterThan(start)
  return source.slice(start, end)
}

function extractStyle(source) {
  const match = source.match(/<style scoped>([\s\S]*?)<\/style>/)

  expect(match).not.toBeNull()
  return match?.[1] || ''
}

function routeLine(component) {
  const componentPattern = new RegExp(`\\bcomponent:\\s*${component}(?=\\s*[,}])`)
  return router.split('\n').find((line) => componentPattern.test(line)) || ''
}

function functionBlock(source, signature) {
  const start = source.indexOf(signature)
  const end = source.indexOf('\n}', start)

  expect(start, signature).toBeGreaterThanOrEqual(0)
  expect(end, signature).toBeGreaterThan(start)
  return source.slice(start, end + 2)
}

function toolbarBlock(template) {
  const match = template.match(/<Toolbar\b[\s\S]*?<\/Toolbar>/)

  expect(match).not.toBeNull()
  return match?.[0] || ''
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
  it('固定九个已迁移页面与 8 个兼容页面，且集合互不重叠', () => {
    const migratedFiles = migratedViews.map(({ file }) => file)

    expect(migratedViews).toHaveLength(9)
    expect(pendingShellViews).toHaveLength(8)
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

  it('8 个待迁移 shell 页面保持 PageHeader 与精确兼容 meta', () => {
    pendingShellViews.forEach((file) => {
      const component = file.replace('.vue', '')
      const template = extractTemplate(readView(file))
      const line = routeLine(component)

      expect(template, file).toContain('<PageHeader')
      expect(line, component).toMatch(exactRoutePattern(component, true))
    })
  })

  it('媒体复核操作保持显式并可由键盘触达', () => {
    const pending = readView('PendingDeleteShorts.vue')
    const collections = readView('ImageCollectionManage.vue')

    expect(pending).toContain('aria-label="待删除短视频队列"')
    expect(pending).toContain('刷新列表')
    expect(pending).not.toMatch(/\.pending-delete-queue,\s*\.pending-delete-player-panel\s*\{[^}]*box-shadow:/s)
    expect(pending).not.toMatch(/\.pending-delete-video-frame\s*\{[^}]*box-shadow:/s)
    expect(collections).toContain('创建合集')
    expect(collections).toContain('<el-drawer')
    expect(collections).not.toMatch(/border-radius:\s*(?:14|16|18)px/)
    expect(collections).not.toMatch(/(?:linear|radial)-gradient\(/)
  })

  it('媒体集合区分读取失败、无缓存加载与真正空态', () => {
    const pending = readView('PendingDeleteShorts.vue')
    const pendingTemplate = extractTemplate(pending)
    const pendingLoad = functionBlock(pending, 'async function loadPage(')
    const pendingCatch = pendingLoad.slice(pendingLoad.indexOf('} catch (error) {'), pendingLoad.indexOf('} finally {'))
    const collections = readView('ImageCollectionManage.vue')
    const collectionTemplate = extractTemplate(collections)
    const collectionLoad = functionBlock(collections, 'async function load()')
    const collectionCatch = collectionLoad.slice(collectionLoad.indexOf('} catch (error) {'), collectionLoad.indexOf('} finally {'))

    expect(pending).toContain('const listLoading = ref(true)')
    expect(pending).toContain("const listError = ref('')")
    expect(pending).toContain('const initialLoading = computed(() => listLoading.value && items.value.length === 0)')
    expect(pendingLoad.indexOf("listError.value = ''")).toBeLessThan(pendingLoad.indexOf('try {'))
    expect(pendingCatch).toContain("listError.value = error?.response?.data?.msg || error?.message || '待删除短视频加载失败'")
    expect(pendingCatch).not.toContain('items.value =')
    expect(pendingTemplate.indexOf('<el-alert v-if="listError"')).toBeLessThan(pendingTemplate.indexOf('<el-skeleton v-if="initialLoading"'))
    expect(pendingTemplate).toMatch(/<section\s+v-else-if="!listError \|\| hasItems"[\s\S]*?aria-label="待删除短视频队列"/)

    expect(collections).toContain("import { shouldShowCrudCollectionSkeleton } from './crudCollectionState'")
    expect(collections).toContain("const loaded = ref(false)")
    expect(collections).toContain("const loadError = ref('')")
    expect(collections).toMatch(
      /const initialLoading = computed\(\(\) => shouldShowCrudCollectionSkeleton\(\{\s*loading: loading\.value,\s*rowCount: list\.value\.length\s*\}\)\)/
    )
    expect(collectionLoad.indexOf("loadError.value = ''")).toBeLessThan(collectionLoad.indexOf('try {'))
    expect(collectionCatch).toContain("loadError.value = extractErrorMessage(error, '加载图片合集列表失败')")
    expect(collectionCatch).not.toContain('list.value =')
    expect(collectionCatch).not.toContain('total.value =')
    expect(collectionLoad).toMatch(/finally \{[\s\S]*loaded\.value = true[\s\S]*loading\.value = false/)
    expect(collections).toContain("const hasFilters = computed(() => String(query.q || '').trim() !== '' || String(query.active || '') !== '')")
    expect(collectionTemplate.indexOf('<el-alert v-if="loadError"')).toBeLessThan(collectionTemplate.indexOf('<el-skeleton v-if="initialLoading"'))
    expect(collectionTemplate).toContain('<SectionCard v-else-if="!loadError || list.length > 0" dense>')
    expect(collectionTemplate).toContain(":title=\"hasFilters ? '当前筛选无结果' : '暂无图片合集'\"")
    expect(collectionTemplate).toContain('<el-button v-if="hasFilters" @click="resetFilters">重置筛选</el-button>')
  })

  it('图片合集保留预览请求契约并使用紧凑媒体网格', () => {
    const source = readView('ImageCollectionManage.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const helpers = readFileSync(new URL('./imageCollectionManage.helpers.js', import.meta.url), 'utf8')

    expect(helpers).toMatch(/IMAGE_COLLECTION_PREVIEW_PARAMS = Object\.freeze\(\{\s*w: 240,\s*h: 240,\s*fit: 'cover',\s*q: 82\s*\}\)/)
    expect(source.match(/IMAGE_COLLECTION_PREVIEW_PARAMS/g)).toHaveLength(3)
    expect(style).toContain('grid-template-columns: repeat(auto-fill, minmax(184px, 1fr));')
    expect(style).toContain('gap: 12px;')
    expect(style).toMatch(/\.thumb-media\s*\{[^}]*aspect-ratio:\s*4 \/ 3;/s)
    expect(style).toMatch(/\.thumb-image\s*\{[^}]*object-fit:\s*contain;/s)
    expect(style).toMatch(/\.image-collection-table\s+:deep\(\.el-table__row\)\s*\{[^}]*height:\s*40px;/s)
    expect(source).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(template).toContain('<StatusIndicator')
    expect(style).not.toMatch(/#[0-9a-fA-F]{3,8}\b/)
    expect(style).not.toMatch(/rgba?\(\s*\d/)
    expect(style).not.toMatch(/(?:linear|radial)-gradient\(/)
  })

  it('图片合集两个 Drawer 复用共享关闭入口且编辑关闭保留脏数据保护', () => {
    const source = readView('ImageCollectionManage.vue')
    const template = extractTemplate(source)

    expect(source).toContain("import AdminDrawerHeader from '../components/base/AdminDrawerHeader.vue'")
    expect(template.match(/<AdminDrawerHeader\b/g)).toHaveLength(2)
    expect(template.match(/<template #header="\{ close, titleId, titleClass \}">/g)).toHaveLength(2)
    expect(template.match(/:close="close"/g)).toHaveLength(2)
    expect(template).toContain(':before-close="handleEditDrawerBeforeClose"')
    expect(template).toContain('@click="requestEditDrawerClose">取消</el-button>')
    expect(template).not.toContain('@click="editDrawerVisible = false">取消</el-button>')
  })

  it('基础 CRUD 列表区分读取失败、首次加载、筛选零结果与真正空态', () => {
    crudViews.forEach(({ file, error, createHandler }) => {
      const source = readView(file)
      const template = extractTemplate(source)
      const load = functionBlock(source, 'async function load()')
      const catchBlock = load.slice(load.indexOf('} catch (error) {'), load.indexOf('} finally {'))
      const alertIndex = template.indexOf('<el-alert v-if="loadError"')
      const skeletonIndex = template.indexOf('<el-skeleton v-if="initialLoading"')
      const sectionIndex = template.indexOf('<SectionCard v-else-if="!loadError || list.length > 0" dense>')

      expect(source, file).toContain("const loaded = ref(false)")
      expect(source, file).toContain("const loadError = ref('')")
      expect(source, file).toContain("import { shouldShowCrudCollectionSkeleton } from './crudCollectionState'")
      expect(source, file).toMatch(
        /const initialLoading = computed\(\(\) => shouldShowCrudCollectionSkeleton\(\{\s*loading: loading\.value,\s*rowCount: list\.value\.length\s*\}\)\)/
      )
      expect(source, file).not.toContain('computed(() => loading.value && !loaded.value)')
      expect(load.indexOf("loadError.value = ''"), file).toBeGreaterThanOrEqual(0)
      expect(load.indexOf("loadError.value = ''"), file).toBeLessThan(load.indexOf('try {'))
      expect(load, file).toContain(`loadError.value = extractErrorMessage(error, '${error}')`)
      expect(catchBlock, file).not.toContain('list.value =')
      expect(catchBlock, file).not.toContain('total.value =')
      expect(load, file).toMatch(/finally \{[\s\S]*loaded\.value = true[\s\S]*loading\.value = false/)
      expect(alertIndex, file).toBeGreaterThanOrEqual(0)
      expect(skeletonIndex, file).toBeGreaterThan(alertIndex)
      expect(sectionIndex, file).toBeGreaterThan(skeletonIndex)

      if (file === 'UserManage.vue') {
        expect(source).toContain('const hasFilters = computed(() => false)')
        expect(template).toContain('title="暂无用户"')
        expect(template).toContain(`@click="${createHandler}"`)
        return
      }

      expect(source).toContain("const hasFilters = computed(() => String(query.q || '').trim() !== '' || String(query.active || '') !== '')")
      expect(template).toContain(`:title="hasFilters ? '当前筛选无结果' : '${file === 'ActorManage.vue' ? '暂无演员' : '暂无合集'}'"`)
      expect(template).toContain('<el-button v-if="hasFilters" @click="resetFilters">重置筛选</el-button>')
      expect(template).toContain(`<el-button v-else type="primary" @click="${createHandler}">`)
    })
  })

  it('基础 CRUD 页头承载刷新和创建动作，工具条只保留查询上下文', () => {
    crudViews.forEach(({ file, createHandler }) => {
      const template = extractTemplate(readView(file))
      const toolbar = toolbarBlock(template)

      expect(template, file).toContain('<template #header-actions>')
      expect(template, file).toContain('@click="load">刷新</el-button>')
      expect(template, file).toContain(`@click="${createHandler}"`)
      expect(toolbar, file).toContain('共 {{ total }}')
      expect(toolbar, file).not.toContain(`@click="${createHandler}"`)
      expect(toolbar, file).not.toContain('刷新</el-button>')
      expect(toolbar, file).not.toContain('重新加载</el-button>')

      if (file === 'UserManage.vue') {
        expect(toolbar).not.toContain('<el-input')
        expect(toolbar).not.toContain('<el-select')
        expect(readView(file)).toContain('const query = reactive({ page: 1, page_size: 20 })')
      } else {
        expect(toolbar).toContain('<el-input')
        expect(toolbar).toContain('<el-select')
        expect(toolbar).toContain('@click="load">查询</el-button>')
        expect(toolbar).toContain('@click="resetFilters">重置</el-button>')
      }
    })
  })

  it('演员和合集使用状态指示器，用户保留可编辑角色选择器', () => {
    for (const file of ['ActorManage.vue', 'CollectionManage.vue']) {
      const source = readView(file)

      expect(source, file).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
      expect(extractTemplate(source), file).toContain('<StatusIndicator')
    }

    const userSource = readView('UserManage.vue')
    const userTemplate = extractTemplate(userSource)
    expect(userSource).not.toContain('StatusIndicator')
    expect(userTemplate).toContain(':model-value="row.role"')
    expect(userTemplate).toContain('@change="(value) => onRoleChange(row, value)"')
  })

  it('基础 CRUD 编辑器使用共享关闭入口的上下文 Drawer', () => {
    crudViews.forEach(({ file, saveHandler, titleAttribute }) => {
      const source = readView(file)
      const template = extractTemplate(source)
      const drawer = template.match(/<el-drawer\b[\s\S]*?>/)?.[0] || ''
      const header = template.match(/<AdminDrawerHeader\b[\s\S]*?\/>/)?.[0] || ''

      expect(source, file).toContain("import AdminDrawerHeader from '../components/base/AdminDrawerHeader.vue'")
      expect(template, file).not.toContain('<el-dialog')
      expect(drawer, file).toContain('v-model="dialogVisible"')
      expect(drawer, file).toContain('class="crud-drawer"')
      expect(drawer, file).toContain(titleAttribute)
      expect(drawer, file).toContain('direction="rtl"')
      expect(drawer, file).toContain('size="min(100vw, 560px)"')
      expect(drawer, file).toContain('destroy-on-close')
      expect(drawer, file).toContain(':show-close="false"')
      expect(template, file).toContain('<template #header="{ close, titleId, titleClass }">')
      expect(header, file).toContain(titleAttribute)
      expect(header, file).toContain(':title-id="titleId"')
      expect(header, file).toContain(':title-class="titleClass"')
      expect(header, file).toContain(':close="close"')
      expect(template, file).toContain(`:loading="saving" @click="${saveHandler}"`)
    })
  })
})
