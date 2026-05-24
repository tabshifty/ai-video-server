# Review：admin 重设计 Phase 4 · 三巨头 IA 改造

- 日期：2026-05-24
- 关联 PRD：`prd.md`
- 关联 Implement：`implement.md`

## 1. 自动化必跑

```bash
cd admin-web
npm run build       # 必须通过
npm test            # 必须全绿（≥ 82 用例：79 + 视图层 audit 扩 + wizard helpers spec）
```

任一红则 review 失败，回 Implement 阶段修复，不得 skip 测试。

### 1.1 spec 扩展核查

| spec 文件 | 必须包含的断言组 |
|----------|----------------|
| `assets/themeTokens.spec.js`（扩展） | `VIEW_HEX_AUDIT_TARGETS` 数组共 15 项：Phase 2 8 + Phase 3 4 + Phase 4 3（VideoList / ImageManage / VideoUpload）；15 个用例分别断言不含 `#881337` / `#be123c` / `#7f1d1d` |
| `views/videoUpload.wizard.helpers.spec.js`（新增） | 11 个真值表用例：`canAdvanceStep` 5 case（step 0 文件0/≥1、step 1 type=''/'short'、step 2 任意）+ `getNextStep` / `getPrevStep` 各 3 case + `shouldShowAVFields` / `shouldShowCollectionField` 各 1 case |

### 1.2 既有测试不得回归

- Phase 1 既有 `themeTokens.spec.js`（外壳层 audit）/ `Layout.spec.js` / `commandPalette.helpers.spec.js` 保持绿
- Phase 2 视图层 audit 8 文件保持绿
- Phase 3 视图层 audit 4 文件保持绿
- 业务 helpers spec 全部保持绿（`videoList.helpers.spec.js` / `videoUpload.remote.spec.js` 等）

## 2. 手测脚本

**预条件**：`cd admin-web && npm run dev`，浏览器登录 admin（1440px 桌面分辨率）。

按「外观锚点 + 功能不破」二维核对，任一不达预期即视为 review 失败。

### H41 / ImageManage —— 最广模式覆盖

**外观锚点**：
- [ ] 顶部 `PageHeader` 显示「图片管理」，Layout 顶栏不再显示同名标题
- [ ] `Toolbar` 含「上传图片」+ chip 筛选（状态 / 演员 / 图片合集等）+ 视图切换（网格/列表）
- [ ] 默认进入网格视图（无既有 localStorage 偏好时）；切到列表 → 刷新页面 → 仍是列表
- [ ] 列表清空时 `EmptyState`
- [ ] 网格卡片 hover 显示「编辑 / 删除 / 切换启用」+ 左上 selection checkbox
- [ ] 编辑入口 + 上传入口都是 drawer（560px / 全屏自适应）
- [ ] Scoped CSS 不含 `#881337` / `#7f1d1d`

**功能不破**：
- [ ] 图片 CRUD + 启用切换 + 演员/合集绑定全可用
- [ ] 上传 drawer 多文件批量选图 + 进度反馈 + 上传完成回填
- [ ] 网格/列表切换不重置分页与筛选

### H42 / VideoList —— 列设置首次试点

**外观锚点**：
- [ ] 顶部 `PageHeader`「视频管理」，Layout 顶栏不再显示同名标题
- [ ] `Toolbar` 含 chip 筛选 + 「更多筛选」抽屉触发 + 列设置按钮
- [ ] 表格仅显示用户配置或系统默认列
- [ ] 窗口拖到 < 1280px，未显式开启的次要列自动隐藏；开启后仍尊重用户偏好
- [ ] 行操作 hover 显示快捷操作 + 「⋯」更多菜单
- [ ] 表格清空时 `EmptyState`
- [ ] 编辑入口 drawer 560px，drawer 内 SectionCard 分区可见（基础 / 季集 / 演员标签 / 图片合集 / 字幕 / 播放预览）

**功能不破**：
- [ ] 表格分页 / 排序 / 搜索 / 筛选全可用
- [ ] 编辑 drawer 内 episode / AV / short / movie 各类型 conditional 字段保留原 schema
- [ ] 播放预览仍可用
- [ ] 字幕管理仍可用

### H43 / VideoUpload —— 3 步 Wizard

**外观锚点**：
- [ ] `PageHeader`「上传中心」+ `el-steps` 显示 3 步
- [ ] Step 1 显示文件拖拽 + 已选清单
- [ ] Step 2 显示类型选择；type=av 时显示「AV 地区分类」+ 标题 + 描述
- [ ] Step 3 显示标签 / 演员 / 图片合集 / 所属合集（仅 short）+ 「开始上传」+ 进度 + 历史
- [ ] 底部固定 Toolbar 含下一步 / 上一步 / 提交

**功能不破**：
- [ ] Step 1 必选 ≥ 1 文件（否则下一步禁用）
- [ ] Step 2 type=video 时 AV 字段隐藏
- [ ] Step 1 ⇄ 2 ⇄ 3 切换字段全部保留
- [ ] Step 3 开始上传 → 进度反馈 → 切回 step 1/2 再回 step 3 进度与历史不丢失
- [ ] 取消上传 / 清空已选 / 清空记录三按钮在 step 3 底部

### H44 / admin 编辑入口 Drawer dirty 通用契约

- [ ] VideoList 编辑 drawer / ImageManage 编辑 drawer / ImageManage 上传 drawer 在 dirty 状态点取消 / Esc / 遮罩 → 弹「未保存」确认
- [ ] save 成功后立即关闭，不弹假确认（snapshot 已对齐 form）
- [ ] cancel 确认丢弃后不二次弹同款确认（snapshot 已复位）

### H45 / admin 批量操作浮条 BulkActionBar 首次试点

- [ ] ImageManage / VideoList 选中 ≥ 1 项 → BulkActionBar 顶部浮起
- [ ] 显示「已选 N 项」+ 主操作（批量删除 / 批量改状态 / 批量打标）
- [ ] 取消选择即消失
- [ ] 关闭 tab 不持久化选择

### H46 / localStorage 持久化

- [ ] VideoList 列设置写入 `localStorage` key `admin-videolist-columns`；刷新页面后保持
- [ ] ImageManage 视图模式写入 `localStorage` key `admin-imagemanage-view`；刷新页面后保持
- [ ] 隐私模式 / quota exceeded 时不破坏视图（兜底默认）

### H47 / admin 路由 hideShellPageHeader 标记

- [ ] DevTools Vue Router 检查 `/videos` / `/images` / `/upload` 3 个 route 的 `meta.hideShellPageHeader=true`
- [ ] 进入这 3 个路由时 Layout shell 顶栏不渲染同名标题
- [ ] 与 Phase 2-3 已落的 11 视图行为一致

### H48 / 全局自动化

- [ ] `cd admin-web && npm run build` 通过
- [ ] `cd admin-web && npm test` 全绿（≥ 82 用例）
- [ ] `themeTokens.spec.js` 视图层 audit 15 文件全部命中
- [ ] `videoUpload.wizard.helpers.spec.js` 11 用例全绿
- [ ] `CONTEXT.md`「admin 设计系统术语」段新增 5 条
- [ ] `plan.md` 顶部追加 Phase 4 实施条目

## 3. 截图归档

`tasks/2026-05-23-admin-three-pillars/screenshots/` 必须包含：

- `before-image-manage.png` / `after-image-manage.png`（列表态）
- `after-image-manage-grid.png`（网格视图）
- `after-image-manage-drawer.png`（上传或编辑 drawer 打开形态）
- `after-image-manage-bulk.png`（BulkActionBar 浮起形态）
- `before-video-list.png` / `after-video-list.png`（列表态）
- `after-video-list-drawer.png`（编辑 drawer 打开形态）
- `after-video-list-columns.png`（列设置 popover 打开形态）
- `before-video-upload.png` / `after-video-upload.png`（Wizard 任一 step）

共 11 张，统一 1440px 桌面分辨率。

after 必须显示：
- PageHeader 在视图顶部、字号字重符合 Phase 1 token
- 应接入的 wrapper（Toolbar / SectionCard / EmptyState / BulkActionBar）肉眼可见
- VideoList drawer + 列设置 + chip 筛选 + BulkActionBar 4 个新模式肉眼可见
- ImageManage 网格 / 列表两种视图 + drawer + BulkActionBar 4 种状态肉眼可见
- VideoUpload 3 步 Wizard 顶部 el-steps + 底部 Toolbar 肉眼可见
- 整页色调与 Phase 1-3 一致，无残留玫红描边 / 玫红色文字 / 玫红按钮

不做像素级 diff；目视判断「IA 改造可见 + 视觉切到新设计系统」即可通过。

## 4. Done definition

- [ ] 第 1 节自动化全绿
- [ ] 第 1.1 节 spec 扩展核查全过
- [ ] 第 1.2 节既有测试零回归
- [ ] 第 2 节手测 H41–H48 全部通过
- [ ] 第 3 节 11 张截图归档完成
- [ ] `CONTEXT.md` 新增 5 条术语（Phase 1 6 + Phase 3 2 + Phase 4 5 = 13 条收口）
- [ ] `plan.md` 顶部追加 Phase 4 实施条目
- [ ] 全部改动落在 1 commit（中文 subject + body）；如有 code review followup 走「实现 + followup」二次 commit
- [ ] `tasks/2026-05-23-admin-three-pillars/DONE.md` 含完成日期 + commit hash + 验证摘要 + admin 重设计 4 阶段全部 Done 闭环标记

## 5. review 后流程

review 通过 → 用户多视图点查一轮 → 新增 `DONE.md` → admin 重设计 4 阶段全部完成 → 可启动新的产品 / 后端 / TV 端任务，不再属于 admin 重设计范围。

## 6. review 失败处理

- **自动化红**：回 Implement 修复，不得 skip 或注释掉用例
- **手测外观锚点不达预期**：判断是 wrapper 不够（回 Phase 1 followup commit 补 wrapper API）还是接入错（回 Implement 改视图）
- **手测功能不破红**：**最高优先级**回退或修复——Phase 4 包含 IA 改造，但不允许引入功能回归；若回归不能立刻定位，先把对应视图回滚到 Phase 3 状态（Phase 3 完成后的 master），再单独 follow-up
- **drawer / wizard 问题集中**：若 ImageManage 双 drawer 或 VideoUpload Wizard 出现多个回归，临时回退该视图的 IA 改造（保留 PageHeader / Toolbar / hex 清零等视觉部分），延后到单独 polish 任务；本期视为部分完成而非整体失败
- **scope creep**：如发现需要新增 wrapper、改 store / API、改路由 URL、引入新依赖（Tailwind / shadcn-vue 等），**严格拒绝**进入本期，记入 plan.md 下一期 backlog
