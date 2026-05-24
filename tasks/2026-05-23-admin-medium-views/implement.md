# Implement：admin 重设计 Phase 3 · 中等视图重排

- 日期：2026-05-24
- 关联 PRD：`prd.md`
- 关联 Phase 1：`tasks/2026-05-23-admin-shell-redesign/DONE.md`
- 关联 Phase 2：`tasks/2026-05-23-admin-simple-views/DONE.md`
- 任务三段执行流：当前阶段 = Implement

## 1. 实施顺序（TDD 友好）

1. **扩展 spec 先**：`assets/themeTokens.spec.js` 视图层 audit 数组从 8 文件扩到 10（追加 ScrapePreview + AVManualScrape）；预期 2 个新用例红 → 实施后绿
2. **`ImageCollectionManage.vue` 第一**（drawer 改造试点）
3. **`ScrapePreview.vue`** → **`AVManualScrape.vue`**（视觉重排 + hex 清零）
4. **`TvSeriesManage.vue`**（深嵌套 SectionCard）
5. **`router/index.js`** 给 4 路由设 `meta.hideShellPageHeader = true`
6. **`CONTEXT.md`** 新增 2 条术语
7. **手测 H31–H37** + 截图归档
8. **`npm run build` + `npm test`** 全绿
9. **追加 `plan.md` + commit**

## 2. ImageCollectionManage drawer 改造细节（核心）

### 2.1 当前形态（dialog）

```vue
<el-dialog v-model="uploadDialogVisible" title="批量上传图片" width="640">
  <el-form>...</el-form>
  <template #footer>
    <el-button @click="onCancel">取消</el-button>
    <el-button type="primary" @click="onSubmit">上传</el-button>
  </template>
</el-dialog>
```

### 2.2 改造后形态（drawer）

```vue
<el-drawer
  v-model="uploadDrawerVisible"
  :size="drawerSize"
  direction="rtl"
  :before-close="onBeforeClose"
  destroy-on-close
>
  <template #header>
    <span>批量上传图片</span>
  </template>
  <div class="upload-drawer-body">
    <SectionCard>
      <template #title>批量选图</template>
      <!-- 既有 el-upload 区域 -->
    </SectionCard>
    <SectionCard>
      <template #title>默认元数据</template>
      <!-- 默认演员 / 默认合集 / 默认备注字段 -->
    </SectionCard>
    <SectionCard v-if="uploadSummary?.items?.length">
      <template #title>上传队列</template>
      <!-- 既有 uploadSummary 表格 -->
    </SectionCard>
  </div>
  <template #footer>
    <Toolbar dense>
      <template #actions>
        <el-button @click="onCancel">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="onSubmit">上传</el-button>
      </template>
    </Toolbar>
  </template>
</el-drawer>
```

### 2.3 关键字段与逻辑

- `drawerSize`：computed，`viewportWidth < 1024 ? '100%' : 560`（建议提到共享 helper 或就地实现）
- `onBeforeClose(done)`：实现 dirty 检查；dirty 时 `ElMessageBox.confirm('未保存的修改将丢失', '确认关闭', { confirmButtonText: '确认丢弃', cancelButtonText: '取消', type: 'warning' }).then(() => done()).catch(() => {})`
- dirty 信号：监听 `selectedFiles.length > 0 || defaultActorIds.length || defaultCollectionId || defaultRemark.length` 任一为真即 dirty
- 保存成功后调 `done()` 并复位 dirty 触发字段
- drawer 底部 footer 用 `Toolbar` wrapper 而非 `el-button` 直接平铺，统一视觉

### 2.4 Phase 4 复用契约

ImageCollectionManage drawer 改造完后，Phase 4 三巨头（VideoList / ImageManage / VideoUpload）的 drawer 改造直接套同款 pattern：
- 宽 560px / `< 1024px` 全屏
- `before-close` dirty 检查
- 内部 `SectionCard` 折叠区块
- 底部 `Toolbar dense` 含取消 + 主操作

可选优化（**本期不做**，留 Phase 4 视情况）：抽 `useDrawerDirtyGuard(visibleRef, dirtyComputed)` composable；本期 ImageCollectionManage 就地实现，跑通后再判断是否抽 composable。

## 3. 通用重排配方（4 视图按此模板）

```vue
<script setup>
import PageHeader from '../components/base/PageHeader.vue'
import Toolbar from '../components/base/Toolbar.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
// ImageCollectionManage 才需要 drawer 相关 import
</script>

<template>
  <Layout>
    <div class="page-shell">
      <PageHeader title="<视图中文名>">
        <template #actions>...</template>
      </PageHeader>

      <Toolbar>
        <template #filters>...</template>
        <template #actions>...</template>
      </Toolbar>

      <SectionCard><template #title>...</template>...</SectionCard>
      <!-- 多个 SectionCard 按视图实际分区 -->

      <template v-if="<有数据>">...</template>
      <EmptyState v-else title="..." description="..." />

      <!-- ImageCollectionManage 才有的 drawer，其它视图无 -->
    </div>
  </Layout>
</template>

<style scoped>
/* 删除原 .page-header / .section-head / 自定义卡片 CSS（已被 wrapper 接管） */
/* 删除/替换所有硬编码 hex */
</style>
```

## 4. 每视图具体改动点

### 4.1 `views/ImageCollectionManage.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="图片合集" />` |
| 操作行 | `<Toolbar><template #filters><el-input v-model="search" /></template><template #actions><el-button type="primary" @click="onUploadClick">上传图片</el-button><el-button @click="onCreateClick">创建合集</el-button></template></Toolbar>` |
| 列表 | 无合集时 `<EmptyState title="暂无图片合集" />` |
| 上传 dialog | **删除整段 `<el-dialog>`，改 `<el-drawer>`**（见 §2.2 模板） |
| dirty guard | 新增 `dirty` computed + `onBeforeClose` 处理 |
| Scoped CSS | 删除 `.upload-form` / `.upload-dialog-footer` 等 dialog 相关样式，新增 `.upload-drawer-body` 容器（gap-section + overflow-y auto）|

### 4.2 `views/ScrapePreview.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="通用刮削" />` |
| 操作行 | `<Toolbar><template #filters><!-- type 切换 / video_id / title / year 等已有筛选挪进来 --></template></Toolbar>` |
| SectionCard ×3 | (a) 查询表单 / (b) 候选列表 + 候选详情（左右二栏可在 SectionCard 内布局，本期不强切分） / (c) 编辑并确认 |
| 列表无数据 | 候选列表 `EmptyState` 替换 `<el-empty>` |
| Scoped CSS | 删 `#7f1d1d` / `#be123c` ×2 / `#881337`，全部换 token |

### 4.3 `views/AVManualScrape.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="AV 手动刮削" />` |
| 操作行 | `<Toolbar><template #filters><!-- 站点切换 --></template><template #actions><el-button @click="loadPending">刷新待确认</el-button></template></Toolbar>` |
| SectionCard ×N | 手动预览 / 候选列表 / 候选详情 / 确认表单 / 待确认列表 |
| 列表无数据 | `EmptyState` |
| Scoped CSS | 删 `#7f1d1d`，全部换 `var(--primary)` |

### 4.4 `views/TvSeriesManage.vue`

| 改动 | 描述 |
|------|------|
| 顶部 | `<PageHeader title="电视剧管理" />` |
| 操作行 | `<Toolbar><template #filters><el-input v-model="search" /></template><template #actions><el-button type="primary" @click="onCreateSeries">创建电视剧</el-button></template></Toolbar>` |
| 列表无剧时 | `EmptyState` |
| SectionCard 三层嵌套 | (a) 剧基础信息 → (b) 季列表（每季一个 SectionCard）→ (c) 集列表（每集 inline form）；嵌套时确认 SectionCard 视觉在二层 / 三层 padding / border 不冲突 |

### 4.5 `router/index.js`（4 路由加 meta）

```js
{
  path: '/image-collections',
  component: ImageCollectionManage,
  meta: { hideShellPageHeader: true }
},
{
  path: '/scrape',
  component: ScrapePreview,
  meta: { hideShellPageHeader: true }
},
{
  path: '/av-scrape',
  component: AVManualScrape,
  meta: { hideShellPageHeader: true }
},
{
  path: '/tv-series',
  component: TvSeriesManage,
  meta: { hideShellPageHeader: true }
}
```

## 5. spec 扩展

### `admin-web/src/assets/themeTokens.spec.js`

把 Phase 2 已扩的 `PHASE2_VIEWS` 数组改名为 `VIEW_HEX_AUDIT_TARGETS` 并扩展到 10 项：

```js
const VIEW_HEX_AUDIT_TARGETS = [
  // Phase 2 已覆盖
  'src/views/Dashboard.vue',
  'src/views/SystemSettings.vue',
  'src/views/UserManage.vue',
  'src/views/TaskMonitor.vue',
  'src/views/IPTVManage.vue',
  'src/views/CollectionManage.vue',
  'src/views/ActorManage.vue',
  'src/components/UploadProgress.vue',
  // Phase 3 新增
  'src/views/ScrapePreview.vue',
  'src/views/AVManualScrape.vue',
]
```

`describe('phase 2 views hex audit')` 块名同步更名为 `'view hex audit'` 表达跨期意图。

预期：
- 实施前：2 个新用例红（ScrapePreview / AVManualScrape 仍含 hex）
- 实施后：10 个用例全绿

## 6. CONTEXT.md 追加（一并入本期 commit）

在「admin 设计系统术语」段尾追加 2 条新术语（PRD 第 7 节列出原文照搬）。不改动既有 6 条术语。

## 7. plan.md 追加模板

```markdown
## 2026-05-24 · admin Phase 3：中等视图重排

- 摘要：admin-web 全面重设计第 3 阶段——4 个中等复杂度视图（ImageCollectionManage / ScrapePreview / AVManualScrape / TvSeriesManage）按 Phase 1 设计系统接入 PageHeader / Toolbar / SectionCard / EmptyState；ImageCollectionManage 上传 dialog → drawer 改造作为 Phase 4 三巨头复用 ground truth；router/index.js 4 个新路由加 hideShellPageHeader meta；ScrapePreview + AVManualScrape 视图层硬编码 hex 5 处清零；CONTEXT.md 新增 2 条 admin 设计系统术语。
- 受影响文件：admin-web/src/views/{ImageCollectionManage,ScrapePreview,AVManualScrape,TvSeriesManage}.vue、admin-web/src/router/index.js、admin-web/src/assets/themeTokens.spec.js（扩 10 文件 audit）、CONTEXT.md（+2 术语）、tasks/2026-05-23-admin-medium-views/{prd.md,implement.md,review.md,screenshots/,DONE.md}、plan.md
- 验证：cd admin-web && npm run build 通过；cd admin-web && npm test 全绿（≥ 77 用例）；H31–H37 手测通过；9 张截图归档。
```

## 8. 风险与回退

- **风险 1**：ImageCollectionManage drawer 内 `<el-upload>` 在 drawer body overflow-y auto 容器内拖拽 / 进度反馈是否还正常——对策：实施时观察实际 DOM 行为；若失败暂时 fallback 到 dialog
- **风险 2**：`onBeforeClose` dirty 守卫与原 dialog 关闭路径行为差异——dialog 没 dirty 守卫；新 drawer 引入后用户可能误以为「点 Esc 即取消」的体验改变，PRD H35 已显式 cover 这条
- **风险 3**：TvSeriesManage 三层嵌套 SectionCard 视觉拥挤——对策：实施时给中层 / 内层 SectionCard 加 `dense` prop 或 reduced padding；如 wrapper 不支持，scoped CSS 局部覆写
- **风险 4**：CONTEXT.md 「admin 编辑入口 Drawer」条目宣称「drawer 内不嵌套 dialog」，但 Element Plus 某些场景（如 `ElMessageBox.confirm`）本质是 dialog——澄清：dirty 确认弹是 `ElMessageBox` 全局 modal，不属于 drawer 嵌套 dialog 范畴；术语描述精确到「业务编辑 dialog」
- **回退**：所有改动是「替换 wrapper + dialog 改 drawer + 删除 scoped CSS 硬编码」纯替换改动；恢复原文件即可回到 Phase 3 实施前（保留 Phase 1 / Phase 2 全部成果）
