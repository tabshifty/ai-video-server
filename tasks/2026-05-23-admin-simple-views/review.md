# Review：admin 重设计 Phase 2 · 简单视图重排

- 日期：2026-05-24
- 关联 PRD：`prd.md`
- 关联 Implement：`implement.md`

## 1. 自动化必跑

```bash
cd admin-web
npm run build       # 必须通过
npm test            # 必须全绿（≥ 75 用例：原 67 + 新 audit 8）
```

任一红则 review 失败，回 Implement 阶段修复，不得 skip 测试。

### 1.1 spec 扩展核查

| spec 文件 | 必须包含的断言组 |
|----------|----------------|
| `assets/themeTokens.spec.js`（扩展） | `describe('phase 2 views hex audit')` 块；对 `Dashboard.vue` / `SystemSettings.vue` / `UserManage.vue` / `TaskMonitor.vue` / `IPTVManage.vue` / `CollectionManage.vue` / `ActorManage.vue` / `components/UploadProgress.vue` 共 8 文件分别断言不含 `#881337` / `#be123c` / `#7f1d1d` |

### 1.2 既有测试不得回归

- Phase 1 既有 `themeTokens.spec.js` / `Layout.spec.js` / `commandPalette.helpers.spec.js` 必须保持绿
- 既有业务 helpers spec（`videoList.helpers.spec.js` / `scrapePreview.helpers.spec.js` / `avManualScrape.helpers.spec.js` / `tvSeriesManage.helpers.spec.js` / `imageCollectionManage.helpers.spec.js` / `videoUpload.remote.spec.js` / `adminTablePagination.helpers.spec.js` / `admin.spec.js` / `router/transition.spec.js`）全部保持绿

## 2. 手测脚本

**预条件**：`cd admin-web && npm run dev`，浏览器登录 admin（1440px 桌面分辨率）。

每条手测按「外观锚点 + 功能不破」二维核对，任一不达预期即视为 review 失败。

### H21 / Dashboard

**外观锚点**：
- [ ] 顶部 `PageHeader` 显示「系统仪表盘」，无副标题
- [ ] 8 个 `StatCard` 替换原 `.metric-card`，数字 `tabular-nums` 对齐
- [ ] 1 个 `SectionCard` 包裹 ECharts 趋势图
- [ ] ECharts 主线色 / 区域填充派生自 `--primary`（视觉上是冷蓝、不是玫红 `#e11d48`）
- [ ] Scoped CSS 不含 `#cad8f5` / `#e11d48` / `#fda4af` 任何硬编码

**功能不破**：
- [ ] `stats` 接口返回数字与卡片一一对应（短视频 / 电影 / 电视剧集 / AV / 总用户数 / 今日上传 / 转码队列长度 / 磁盘剩余 8 项）
- [ ] 拖动浏览器宽度，ECharts 跟随 resize
- [ ] 手动断开后端模拟加载失败，显示 `EmptyState` 含「重试」按钮

### H22 / SystemSettings

**外观锚点**：
- [ ] 顶部 `PageHeader` 显示「系统设置」
- [ ] 表单按 N 个 `SectionCard` 分组（基础 / 转码 / 刮削 / 安全等）
- [ ] Scoped CSS 不含 `#7f1d1d` / `#9f1239` / `#6b7280` / `#e5e7eb`

**功能不破**：
- [ ] 任一字段修改 → 保存 → 刷新页面后保留
- [ ] 必填校验仍弹错

### H23 / UserManage

**外观锚点**：
- [ ] 顶部 `PageHeader` + `Toolbar`（刷新 + 添加用户两按钮，对齐 Phase 1 设计）
- [ ] 表格清空数据时显示 `EmptyState`
- [ ] Scoped CSS 不含 `#6b7280` / `#7f1d1d`

**功能不破**：
- [ ] 表格分页正常
- [ ] 点「添加用户」弹出**dialog**（保留 dialog 形态，不改 drawer）
- [ ] 创建 + 改角色 + 删除全可用，操作后表格刷新

### H24 / TaskMonitor

**外观锚点**：
- [ ] 顶部 `PageHeader` + `Toolbar`（刷新 + 状态筛选 chip）
- [ ] 顶部 3 个 `StatCard`（队列 / 失败 / 处理中）
- [ ] 任务列表清空时 `EmptyState`

**功能不破**：
- [ ] 自动刷新计时器正常工作（页面打开等待计时器周期后列表自动更新）
- [ ] 状态筛选 chip 切换正确过滤列表

### H25 / IPTVManage

**外观锚点**：
- [ ] 顶部 `PageHeader` + `Toolbar`（刷新 + 上传 M3U + 远程拉取三按钮）
- [ ] N 个 `StatCard`（频道总数 / 分组数等）
- [ ] M3U 源配置在 `SectionCard` 内
- [ ] 频道列表清空时 `EmptyState`
- [ ] Scoped CSS 不含 `#7f1d1d`

**功能不破**：
- [ ] 上传 M3U 文件成功 + 频道列表刷新
- [ ] 填入远程 URL + 远程拉取成功
- [ ] 手动刷新解析正常

### H26 / CollectionManage

**外观锚点**：
- [ ] 顶部 `PageHeader` + `Toolbar`（创建合集 + 搜索框）
- [ ] 合集列表清空时 `EmptyState`

**功能不破**：
- [ ] 创建 / 编辑 / 删除合集
- [ ] 视频绑定 / 解绑

### H27 / ActorManage

**外观锚点**：
- [ ] 顶部 `PageHeader` + `Toolbar`（搜索 + 创建演员）
- [ ] 演员列表清空时 `EmptyState`
- [ ] Scoped CSS 不含 `#9ca3af` / `#6b7280` / `#4b5563`

**功能不破**：
- [ ] 创建 / 编辑 / 删除演员
- [ ] 搜索过滤
- [ ] 头像上传

### H28 / 全局自动化

- [ ] `cd admin-web && npm run build` 通过
- [ ] `cd admin-web && npm test` 全绿（≥ 75 用例）
- [ ] `themeTokens.spec.js` 视图层 audit 8 个用例全绿

## 3. 截图归档

`tasks/2026-05-23-admin-simple-views/screenshots/` 必须包含：

- `before-dashboard.png` / `after-dashboard.png`
- `before-system-settings.png` / `after-system-settings.png`
- `before-user-manage.png` / `after-user-manage.png`
- `before-task-monitor.png` / `after-task-monitor.png`
- `before-iptv-manage.png` / `after-iptv-manage.png`
- `before-collection-manage.png` / `after-collection-manage.png`
- `before-actor-manage.png` / `after-actor-manage.png`

共 14 张，统一 1440px 桌面分辨率。

after 必须显示：
- PageHeader 在顶部、字号字重符合 Phase 1 token
- 应接入的 wrapper（Toolbar / StatCard / SectionCard / EmptyState）肉眼可见
- 整页色调与 Phase 1 外壳一致，无残留玫红描边 / 玫红色文字 / 玫红按钮

不做像素级 diff；目视判断「视觉已切到新设计系统」即可通过。

## 4. Done definition

- [ ] 第 1 节自动化全绿
- [ ] 第 1.1 节 spec 扩展核查全过
- [ ] 第 1.2 节既有测试零回归
- [ ] 第 2 节手测 H21–H28 全部通过
- [ ] 第 3 节 14 张截图归档完成
- [ ] `plan.md` 顶部已追加 Phase 2 实施条目（implement.md 第 5 节模板）
- [ ] 全部改动落在 1 commit（中文 subject + body）；如有 code review followup 走「实现 + followup」二次 commit
- [ ] `tasks/2026-05-23-admin-simple-views/DONE.md` 含完成日期 + commit hash + 验证摘要

## 5. review 后流程

review 通过 → 用户多视图点查一轮 → 新增 `DONE.md` → 进入 Phase 3（`tasks/2026-05-23-admin-medium-views/`）。

## 6. review 失败处理

- **自动化红**：回 Implement 修复，不得 skip 或注释掉用例
- **手测外观锚点不达预期**：判断是 wrapper 不够（回 Phase 1 followup commit 补 wrapper API）还是接入错（回 Implement 改视图）
- **手测功能不破红**：**最高优先级**回退或修复——Phase 2 是视觉重排不允许引入功能回归；若回归不能立刻定位，先把对应视图回滚到 Phase 1 状态，再单独 follow-up
- **截图缺失或不达预期**：补截图；不影响代码已合入的事实
- **scope creep**：如发现需要改 Phase 3 / Phase 4 视图、新增 wrapper、改 IA、改 store / API，**严格拒绝**进入本期，记入 plan.md 下一期 backlog
