# PRD：TV 长视频播放器操作 UI 跟随遥控器互动

- 日期：2026-05-24
- 目标端：`android-tv-app/`（TV App，独立工程，与 `android-app/` 源码隔离）
- 范围：TV 长视频播放器的"操作 UI"层（`core/ui/LongFormVideoPlayer.kt`，调用方 `feature/tv/TvLongFormPlayerScreen.kt`、`feature/tv/TvSeriesPlayerScreen.kt`），**仅** `tvMode = true` 路径

## 1. 用户故事

作为 TV 用户，我在长视频播放（电影 / 电视剧 / `18+`）过程中希望：

1. **遥控器一互动就看到操作 UI**：任意遥控器键事件都把"操作 UI"层（底部 controls + 左上信息层）唤起，不需要先盲操作再猜按钮在哪。
2. **互动时 5 秒倒计时重置**：每一次遥控器互动都把 5 秒自动隐藏的计时器重置，按按钮、左右切焦点也算互动。
3. **左右键能在控制按钮之间循环**：焦点进入底部控制条后，左/右按到尽头自动绕回另一端，避免"再按一次没反应"的死角。
4. **左上角能看到我在看什么**：左上角始终能看到主标题；电视剧追加"第 X 季 · 第 Y 集 [单集标题]"副行，5 秒不动随 controls 一起隐藏。
5. **BACK 键先收 UI 再退出**：UI 亮着时按 BACK 立即收 UI（不等 5 秒、不进退出确认）；UI 已收时再按 BACK 才进 [[TV 播放器退出确认]] 的"两段式提示"。

## 2. 作用域

- **仅** `tvMode = true` 路径下的 `LongFormVideoPlayer`：覆盖 `TvLongFormPlayerScreen`（电影 / `18+`）和 `TvSeriesPlayerScreen`（电视剧）
- **仅** 操作 UI 层（底部 controls 玻璃条 + 左上信息层）+ 遥控器按键路由 + 5 秒自动隐藏计时
- **不** 改动 `tvMode = false` 路径（`UnifiedPlayerScreen` 等手机端短视频全屏），触屏 tap/拖拽/长按交互完全保留
- **不** 改动 IPTV 播放（独立 LibVLC 播放器，非 `LongFormVideoPlayer`）
- **不** 改动现有 `控制条 → 字幕/音轨夜台玻璃面板`、`续播提示卡`、`连播提示卡` 的视觉与交互
- **不** 改动 `LongFormVideoPlayer` 的非 TV 模式公共 API 形状（仅新增可选 props）

## 3. 非目标（明确不做）

- N1：手机短视频全屏路径同步采用左上信息层——保持现状玻璃顶栏，避免触动 `CONTEXT.md` line 63–66 短视频全屏 surface 归属约束
- N2：focus 进入控制条后，OK 在「播放/暂停」之外按钮上的行为重写——沿用 Compose 默认 `onClick`，每个按钮原有 onClick 即触发
- N3：Slider scrub 能力保留——本期把 Slider 从焦点链剔除，scrub 通过左右键 seek + 快退/快进按钮承载
- N4：5 秒计时长度可配置——固定 5 秒（与 `CONTEXT.md` 现有 `scheduleAutoHideControls` 行为一致）
- N5：长按左右键的"3 倍步长"快速跳转语义改动——沿用 `CONTEXT.md` 的 `快进/快退步长` 与 `连按合并跳转`
- N6：增加新的"全局开关"或"用户偏好"——本任务无任何持久化偏好
- N7：UI 收起动画时长改动——继续走 `TvMotionTokens.DurationStandardMs` + `EasingStandard`
- N8：左上信息层位置可配置（左上 / 居中 / 右上）——固定左上
- N9：底部 controls 横幅玻璃面板视觉重做——保留现有 `Color(0x910D1016)` + `AppChrome.SurfaceShape` 玻璃条
- N10：把"操作 UI 层"从 `LongFormVideoPlayer` 内部拆出去——本期不重构组件边界

## 4. 术语对应（详见 `CONTEXT.md`「TV 播放术语」区）

本任务一并沉淀以下新术语到 `CONTEXT.md`：

| 术语 | 一句话定义 |
|------|----------|
| `TV 操作 UI 层` | 长视频 TV 模式下的"操作可见性单元"：底部 controls 玻璃条 + 左上信息层。两者由同一个 `controlsVisible` 状态驱动，同一个 5 秒计时控制，同一组 240ms tween fade 动效。 |
| `左上信息层` | 左上角无背板的纯文字标题层：主行 = 剧名 / 电影标题，电视剧追加副行「第 X 季 · 第 Y 集 单集标题」。文字带阴影保证亮场景对比度，可见性与 [[TV 操作 UI 层]] 共生。 |
| `操作 UI 互动唤起` | 任意遥控器键事件（含 LEFT/RIGHT seek、DOWN 进入焦点、OK 切换播放暂停）都唤起 [[TV 操作 UI 层]] 并重置 5 秒计时；UP / 在控制条内左右切焦点也算互动重置。 |
| `controls 焦点入口` | TV 模式下，**仅** DPad DOWN 键能把焦点送入底部控制条；其它键（LEFT / RIGHT / OK / UP）唤起 UI 时焦点保持在播放器根。这是与 phone-era "controls 一出现就抢焦点" 模式的显式分离。 |
| `controls 焦点环绕` | TV 控制条焦点链按可见按钮的顺序首尾相连：在最左侧按钮按 LEFT 跳到最右侧按钮，反之亦然。Slider 不参与焦点链。 |
| `BACK 优先收 UI` | TV 长视频播放器中 BACK 键的两段式语义：[[TV 操作 UI 层]] 可见时优先收 UI（不等 5 秒、不进退出确认）；UI 已收起时按 BACK 才进 [[TV 播放器退出确认]]。这是对 [[TV 播放器退出确认]] 的前置例外，不破坏其"两段式提示"主流程。 |
| `controls 焦点退出键` | 焦点在 [[TV 操作 UI 层]] 内时，UP 键把焦点退回播放器根；UI 不被 UP 收起，5 秒计时继续。其它"焦点离开 controls"的方式：点 BACK（收 UI）/ 5 秒未互动自动隐藏。 |

## 5. 关键决策表（grill-with-docs 八问的结果）

| # | 决策点 | 选项 | 理由 |
|---|--------|------|------|
| Q1 | controls 可见时左右键语义 | **B：焦点环绕** | 用户原话"按左右键循环"指焦点几何而非 seek 持续；Slider 中间夹着如果焦点穿过会触发 scrub，需把 Slider 剔除焦点链（见 Q6） |
| Q2 | controls 隐藏时左右键 + UI 唤起耦合 | **「seek + 亮 UI，焦点不进入」**（用户修订版本，非推荐版本） | 用户明确要求"只有按下键才能进入 controls 的焦点"——LEFT/RIGHT 同时承担 seek 与点亮职责，但**不**送焦点；这把"UI 可见"和"焦点是否在 controls"严格解耦 |
| Q3 | UP / OK / BACK 语义 | **UP = 焦点退出 / OK = 焦点不在 controls 时切换播放暂停 / BACK = UI 可见时收 UI**（用户对 BACK 修订） | BACK 修订把 [[TV 播放器退出确认]] 调整为"UI 已收起后才进入"，是显式扩展，需同步写进 `CONTEXT.md` |
| Q4 | 左上信息层视觉 | **布局 ③：左上角无背板纯文字** | 用户明确选 ③；与控制条玻璃条视觉解耦，但**显示/隐藏时机与动效**继续与 controls 同步 |
| Q5a | 文本内容组合 | **组合 ①：两行（主：剧名 / 副：第 X 季 · 第 Y 集 单集标题）** | 剧名是 10-foot 视距下用户最依赖的锚点；季集 + 单集标题走副行，被 ellipsis 不影响主判断 |
| Q5b | 可读性处理 | **处理 ①：文字阴影（Compose 原生 `Shadow`）** | 跨亮 / 暗背景自适应；无新视觉资产；与 `LongFormVideoPlayer` 现有 `centerFeedbackText` 风格自洽 |
| Q6 | Slider 是否参与焦点链 | **走法 ①：Slider 不可聚焦，焦点环绕只走按钮** | 控制条已有快退 / 快进按钮覆盖 ±10s seek；TV 遥控器粒度对 scrub 不友好；不剔除会破坏焦点环绕承诺 |
| Q7 | 首次进入视频的初始状态 | **走法 ②：自动亮 controls + 左上信息层 5 秒，焦点停留在播放器根** | 提供 affordance（让用户知道有可操作 UI）；同时严守"只有 ↓ 送焦点"承诺 |
| Q8 | 改动范围 | **走法 ①：仅 TV 模式** | 用户原话"遥控器互动"锁定 TV 语境；非 TV 模式触屏交互与 surface 归属不动 |

## 6. 验收标准（QA scenarios）

### 6.1 任意遥控器互动唤起操作 UI（PRD §1.1）

进入任一 TV 长视频（电影 / `18+` / 电视剧任一集）播放器，等首次进入的 5 秒自动隐藏完成（[[TV 操作 UI 层]] 收起）。

- **A1**：按 ← 一次
  - ✓ 视频后退 10 秒（或用户设置的 `快进/快退步长`）
  - ✓ 底部控制条玻璃条出现 + 左上信息层出现，焦点**不**进入控制条（依旧停在播放器根）
  - ✓ 5 秒计时从 0 开始
- **A2**：按 → 一次
  - ✓ 视频前进同样步长
  - ✓ 操作 UI 出现，焦点不进入
- **A3**：按 ↓ 一次
  - ✓ 操作 UI 出现
  - ✓ 焦点**进入**控制条，落在「播放/暂停」按钮（焦点蓝青色光感可见）
- **A4**：按 OK 一次
  - ✓ 播放暂停切换，中央反馈文字「已暂停」/「继续播放」出现
  - ✓ 操作 UI **不**额外亮起（OK 不属于"互动唤起"，由中央反馈层独立承载视觉）

> 注：A4 是 Q3 决议的语义——OK 在焦点不在 controls 时承担"快捷切换播放暂停"，由中央 toast 承载视觉，不重复亮整套 UI。

### 6.2 控制条焦点入口与左右键焦点环绕（PRD §1.3）

接 A3 完成（焦点已落在「播放/暂停」按钮，UI 可见）。

- **B1**：按 → 一次
  - ✓ 焦点移到下一个按钮（电视剧场景下一般是「快退」）；操作 UI 继续可见，5 秒计时重置
- **B2**：连续按 → 直到焦点到最右侧按钮（典型电视剧场景下一路走到「退出播放」按钮）
  - ✓ 一直能穿过所有可见按钮，不被 Slider 拦截（Slider 不获得焦点）
- **B3**：在最右侧按钮上再按 → 一次
  - ✓ 焦点**绕回**最左侧按钮（「播放/暂停」）
- **B4**：在「播放/暂停」按钮上按 ← 一次
  - ✓ 焦点**绕回**最右侧按钮
- **B5**：在 B3 状态下按 OK
  - ✓ 该按钮原有 onClick 触发（如「退出播放」会调 `onRequestExitPlayback ?: onExitPlayback`，再走「TV 播放器退出确认」流程）

### 6.3 5 秒计时重置（PRD §1.2）

接 B1 完成。

- **C1**：在焦点已进入 controls 状态下，每 4 秒按 → 一次（在按钮间环绕）
  - ✓ 操作 UI 始终不消失（焦点切换重置计时）
- **C2**：停止操作，开始数秒
  - ✓ 5 秒后操作 UI 完整淡出（240ms fade）；焦点在 controls 消失瞬间**回到**播放器根，整个隐藏过程不留焦点光感残影
- **C3**：在 controls 可见、焦点未进入状态下（接 A1），4 秒内连续按 ← 三次
  - ✓ seek 累计触发（走 [[连按合并跳转]] 的 300ms 防抖合并）；操作 UI 全程可见，每次按键重置 5 秒计时

### 6.4 左上信息层（PRD §1.4）

- **D1**：电影播放器
  - ✓ UI 可见时左上角显示主行：电影标题（白色，带 `Shadow(Color.Black, offset=(0,2), blurRadius=4f)`），**无**副行
  - ✓ UI 隐藏时左上信息层完全淡出（与 controls 同步，同 240ms tween）
- **D2**：`18+` 长视频播放器
  - ✓ 同电影：主行 = `detail.title`，无副行
- **D3**：电视剧 S1E7 播放器（假设单集标题为「朱元璋登基」）
  - ✓ 主行 = `series.title`（如「明朝那些事」）
  - ✓ 副行 = `第 1 季 · 第 7 集 朱元璋登基`，比主行字号略小、alpha 略低（约 0.7）
  - ✓ 副行受 [[TV 文本溢出保护]] 约束：`maxLines = 1` + ellipsis
- **D4**：电视剧切到下一集（如 [[电视剧自动连播]] 触发或手动「下一集」）
  - ✓ 左上信息层副行自动更新为新集号 + 新单集标题
- **D5**：电视剧单集标题为空字符串或 null
  - ✓ 副行仅显示「第 1 季 · 第 7 集」，无尾部空格
- **D6**：长背景画面（如雪景纯白镜头）
  - ✓ 文字仍可读（阴影提供对比度）；不出现"白底白字"完全消失

### 6.5 BACK 优先收 UI（PRD §1.5）

进入任一 TV 长视频播放器。

- **E1**：[[TV 操作 UI 层]] 可见（不论焦点在不在 controls），按 BACK 一次
  - ✓ 操作 UI 立即收起（同步、不等 240ms 完整动画？至少视觉上立即触发收起动画）
  - ✓ **不**进入「TV 播放器退出确认」流程
  - ✓ 焦点回到播放器根
- **E2**：UI 已收起，按 BACK 一次
  - ✓ 进入 [[TV 播放器退出确认]] 第一次提示
- **E3**：E2 状态下再按 BACK 一次（5 秒内）
  - ✓ 真正退出播放器（沿用 [[TV 播放器退出确认]] 现有第二次确认行为）
- **E4**：控制条上的「返回详情」/「退出播放」按钮（点击触发，不是 BACK 键）
  - ✓ 继续走 [[TV 播放器退出确认]]（不被 BACK 优先收 UI 影响）

### 6.6 首次进入视频的初始状态（PRD Q7）

冷启进入任一 TV 长视频播放器。

- **F1**：进入瞬间
  - ✓ [[TV 操作 UI 层]] 自动亮起（5 秒）
  - ✓ 焦点**停留在播放器根**，**不**进入控制条
- **F2**：5 秒内不操作
  - ✓ 操作 UI 自动隐藏，焦点保持在播放器根
- **F3**：5 秒内按 ↓ 一次
  - ✓ 焦点进入「播放/暂停」按钮
- **F4**：5 秒内按 ← 或 →
  - ✓ seek 触发 + 5 秒计时重置（焦点仍在播放器根）

### 6.7 与既有特性不冲突

- **G1**：[[续播提示卡]] 显示期间，按 ↓ 不应被 UI 收起干扰，应按现有 [[续播提示卡]] 焦点契约由 `TvResumePromptCard` 处理；左上信息层与续播提示卡同时可见时不重叠
- **G2**：[[连播提示卡]] 显示期间（电视剧片尾 T-10），按 ↓ 仍按现有契约处理，左上信息层与连播提示卡同时可见时不重叠
- **G3**：字幕 / 音轨 [[夜台玻璃面板]] 可见时（用户在选轨），左上信息层应继续可见（让用户记得自己在选哪一集的轨），但 5 秒计时不应在面板可见期间倒计触发隐藏——沿用现有 `onTrackSheetVisibilityChanged` 信号在调用方处理（**本任务不修改该信号契约**，仅保证左上信息层与控制条共享 `controlsVisible` 状态后行为一致）
- **G4**：[[运行时切轨]] / [[字幕处理约定]] 不受影响
- **G5**：[[快进/快退步长]] / [[连按合并跳转]] 行为不变（步长和防抖逻辑沿用既有纯函数）
- **G6**：BACK 优先收 UI **不**绕过 [[TV 播放器退出确认]]——只是在 UI 可见时插入一段"先收 UI"的前置步骤

## 7. 非功能要求

- I1：左上信息层主行与副行的字号、行距、阴影参数必须收口为 `core/ui/TvLongFormTitleOverlay.kt` 内的 `object TvLongFormTitleOverlayTokens`，禁止在调用点出现 `2.dp` / `4f` 等阴影字面量（遵循「TV 圆角语言收口」式的 token-only 调用模式）
- I2：所有 fade 动画使用 `TvMotionTokens.DurationStandardMs` + `EasingStandard`，禁止裸 `fadeIn()` / `fadeOut()` 默认 400ms
- I3：左上信息层不引入新的 `Surface` / 背板 / `Shape`，纯 `Box + Column + Text` + `TextStyle.shadow`（遵循布局 ③ 决议）
- I4：焦点环绕实现使用 Compose 原生 `Modifier.focusProperties { left = ...; right = ... }` 配 `FocusRequester`，禁止在 root `onPreviewKeyEvent` 写"判定当前焦点是哪个按钮"的硬编码 if-else 路由
- I5：Slider 不可聚焦使用 `Modifier.focusProperties { canFocus = false }`，禁止在调用方 `Modifier.focusable(false)` 反向覆盖
- I6：5 秒计时器**统一**为既有 `scheduleAutoHideControls()`，不引入第二个 `hideJob`；左上信息层与控制条共用 `controlsVisible` 状态
- I7：BACK 优先收 UI 的实现必须在 `LongFormVideoPlayer` 的 `onPreviewKeyEvent` 内拦截，**不**侵入 `TvLongFormPlayerScreen` / `TvSeriesPlayerScreen` 各自的 BackHandler；UI 收起后下一次 BACK 由调用方现有 BackHandler 接管（[[TV 播放器退出确认]] 现有逻辑无需改）
- I8：版本号：`android-tv-app/tv-app/build.gradle.kts` `versionCode +1`、`versionName` 末位 +1（遵循 `AGENTS.md`）
- I9：动画时长与对比度的回归测试：`LongFormVideoPlayerStyleTest` / `TvMotionTokensTest` 现有 audit 不破规

## 8. 测试策略

### 8.1 纯函数 / 状态机单测（必须）

新增到 `tv-app/src/test/java/com/chee/videos/core/ui/`：

- `TvLongFormRemoteKeyRoutingTest.kt`：
  - `resolveTvRemoteKeyAction(visible: Boolean, focusInControls: Boolean, keyCode: Int, repeatCount: Int, seekStepSec: Int)` 纯函数返回 `TvRemoteKeyAction`（密封类：`Seek(deltaMs) | EnterFocus | ExitFocus | TogglePlayPause | DismissUi | FocusWrap(Direction) | PassThrough`）
  - 覆盖 8 个问题的所有组合行为：隐藏态下 ← / → / ↓ / ↑ / OK / BACK；可见+焦点未进入下同 6 键；可见+焦点进入下同 6 键
  - 覆盖 `repeatCount > 0` 的 3 倍步长（沿用 [[快进/快退步长]]）
- `TvLongFormTitleOverlayDataTest.kt`：
  - `buildTvLongFormTitleOverlayData(primary, seasonNumber?, episodeNumber?, episodeTitle?): TvLongFormTitleOverlayData` 纯函数
  - case：电影 / `18+` → 单行
  - case：电视剧 + 完整 episode.title → 两行尾部含单集标题
  - case：电视剧 + episode.title 为空 → 两行无尾部空格
  - case：电视剧 + episode.title 为 null → 两行无尾部空格
- `TvLongFormControlsAutoHideTest.kt`：
  - `shouldResetAutoHideTimer(prevAction: TvRemoteKeyAction?, nextAction: TvRemoteKeyAction): Boolean` 纯函数
  - 任何"互动"动作（Seek / EnterFocus / FocusWrap / TogglePlayPause）→ true
  - ExitFocus / DismissUi / PassThrough → false（已在路由层处理）

### 8.2 焦点环绕集成测试（必须）

新增 `tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerFocusWrapTest.kt`：

- 渲染 TV 模式控制条，模拟按 ↓ 进入焦点，再连续 → 至环绕，验证焦点回到「播放/暂停」
- 模拟 ←/→ 在 Slider 位置时焦点不卡（确认 Slider `canFocus = false`）

### 8.3 源文审计（必须）

新增 `tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerStyleTest.kt`（扩展现有）或独立 `TvLongFormTitleOverlaySpecTest.kt`：

- `LongFormVideoPlayer.kt` 在 `tvMode = true` 路径上必须 `import` `TvLongFormTitleOverlay`
- `LongFormVideoPlayer.kt` 不能在调用点出现 `fadeIn()` / `fadeOut()` 不带 `tween(TvMotionTokens.DurationStandardMs, ...)`
- `TvLongFormTitleOverlay.kt` 调用点不出现裸 `0xCC` / 形状字面量 / shadow 字面量
- `TvLongFormTitleOverlay.kt` 必须使用 `TextStyle(shadow = Shadow(...))` 而非 `Modifier.shadow(...)`
- 调用方 `TvLongFormPlayerScreen.kt` 和 `TvSeriesPlayerScreen.kt` 必须用同一个 `buildTvLongFormTitleOverlayData(...)` 构造数据

### 8.4 手测（review.md 详写）

进入 TV 模拟器或真机覆盖 §6 所有 A–G 场景。

## 9. 不引入的依赖 / 改动

- 不引入新的 ExoPlayer / Media3 模块
- 不引入新的 DataStore key
- 不修改 `TvSeriesUiModel` / `TvLongFormUiModel` 数据结构
- 不修改 `TvSeriesPlayerViewModel` / `TvLongFormPlayerViewModel` state shape（仅 UI 层读取既有 state）
- 不修改 [[TV 播放器退出确认]] 主流程（仅前置 BACK 优先收 UI 例外）
- 不修改 [[字幕处理约定]] / [[运行时切轨]] / [[续播提示卡]] / [[连播提示卡]] 行为
- 不修改 `UnifiedPlayerScreen` 等非 TV 模式调用方代码（除非有可空 props 引入需要默认值，否则零改动）
- 不修改 `pkg/ffmpeg` / Go 后端 / admin-web
