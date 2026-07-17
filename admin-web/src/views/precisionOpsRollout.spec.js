import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import ActorManage from './ActorManage.vue'
import AVManualScrape from './AVManualScrape.vue'
import CollectionManage from './CollectionManage.vue'
import Dashboard from './Dashboard.vue'
import ImageCollectionManage from './ImageCollectionManage.vue'
import ImageManage from './ImageManage.vue'
import IPTVManage from './IPTVManage.vue'
import PendingDeleteShorts from './PendingDeleteShorts.vue'
import ScrapePreview from './ScrapePreview.vue'
import SystemSettings from './SystemSettings.vue'
import TaskMonitor from './TaskMonitor.vue'
import Toolbox from './Toolbox.vue'
import TvAppManage from './TvAppManage.vue'
import TvSeriesManage from './TvSeriesManage.vue'
import UserManage from './UserManage.vue'
import VideoList from './VideoList.vue'
import VideoUpload from './VideoUpload.vue'

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
  { file: 'UserManage.vue', component: 'UserManage', density: 'compact', compiled: UserManage },
  { file: 'IPTVManage.vue', component: 'IPTVManage', density: 'compact', compiled: IPTVManage },
  { file: 'TvAppManage.vue', component: 'TvAppManage', density: 'compact', compiled: TvAppManage },
  { file: 'VideoUpload.vue', component: 'VideoUpload', density: 'form', compiled: VideoUpload },
  { file: 'ScrapePreview.vue', component: 'ScrapePreview', density: 'form', compiled: ScrapePreview },
  { file: 'AVManualScrape.vue', component: 'AVManualScrape', density: 'form', compiled: AVManualScrape },
  { file: 'TvSeriesManage.vue', component: 'TvSeriesManage', density: 'form', compiled: TvSeriesManage },
  { file: 'SystemSettings.vue', component: 'SystemSettings', density: 'form', compiled: SystemSettings },
  { file: 'Toolbox.vue', component: 'Toolbox', density: 'form', compiled: Toolbox }
]
const phaseTwoFiles = [
  'PendingDeleteShorts.vue',
  'ActorManage.vue',
  'CollectionManage.vue',
  'ImageCollectionManage.vue',
  'UserManage.vue',
  'IPTVManage.vue',
  'TvAppManage.vue'
]
const pendingShellViews = []
const standaloneViews = {
  'ToolboxArchiveImport.vue': 'form',
  'ToolboxEd2k.vue': 'form',
  'ToolboxEd2kDownload.vue': 'form',
  'ToolboxOrphanFiles.vue': 'form',
  'ToolboxPasswordVault.vue': 'form',
  'Login.vue': 'form'
}

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

function rootMainStartTag(source) {
  const template = extractTemplate(source).trimStart()
  const match = template.match(/^<main\b[^>]*>/)

  expect(match).not.toBeNull()
  return match?.[0] || ''
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

function headerActionsBlock(template) {
  const match = template.match(/<template #header-actions>[\s\S]*?<\/template>/)

  expect(match).not.toBeNull()
  return match?.[0] || ''
}

function candidateOpeningTag(template) {
  const match = template.match(/<div\s+v-for="\(item, index\) in candidates"[\s\S]*?class="candidate-item"[\s\S]*?>/)

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
  it('固定 17 个已迁移页面与 0 个兼容页面，且集合互不重叠', () => {
    const migratedFiles = migratedViews.map(({ file }) => file)

    expect(migratedViews).toHaveLength(17)
    expect(pendingShellViews).toHaveLength(0)
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

  it('不存在仍依赖 PageHeader 兼容 meta 的待迁移 shell 页面', () => {
    expect(pendingShellViews).toEqual([])
  })

  it('电视剧管理把唯一主动作合并到页头并保持紧凑列表与原编辑命令', () => {
    const source = readView('TvSeriesManage.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const headerActions = headerActionsBlock(template)
    const toolbar = toolbarBlock(template)
    const models = [
      'detail.title',
      'detail.overview',
      'detail.poster_url',
      'detail.backdrop_url',
      'detail.first_air_date',
      'detail.active',
      'season.season_number',
      'season.title',
      'season.air_date',
      'season.overview',
      'season.poster_url',
      'episode.episode_number',
      'episode.title',
      'episode.overview',
      'episode.runtime',
      'episode.air_date',
      'episode.still_url',
      'episode.video_id'
    ]

    expect.soft(template).toContain('class="page-shell page-shell--medium tv-manage-shell" data-density="form"')
    expect.soft(template).not.toContain('<PageHeader')
    expect(headerActions).toContain(':icon="Plus" @click="openCreateSeries">新建系列</el-button>')
    expect(headerActions).toContain(':icon="Search" @click="query.page = 1; loadList()">筛选</el-button>')
    expect(template.match(/@click="openCreateSeries">新建系列<\/el-button>/g)).toHaveLength(1)
    expect(template.match(/>筛选<\/el-button>/g)).toHaveLength(1)
    expect(toolbar).toContain('v-model="query.q"')
    expect(toolbar).toContain('v-model="query.active"')
    expect(toolbar).toContain('v-model="query.has_playable"')
    expect(toolbar).toContain('>重置</el-button>')
    expect(toolbar).not.toContain('新建系列')
    expect(toolbar).not.toContain('>筛选</el-button>')
    expect(template).toMatch(/<SectionCard class="tv-series-list-card" data-density="compact">/)
    expect(template.match(/<SectionCard\b/g)).toHaveLength(2)
    expect(template).toContain('<section class="editor-section" aria-labelledby="series-basics-title">')
    expect(template).toContain('<section class="editor-section" aria-labelledby="season-episode-title">')
    expect(template).toMatch(/<fieldset\b[^>]*class="field-group season-field-group"/)
    expect(template).toMatch(/<fieldset\b[^>]*class="field-group episode-field-group"/)
    expect(template).toContain('<legend class="field-group__legend">')
    models.forEach((model) => expect(template, model).toContain(`v-model="${model}"`))
    for (const handler of ['saveSeries', 'removeSeries', 'addSeason', 'saveSeason(season)', 'addEpisode(season)', 'removeSeason(season)', 'saveEpisode(season, episode)', 'removeEpisode(season, episode)']) {
      expect(template, handler).toContain(`@click="${handler}"`)
    }
    expect(style).toMatch(/\.field-group\s*\{[^}]*border:\s*0;[^}]*box-shadow:\s*none;/s)
    expect(style).toMatch(/\.field-group__legend\s*\{[^}]*border-bottom:\s*1px solid var\(--line-soft\);/s)
    expect.soft(style).toMatch(/\.tv-manage-shell :deep\(\.el-switch\)\s*\{[^}]*min-height:\s*var\(--control-height\);/s)
    expect.soft(style).toMatch(/@media \(max-width: 63\.9375rem\)\s*\{[\s\S]*?\.tv-manage-shell :deep\(\.el-switch\)\s*\{[^}]*min-height:\s*44px;/s)
  })

  it('电视剧系列项使用原生按钮保留选择动作与准确选中语义', () => {
    const template = extractTemplate(readView('TvSeriesManage.vue'))
    const seriesItem = template.match(/<(?:button|div)\s+v-for="item in list"[\s\S]*?@click="selectSeries\(item\.id\)"[\s\S]*?>/)?.[0] || ''

    expect.soft(seriesItem).toMatch(/^<button\b/)
    expect.soft(seriesItem).toContain('type="button"')
    expect.soft(seriesItem).toContain(':aria-pressed="String(item.id) === selectedSeriesId"')
    expect.soft(seriesItem).toContain('@click="selectSeries(item.id)"')
  })

  it('电视剧筛选控件提供与用途对应的中文可访问名称', () => {
    const toolbar = toolbarBlock(extractTemplate(readView('TvSeriesManage.vue')))
    const titleFilter = toolbar.match(/<el-input\b[^>]*v-model="query\.q"[^>]*\/>/)?.[0] || ''
    const activeFilter = toolbar.match(/<el-select\b[^>]*v-model="query\.active"[^>]*>/)?.[0] || ''
    const playableFilter = toolbar.match(/<el-select\b[^>]*v-model="query\.has_playable"[^>]*>/)?.[0] || ''

    expect.soft(titleFilter).toContain('aria-label="系列标题筛选"')
    expect.soft(activeFilter).toContain('aria-label="启用状态筛选"')
    expect.soft(playableFilter).toContain('aria-label="可播分集筛选"')
  })

  it('电视剧系列按钮重置原生外观并提供可见键盘焦点', () => {
    const style = extractStyle(readView('TvSeriesManage.vue'))
    const buttonRule = style.match(/\.series-card\s*\{[^}]*\}/s)?.[0] || ''
    const focusRule = style.match(/\.series-card:focus-visible\s*\{[^}]*\}/s)?.[0] || ''

    expect.soft(buttonRule).toContain('width: 100%;')
    expect.soft(buttonRule).toContain('font: inherit;')
    expect.soft(buttonRule).toContain('text-align: left;')
    expect.soft(focusRule).toContain('outline: 2px solid var(--line-focus);')
    expect.soft(focusRule).toContain('outline-offset: -2px;')
  })

  it('电视剧系列标题与绑定视频文本在自身容器内安全换行', () => {
    const style = extractStyle(readView('TvSeriesManage.vue'))
    const titleRule = style.match(/\.series-card__title\s*\{[^}]*\}/s)?.[0] || ''
    const bindingRule = style.match(/\.episode-binding-tip\s*\{[^}]*\}/s)?.[0] || ''

    for (const rule of [titleRule, bindingRule]) {
      expect.soft(rule).toContain('min-width: 0;')
      expect.soft(rule).toContain('overflow-wrap: anywhere;')
    }
  })

  it('系统设置使用无阴影表单区块并在清理与日志区分别保留可恢复错误', () => {
    const source = readView('SystemSettings.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const loadLogs = functionBlock(source, 'async function loadLogs()')
    const runCleanup = functionBlock(source, 'async function runCleanup()')

    expect.soft(template).toContain('class="page-shell settings-page" data-density="form"')
    expect.soft(template).toContain('<p class="page-context-note">执行临时文件清理并查看最近系统日志。</p>')
    expect.soft(template).not.toContain('<PageHeader')
    expect(template.match(/<SectionCard\b/g)).toHaveLength(2)
    expect(source).toContain("const logsError = ref('')")
    expect(source).toContain("const cleanupError = ref('')")
    expect(loadLogs.indexOf("logsError.value = ''")).toBeLessThan(loadLogs.indexOf('try {'))
    expect(runCleanup.indexOf("cleanupError.value = ''")).toBeLessThan(runCleanup.indexOf('try {'))
    expect(loadLogs).toMatch(/catch \(error\) \{[\s\S]*logsError\.value = extractErrorMessage\(error,/)
    expect(runCleanup).toMatch(/catch \(error\) \{[\s\S]*cleanupError\.value = extractErrorMessage\(error,/)
    expect(template).toMatch(/<el-alert\s+v-if="cleanupError"[\s\S]*?:title="cleanupError"[\s\S]*?closable[\s\S]*?@close="cleanupError = ''"/)
    expect(template).toMatch(/<el-alert\s+v-if="logsError"[\s\S]*?:title="logsError"[\s\S]*?closable[\s\S]*?@close="logsError = ''"/)
    expect(template).toContain('v-if="!hasLogs && !logsError"')
    expect(template).toContain('<el-scrollbar v-else-if="hasLogs" max-height="60vh" class="log-box">')
    expect(style).toMatch(/\.settings-page :deep\(\.section-card\)\s*\{[^}]*border-radius:\s*var\(--radius-md\);[^}]*box-shadow:\s*none;/s)
    expect(style).toMatch(/\.log-box\s*\{[^}]*overflow:\s*hidden;/s)
    expect(style).toMatch(/\.log-text\s*\{[^}]*font-family:\s*var\(--font-mono\);/s)
  })

  it('工具箱移除外层装饰卡并保留六个 8px 独立工具入口', () => {
    const source = readView('Toolbox.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)

    expect.soft(template).toContain('class="page-shell toolbox-page" data-density="form"')
    expect.soft(template).toContain('<p class="page-context-note">工具会在独立标签页打开，并保持当前管理端上下文。</p>')
    expect.soft(template).not.toContain('<PageHeader')
    expect(template).not.toContain('<SectionCard')
    expect(template.match(/<a class="tool-menu-item"/g)).toHaveLength(6)
    expect(template.match(/target="_blank"/g)).toHaveLength(6)
    expect(template.match(/rel="noopener noreferrer"/g)).toHaveLength(6)
    expect(template.match(/<span>新标签页打开<\/span>/g)).toHaveLength(6)
    expect(style).toMatch(/\.tool-menu-item\s*\{[^}]*min-height:\s*44px;[^}]*border-radius:\s*8px;[^}]*box-shadow:\s*none;/s)
    expect(style).toMatch(/\.toolbox-page\s*\{[^}]*min-width:\s*0;[^}]*overflow-x:\s*clip;/s)
  })

  it('两页刮削工作台把唯一查询动作合并到壳层页头', () => {
    for (const file of ['ScrapePreview.vue', 'AVManualScrape.vue']) {
      const template = extractTemplate(readView(file))
      const headerActions = headerActionsBlock(template)
      const toolbar = toolbarBlock(template)

      expect.soft(template, file).toContain('data-density="form"')
      expect.soft(template, file).not.toContain('<PageHeader')
      expect(headerActions, file).toContain('type="primary"')
      expect(headerActions, file).toContain(':icon="Search"')
      expect(headerActions, file).toContain(':loading="previewLoading"')
      expect(headerActions, file).toContain('@click="doPreview">查询预览</el-button>')
      expect(template.match(/@click="doPreview">查询预览<\/el-button>/g), file).toHaveLength(1)
      expect(toolbar, file).not.toContain('查询预览')
    }
  })

  it('通用刮削保留完整筛选字段并按 5/2/1 栅格响应', () => {
    const source = readView('ScrapePreview.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)

    for (const model of ['video_id', 'title', 'year', 'type', 'season_number', 'episode_number']) {
      expect(template).toContain(`v-model="form.${model}"`)
    }
    expect(style).toMatch(/\.scrape-filter-form\s*\{[^}]*display:\s*grid;[^}]*grid-template-columns:\s*repeat\(5,\s*minmax\(0,\s*1fr\)\);/s)
    expect(style).toMatch(/@media \(max-width: 1024px\)[\s\S]*?\.scrape-filter-form\s*\{[^}]*grid-template-columns:\s*repeat\(2,\s*minmax\(0,\s*1fr\)\);/)
    expect(style).toMatch(/@media \(max-width: 768px\)[\s\S]*?\.scrape-filter-form\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\);/)
    expect(style).not.toContain('var(--text-on-inverse,')
  })

  it('两页候选继续使用原选择 handler 并支持键盘操作', () => {
    const cases = [
      ['ScrapePreview.vue', 'choose(item, index)'],
      ['AVManualScrape.vue', 'chooseCandidate(item, index)']
    ]

    cases.forEach(([file, handler]) => {
      const tag = candidateOpeningTag(extractTemplate(readView(file)))

      expect(tag, file).toContain('role="button"')
      expect(tag, file).toContain('tabindex="0"')
      expect(tag, file).toContain(':aria-pressed="index === selectedIndex"')
      expect(tag, file).toContain(`@click="${handler}"`)
      expect(tag, file).toContain(`@keydown.enter.prevent="${handler}"`)
      expect(tag, file).toContain(`@keydown.space.prevent="${handler}"`)
    })
  })

  it('AV 刮削保留站点工作流、双栏结果和非纯色候选选中态', () => {
    const source = readView('AVManualScrape.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)

    for (const model of ['video_id', 'title', 'site_category', 'site_source', 'bypass_cache']) {
      expect(template).toContain(`v-model="form.${model}"`)
    }
    expect(template).toContain('<template #title>AV 刮削配置</template>')
    expect(template).toContain('v-model="configForm.enabled_sites"')
    expect(template).toMatch(/<el-tag\s+v-if="index === selectedIndex"[\s\S]*?class="candidate-selected"[\s\S]*?role="status"[\s\S]*?>已选中<\/el-tag>/)
    expect(style).toMatch(/\.result-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\)\s+minmax\(0,\s*1fr\);/s)
    expect(style).toMatch(/\.candidate-item\.active\s*\{[^}]*border(?:-color)?:\s*var\(--primary\);/s)
    expect(style).toMatch(/\.candidate-selected\s*\{[^}]*font-weight:\s*600;/s)
    expect(style).toMatch(/@media \(max-width: 900px\)[\s\S]*?\.result-grid\s*\{[^}]*grid-template-columns:\s*minmax\(0,\s*1fr\);/)
    expect.soft(style).toMatch(/\.av-manual-scrape-page :deep\(\.el-switch\)\s*\{[^}]*min-height:\s*var\(--control-height\);/s)
    expect.soft(style).toMatch(/@media \(max-width: 63\.9375rem\)\s*\{[\s\S]*?\.av-manual-scrape-page :deep\(\.el-switch\)\s*\{[^}]*min-height:\s*44px;/s)
    expect(style).not.toContain('var(--text-on-inverse,')
  })

  it('上传流程保持在单个中密度工作区并保留原有五段顺序', () => {
    const source = readView('VideoUpload.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const stages = ['文件与基础信息', '关联信息', '上传控制', '进度区', '结果区']
    const stageIndexes = stages.map((stage) => template.indexOf(`<template #title>${stage}</template>`))

    expect.soft(template).toContain('data-density="form"')
    expect.soft(template).toContain('<p class="page-context-note">选择文件与媒体类型，补充关联信息后开始上传。</p>')
    expect.soft(template).not.toContain('<PageHeader')
    expect(template).toContain('uploadFileList')
    expect(template).toContain('uploadResults')
    expect(stageIndexes.every((index) => index >= 0)).toBe(true)
    expect(stageIndexes).toEqual([...stageIndexes].sort((left, right) => left - right))
    expect(template.match(/@click="submit"/g)).toHaveLength(1)
    expect(template).toContain('class="table-wrap upload-result"')
    expect(template).toMatch(/prop="name"[^>]*show-overflow-tooltip/)
    expect(template).toMatch(/prop="message"[^>]*show-overflow-tooltip/)
    expect(template).toMatch(/prop="videoId"[^>]*show-overflow-tooltip/)
    expect(style).toMatch(/\.upload-actions :deep\(\.el-button\)\s*\{[^}]*min-height:\s*44px/s)
    expect(style).toMatch(/\.upload-drop :deep\(\.el-upload-dragger\)\s*\{[^}]*min-height:\s*44px/s)
    expect(style).toMatch(/@media \(min-width: 64rem\)\s*\{\s*\.upload-page :deep\(\.el-radio-button__inner\)\s*\{[^}]*min-height:\s*var\(--control-height\)/s)
    expect(style).toMatch(/@media \(max-width: 768px\)[\s\S]*\.upload-page :deep\(\.el-form-item\)\s*\{[^}]*grid-template-columns:\s*minmax\(0, 1fr\)/)
  })

  it('服务资源页使用指标条替代重复统计卡', () => {
    for (const file of ['IPTVManage.vue', 'TvAppManage.vue']) {
      const source = readView(file)

      expect(source, file).toContain('<MetricStrip')
      expect(source, file).not.toContain('<StatCard')
    }
  })

  it('IPTV 区分读取失败、无缓存加载与成功空态并保留缓存频道', () => {
    const source = readView('IPTVManage.vue')
    const template = extractTemplate(source)
    const load = functionBlock(source, 'async function loadPlaylist()')
    const catchBlock = load.slice(load.indexOf('} catch (error) {'), load.indexOf('} finally {'))
    const alertIndex = template.indexOf('<el-alert v-if="loadError"')
    const sourceIndex = template.indexOf('播放列表来源')
    const skeletonIndex = template.indexOf('<el-skeleton v-if="initialLoading"')
    const contentIndex = template.indexOf('<template v-else-if="!loadError || hasChannels">')
    const dataContent = template.slice(contentIndex)

    expect(source).toContain("import { shouldShowCrudCollectionSkeleton } from './crudCollectionState'")
    expect(source).toContain("const loading = ref(true)")
    expect(source).toContain("const loadError = ref('')")
    expect(source).toMatch(
      /const initialLoading = computed\(\(\) => shouldShowCrudCollectionSkeleton\(\{\s*loading: loading\.value,\s*rowCount: channels\.value\.length\s*\}\)\)/
    )
    expect(load.indexOf("loadError.value = ''")).toBeGreaterThanOrEqual(0)
    expect(load.indexOf("loadError.value = ''")).toBeLessThan(load.indexOf('try {'))
    expect(catchBlock).toContain("loadError.value = extractErrorMessage(error, '加载 IPTV 状态失败')")
    expect(catchBlock).not.toContain('applyPlaylist(')
    expect(catchBlock).not.toContain('ElMessage.error')
    expect(alertIndex).toBeGreaterThanOrEqual(0)
    expect(sourceIndex).toBeGreaterThan(alertIndex)
    expect(sourceIndex).toBeLessThan(skeletonIndex)
    expect(skeletonIndex).toBeGreaterThan(alertIndex)
    expect(contentIndex).toBeGreaterThan(skeletonIndex)
    expect(dataContent).toContain('<MetricStrip :items="stats" aria-label="IPTV 摘要" />')
    expect(dataContent).toContain('<template #title>频道预览</template>')
    expect(dataContent).not.toContain('<template #title>播放列表来源</template>')
    expect(dataContent).toContain('v-if="!hasChannels"')
  })

  it('IPTV 保留来源与频道预览命令并使用紧凑无阴影区块', () => {
    const source = readView('IPTVManage.vue')
    const template = extractTemplate(source)
    const style = extractStyle(source)
    const sourceSection = template.match(/<section class="source-section" aria-labelledby="iptv-source-title">[\s\S]*?<\/section>/)?.[0] || ''
    const sourceSectionRule = style.match(/\.source-section\s*\{[^}]*\}/s)?.[0] || ''
    const sourceTitleRule = style.match(/\.source-section__title\s*\{[^}]*\}/s)?.[0] || ''
    const sourcePanelRule = style.match(/\.source-panel\s*\{[^}]*\}/s)?.[0] || ''

    expect(source).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(template).toContain('<template #header-actions>')
    expect(template).toContain('<Toolbar dense>')
    expect(template).toContain(':label="`最后更新时间：${updatedAtText}`"')
    expect(template).toContain('@click="refreshPlaylist">远程拉取</el-button>')
    expect(template).toContain('@click="uploadPlaylist">上传 M3U</el-button>')
    expect(template).toContain('@click="saveSourceUrl">保存 URL</el-button>')
    expect(template).toContain('label="播放地址" min-width="260" show-overflow-tooltip')
    expect(template).not.toContain('<template #title>播放列表来源</template>')
    expect(sourceSection).toContain('<h2 id="iptv-source-title" class="source-section__title">播放列表来源</h2>')
    expect(sourceSection.match(/<article class="source-panel">/g)).toHaveLength(2)
    expect(sourceSectionRule).toContain('width: 100%;')
    expect(sourceSectionRule).not.toMatch(/(?:^|[;{]\s*)(?:border|background|box-shadow)\s*:/)
    expect(sourceTitleRule).toContain('font-size: var(--text-h2);')
    expect(sourceTitleRule).toContain('line-height: var(--leading-h2);')
    expect(sourcePanelRule).toContain('border-radius: var(--radius-md);')
    expect(sourcePanelRule).not.toContain('box-shadow:')
    expect(style).toMatch(/\.channel-table\s+:deep\(\.el-table__row\)\s*\{[^}]*height:\s*var\(--table-row-height\);/s)
  })

  it('服务资源页样式只使用既有语义颜色且没有装饰性表面效果', () => {
    for (const file of ['IPTVManage.vue', 'TvAppManage.vue']) {
      const style = extractStyle(readView(file))

      expect(style, file).not.toMatch(/#[0-9a-fA-F]{3,8}\b/)
      expect(style, file).not.toMatch(/rgba?\(\s*\d/)
      expect(style, file).not.toMatch(/(?:linear|radial)-gradient\(/)
      expect(style, file).not.toMatch(/border-radius:\s*(?:14|16|18)px/)
    }
  })

  it('媒体复核操作保持显式并可由键盘触达', () => {
    const pending = readView('PendingDeleteShorts.vue')
    const pendingTemplate = extractTemplate(pending)
    const queueItem = pendingTemplate.match(/<button\s+v-for="\(item, index\) in items"[\s\S]*?<\/button>/)?.[0] || ''
    const collections = readView('ImageCollectionManage.vue')

    expect(pending).toContain('aria-label="待删除短视频队列"')
    expect(pending).toContain('刷新列表')
    expect(queueItem).toMatch(
      /:aria-current="index === currentIndex \? 'true' : undefined"[\s\S]*?<span class="pending-delete-item__thumb">[\s\S]*?<CircleCheck v-if="index === currentIndex" \/>[\s\S]*?<VideoCamera v-else \/>[\s\S]*?<\/span>/
    )
    expect(pending).not.toMatch(/\.pending-delete-queue,\s*\.pending-delete-player-panel\s*\{[^}]*box-shadow:/s)
    expect(pending).not.toMatch(/\.pending-delete-video-frame\s*\{[^}]*box-shadow:/s)
    expect(collections).toContain('创建合集')
    expect(collections).toContain('<el-drawer')
    expect(collections).not.toMatch(/border-radius:\s*(?:14|16|18)px/)
    expect(collections).not.toMatch(/(?:linear|radial)-gradient\(/)
  })

  it('待删除队列项遵循 compact 媒体行高', () => {
    const style = extractStyle(readView('PendingDeleteShorts.vue'))
    const itemRule = style.match(/\.pending-delete-item\s*\{[^}]*\}/s)?.[0] || ''

    expect(itemRule).toContain('height: var(--media-row-height);')
    expect(itemRule).not.toContain('height: 64px;')
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
    expect(collections).not.toContain("const loaded = ref(false)")
    expect(collections).toContain("const loadError = ref('')")
    expect(collections).toMatch(
      /const initialLoading = computed\(\(\) => shouldShowCrudCollectionSkeleton\(\{\s*loading: loading\.value,\s*rowCount: list\.value\.length\s*\}\)\)/
    )
    expect(collectionLoad.indexOf("loadError.value = ''")).toBeLessThan(collectionLoad.indexOf('try {'))
    expect(collectionCatch).toContain("loadError.value = extractErrorMessage(error, '加载图片合集列表失败')")
    expect(collectionCatch).not.toContain('list.value =')
    expect(collectionCatch).not.toContain('total.value =')
    expect(collectionLoad).not.toContain('loaded.value = true')
    expect(collectionLoad).toContain('loading.value = false')
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
    expect(template).toContain('size="min(100vw, 920px)"')
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

describe('Precision Ops 独立工具工作区', () => {
  it('只读取 template 顶层 main 的开始标签', () => {
    const fixture = `
      <template>
        <main class="fixture-root">
          <el-dialog data-density="form" />
        </main>
      </template>
    `

    expect(rootMainStartTag(fixture)).not.toContain('data-density')
  })

  Object.entries(standaloneViews).forEach(([file, density]) => {
    it(`${file} 保持独立标题工作区`, () => {
      const source = readView(file)
      const template = extractTemplate(source)
      const rootMain = rootMainStartTag(source)

      expect(rootMain).toContain(`data-density="${density}"`)
      expect(template).toContain('<PageHeader')
      expect(template).not.toContain('<Layout')
    })
  })

  it('压缩包导入保持业务边界并应用紧凑集合密度', () => {
    const source = readView('ToolboxArchiveImport.vue')
    const template = extractTemplate(source)

    expect(source).toContain('class="archive-batch-panel" data-density="compact"')
    expect(source).toContain('class="archive-file-panel" data-density="compact"')
    expect(source).toContain('<MetricStrip')
    expect(source).toContain('<BulkActionBar')
    expect(template).toContain('<PageHeader')
    expect(template).not.toContain('<Layout')
    expect(source).not.toMatch(/(?:linear|radial)-gradient\(/)
    expect(source).not.toContain('archive-overview-card')
  })

  it('移除登录页装饰渐变与面板常驻阴影', () => {
    const source = readView('Login.vue')

    expect(source).not.toMatch(/(?:linear|radial)-gradient\(/)
    expect(source).not.toMatch(/\.login-card\s*\{[^}]*box-shadow:/s)
  })

  it('密码库列表使用紧凑密度且两个对话框保持表单密度', () => {
    const template = extractTemplate(readView('ToolboxPasswordVault.vue'))
    const dialogs = template.match(/<el-dialog\b[\s\S]*?>/g) || []

    expect(template).toMatch(/<SectionCard\s+data-density="compact">/)
    expect(dialogs).toHaveLength(2)
    dialogs.forEach((dialog) => expect(dialog).toContain('data-density="form"'))
  })

  it('ED2K 下载任务保持独立表单工作台与紧凑全宽列表', () => {
    const source = readView('ToolboxEd2kDownload.vue')
    const template = extractTemplate(source)
    const rootMain = rootMainStartTag(source)
    const style = extractStyle(source)
    const listIndex = template.indexOf('<SectionCard class="task-list-card" data-density="compact">')
    const detailIndex = template.indexOf('<SectionCard v-if="selectedTask">')
    const taskWorkspaceRule = style.match(/\.task-workspace\s*\{[^}]*\}/s)?.[0] || ''

    expect.soft(rootMain).toContain('class="tool-workspace"')
    expect.soft(rootMain).toContain('data-density="form"')
    expect.soft(template).toContain('<PageHeader')
    expect.soft(template).not.toContain('<Layout')
    expect(source).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(template).toContain('<StatusIndicator')
    expect(listIndex).toBeGreaterThanOrEqual(0)
    expect(detailIndex).toBeGreaterThan(listIndex)
    expect(template.match(/<SectionCard\b/g)).toHaveLength(3)
    expect(taskWorkspaceRule).toContain('display: grid;')
    expect(taskWorkspaceRule).toContain('grid-template-columns: minmax(0, 1fr);')
    expect(taskWorkspaceRule).toContain('gap: var(--space-4);')
  })

  it('独立工具无可见标签的主要输入提供中文可访问名称', () => {
    const ed2k = readView('ToolboxEd2k.vue')
    const download = readView('ToolboxEd2kDownload.vue')
    const passwordVault = readView('ToolboxPasswordVault.vue')

    expect.soft(ed2k).toContain('aria-label="ED2K 链接文本"')
    expect.soft(download).toContain('aria-label="ED2K 下载链接"')
    expect.soft(passwordVault).toContain('aria-label="密码库搜索"')
    expect.soft(passwordVault).toContain('aria-label="密码内容"')
  })

  it('独立工作区限制页面横向溢出并收纳窄屏操作', () => {
    const workspaceFiles = ['ToolboxEd2k.vue', 'ToolboxEd2kDownload.vue', 'ToolboxOrphanFiles.vue', 'ToolboxPasswordVault.vue']

    workspaceFiles.forEach((file) => {
      const style = extractStyle(readView(file))
      const rootRule = style.match(/\.tool-workspace\s*\{[^}]*\}/s)?.[0] || ''

      expect.soft(rootRule, file).toContain('min-width: 0;')
      expect.soft(rootRule, file).toContain('overflow-x: clip;')
    })

    const loginStyle = extractStyle(readView('Login.vue'))
    const loginRootRule = loginStyle.match(/\.login-page\s*\{[^}]*\}/s)?.[0] || ''
    const downloadStyle = extractStyle(readView('ToolboxEd2kDownload.vue'))
    const orphanStyle = extractStyle(readView('ToolboxOrphanFiles.vue'))
    const passwordStyle = extractStyle(readView('ToolboxPasswordVault.vue'))

    expect.soft(loginRootRule).toContain('min-width: 0;')
    expect.soft(loginRootRule).toContain('overflow-x: clip;')
    expect.soft(downloadStyle).toMatch(
      /@media \(max-width: 63\.9375rem\)[\s\S]*?\.tool-workspace :deep\(\.el-button\),\s*\.crud-dialog :deep\(\.el-button\)\s*\{[^}]*min-height:\s*44px;/s
    )
    expect.soft(downloadStyle).toMatch(
      /@media \(max-width: 63\.9375rem\)[\s\S]*?\.tool-workspace :deep\(\.page-header-shell__actions\),\s*\.task-list-card :deep\(\.section-card__actions\)\s*\{[^}]*flex-wrap:\s*wrap;/s
    )
    expect.soft(orphanStyle).toMatch(
      /@media \(max-width: 63\.9375rem\)[\s\S]*?\.orphan-tool :deep\(\.section-card__actions\)\s*\{[^}]*width:\s*100%;[^}]*flex-wrap:\s*wrap;/s
    )
    expect.soft(passwordStyle).toMatch(
      /@media \(max-width: 63\.9375rem\)[\s\S]*?\.password-vault-tool__url\s*\{[^}]*min-height:\s*44px;/s
    )
  })

  it('独立工作区原生链接提供可见键盘焦点', () => {
    const ed2kStyle = extractStyle(readView('ToolboxEd2k.vue'))
    const passwordStyle = extractStyle(readView('ToolboxPasswordVault.vue'))

    expect.soft(ed2kStyle).toMatch(/\.ed2k-link:focus-visible\s*\{[^}]*outline:\s*2px solid var\(--line-focus\);/s)
    expect.soft(passwordStyle).toMatch(/\.password-vault-tool__url:focus-visible\s*\{[^}]*outline:\s*2px solid var\(--line-focus\);/s)
  })
})

describe('Precision Ops 第二阶段集成门禁', () => {
  it('锁定阶段二 7 个资源页完整清单并使用紧凑密度', () => {
    expect(phaseTwoFiles).toHaveLength(7)
    expect(new Set(phaseTwoFiles).size).toBe(phaseTwoFiles.length)
    expect(
      phaseTwoFiles.every((file) =>
        migratedViews.some(({ file: migratedFile, density }) => migratedFile === file && density === 'compact')
      )
    ).toBe(true)
  })

  it('阶段二筛选与行内编辑控件提供可访问名称', () => {
    const actor = readView('ActorManage.vue')
    const collection = readView('CollectionManage.vue')
    const imageCollection = readView('ImageCollectionManage.vue')
    const user = readView('UserManage.vue')
    const tvApp = readView('TvAppManage.vue')

    expect(actor).toContain('aria-label="演员姓名筛选"')
    expect(actor).toContain('aria-label="演员状态筛选"')
    expect(collection).toContain('aria-label="合集名称筛选"')
    expect(collection).toContain('aria-label="合集状态筛选"')
    expect(imageCollection).toContain('aria-label="图片合集名称筛选"')
    expect(imageCollection).toContain('aria-label="图片合集状态筛选"')
    expect(user).toContain(':aria-label="`调整用户 ${row.username || row.id} 的角色`"')
    expect(tvApp).toContain('aria-label="安装包版本筛选"')
    expect(tvApp).toContain('aria-label="安装包状态筛选"')
    expect(tvApp).toContain('aria-label="ABI 完整性筛选"')
    expect(tvApp).toContain('aria-label="家庭可见筛选"')
    expect(tvApp).toContain(':aria-label="`${row.version_name} 版本说明`"')
    expect(tvApp).toContain(':aria-label="`${row.version_name} 管理端备注`"')
  })

  it('阶段二动态图片提供替代文本与稳定尺寸', () => {
    const iptv = readView('IPTVManage.vue')
    const iptvStyle = extractStyle(iptv)
    const tvApp = readView('TvAppManage.vue')

    expect(iptv).toContain(':alt="`${row.name || \'未命名频道\'}台标`"')
    expect(iptvStyle).toMatch(/\.logo-image\s*\{[^}]*width:\s*44px;[^}]*height:\s*28px;/s)
    expect(tvApp).toContain(':alt="downloadQRCodeTitle" width="220" height="220"')
  })

  it('阶段二资源名称与待删除详情标题保持紧凑收敛', () => {
    const pendingStyle = extractStyle(readView('PendingDeleteShorts.vue'))

    expect(readView('ActorManage.vue')).toContain('prop="name" label="演员姓名" min-width="160" show-overflow-tooltip')
    expect(readView('CollectionManage.vue')).toContain('prop="name" label="合集名称" min-width="180" show-overflow-tooltip')
    expect(readView('ImageCollectionManage.vue')).toContain('prop="name" label="图片合集名称" min-width="180" show-overflow-tooltip')
    expect(pendingStyle).toMatch(
      /\.pending-delete-detail__copy h2\s*\{[^}]*display:\s*-webkit-box;[^}]*overflow:\s*hidden;[^}]*-webkit-box-orient:\s*vertical;[^}]*-webkit-line-clamp:\s*2;/s
    )
  })

  it('用户身份字段使用表格溢出提示保持 compact 行高', () => {
    const user = readView('UserManage.vue')

    expect(user).toContain('prop="username" label="用户名" min-width="160" show-overflow-tooltip')
    expect(user).toContain('prop="email" label="邮箱" min-width="220" show-overflow-tooltip')
  })
})
