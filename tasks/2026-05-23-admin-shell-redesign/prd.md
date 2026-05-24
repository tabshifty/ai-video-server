# PRD：admin 重设计 Phase 1 · 设计系统与外壳

- 日期：2026-05-23
- 阶段：Phase 1 / 4（设计系统与外壳）
- 目标端：`admin-web/`（Vue 3 + Vite + Element Plus 管理端）
- 范围：全局 token / Element Plus 覆写 / Layout 外壳 / 7 个共享 wrapper 组件 / Login / 字体栈

## 1. 用户故事

作为管理员，登录 admin 后台时希望：

1. **第一眼有审美**——侧栏不再是玫红→蓝渐变 + 红色标题字与蓝色品牌色三方冲突；整站统一冷蓝主色 + 浅灰中性 + 大留白
2. **导航不再扁平**——14 个菜单分 5 组，浏览时一眼就知道当前在哪个功能区
3. **窄屏自动让位**——1280px 以下侧栏自动折叠到 60px icon-only 模式；偏好持久化
4. **快速跳页**——按 `⌘K` / `Ctrl+K` 唤起命令面板，输入中文 / 拼音首字母 / 英文路径都能跳
5. **轻量身份感**——侧栏底部固定 profile chip，显示当前账号与 admin 徽标，点开退出登录
6. **登录页不再是 Element Plus 默认皮肤**——居中卡片，与整站设计语言一致

## 2. 作用域

| 改 | 不改 |
|----|------|
| `admin-web/src/assets/theme.css` 全文重写 | 路由结构 / URL |
| 新增 `admin-web/src/assets/element-overrides.css` | Pinia store 形状 |
| `admin-web/src/components/Layout.vue` 全文重写 | API 接口 / 后端契约 |
| 新增 `admin-web/src/components/base/PageHeader.vue` | 14 个业务视图内部（Phase 2-4） |
| 新增 `admin-web/src/components/base/Toolbar.vue` | dark mode UI 切换 |
| 新增 `admin-web/src/components/base/EmptyState.vue` | Android TV / phone App |
| 新增 `admin-web/src/components/base/StatCard.vue` | Go 后端任何代码 |
| 新增 `admin-web/src/components/base/SectionCard.vue` | 既有 `AdminTablePagination.vue` 接口 |
| 新增 `admin-web/src/components/base/CommandPalette.vue` | 既有 `UploadProgress.vue` 接口 |
| 新增 `admin-web/src/components/base/BulkActionBar.vue` | |
| 新增 `admin-web/src/components/base/commandPalette.helpers.js` + `.spec.js` | |
| `admin-web/src/views/Login.vue` 全文重写 | |
| `admin-web/src/main.js` 引入 overrides | |
| `admin-web/index.html` 字体 preconnect 优化 | |
| 新增 `admin-web/src/assets/themeTokens.spec.js` | |
| 新增 `admin-web/src/components/Layout.spec.js`（源文 audit） | |

## 3. 非目标

- N1：本期**不动** 14 个业务视图内部（Dashboard / VideoList / ImageManage / ... 全部留给 Phase 2-4）；这些视图的 token 自动跟随，视觉会"被动统一"但 IA 不改
- N2：本期**不接 dark mode UI 切换**——`theme.css` 暴露 `:root[data-theme="dark"]` 占位变量供未来使用，但不暴露切换按钮
- N3：本期**不引入新框架**（Tailwind / UnoCSS / Naive UI / shadcn-vue 全不引）
- N4：本期**不替换 Element Plus**——继续使用，仅做 token 覆盖 + 单文件全局覆写
- N5：本期**不重写 `AdminTablePagination.vue` / `UploadProgress.vue`**——保留接口与逻辑，仅样式跟随 token 自动改变
- N6：本期**不引入 Playwright / Percy / Chromatic**——视觉验收靠人肉点查 + 截图归档
- N7：本期**不引入手机端专门 UX**——< 1024px 仅保证不破版可读，不专门优化触摸体验
- N8：本期**不在 cmd+K 命令面板里搜视频内容**——仅做导航搜索
- N9：本期**不接入 `cmdk` / `vue-command-palette` 等三方包**——原生 + 50 行实现
- N10：本期**不引入中文 webfont**（思源黑体 / 阿里巴巴普惠体等）——加载代价不值

## 4. Phase 1 关键决策结论（grill-with-docs 结果）

| # | 决策点 | 选项 |
|---|--------|------|
| Q1 | 改造范围 | Phase 1: 外壳 + token + Login + wrapper；Phase 2-4: 业务视图 |
| Q2 | 重排语义 | 视觉皮肤 + 三巨头 IA 改造（在 Phase 4） |
| Q3 | 设计语言 | Modern Minimal，冷蓝主色，light 优先 |
| Q4 | 导航分组 | 5 组（仪表盘单列 / 媒体库 / 录入处理 / 服务 / 系统） |
| Q5 | 外壳附加功能 | 侧栏可折叠 + cmd+K 命令面板 + 侧栏 profile，删顶栏副标题与顶栏退出登录 |
| Q6 | 字体 | Inter + PingFang SC + 系统 mono；root 14px；6 档字号；删 Fira Code/Sans |
| Q7 | Element Plus 改造 | L1 token + L2 单文件覆写 + L3 七个 wrapper |
| Q8 | 主色与色板 | Tailwind blue-600 #2563eb 主色；slate / blue / emerald / amber / red / sky 色板 |
| Q9 | 侧栏底色 | slate-100 浅灰侧栏 + 白色内容 + slate-200 1px 分隔 |
| Q10 | 三巨头 IA | Phase 4 处理（VideoList drawer / ImageManage 网格 / VideoUpload 3 步 wizard） |
| Q11 | 响应式 | desktop-first 4 断点（768/1024/1280/1536）；< 1024 sidebar 改 drawer；< 768 仅"不破版" |
| Q12 | 验收 | npm run build + npm test + 人肉点查 + 前后截图 |
| Q13 | 推进节奏 | 4 个阶段，本期 Phase 1 是地基 |

## 5. 拟新增的 `CONTEXT.md` 术语（实施阶段一并落入「admin 设计系统术语」新区块）

| 术语 | 一句话定义 |
|------|----------|
| `admin Modern Minimal` | admin-web 整站视觉语言：中性 slate 打底 + 单一冷蓝主色 + 大留白 + 仅 elevation 表达层次；与 TV 端「夜台玻璃面板」语言解耦 |
| `admin 设计 token` | `admin-web/src/assets/theme.css` 中集中导出的颜色/字体/间距/圆角/阴影/动效 CSS vars，是所有 admin 视图视觉的单一来源；调用点不允许出现裸 hex 与裸 px 字面量 |
| `Element Plus 三层架构` | L1=token 覆盖（`--el-*` CSS var）；L2=`element-overrides.css` 全局 `:where()` 覆写；L3=`components/base/` 内 7 个 wrapper 组件。三层各司其职，禁止跨层乱写 |
| `admin 侧栏分组导航` | 14 个菜单按 5 个分组（仪表盘 / 媒体库 / 录入处理 / 服务 / 系统）展示，每组有 11–12px uppercase 小字标题，组间不画 divider |
| `admin 侧栏折叠偏好` | 侧栏 240px 展开与 60px 折叠两态由用户决定，并写入 `localStorage`；窗口 ≤ 1279px 自动折叠（但仍尊重用户既有偏好） |
| `admin 命令面板` | 通过 `⌘K` / `Ctrl+K` / 顶栏按钮唤起，仅做导航搜索；匹配菜单中文 label、拼音首字母、英文 alias、路径前缀；与视频内容搜索严格隔离 |

## 6. 验收（Phase 1 子集，对齐 grill Q12 列出的 H 用例）

| 用例 | 路径 | 期望 |
|------|------|------|
| H1 | 进入任意页面 | 侧栏 slate-100 浅灰 + 5 个分组标题（uppercase 小字）+ 当前页高亮主色 + 底部 profile chip |
| H2 | 浏览器宽度拖拽 1440 → 1200 → 900 → 600 | 1200 自动折叠侧栏到 60px；900 侧栏改 drawer；600 表格水平滚动；任何阶段无破版、无 body 横向滚动 |
| H3 | `⌘K` / `Ctrl+K` 或顶栏按钮唤起命令面板 | input 抢焦点；输入 "上传" / "sj" / "/videos" 任一形式都命中；回车跳转；Esc 关闭 |
| H4 | 侧栏底部 profile 点击 | popper 显示用户名 + admin 徽标 + 退出登录 |
| H5 | 各页面顶部 PageHeader（先在 Login + 任一既有视图验证） | 标题 20px semibold，无"管理员工作台"副标题，右侧 actions slot 对齐 |
| H6 | 任意页面通过 EmptyState 组件渲染空态（先在新增 demo 路由验证） | 图标 + 主文案 + 副文案，不再裸"暂无数据" |
| H13 | Element Plus 弹窗/抽屉/按钮/输入框/标签/分页/Message/Notification 视觉统一 | 圆角 / 阴影 / 字体 / 间距全部跟随 token；无残留 Element Plus 默认皮肤 |
| H14 | 手动 `<html data-theme="dark">` | 切换到暗色（验证 token 占位生效）；UI 不暴露切换 |
| H15 | < 1024px 顶栏按钮唤起 mobile drawer | 唤起 + 菜单项点击跳转 + drawer 自动关闭 |

### 自动化必跑

```bash
cd admin-web
npm run build                                   # 必须通过
npm test                                        # 既有 helpers spec 全绿 + 新增三处 spec 全绿
```

新增的自动化 spec：

- `commandPalette.helpers.spec.js`：命令面板模糊匹配真值表
- `themeTokens.spec.js`：theme.css 文本断言核心 token 存在 + 旧玫红色值（`#881337` / `#7f1d1d` / `#be123c`）零残留
- `Layout.spec.js`：源文 audit，含 `<aside`、`分组标题`、`profile`、`命令面板触发`、新 token 引用；不含 `linear-gradient(180deg, #881337`

### 截图归档

- Phase 1 落地后，在 `tasks/2026-05-23-admin-shell-redesign/screenshots/` 目录下提交以下 PNG：
  - `before-login.png` / `after-login.png`
  - `before-layout-1440.png` / `after-layout-1440.png`
  - `before-layout-1024.png` / `after-layout-1024.png`
  - `before-layout-768.png` / `after-layout-768.png`
  - `after-command-palette.png`
  - `after-profile-popper.png`
  - `after-dark-token-demo.png`（手动 `data-theme=dark`）
