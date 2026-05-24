# Implement：admin 重设计 Phase 4 · 三巨头 IA 改造

- 日期：2026-05-24
- 关联 PRD：`prd.md`
- 关联 Phase 1-3：`tasks/2026-05-23-admin-shell-redesign/DONE.md` + `tasks/2026-05-23-admin-simple-views/DONE.md` + `tasks/2026-05-23-admin-medium-views/DONE.md`
- 任务三段执行流：当前阶段 = Implement

## 1. 实施顺序（TDD 友好）

1. **扩展 spec 先**：
   - `assets/themeTokens.spec.js` 视图层 audit 数组从 12 文件扩到 15（追加 VideoList / ImageManage / VideoUpload + 后续 step 子文件按需）；预期 3 个新用例红 → 实施后绿
   - 新增 `views/videoUpload.wizard.helpers.spec.js`（先红）：step 切换、字段保留、类型驱动条件字段
2. **`ImageManage.vue` 第一**（覆盖最多新模式）：BulkActionBar + chip + 视图切换 + 2 个 drawer + hex 清零
3. **`VideoList.vue`** 第二（复用 + 增量列设置）：chip + 列设置 + drawer + BulkActionBar
4. **`VideoUpload.vue` + step 子组件** 第三（Wizard 拆分）：拆 `StepFile.vue` / `StepBasic.vue` / `StepRelate.vue` + `videoUpload.wizard.helpers.js`
5. **`router/index.js`** 给 3 路由设 `meta.hideShellPageHeader = true`
6. **`CONTEXT.md`** 新增 5 条术语
7. **手测 H41–H48** + 截图归档
8. **`npm run build` + `npm test`** 全绿
9. **追加 `plan.md` + commit**

## 2. ImageManage 改造细节（核心：覆盖最多新模式）

### 2.1 视图切换（网格 / 列表）+ localStorage

```vue
<script setup>
import { computed, ref } from 'vue'

const IMAGEMANAGE_VIEW_KEY = 'admin-imagemanage-view'

function readStoredView() {
  if (typeof window === 'undefined') return 'grid'
  try {
    const value = window.localStorage.getItem(IMAGEMANAGE_VIEW_KEY)
    return value === 'list' ? 'list' : 'grid'
  } catch (_) {
    return 'grid'
  }
}

const viewMode = ref(readStoredView())

function setViewMode(mode) {
  viewMode.value = mode === 'list' ? 'list' : 'grid'
  if (typeof window === 'undefined') return
  try {
    window.localStorage.setItem(IMAGEMANAGE_VIEW_KEY, viewMode.value)
  } catch (_) {
    // 隐私模式 / quota exceeded：不持久化但本会话仍生效
  }
}
</script>

<template>
  <Toolbar>
    <template #filters>
      <!-- chip 筛选 -->
    </template>
    <template #actions>
      <el-radio-group :model-value="viewMode" @update:model-value="setViewMode" size="default">
        <el-radio-button value="grid">网格</el-radio-button>
        <el-radio-button value="list">列表</el-radio-button>
      </el-radio-group>
      <el-button type="primary" @click="openUploadDrawer">上传图片</el-button>
    </template>
  </Toolbar>

  <SectionCard v-loading="loading">
    <template v-if="list.length">
      <ImageManageGrid v-if="viewMode === 'grid'" :items="list" v-model:selection="selection" @edit="openEdit" @toggle-active="toggleActive" @delete="doDelete" />
      <el-table v-else :data="list" border>...</el-table>
    </template>
    <EmptyState v-else-if="!loading" title="暂无图片" />
  </SectionCard>
</template>
```

### 2.2 BulkActionBar 接入（首次）

```vue
<BulkActionBar
  :count="selection.length"
  :actions="bulkActions"
  @action="onBulkAction"
/>
```

`BulkActionBar` 接受 `count` (number) + `actions` (array of `{label, icon?, type?}`) + `@action(actionKey)`，本期实际接入：
- 批量删除（type=danger）
- 批量改状态（启用 / 停用）
- 批量打标（弹 popover 选演员 / 图片合集）

### 2.3 上传 / 编辑 drawer 改造（复用 Phase 3 contract）

| 契约 | 落地 |
|------|------|
| cancel + save 都 captureFormSnapshot | `captureUploadSnapshot()` / `captureEditSnapshot()` 各自实现 |
| v-loading 挂外层 | SectionCard 外层挂 `v-loading="loading"` |
| dirty 比对含 id 字段 | snapshot 含 `cover_image_id` / `actor_ids` / `collection_ids` |
| dense 嵌套 SectionCard | drawer 内 SectionCard 全部传 `dense` |
| 宽 560px + 全屏 < 1024 | `:size="viewportWidth < 1024 ? '100%' : '560px'"` |

### 2.4 hex 清零

| 位置 | 改 |
|------|---|
| `:919 #881337` | `var(--primary-strong)` |
| `:937 #7f1d1d` | `var(--primary)` |
| `:986 #7f1d1d` | `var(--primary)` |

### 2.5 chip 筛选折叠

```vue
<Toolbar>
  <template #filters>
    <el-tag v-for="chip in activeFilterChips" :key="chip.key" closable @close="removeFilter(chip.key)">
      {{ chip.label }}: {{ chip.value }}
    </el-tag>
    <el-button text @click="openFilterDrawer">更多筛选</el-button>
  </template>
</Toolbar>

<el-drawer v-model="filterDrawerVisible" direction="rtl" :size="320">
  <!-- 完整筛选表单 -->
</el-drawer>
```

## 3. VideoList 改造细节（复用 + 列设置首次）

### 3.1 列设置（localStorage）

```js
const COLUMN_VISIBILITY_KEY = 'admin-videolist-columns'
const DEFAULT_VISIBLE_COLUMNS = ['title', 'thumbnail', 'type', 'status', 'created_at', 'operations']
const NARROW_VIEWPORT_HIDDEN = ['upload_user', 'created_at']  // < 1280px 自动隐藏

function readStoredColumns() {
  if (typeof window === 'undefined') return new Set(DEFAULT_VISIBLE_COLUMNS)
  try {
    const stored = window.localStorage.getItem(COLUMN_VISIBILITY_KEY)
    if (!stored) return new Set(DEFAULT_VISIBLE_COLUMNS)
    return new Set(JSON.parse(stored))
  } catch (_) {
    return new Set(DEFAULT_VISIBLE_COLUMNS)
  }
}

const columnVisibility = ref(readStoredColumns())
const isNarrowViewport = computed(() => viewportWidth.value < 1280)

function isColumnVisible(key) {
  if (!columnVisibility.value.has(key)) return false
  // 用户显式勾选过该列才会出现在 set；narrow viewport 自动隐藏除非用户显式开启
  if (isNarrowViewport.value && NARROW_VIEWPORT_HIDDEN.includes(key) && !isUserExplicitlyEnabled(key)) {
    return false
  }
  return true
}

function setColumnVisibility(key, visible) {
  if (visible) columnVisibility.value.add(key)
  else columnVisibility.value.delete(key)
  try {
    window.localStorage.setItem(COLUMN_VISIBILITY_KEY, JSON.stringify([...columnVisibility.value]))
  } catch (_) {}
}
```

### 3.2 列设置 popover UI

```vue
<el-popover trigger="click" :width="240">
  <template #reference>
    <el-button :icon="Setting">列设置</el-button>
  </template>
  <el-checkbox-group :model-value="[...columnVisibility]" @update:model-value="onColumnChange">
    <el-checkbox v-for="col in ALL_COLUMNS" :key="col.key" :value="col.key">
      {{ col.label }}
    </el-checkbox>
  </el-checkbox-group>
</el-popover>
```

### 3.3 编辑 drawer

复用 Phase 3 ImageCollectionManage drawer pattern；drawer 内 SectionCard 分区：

- 基础信息（默认展开，不折叠）
- 季集（仅 episode 类型显示；默认展开）
- 演员标签（默认折叠 collapsible）
- 图片合集（默认折叠 collapsible）
- 字幕（默认折叠 collapsible）
- 播放预览（默认折叠 collapsible，置顶或下沉视实际体验调整）

### 3.4 BulkActionBar 复用

同 ImageManage，传 `actions` 列表：
- 批量删除
- 批量改状态
- 批量打标

## 4. VideoUpload 改造细节（Wizard 拆 step）

### 4.1 文件拆分

```
admin-web/src/views/VideoUpload.vue           # 主容器 + step 状态 + 共享 form state
admin-web/src/views/VideoUpload/StepFile.vue  # Step 1: 选文件
admin-web/src/views/VideoUpload/StepBasic.vue # Step 2: 类型 + 基础信息（含 AV 条件字段）
admin-web/src/views/VideoUpload/StepRelate.vue # Step 3: 关联 + 上传 + 历史
admin-web/src/views/videoUpload.wizard.helpers.js # step 切换 / 字段保留 / 类型条件字段纯函数
admin-web/src/views/videoUpload.wizard.helpers.spec.js
```

### 4.2 主容器 `VideoUpload.vue`

```vue
<script setup>
import { ref, reactive } from 'vue'
import StepFile from './VideoUpload/StepFile.vue'
import StepBasic from './VideoUpload/StepBasic.vue'
import StepRelate from './VideoUpload/StepRelate.vue'
import { canAdvanceStep, getNextStep, getPrevStep } from './videoUpload.wizard.helpers'

const activeStep = ref(0)
const form = reactive({
  files: [],
  type: 'short',
  site_category: 'japanese',  // 仅 AV 类型使用
  title: '',
  description: '',
  tags: [],
  actor_ids: [],
  collection_id: '',
  image_collection_id: '',
})
const uploadProgress = ref({ active: false, items: [] })
const uploadHistory = ref([])

function next() {
  if (canAdvanceStep(activeStep.value, form)) {
    activeStep.value = getNextStep(activeStep.value)
  }
}

function prev() {
  activeStep.value = getPrevStep(activeStep.value)
}
</script>

<template>
  <Layout>
    <PageHeader title="上传中心" />
    <el-steps :active="activeStep" finish-status="success">
      <el-step title="选择文件" />
      <el-step title="基础信息" />
      <el-step title="关联与上传" />
    </el-steps>
    <StepFile v-show="activeStep === 0" v-model:files="form.files" />
    <StepBasic v-show="activeStep === 1" v-model:form="form" />
    <StepRelate v-show="activeStep === 2" v-model:form="form" :upload-progress="uploadProgress" :upload-history="uploadHistory" @upload="onUpload" />
    <Toolbar dense>
      <template #actions>
        <el-button :disabled="activeStep === 0" @click="prev">上一步</el-button>
        <el-button v-if="activeStep < 2" type="primary" :disabled="!canAdvanceStep(activeStep, form)" @click="next">下一步</el-button>
      </template>
    </Toolbar>
  </Layout>
</template>
```

**关键**：所有 step 用 `v-show` 而非 `v-if`，保证切换时子组件不卸载、字段不丢失；上传进度与历史在主容器，step 3 切出再切入时数据完整保留。

### 4.3 wizard helpers + spec

```js
// videoUpload.wizard.helpers.js
export const WIZARD_STEPS = ['file', 'basic', 'relate']

export function canAdvanceStep(currentIndex, form) {
  if (currentIndex === 0) return form.files && form.files.length > 0
  if (currentIndex === 1) return Boolean(form.type)
  return false  // last step 不允许"下一步"
}

export function getNextStep(currentIndex) {
  return Math.min(currentIndex + 1, WIZARD_STEPS.length - 1)
}

export function getPrevStep(currentIndex) {
  return Math.max(currentIndex - 1, 0)
}

export function shouldShowAVFields(type) {
  return type === 'av'
}

export function shouldShowCollectionField(type) {
  return type === 'short'
}
```

`videoUpload.wizard.helpers.spec.js` 真值表：

| 测试 | 输入 → 期望 |
|------|------------|
| `canAdvanceStep` step 0 文件 0 | false |
| `canAdvanceStep` step 0 文件 ≥ 1 | true |
| `canAdvanceStep` step 1 type='' | false |
| `canAdvanceStep` step 1 type='short' | true |
| `canAdvanceStep` step 2 任意 | false（last step） |
| `getNextStep(0/1/2)` | 1 / 2 / 2（clamp） |
| `getPrevStep(0/1/2)` | 0 / 0 / 1（clamp） |
| `shouldShowAVFields('av')` | true |
| `shouldShowAVFields('movie')` | false |
| `shouldShowCollectionField('short')` | true |
| `shouldShowCollectionField('movie')` | false |

## 5. router/index.js（3 路由 meta）

```js
{
  path: '/videos',
  component: VideoList,
  meta: { hideShellPageHeader: true }
},
{
  path: '/images',
  component: ImageManage,
  meta: { hideShellPageHeader: true }
},
{
  path: '/upload',
  component: VideoUpload,
  meta: { hideShellPageHeader: true }
}
```

## 6. spec 扩展

### `assets/themeTokens.spec.js`

`VIEW_HEX_AUDIT_TARGETS` 数组从 12 项扩到 15：

```js
const VIEW_HEX_AUDIT_TARGETS = [
  // Phase 2 已覆盖（8）
  '../views/Dashboard.vue', '../views/SystemSettings.vue', '../views/UserManage.vue',
  '../views/TaskMonitor.vue', '../views/IPTVManage.vue', '../views/CollectionManage.vue',
  '../views/ActorManage.vue', '../components/UploadProgress.vue',
  // Phase 3 已覆盖（4）
  '../views/ScrapePreview.vue', '../views/AVManualScrape.vue',
  '../views/TvSeriesManage.vue', '../views/ImageCollectionManage.vue',
  // Phase 4 新增（3）
  '../views/VideoList.vue', '../views/ImageManage.vue', '../views/VideoUpload.vue'
]
```

预期：实施前 1 个用例红（ImageManage 仍含玫红）→ 实施后 15 个用例全绿。

### `views/videoUpload.wizard.helpers.spec.js`

11 个真值表用例（见 §4.3）。

## 7. CONTEXT.md 追加（5 条新术语）

在「admin 设计系统术语」段尾追加 5 条新术语（PRD 第 7 节列出原文照搬）。不改动既有 Phase 1（6 条）+ Phase 3（2 条）。

## 8. plan.md 追加模板

```markdown
## 2026-05-24 · admin Phase 4：三巨头 IA 改造

- 摘要：admin-web 全面重设计第 4 阶段（收尾）—— ImageManage 引入「视图切换 + chip 筛选 + 2 个 drawer + BulkActionBar 首次试点」5 大新模式 + 3 处 hex 清零；VideoList 引入「列设置 localStorage」+ 复用 ImageManage 验证过的 chip + drawer + BulkActionBar 4 大模式；VideoUpload 拆 3 步 Wizard（StepFile / StepBasic / StepRelate）+ wizard helpers 单测；router 3 路由加 hideShellPageHeader meta；CONTEXT.md 新增 5 条 admin 设计系统术语；admin 重设计 4 阶段全部 Done。
- 受影响文件：admin-web/src/views/{VideoList,ImageManage,VideoUpload}.vue、admin-web/src/views/VideoUpload/{StepFile,StepBasic,StepRelate}.vue、admin-web/src/views/videoUpload.wizard.helpers.js + .spec.js、admin-web/src/router/index.js、admin-web/src/assets/themeTokens.spec.js（扩 15 文件）、CONTEXT.md（+5 术语）、tasks/2026-05-23-admin-three-pillars/{prd.md,implement.md,review.md,screenshots/,DONE.md}、plan.md
- 验证：cd admin-web && npm run build 通过；cd admin-web && npm test 全绿（≥ 82 用例）；H41–H48 手测通过；11 张截图归档；admin 重设计 4 阶段全部 DONE.md 闭环。
```

## 9. 风险与回退

- **风险 1**：BulkActionBar 是 Phase 1 留下的"现成" wrapper 但首次实接入，可能发现 API 缺 prop（如 disabled 状态 / 单条 action loading 反馈）；对策：在 ImageManage 落地时观察，若需补 prop 回 Phase 1 followup commit
- **风险 2**：VideoList 表格 column visibility 与 `el-table` 内置列 prop 的兼容性——`el-table-column` 的 `v-if` 切换在 1422 行表格里频繁挂载/卸载可能引起 layout 抖动；对策：用 `v-show` 而非 `v-if`（如果 el-table-column 支持）；若不支持则改用渲染时 filter 列定义数组
- **风险 3**：VideoUpload Wizard 用 `v-show` 而非 `v-if`，3 个 step 子组件都常驻 DOM；StepRelate 内的 `el-upload` / `UploadProgress` 在 step 1/2 时也挂载着但隐藏，可能引起 hover 误触发 / 内存占用；对策：观察实际体验，必要时改 `v-if` + 把上传状态提到主容器
- **风险 4**：ImageManage 视图切换瞬间分页与筛选状态丢失——v-model 切换时如果重 mount 子组件会重置；对策：把视图状态提到主容器，list / pagination / filter 都在主组件，子组件只负责渲染
- **风险 5**：BulkActionBar 浮起时与 PageHeader / Toolbar 的 z-index 冲突——对策：BulkActionBar 的 z-index 高于 Toolbar，但低于 drawer
- **回退**：所有改动是「替换 wrapper + dialog 改 drawer + 拆 step 子组件 + localStorage 持久化」纯加法/重构改动；恢复原 3 文件即可回到 Phase 4 实施前（保留 Phase 1-3 全部成果）；step 子组件直接删目录即可清理
