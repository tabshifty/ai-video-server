# Feedback：TV 电视剧自动连播下一集 · 实现后 Review

- 日期：2026-05-23
- 审查对象：`0db65e40 实现TV电视剧自动连播下一集`（HEAD）
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`
- 关联 Review 验收脚本：`./review.md`

## 总览

- 改动规模：20 个文件、+916/-33
- 自动化验证：`:tv-app:testDebugUnitTest` / `:tv-app:assembleDebug` / `:tv-app:assembleRelease` 三档均绿（按 `plan.md` 2026-05-23 12:04 条目）
- 沉淀：`CONTEXT.md` 9 个新术语全部到位，含 `[[name]]` 交叉引用
- 版本：`versionCode 61→62`、`versionName 0.1.60→0.1.61`
- `plan.md` 4 条 reverse-chronological 进度条目

整体质量：实现紧扣 PRD / CONTEXT 约束，测试覆盖关键路径，命名一致，token 复用正确。**但发现 1 条 Critical、2 条 Medium、2 条 Low 待修复**，其中 Critical 必须在上线前修掉。

---

## 🔴 Critical：自动切的双触发 race

### 位置
- `TvSeriesPlayerScreen.kt` 的 `advanceFromAutoplay()` + `handlePlaybackEnded()` + `LaunchedEffect(... remainingMs ...)`

### 问题
倒计时归零路径与 ExoPlayer `STATE_ENDED` 兜底路径之间存在窗口竞态：

1. `remainingMs ≤ 0` 时 `LaunchedEffect` 触发 `advanceFromAutoplay()` —— 同步路径：
   - 写 `lastAutoplaySwitchedVideoId = "video-1"`（切前的 videoId）
   - 调 `reportTvSeriesHistory(..., completedOverride = true)`
   - 调 `viewModel.advanceToNextEpisodeFromAutoplay()` → `currentVideoId = "video-2"`、`autoplayCanceledForCurrentEpisode = false`（已被 `updateSelectedEpisode` 复位）
2. **同一主线程 tick 之后**，ExoPlayer 旧 source 残留的 `STATE_ENDED` 事件被处理 —— `handlePlaybackEnded()` 用 `latestUiState`（已是新集 state）：
   - `autoplayEnabled = true`、`!canceledForCurrentEpisode`、新集视角下 `hasNext` 仍 true → 命中第一分支再次调 `advanceFromAutoplay()`
3. `advanceFromAutoplay` 内的重入保护 `lastAutoplaySwitchedVideoId == videoId` 比较的是 `"video-1" == "video-2"` → **不命中** → 切到下下集

### 复现条件
ExoPlayer 在 `remainingMs` ceiling 归零的 250ms tick 之前已 fire `STATE_ENDED` 事件入主线程队列（实际媒体 EOF 略早于显示的"剩余 0 秒"）。

### 危害
用户看到 S1E1 → S1E2 还没出第一帧就跳到 S1E3，完全失控；行为不可预测，无法靠"取消本次"挽回。

### 修法（推荐 A）

**A. `handlePlaybackEnded` 入口短路**（最简）：

```kotlin
fun handlePlaybackEnded() {
    val state = latestUiState
    // 旧集 STATE_ENDED 残留：已经经过 advanceFromAutoplay 切走的集
    if (
        lastAutoplaySwitchedVideoId.isNotBlank() &&
        state.currentVideoId.isNotBlank() &&
        state.currentVideoId != lastAutoplaySwitchedVideoId
    ) {
        return
    }
    // ... 现有分支逻辑
}
```

语义：`lastAutoplaySwitchedVideoId` 记录"刚由 advanceFromAutoplay 切走的旧 videoId"；如果 STATE_ENDED 进 handler 时 state 已经是更新后的新集（currentVideoId 与"切走的旧集"不等），说明这次 STATE_ENDED 是切前残留事件，**忽略**。

**B. 替代方案 `advanceFromAutoplay` 完成后立即 `exoPlayer.removeListener` 再重新 add**：成本高、改动面广，不推荐。

### 单测建议
新增 `TvSeriesAutoplayHandlerSpecTest`（纯函数风格）：抽出"是否应该响应 STATE_ENDED"判定为纯函数 `shouldHandlePlaybackEnded(currentVideoId, lastAutoplaySwitchedVideoId): Boolean`，单测覆盖：
- 首次 STATE_ENDED：lastAutoplay 为空 → true
- 旧集残留 STATE_ENDED：lastAutoplay = currentVideoId 不同 → false
- 自然 STATE_ENDED（手动场景）：lastAutoplay = currentVideoId（同一 ID）→ true（兜底走自动切或覆盖层）

---

## 🟠 Medium #1：自动切瞬间的"双重历史上报"

### 位置
- `TvSeriesPlayerScreen.kt:194-199` 的 `LaunchedEffect(uiState.currentVideoId)` 切换上报通道
- `advanceFromAutoplay` 内的 `reportTvSeriesHistory(..., completedOverride = true)`

### 问题
自动切路径上当前集会被上报两次：

1. `advanceFromAutoplay` 同步调用 `reportTvSeriesHistory(viewModel, videoId, exoPlayer, completedOverride = true)` → 落库 `completed = true`
2. `viewModel.advanceToNextEpisodeFromAutoplay()` 切换 `currentVideoId` → 触发 `LaunchedEffect(uiState.currentVideoId)` → 第二次调 `reportTvSeriesHistory(lastHistoryVideoId, exoPlayer)` **不带 `completedOverride`**

第二次上报时 `lastHistoryVideoId` 仍是上一集（"video-1"），但 `exoPlayer.currentPosition / duration` 可能已经被异步 `setMediaSource` 影响，算出来的 `completed` 不可控：可能维持 true（位置接近 duration），也可能被算成 false（ExoPlayer reset 为 0），**把后端的 `completed=true` 覆盖回 `false`**。

### 危害
用户视角："我看完了 S1E1 自动切到了 S1E2"，但历史页可能显示 S1E1 未看完；影响"最近播放"列表的语义和后续推荐。

### 修法
`advanceFromAutoplay` 同步把 `lastHistoryVideoId` 设为 `currentVideoId`（提前于 LaunchedEffect 切换上报路径），让该路径的 `lastHistoryVideoId.isNotBlank() && lastHistoryVideoId != uiState.currentVideoId` 守卫不再命中：

```kotlin
fun advanceFromAutoplay() {
    val state = latestUiState
    val videoId = state.currentVideoId
    if (videoId.isBlank() || lastAutoplaySwitchedVideoId == videoId || !state.hasNextPlayableEpisode()) {
        return
    }
    lastAutoplaySwitchedVideoId = videoId
    reportTvSeriesHistory(viewModel, videoId, exoPlayer, completedOverride = true)
    lastHistoryVideoId = videoId  // 👈 告知切换上报路径"已经报过了，别再报"
    viewModel.advanceToNextEpisodeFromAutoplay()
}
```

注意：`lastHistoryVideoId` 在 LaunchedEffect 里最后还是会被赋为新集 `currentVideoId`，保持后续切换上报的正常流转。

### 单测建议
TvSeriesPlayerViewModelTest 增加一个 FakeRepository spy 计数 `reportHistory` 调用次数，验证自动切路径下当前集只调一次（completed=true）。但因为这是 Screen 层逻辑（不在 ViewModel 里），需要 Compose UI 测试或抽出 Screen-level pure handler。

---

## 🟠 Medium #2：暂停时提示卡消失而非"数字停留"——与 review.md E1 文案不一致

### 位置
- `TvSeriesAutoplay.kt:76` `shouldShowAutoplayPromptCard` 的 `if (!input.isPlaying) return false`
- `tasks/.../review.md` §1.5 E1 描述

### 问题
**实现选择**：`!isPlaying → shouldShow=false` → 提示卡 fade out 消失；用户恢复播放后若仍在 T-10 区间，卡 fade in 重新出现，数字基于 remainingMs 接续（不是从 10 重新开始）。

**review.md E1 文案**："✓ 提示卡如已出现，倒计时数字停在当前值，不再递减" —— 暗示卡仍然可见。

`CONTEXT.md` 的 `连播倒计时窗口` 写"用户暂停时冻结倒计时（不消耗剩余秒数）" —— 这层与实现一致（remainingMs 不变 = 数字不变），但 CONTEXT 没明说"卡是否消失"。

### 危害
review 阶段按手测脚本走 E1 时会观察到"卡消失"而不是"卡停留"，引起"是不是 bug"的判断分歧；如不澄清，后续维护者可能误改回"暂停时卡仍可见"反而引入新问题（卡停留 + 数字不动 = 视觉上像 hang）。

### 推荐：修文档（实现是合理设计，不动）
把 review.md §1.5 E1 改为：

```markdown
13. **E1**：S1E1 seek 到 T-7（在 10 秒倒计时窗口内）后按暂停
    - ✓ 提示卡 fade out 隐藏（不打扰暂停状态）
    - ✓ 倒计时不消耗剩余秒数（基于 remainingMs，暂停时 position 不变）
    - ✓ 继续播放后提示卡 fade in 重新出现，数字从暂停那刻接续（不是从 10 重新开始）
```

CONTEXT.md `连播倒计时窗口` 也建议补一句"暂停时提示卡隐藏，恢复后重新出现并按 remainingMs 接续显示数字"，与 `连播提示卡` 内的可见性规则对齐。

### 替代：改实现
如果你坚持"暂停时卡停留"，把 `if (!input.isPlaying) return false` 守卫拿掉，让卡在暂停时仍渲染。但此时卡仍然抢焦点 / 占屏，且数字不动给用户"卡死了"的错觉，**不推荐**。

---

## 🟡 Low #1：`shouldShowAutoplayPromptCard` 守卫缺 `pendingEndOverlayKind`

### 位置
`TvSeriesPlayerScreen.kt` shouldShowPromptCard 派生：

```kotlin
val shouldShowPromptCard = shouldShowAutoplayPromptCard(...) && uiState.pendingEndOverlayKind == null
```

### 问题
`pendingEndOverlayKind == null` 守卫被手动叠加在 Screen 层，没纳入 `AutoplayPromptGuardInput`。`AutoplayPromptGuardInput` 的设计意图是"集中守卫真值表"；外挂一条让单测覆盖不到、维护时容易遗漏。

### 修法
`AutoplayPromptGuardInput` 加 `isEndOverlayVisible: Boolean` 字段，纯函数内 `if (input.isEndOverlayVisible) return false`，单测 `shouldShowAutoplayPromptCard_requiresAllowedStateAndPositiveRemainingWindow` 一并覆盖。Screen 层调用方传 `isEndOverlayVisible = uiState.pendingEndOverlayKind != null`，去掉外挂的 `&& ...` 表达式。

---

## 🟡 Low #2：`tvFocusableGlow + clickable` 双重 focusable 担忧

### 位置
- `TvAutoplayPromptCard.kt:97-99`（按钮）
- `TvSeriesEndOverlay.kt:113-115`（按钮）
- `TvCatalogScreen.kt:541-543`（自动连播开关 row）

### 问题
`CONTEXT.md` 「TV 一级菜单 → 内容区焦点回流陷阱」与"TV 首页左侧菜单按钮"条目明确禁止 `tvFocusableGlow() + .focusable()` 叠加（导致"按两次"bug）。

本期新代码使用 `tvFocusableGlow + .clickable()`：`.clickable()` 内部也会附加 `focusable`，行为是否触发同类 bug 取决于 Compose 当前版本对 clickable 的处理方式。

### 评估
CatalogScreen 里现有按钮也用 `tvFocusableGlow + clickable`（如季选择 row、SeekStep selector），说明项目实际接受这个模式而未出"按两次"。但 CONTEXT 明文只警告 `.focusable()`，没明说 `.clickable()`。

### 建议
**手测必做项**：review.md §1.1 A3 "提示出现按 OK" 测试时**显式记录**是否单次 OK 就触发"立即播放"切集；同样在 §1.4 D2 "全剧已播完覆盖层点击返回详情"也观察是否单次 OK 生效。

如果手测发现"按两次"现象，把按钮改造成不叠 `.clickable()`、而是用 `tvFocusableGlow + .onPreviewKeyEvent { ... DirectionCenter/Enter ... }` 自行处理点击，或者寻找项目里已有的统一可点击 TV 按钮（如 `TvIconActionButton`）替换。

---

## ✅ 已正确实现 / 无需调整

- **跨季顺接**：`resolveNextPlayableEpisode` 用 `sortedWith(compareBy { it.number })` 防御性升序，单测 `resolveNextPlayableEpisode_crossesToNextSeasonAndReturnsNullAtSeriesEnd` 覆盖
- **跳过不可播放集**：`isPlayableEpisode = playable && videoId.isNotBlank()` 双重判定，单测 `resolveNextPlayableEpisode_sortsByNumberAndSkipsUnplayableEpisodes` 覆盖
- **自动切 `completedOverride=true` 显式注入**：`advanceFromAutoplay` 内传 `true`；手动按钮路径不传，保留位置阈值语义（CONTEXT.md `手动下一集按钮` 落实）
- **整剧末尾手动按钮不渲染**：`onNextEpisode = if (hasNextEpisode) viewModel::nextEpisode else null`，配合 LongFormVideoPlayer 内 `onNextEpisode?.let { ... }` 守卫
- **「下一集」按钮图标 `SkipNext`**：源文 audit 测试 `next episode control uses skip next icon` 锁定不回归
- **`startCurrentEpisodeFromBeginning` 分流**：自动切 + 手动切都从头播放，手动选集面板 / 历史 resume 路径保留 `watchSeconds`，符合 CONTEXT.md `连播链路` "链路触发的切集一律从头开始"
- **取消本次 / 全剧末尾 / STATE_ENDED 兜底**：`handlePlaybackEnded` 三分支覆盖 PRD §6.1 A4 / §6.4 D2 / §6.6 F3
- **焦点抢占**：全部走 `LaunchedTvInitialFocus + tryRequestFocus`，含 spec 测试锁定
- **token 复用**：圆角 / 动效 / typography 全部走 `AppChrome` / `TvMotionTokens`，源文 audit 锁定无裸 `RoundedCornerShape(` / `tween(`
- **9 个 CONTEXT 术语**：全到位含 `[[name]]` 交叉引用；用户加的 `手动下一集按钮` 把"手动 vs 自动"的语义边界沉淀清楚
- **版本号、plan.md**：均符合 `AGENTS.md` 约定

---

## 修复落地建议顺序

1. 🔴 **必修**：`handlePlaybackEnded` 入口短路（Critical 双触发 race）
2. 🟠 **必修**：`advanceFromAutoplay` 完成后同步赋值 `lastHistoryVideoId = videoId` 防双重上报覆盖 `completed`
3. 🟠 **必修**：review.md §1.5 E1 措辞校准（文档对齐实现）；CONTEXT.md `连播倒计时窗口` 补可见性规则
4. 🟡 **建议**：`AutoplayPromptGuardInput` 加 `isEndOverlayVisible` 字段，纯函数 + 单测一并覆盖
5. 🟡 **手测必做**：review.md §1.1 A3 / §1.4 D2 显式记录"按 OK 是否单次生效"，确认无 `tvFocusableGlow + clickable` 双 focusable 回归

---

## 修复后的回归要求

- 新增 `shouldHandlePlaybackEnded` 纯函数 + 单测覆盖三种 case（首次 / 旧集残留 / 自然 ended）
- 新增 `shouldShowAutoplayPromptCard` 测试 case `blockedByEndOverlay`
- 手测 review.md §1.1 A2 完成后**检查后端历史接口**，确认 S1E1 只收到一次 `completed=true` 上报（而非两次，第二次可能为 false）
- 手测 review.md §1.5 E1 按修订后的文案验证"暂停时卡消失、恢复时卡接续"
- 修复提交需保留 `versionCode +1`、`versionName +0.0.1`（继续 bump 到 63 / 0.1.62）；plan.md 追加 reverse-chronological 进度条目；CONTEXT.md 如有补充也同步沉淀
