# Phase 1 完成记录 · admin 外壳重设计

- 完成日期：2026-05-24
- 关联提交：
  - `12bf7c93` 完成 admin 外壳重设计一期（首版实现）
  - `5ae43fa5` 修复 admin 外壳一期 code review 反馈（10 项审查发现的 7 项必修/强烈建议）

## 验证摘要

### 自动化必跑
- `cd admin-web && npm test` —— 12 测试文件 / 67 用例全绿
- `cd admin-web && npm run build` —— 通过（vite 6.4.2，2272 modules，dist 397KB CSS / 2.3MB JS）
- 新增 spec 三件套（`themeTokens.spec.js` / `Layout.spec.js` / `commandPalette.helpers.spec.js`）全部命中断言

### 源文 audit
- `theme.css` 删除全部旧玫红 hex（`#881337` / `#be123c` / `#7f1d1d`）与 `Fira Code` / `Fira Sans` / `--font-code` 字体引用
- `Layout.vue` 不含 `linear-gradient(180deg, #881337` 与 `'管理员工作台'` 副标题字面量
- 7 个 `components/base/*` 全部使用 token，零裸 hex

### 手测覆盖
- H1（侧栏 slate-100 + 5 组分组 + profile chip）
- H2 + H2.1（1440 / 1024 / 768 三档断点 + 折叠偏好持久化）
- H3.x（`⌘K` / `Ctrl+K` 命令面板 + 拼音 / 别名 / 路径 / Esc 关闭）
- H4 + H4.1（profile popper + 退出登录）
- H5（PageHeader 应用于 Login，无 `管理员工作台` 副标题）
- H6（EmptyState 形态可用）
- H13（Element Plus 弹窗 / 按钮 / 输入 / 表格 token 跟随）
- H14（手动 `data-theme="dark"` 整站切暗）
- H15（< 1024px mobile drawer）

### 截图归档（`screenshots/`）
- before / after 11 张 PNG：
  - login（before / after）
  - layout 三档断点 1440 / 1024 / 768（before / after × 3）
  - command palette（after）
  - profile popper（after）
  - dark token demo（after）

### CONTEXT.md 沉淀
- 「admin 设计系统术语」新区块共 6 条：
  - `admin Modern Minimal`
  - `admin 设计 token`
  - `Element Plus 三层架构`
  - `admin 侧栏分组导航`
  - `admin 侧栏折叠偏好`
  - `admin 命令面板`

### plan.md 追加条目
- 2026-05-23 22:30 +0800 Phase 1 实施记录
- 2026-05-24 实施期间随实施补的代码 review 反馈条目

## Code review followup 落地清单

| 等级 | 修复 | 状态 |
|------|------|------|
| 必修 | Layout `loadProfile` 401 / 非 401 分类处理 | ✅（401 已由 axios 拦截器 handleAuthExpired 跳登录；非 401 加 `console.warn`） |
| 必修 | Layout `localStorage` SecurityError / QuotaExceededError 兜底 | ✅ try/catch 包裹 read + write |
| 必修 | theme.css `--surface-muted` token 孤儿 | ✅ 加 alias `--surface-muted: var(--bg-surface-muted)` |
| 强烈建议 | CommandPalette `⌘K` 在 input / IME / repeat 时不抢键 | ✅ event.target + isComposing + repeat 三重守卫 |
| 强烈建议 | mobile drawer body / html `overflow: hidden` scroll-lock 还原 | ✅ watch + onUnmounted 兜底 |
| 强烈建议 | 侧栏激活 RouterLink `aria-current="page"` | ✅（WCAG 2.4.8 Location 满足） |
| 强烈建议 | CommandPalette `Tab` / `Shift+Tab` focus trap + Esc 提到 document 级 + 关闭还原焦点 | ✅ |
| 可选 | PINYIN_INITIAL 字典 `频: 'j' → 'p'` 修正 + 补 5 字 | ✅ |
| 可选 | `orientationchange` 监听（Android 转屏） | ✅ |
| 可选 | Dashboard 与 `#app :where()` 全局 specificity 兼容 | ⏸️ 留 Phase 2（涉及 13 处 `#app` 前缀重构、本期 Dashboard 视觉未报回归） |

## 已知遗留（移交 Phase 2）

- 14 个业务视图（含 Phase 2 域内的 Dashboard / SystemSettings / UserManage / TaskMonitor / IPTVManage / CollectionManage / ActorManage）的 scoped CSS 内仍有硬编码颜色 hex / token 违规，需在 Phase 2 视图重排时一并清理
- `components/UploadProgress.vue`（共享组件，被 VideoUpload 等多视图使用）也含 `color: #7f1d1d` 一处，需要在它被实际复用的视图重排时一并清理
- `#app :where()` 全局选择器 specificity（13 处 `#app` 前缀规则）与 scoped CSS 的优先级冲突的 specificity 收口
