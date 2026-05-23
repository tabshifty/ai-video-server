# PRD：TV 电视剧自动连播下一集

- 日期：2026-05-23
- 目标端：`android-tv-app/`（TV App，独立工程，与 `android-app/` 源码隔离）
- 范围：TV 电视剧播放器（`feature/tv/TvSeriesPlayerScreen.kt` 及关联 ViewModel / Repository / 设置界面）

## 1. 用户故事

作为 TV 用户，我在电视剧播放器里看完一集后，希望：

1. **手动接下一集**：随时通过控制条上的「下一集」按钮主动跳到下一集（可跨季，保留手动语义）
2. **自动接下一集**：上一集播完时无需操作，自动衔接到下一集（包括跨季）
3. **临近结束有提示**：上一集即将结束时，屏幕右下角出现「即将播放第 N 集」提示卡 + 倒计时，可一键「立即播放」或「取消本次」，且不会遮住底部控制条
4. **可关掉**：在设置里能彻底关掉自动连播（关闭后手动按钮和选集面板仍然能用）
5. **整剧看完有出口**：整剧最后一集播完时不自动循环也不卡黑屏，给一个明确的「全剧已播完 · 返回详情」覆盖层

## 2. 作用域

- **仅** `TvSeriesPlayerScreen` 及其 ViewModel / Repository / TV 设置入口
- **仅** 电视剧；电影 / `18+` / IPTV / 短视频 / 手机端全部不在范围
- **不** 改动 `LongFormVideoPlayer` 公共 API 形状（仅新增 props + 替换"下一集"按钮图标）
- **不** 改动后端、admin-web、Go 服务

## 3. 非目标（明确不做）

- N1：电影 / `18+` 的"自动播下一部"——本期不做，未来如要做需新建独立语义，不复用本期开关
- N2：per-series 的连播开关——仅做全局
- N3：连播提示卡上显示下一集**缩略图**——简化实现，仅文字
- N4：自动连播切集时**预加载**下一集的媒体源——交给现有 `LaunchedEffect(uiState.currentSourceUrl)` 切换路径，新集触发 `setMediaSource + prepare`，不引入预加载机制
- N5：倒计时秒数**可配置**（5 / 10 / 15 / 30）——固定 10 秒；如未来用户反馈再加设置
- N6：连播链路**循环**（最后一集 → 第一集）——整剧末尾即停，不循环
- N7：取消本次连播**记忆**到下一集（"我说取消了你怎么又问"）——单集隔离，下一集到达 T-10 时重新提示
- N8：进入播放器时弹"首次引导"提示——不引入额外打扰

## 4. 术语对应（详见 `CONTEXT.md`「TV 播放术语」区）

本任务一并沉淀以下新术语到 `CONTEXT.md`：

| 术语 | 一句话定义 |
|------|----------|
| `电视剧自动连播` | 全局可关、仅电视剧、上一集播完自动接下一集的能力 |
| `连播链路` | 从当前 episode 解析"下一集"的统一规则（跨季顺接、跳过 `playable=false`、整剧末尾即"无下一集"；编号重复或缺失时只用接口原始顺序兜底） |
| `连播倒计时窗口` | `remainingMs ≤ 10_000` 触发的 10 秒倒计时；倒计时按播放器位置/时长推导，显示剩余整秒；带状态守卫 |
| `连播提示卡` | 右下角告知 + 操作卡，右侧保留 48dp，底部优先避开控制条安全区，抢焦点到「立即播放」 |
| `取消本次连播` | 「取消本次」按钮语义：仅取消本集自动切，不影响全局开关；取消状态绑定当前播放目标，seek 出 T-10 再进也不撤销 |
| `连播覆盖层` | 两种结尾覆盖层：「本集已播完」（有下一集但不自动切）和「全剧已播完」（无下一集） |
| `连播自动切上报` | 自动切时上报当前集 `completed=true`，保留真实 `watchSeconds`，不依赖位置阈值 |
| `手动下一集按钮` | 控制条上的「下一集」按钮，保留现有历史阈值判断，不强制 `completed=true` |
| `TV 下一集按钮图标` | `Icons.Filled.SkipNext`，不复用 `FastForward` |

## 5. 关键决策表（grill-with-docs 九问的结果）

| # | 决策点 | 选项 | 理由 |
|---|--------|------|------|
| Q1 | 「下一集」边界 | **跨季顺接** | TV 用户对"剧"心智是连续的；手动 + 自动共用同一语义 |
| Q2 | 不可播放集 | **跳过 `playable=false`** | "可播放"才是用户视角的下一集；自动连播停在不可播会黑屏 |
| Q3 | 提示与自动切的耦合 | **倒计时绑定** | 国产剧片尾 30-90s，等 STATE_ENDED 太慢；倒计时归零是主触发，STATE_ENDED 兜底 |
| Q4 | 倒计时长度 + 触发 | **固定 10s + 状态守卫**（暂停冻结 / seek 出窗口复位 / 整剧末尾 / 错误 / 选集面板 / 退出提示 / 开关关闭任一即不显示） | 10s = Netflix / Disney+ 默认；seek 出窗口要允许再次自然进入时重新提示 |
| Q5 | 全局开关 | **设置项，默认开** | 默认开 = 用户能立即体验无缝；开关位置贴近"快进/快退步长" |
| Q6 | 提示卡形态 | **右下角小卡 + 抢焦点 + 两按钮**（立即播放 / 取消本次） | 右下角是 TV 端"非关键告知"成熟位置；抢焦点让 OK 键即可立即切；BACK 仍走全局退出确认 |
| Q7 | 整剧末尾 | **隐藏手动按钮 + 「全剧已播完」覆盖层 + 顺手改"下一集"图标为 SkipNext** | 隐藏比灰显在 10-foot 更清晰；覆盖层比黑屏 / 强制 onBack 更友好；图标 bug 顺手修 |
| Q8 | 取消本次语义 + STATE_ENDED 兜底 + 历史上报 | **取消本次 = 仅取消本集自动切；两种覆盖层；自动切显式 `completed=true`** | 单集隔离 + 显式 completed 避免依赖位置阈值的脆弱性 |
| Q9 | 开关存储契约 | **`tv_series_autoplay_enabled` Boolean + `TvSeriesAutoplaySetting` object + 复用 `TvRepository` DataStore 通道** | key 名锁定剧范围避免未来误用；object 形态与 `TvPlaybackSeekStepSetting` 对称 |

## 6. 验收标准（QA scenarios）

### 6.1 自动连播开启路径（默认状态）

- A1：进入 S1E1 播放器，让媒体自然播到剩余 10 秒
  - **✓ 右下角出现「连播提示卡」**，右侧保持 48dp，底部优先避开控制条安全区
  - **✓ 上行显示「即将播放 · 第 2 集 [E2 标题]」**，标题超长 ellipsis
  - **✓ 下行显示「立即播放 (10)」**（数字嵌在 label 内）和「取消本次」两按钮
  - **✓ 焦点抢占到「立即播放」**（TV 焦点蓝青色光感可见）
  - **✓ 数字每秒递减**：10 → 9 → ... → 1
- A2：A1 状态下不操作，倒计时归零
  - **✓ 立即切换到 S1E2 从头开始播放**
  - **✓ 后端历史接收到 S1E1 `completed=true` 的上报**
- A3：A1 状态下按 OK
  - **✓ 立即切到 S1E2**（不等倒计时归零）
  - **✓ 历史上报 S1E1 `completed=true`**
- A4：A1 状态下 DPad LEFT → OK（取消本次）
  - **✓ 提示卡消失**
  - **✓ 本集继续播完直到 STATE_ENDED**
  - **✓ STATE_ENDED 时显示「本集已播完」覆盖层**，含「播放下一集」（抢焦点）和「返回详情」两按钮
  - **✓ 点「播放下一集」→ 切到 S1E2 从头开始**
  - **✓ S1E2 再次到达 T-10 时重新出现提示卡**（不记忆"上集已取消"）

### 6.2 跨季顺接

- B1：S1E10（本季最后一集，下季 S2E1 存在且 `playable=true`），自然播到 T-10
  - **✓ 提示卡显示「即将播放 · 第 1 集 [S2E1 标题]」**（无"第二季"前缀也可，至少集数和标题正确）
  - **✓ 倒计时归零 / 立即播放 → 切到 S2E1**
- B2：在 S1E10 控制条点「下一集」按钮（不等倒计时）
  - **✓ 立即切到 S2E1**（手动按钮也跨季）

### 6.3 跳过不可播放集

- C1：S1E5 后下一可播放是 S1E7（S1E6 `playable=false`），S1E5 自然播到 T-10
  - **✓ 提示卡显示「即将播放 · 第 7 集 [S1E7 标题]」**（跳过 E6）
  - **✓ 倒计时归零 → 切到 S1E7**
- C2：手动按钮在 S1E5 状态点击
  - **✓ 立即切到 S1E7**
- C3：选集面板内 S1E6 仍可见，副标题显示"待绑定 / 未就绪"（现状行为不变）

### 6.4 整剧末尾

- D1：S2E10（整剧最后一集，无更后），自然播到 T-10
  - **✓ 不出现提示卡**（链路解析为"无下一集"）
  - **✓ 控制条上「下一集」按钮不渲染**
- D2：D1 继续播到 STATE_ENDED
  - **✓ 显示「全剧已播完」覆盖层**，仅含「返回详情」按钮（抢焦点）
  - **✓ BACK 键仍走「TV 播放器退出确认」**（第一次提示、第二次退出）

### 6.5 状态守卫

- E1：T-10 期间用户按暂停
  - **✓ 提示卡不出现 / 已出现则倒计时数字停在当前值，不再递减**
  - **✓ 用户继续播放后倒计时从停止处继续**
- E2：T-10 出现提示后，用户 DPad LEFT seek 回 T-30 之前
  - **✓ 提示卡消失，倒计时复位**
  - **✓ 用户再次播到 T-10 时提示卡重新出现并从 10 开始倒计时**
- E3：T-10 期间用户打开选集面板
  - **✓ 提示卡不出现 / 已出现则隐藏**
  - **✓ 关闭选集面板后，若仍在 T-10 区间内则重新出现**
- E4：T-10 期间用户按 BACK（触发退出确认）
  - **✓ 退出确认提示出现，连播提示卡隐藏**
  - **✓ 5 秒内未再按 BACK 则退出确认消失；若仍在 T-10 区间内则连播提示卡重新出现**
- E5：播放过程中发生 `onPlayerError`
  - **✓ 提示卡不出现 / 已出现则隐藏**

### 6.6 全局开关

- F1：在 TV 设置 → 播放 分组下能看到「自动连播下一集」开关，默认开
- F2：关闭开关，重启 App 或返回播放器
  - **✓ 任何集到 T-10 时都不出现提示卡**
  - **✓ STATE_ENDED 时显示「本集已播完」覆盖层（如有下一集）/「全剧已播完」覆盖层（无下一集）**
  - **✓ 控制条上「下一集」按钮**仍然可用（手动操作不受开关影响）
- F3：开关切换状态会持久化（重启 App 后保持）
- F4：开关的视觉与现有 TV 设置项对齐（焦点视觉、字体、对比度）

### 6.7 手动「下一集」按钮

- G1：控制条上「下一集」按钮使用 `Icons.Filled.SkipNext` 图标，不与「快进」混淆
- G2：手动按钮在以下情况**不渲染**：整剧末尾时
- G3：手动按钮在以下情况**正常渲染**：当前集后还有任意 `playable=true` 集（含跨季）
- G4：自动连播开关关闭**不**影响手动按钮渲染（手动操作不被开关绑住）

### 6.8 自动切的历史上报

- H1：倒计时归零自动切 / 点「立即播放」切，**都**上报当前集 `completed=true`
- H2：用户取消本次后 STATE_ENDED 时也算"看完"，上报 `completed=true`
- H3：手动按「下一集」切集，沿用现有 `tvPlaybackHistorySnapshot` 的位置阈值判断（不强制 completed）

## 7. 非功能要求

- I1：倒计时数字每秒更新一次，不抖动、不卡顿
- I2：连播链路解析（跨季 / 跳过 / 末尾判定）必须在主线程同步完成，复杂度 O(n) n=总集数；不引入协程
- I3：提示卡的进场动画沿用 `TvMotionTokens.DurationStandardMs` + `EasingStandard`，**不**自定义 spring/tween 参数
- I4：提示卡的焦点视觉沿用 `tvFocusableGlow` / `tvFocusableScaleOnly`；按钮的按下反馈沿用 `TvFocusMotionTokens` 的 `Press*` 参数
- I5：提示卡半透明深色背景必须保证文本前景对比度 ≥ 7.0:1（WCAG AAA，遵循「TV 10-foot 对比度收口」）
- I6：提示卡圆角使用 `AppChrome.SurfaceShape`（16dp），按钮圆角同；不引入新的 `RoundedCornerShape(N.dp)` 字面量（遵循「TV 圆角语言收口」白名单审计 `TvShapeAuditTest`）
- I7：「TV App version 必须 +1」：`android-tv-app/tv-app/build.gradle.kts` 的 `versionCode +1`、`versionName` 末位 +1（遵循 `AGENTS.md` 的 TV 版本号约定）

## 8. 测试策略

### 8.1 纯函数单测（必须）

- `resolveNextPlayableEpisode(series, currentSeason, currentEpisode): TvNextEpisodeRef?` —— 连播链路核心
  - case 1: 当前季内向后找到 `playable=true` → 返回该集
  - case 2: 当前季内全部 `playable=false` 或已是本季最后一集 → 跨到下一季找
  - case 3: 整剧末尾 → 返回 `null`
  - case 4: 当前集已是整剧最后一集 → 返回 `null`
  - case 5: 跳过中间 `playable=false` 集
  - case 6: 当前集所在季在 `series.seasons` 列表中不是按 `number` 升序排列 → 链路按 `number` 升序找下一季（防御性）
- `TvSeriesAutoplaySetting.parse(raw: Boolean?): Boolean` —— `null` → `true`；`true` → `true`；`false` → `false`
- `shouldShowAutoplayPromptCard(state: TvSeriesPlayerUiState, isPlaying, hasNextEpisode, hasTriggeredOnce, otherSheetsVisible): Boolean` —— 状态守卫合并判定
- `autoplayCountdownTickRemaining(startSeconds, isPaused, elapsedMillis): Int` —— 倒计时数字递减纯函数

### 8.2 源文审计（必须）

- `TvSeriesPlayerScreen` 必须 `import` 共享的 `TvAutoplayPromptCard` / `TvSeriesEndOverlay` 组件
- `LongFormVideoPlayer.kt` 中"下一集"按钮的 `Icons.Filled.SkipNext` 出现且没有遗留 `FastForward` 的同名描述
- 调用点没有裸 `RoundedCornerShape(N.dp)` 字面量（`TvShapeAuditTest` 已有，加入新组件文件不破规则即可）
- 提示卡和覆盖层文件中没有 `tween(...)` 字面量绕过 `TvMotionTokens`

### 8.3 手测（review.md 详写）

进入 TV 模拟器或真机覆盖 §6 所有 A-H 场景。

## 9. 不引入的依赖 / 改动

- 不引入新的 ExoPlayer / Media3 模块
- 不引入新的 DataStore key 类型（继续 Boolean / Int）
- 不修改 `TvSeriesUiModel` / `TvSeasonUiModel` / `TvEpisodeUiModel` 数据结构
- 不改 `TvRepository.fetchSeriesDetail` 接口
- 不改 `tvPlaybackHistorySnapshot` 算法（仅在调用方传 `completed=true` 覆盖）
- 不改 `TvPlayerBackConfirm` 退出确认流程
