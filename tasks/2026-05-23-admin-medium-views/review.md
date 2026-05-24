# Review：admin 重设计 Phase 3 · 中等视图重排

- 日期：2026-05-24
- 关联 PRD：`prd.md`
- 关联 Implement：`implement.md`

## 1. 自动化必跑

```bash
cd admin-web
npm run build       # 必须通过
npm test            # 必须全绿（≥ 77 用例：原 75 + 新增 2 个 audit）
```

任一红则 review 失败，回 Implement 阶段修复，不得 skip 测试。

### 1.1 spec 扩展核查

| spec 文件 | 必须包含的断言组 |
|----------|----------------|
| `assets/themeTokens.spec.js`（扩展） | `VIEW_HEX_AUDIT_TARGETS` 数组共 10 项：Phase 2 已覆盖 8（Dashboard / SystemSettings / UserManage / TaskMonitor / IPTVManage / CollectionManage / ActorManage / UploadProgress）+ Phase 3 新增 2（ScrapePreview / AVManualScrape）；10 个用例分别断言不含 `#881337` / `#be123c` / `#7f1d1d` |

### 1.2 既有测试不得回归

- Phase 1 既有 `themeTokens.spec.js`（外壳层 audit）/ `Layout.spec.js` / `commandPalette.helpers.spec.js` 必须保持绿
- Phase 2 既有 `themeTokens.spec.js`（Phase 2 视图层 audit）保持绿（Phase 3 仅扩 2 个用例不覆盖原 8）
- 业务 helpers spec 全部保持绿
- `commandPalette.helpers.spec.js` / `scrapePreview.helpers.spec.js` / `avManualScrape.helpers.spec.js` / `tvSeriesManage.helpers.spec.js` / `imageCollectionManage.helpers.spec.js` 全部保持绿

## 2. 手测脚本

**预条件**：`cd admin-web && npm run dev`，浏览器登录 admin（1440px 桌面分辨率）。

按「外观锚点 + 功能不破」二维核对，任一不达预期即视为 review 失败。

### H31 / ImageCollectionManage —— drawer 首次试点

**外观锚点**：
- [ ] 顶部 `PageHeader` 显示「图片合集」，无副标题；Layout 顶栏不再显示同名标题
- [ ] `Toolbar` 内显示「上传图片」+「创建合集」+ 搜索框
- [ ] 合集列表清空数据后显示 `EmptyState`
- [ ] 点「上传图片」打开右侧 `<el-drawer>`：宽 560px（拖窗口到 800px 验证全屏自适应）
- [ ] drawer 内 `SectionCard` ×3：批量选图（默认展开）/ 默认元数据（默认展开）/ 上传队列（执行后展开）
- [ ] drawer 底部 `Toolbar` 固定：左「取消」、右「上传」(primary)

**功能不破**：
- [ ] 合集 CRUD + 视频绑定 + 取消绑定全可用
- [ ] drawer 内上传 ≥ 1 张图片 + 进度反馈正常 + 上传完成回填表格
- [ ] 默认演员 / 默认合集 / 默认备注三字段保存生效
- [ ] drawer dirty 时点取消 / Esc / 遮罩 → 弹「未保存确认」 → 选「取消」回到 drawer / 选「确认丢弃」关 drawer
- [ ] drawer 内无嵌套 dialog（演员选择 / 合集选择走 popover 或 inline list）

### H32 / ScrapePreview

**外观锚点**：
- [ ] `PageHeader`「通用刮削」+ `Toolbar`
- [ ] `SectionCard` ×3 包裹：查询表单 / 候选列表 + 候选详情 / 编辑确认
- [ ] 候选列表清空时 `EmptyState`
- [ ] Scoped CSS 不含 `#7f1d1d` / `#be123c` / `#881337`

**功能不破**：
- [ ] 电影 / 剧集两种类型切换
- [ ] 查询预览 + 选择候选 + 编辑 metadata + 提交确认
- [ ] `bypass_cache` 重抓
- [ ] 季 / 集绑定正确

### H33 / AVManualScrape

**外观锚点**：
- [ ] `PageHeader`「AV 手动刮削」+ `Toolbar`（站点切换 + 查询按钮）
- [ ] `SectionCard` ×N 包裹：手动预览 / 候选列表 / 候选详情 / 待确认列表
- [ ] 候选无数据时 `EmptyState`
- [ ] Scoped CSS 不含 `#7f1d1d`

**功能不破**：
- [ ] 切换 JavDB / JavBus / JavLibrary / ThePornDB / MDCX 五站点查询
- [ ] 候选选择 + 确认刮削
- [ ] 弃刮路径
- [ ] `av_scrape_pending` 待确认列表加载

### H34 / TvSeriesManage

**外观锚点**：
- [ ] `PageHeader`「电视剧管理」+ `Toolbar`（创建电视剧 + 搜索）
- [ ] `SectionCard` 三层嵌套：剧基础 → 季列表 → 每季嵌套集列表
- [ ] 列表无剧时 `EmptyState`
- [ ] 二层 / 三层嵌套 SectionCard 视觉不拥挤（padding / border 合理）

**功能不破**：
- [ ] 剧创建 / 编辑 / 删除
- [ ] 季创建 / 编辑 / 删除
- [ ] 集创建 / 编辑 / 删除
- [ ] 嵌套表单字段保存 / 读取正确

### H35 / admin 编辑入口 Drawer dirty 关闭契约

- [ ] ImageCollectionManage 上传 drawer 在 batch 选了 1 张图片 / 填了默认演员 / 填了默认合集 / 填了默认备注 4 种 dirty 触发条件，任一为真 → 关闭弹确认
- [ ] drawer 未 dirty 时（刚打开未填任何字段）直接关闭无确认提示
- [ ] drawer 内保存成功后 dirty 标志复位，再点关闭无确认
- [ ] drawer 关闭后再次打开是干净状态（dirty=false）

### H36 / admin 路由 hideShellPageHeader 标记

- [ ] 打开 DevTools，Vue Router 检查 `/image-collections` / `/scrape` / `/av-scrape` / `/tv-series` 4 个 route 的 `meta.hideShellPageHeader` 都是 `true`
- [ ] 进入这 4 个路由时 Layout shell 顶栏不渲染同名标题，仅视图体内 PageHeader 渲染
- [ ] 切换到 Phase 2 已覆盖的视图（如 `/users`）行为一致

### H37 / 全局自动化

- [ ] `cd admin-web && npm run build` 通过
- [ ] `cd admin-web && npm test` 全绿（≥ 77 用例）
- [ ] `themeTokens.spec.js` 视图层 audit 10 文件全部命中
- [ ] `CONTEXT.md`「admin 设计系统术语」段新增 2 条术语：`admin 编辑入口 Drawer` + `admin 路由 hideShellPageHeader 标记`
- [ ] `plan.md` 顶部追加 Phase 3 实施条目

## 3. 截图归档

`tasks/2026-05-23-admin-medium-views/screenshots/` 必须包含：

- `before-image-collection.png` / `after-image-collection.png`
- `after-image-collection-drawer.png`（drawer 打开状态专属，首次试点）
- `before-scrape-preview.png` / `after-scrape-preview.png`
- `before-av-manual-scrape.png` / `after-av-manual-scrape.png`
- `before-tv-series-manage.png` / `after-tv-series-manage.png`

共 9 张，统一 1440px 桌面分辨率。

after 必须显示：
- PageHeader 在视图顶部、字号字重符合 Phase 1 token
- 应接入的 wrapper（Toolbar / SectionCard / EmptyState）肉眼可见
- ImageCollectionManage 单独额外展示 drawer 打开形态（560px 右侧滑入 + 底部 Toolbar 固定）
- 整页色调与 Phase 1 / Phase 2 一致，无残留玫红描边 / 玫红色文字 / 玫红按钮

不做像素级 diff；目视判断「视觉切到新设计系统 + drawer 改造可见」即可通过。

## 4. Done definition

- [ ] 第 1 节自动化全绿
- [ ] 第 1.1 节 spec 扩展核查全过
- [ ] 第 1.2 节既有测试零回归
- [ ] 第 2 节手测 H31–H37 全部通过
- [ ] 第 3 节 9 张截图归档完成
- [ ] `CONTEXT.md` 新增 2 条术语
- [ ] `plan.md` 顶部追加 Phase 3 实施条目
- [ ] 全部改动落在 1 commit（中文 subject + body）；如有 code review followup 走「实现 + followup」二次 commit
- [ ] `tasks/2026-05-23-admin-medium-views/DONE.md` 含完成日期 + commit hash + 验证摘要

## 5. review 后流程

review 通过 → 用户多视图点查一轮 → 新增 `DONE.md` → 进入 Phase 4（三巨头 IA 改造，`tasks/2026-05-23-admin-three-pillars/`）。

Phase 4 时 ImageCollectionManage drawer 已落地的 pattern 直接复用：
- 宽 560px / `< 1024px` 全屏 → 三巨头同款
- `before-close` dirty guard → 抽 composable 或就地实现
- 内部 `SectionCard` 折叠 → 三巨头 drawer 各自分区
- 底部 `Toolbar dense` 含取消 + 主操作 → 三巨头同款

## 6. review 失败处理

- **自动化红**：回 Implement 修复，不得 skip 或注释掉用例
- **手测外观锚点不达预期**：判断是 wrapper 不够（回 Phase 1 followup commit 补 wrapper API）还是接入错（回 Implement 改视图）
- **手测功能不破红**：**最高优先级**回退或修复——Phase 3 是视觉重排 + 单点 IA 改造（ImageCollectionManage drawer），不允许引入功能回归；若回归不能立刻定位，先把对应视图回滚到 Phase 2 状态，再单独 follow-up
- **drawer 改造问题集中**：若 ImageCollectionManage drawer 出现多个回归，临时回退 drawer 改造到 dialog（保留视觉重排），延后到 Phase 4 与三巨头一起改 drawer；本期视为 Phase 3 部分完成而非整体失败
- **scope creep**：如发现需要改 Phase 4 视图、新增 wrapper、改 IA、改 store / API，**严格拒绝**进入本期，记入 plan.md 下一期 backlog
