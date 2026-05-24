# Review：admin 重设计 Phase 1 · 设计系统与外壳

- 日期：2026-05-23
- 关联 PRD：`prd.md`
- 关联 Implement：`implement.md`

## 1. 自动化必跑

```bash
cd admin-web
npm run build       # AGENTS.md 要求的必跑检查
npm test            # 既有 helpers spec + 新增 themeTokens / Layout / commandPalette spec
```

任一红则 review 失败，回 Implement 阶段修复，不得 skip 测试。

### 新增 spec 覆盖核查

| spec 文件 | 必须包含的断言组 |
|----------|----------------|
| `assets/themeTokens.spec.js` | 核心 token 存在（`--primary` / `--text-primary` / `--bg-canvas` / `--bg-sidebar`）；6 档字号；旧玫红色值 `#881337` / `#be123c` / `#7f1d1d` 零残留；`Fira Code` / `Fira Sans` / `--font-code` 零残留 |
| `components/Layout.spec.js`（源文 audit） | 含「分组」/「媒体库」/「录入处理」/「系统」字面量；含 `admin-sidebar-collapsed`（localStorage key）；含 `profile` 与 `CommandPalette` 调用点；不含 `linear-gradient(180deg, #881337`、不含 `'管理员工作台'` 字面量 |
| `components/base/commandPalette.helpers.spec.js` | 真值表：`'上传' → 「上传视频」`、`'sj' → 「视频管理」`（拼音首字母）、`'/videos' → 「视频管理」`（路径前缀）、`'upload' → 「上传视频」`（英文 alias）、`''` 返回所有项 score=0 |

## 2. 源文 audit 断言点

| # | 断言（出现 = 通过） | 文件 |
|---|---------------------|------|
| 1 | 含 `var(--primary)`、`var(--bg-sidebar)`、`var(--text-primary)`、`var(--font-sans)` | `Layout.vue` |
| 2 | 不含裸 hex 字面量（除 `theme.css` 本身） | `Layout.vue` / 7 个 base 组件 / `Login.vue` |
| 3 | 不含裸 px 字面量用于间距 / 圆角 / 阴影（除 `theme.css` 与 `element-overrides.css` 自身） | 同上 |
| 4 | 含 `localStorage.getItem('admin-sidebar-collapsed')` 与 `setItem` | `Layout.vue` |
| 5 | 含 `useEventListener` 或等价的 keydown 监听 `KeyK` + `metaKey || ctrlKey` | `CommandPalette.vue` |
| 6 | 含 `tabular-num` class 或 `font-variant-numeric: tabular-nums` | `StatCard.vue` |
| 7 | 含 Element Plus icon 引用（`@element-plus/icons-vue`） | `Layout.vue` / `EmptyState.vue` |

## 3. 手测脚本

**预条件**：`cd admin-web && npm run dev`，浏览器登录 admin。

| 用例 | 路径 | 期望 |
|------|------|------|
| H1 | 进入仪表盘 | 侧栏 slate-100 + 5 组分组标题（uppercase 小字、`--text-caption`）+ 仪表盘高亮主色 + 底部 profile chip |
| H2 | 浏览器宽度依次拖至 1440 / 1200 / 900 / 600 | 1440 完整 240px 侧栏；1200 自动折叠到 60px（icon-only + tooltip 显示菜单名）；900 侧栏改 drawer 浮层；600 主区水平滚动但不破版 |
| H2.1 | 1440 下手动折叠侧栏，刷新 | 重新加载后保持折叠（localStorage 持久化）|
| H3 | 按 `⌘K`（mac）或 `Ctrl+K`（win/linux） | 命令面板居中弹出 + 输入框抢焦点 |
| H3.1 | 输入「上传」 | 命中「上传视频」，按回车跳转 `/upload` |
| H3.2 | 输入「sj」 | 命中含「视」字开头的菜单（视频管理 / 视频上传 / 视频…）|
| H3.3 | 输入「/iptv」 | 命中「IPTV 管理」 |
| H3.4 | 按 Esc | 命令面板关闭 |
| H4 | 点击侧栏底部 profile chip | 弹出 popper 显示用户名 + admin 徽标 + 退出登录 |
| H4.1 | popper 内点击退出登录 | 跳回登录页 |
| H5 | 任一既有视图（如 Dashboard / VideoList）的 PageHeader | 标题 20px semibold，无副标题 "管理员工作台"，右侧 actions slot 对齐（Phase 1 内可暂时空 slot） |
| H6 | 在新增的 demo 路由（或直接挂在 Dashboard）渲染 EmptyState | 图标 + 主文案 + 副文案，不再裸 "暂无数据" |
| H13 | 任意视图弹窗 / 抽屉 / 按钮 / 输入框 / 标签 / 分页 / Message / Notification | 视觉跟随 token：圆角 `--radius-*`、阴影 `--shadow-*`、字体 Inter、间距 `--space-*` |
| H14 | DevTools Console 执行 `document.documentElement.dataset.theme = 'dark'` | 整站切到暗色（验证 token 占位生效）；UI 不暴露切换按钮 |
| H15 | 浏览器拖到 < 1024px，点顶栏左侧菜单按钮 | mobile drawer 唤起；点任一菜单项跳转后 drawer 自动关闭 |

## 4. 截图归档

`tasks/2026-05-23-admin-shell-redesign/screenshots/` 必须包含：

- `before-login.png` / `after-login.png`
- `before-layout-1440.png` / `after-layout-1440.png`
- `before-layout-1024.png` / `after-layout-1024.png`
- `before-layout-768.png` / `after-layout-768.png`
- `after-command-palette.png`
- `after-profile-popper.png`
- `after-dark-token-demo.png`

review 不强制比对像素差异；目视判断「视觉语言已彻底切换」即可通过。

## 5. Done definition

- [ ] 第 1 节自动化全绿
- [ ] 第 2 节源文 audit 全部命中
- [ ] 第 3 节手测 H1–H6 / H13 / H14 / H15 全部通过
- [ ] 第 4 节截图归档完成
- [ ] `CONTEXT.md` 已追加 PRD 第 5 节列出的 6 条新术语（在「admin 设计系统术语」新区块）
- [ ] `plan.md` 顶部已追加本期条目（implement.md 第 3 节模板）
- [ ] 全部改动落在**一个** git commit（中文 subject + 中文 body）

## 6. review 后流程

review 通过 → 用户在 TV / phone / desktop 多设备点查一轮 → 新增 `DONE.md` 含完成日期 / 关联 commit hash / 验证摘要 → 进入 Phase 2。
