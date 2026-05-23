# Implement：TV 播放续播提醒

- 日期：2026-05-23
- 关联 PRD：`prd.md`
- 任务三段执行流：当前阶段 = Implement（CONTEXT.md「tasks 任务三段执行流」定义）

## 1. 代码改动清单

### 1.1 新增文件

#### `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePrompt.kt`

纯函数 + token 单一来源。骨架：

```kotlin
package com.chee.videos.feature.tv

import androidx.compose.ui.unit.dp
import kotlin.math.ceil

object TvResumePromptTokens {
    const val CountdownDurationMs: Long = 5_000L
    const val MinResumeMs: Long = 10_000L
    val CardMinWidthDp = 320.dp
    val CardMaxWidthDp = 380.dp
    val HorizontalPaddingDp = 48.dp
    val BottomPaddingDp = 156.dp
}

data class ResumePromptGuardInput(
    val hasResumeSeekTriggered: Boolean,
    val promptPermanentlyDismissed: Boolean,
    val isPlayerError: Boolean,
    val isBackConfirmVisible: Boolean,
    val isEpisodeSelectorVisible: Boolean,
    val isEndOverlayVisible: Boolean,
    val isAutoplayPromptVisible: Boolean,
    val isPausedByUser: Boolean,
    val remainingMs: Long,
)

fun shouldShowResumePromptCard(input: ResumePromptGuardInput): Boolean =
    shouldTickResumePromptCountdown(input) && input.remainingMs > 0L

fun shouldTickResumePromptCountdown(input: ResumePromptGuardInput): Boolean =
    input.hasResumeSeekTriggered
        && !input.promptPermanentlyDismissed
        && !input.isPlayerError
        && !input.isBackConfirmVisible
        && !input.isEpisodeSelectorVisible
        && !input.isEndOverlayVisible
        && !input.isAutoplayPromptVisible
        && !input.isPausedByUser

fun resumePromptCountdownTickRemaining(remainingMs: Long): Int {
    if (remainingMs <= 0L) return 0
    val seconds = ceil(remainingMs / 1000.0).toInt()
    return seconds.coerceIn(0, 5)
}

fun shouldTriggerResumePrompt(initialResumePositionMs: Long): Boolean =
    initialResumePositionMs >= TvResumePromptTokens.MinResumeMs

fun formatResumePromptTimestamp(positionMs: Long): String {
    val totalSeconds = (positionMs.coerceAtLeast(0L) / 1000L).toInt()
    val hours = totalSeconds / 3600
    val minutes = (totalSeconds % 3600) / 60
    val seconds = totalSeconds % 60
    return if (hours > 0) {
        "%d:%02d:%02d".format(hours, minutes, seconds)
    } else {
        "%d:%02d".format(minutes, seconds)
    }
}
```

#### `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePromptCard.kt`

UI Composable + 私有按钮副本。骨架：

```kotlin
@Composable
fun TvResumePromptCard(
    lastPositionMs: Long,
    visible: Boolean,
    remainingSeconds: Int,
    onContinue: () -> Unit,
    onStartFromBeginning: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val continueFocusRequester = remember { FocusRequester() }
    LaunchedTvInitialFocus(visible) {
        if (visible) {
            continueFocusRequester.tryRequestFocus()
        }
    }
    AnimatedVisibility(
        visible = visible,
        enter = fadeIn(animationSpec = TweenSpec(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
        exit = fadeOut(animationSpec = TweenSpec(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
        modifier = modifier,
    ) {
        Surface(
            color = AppChrome.SurfaceMuted.copy(alpha = 0.94f),
            contentColor = AppChrome.TextPrimary,
            shape = AppChrome.SurfaceShape,
            modifier = Modifier.widthIn(
                min = TvResumePromptTokens.CardMinWidthDp,
                max = TvResumePromptTokens.CardMaxWidthDp,
            ),
        ) {
            Column(
                modifier = Modifier.padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text(
                    text = "上次播放至 ${formatResumePromptTimestamp(lastPositionMs)}",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    TvResumePromptActionButton(
                        text = "继续观看 (${remainingSeconds.coerceAtLeast(1)})",
                        onClick = onContinue,
                        modifier = Modifier.focusRequester(continueFocusRequester),
                    )
                    TvResumePromptActionButton(
                        text = "从头播放",
                        onClick = onStartFromBeginning,
                    )
                }
            }
        }
    }
}

@Composable
private fun TvResumePromptActionButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = AppChrome.SurfaceStrong,
        contentColor = AppChrome.TextPrimary,
        shape = AppChrome.ChipShape,
        modifier = modifier
            .tvFocusableGlow(shape = AppChrome.ChipShape, focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Text(
            text = text,
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.labelLarge,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
        )
    }
}
```

#### `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvResumePromptTest.kt`

| 测试组 | 用例 |
|--------|------|
| `shouldShowResumePromptCard` | 真值表：9 个守卫每个翻转一次确认能 false 化；全部正向 + remainingMs > 0 → true；remainingMs ≤ 0 → false |
| `shouldTickResumePromptCountdown` | 复用守卫真值表；全部正向时 remainingMs = 0 仍返回 true，由倒计时 loop 负责设置永久 dismiss |
| `resumePromptCountdownTickRemaining` | `5_000 → 5`、`4_999 → 5`、`4_001 → 5`、`4_000 → 4`、`1_001 → 2`、`1_000 → 1`、`1 → 1`、`0 → 0`、`-100 → 0` |
| `shouldTriggerResumePrompt` | `9_999 → false`、`10_000 → true`、`10_001 → true`、`Long.MAX_VALUE → true`、`0 → false`、`-1 → false` |
| `formatResumePromptTimestamp` | `0 → "0:00"`、`754_000 → "12:34"`、`5_025_000 → "1:23:45"`、`-100 → "0:00"`、`3_599_000 → "59:59"`、`3_600_000 → "1:00:00"` |

#### `android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/TvResumePromptCardSpecTest.kt`

源文 audit（字符串断言）：

| 断言 | 文件 |
|------|------|
| 含 `TvResumePromptCard(` | `TvLongFormPlayerScreen.kt` |
| 含 `TvResumePromptCard(` | `TvSeriesPlayerScreen.kt` |
| 含 `Alignment.BottomStart` | 两个 player 文件（调用点） |
| 含 `TvResumePromptTokens.HorizontalPaddingDp` 与 `TvResumePromptTokens.BottomPaddingDp` | 两个 player 文件（调用点） |
| 含 `LaunchedTvInitialFocus(` 与 `.tryRequestFocus()` | `TvResumePromptCard.kt` |
| 含 `AppChrome.SurfaceShape` 与 `AppChrome.ChipShape` | `TvResumePromptCard.kt` |
| **不含** 裸 `5_000L` / `10_000L` / `RoundedCornerShape(48.dp)` / `156.dp`（必须走 token） | 两个 player 文件、`TvResumePromptCard.kt`（除 `TvResumePromptTokens` 内部） |

### 1.2 修改文件

#### `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`

**新增 state**（紧邻 `resumedFromHistoryVideoId` 行）：
```kotlin
var resumePromptLastPositionMs by remember(detail.id, uiState.accessToken) { mutableStateOf(0L) }
var resumePromptRemainingMs by remember(detail.id, uiState.accessToken) { mutableStateOf(0L) }
var resumePromptDismissed by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
```
| 3 | 含 `Alignment.BottomStart` | `feature/tv/TvLongFormPlayerScreen.kt` / `feature/tv/TvSeriesPlayerScreen.kt` |
```

**触发点**（在第 200 行的 `LaunchedEffect(playUrl, dataSourceFactory, ...)` 内）：
```kotlin
if (updateDecision.shouldReplaceSource) {
    ...
    if (restorePositionMs > 0L) {
        exoPlayer.seekTo(restorePositionMs)
    }
    if (!updateDecision.preservePosition && detail.id.isNotBlank()) {
        resumedFromHistoryVideoId = detail.id
        if (shouldTriggerResumePrompt(initialResumePositionMs)) {
            resumePromptLastPositionMs = initialResumePositionMs
            resumePromptRemainingMs = TvResumePromptTokens.CountdownDurationMs
            resumePromptDismissed = false
        }
    }
    ...
}
```

**倒计时驱动**（新增 LaunchedEffect；只在 tick 条件成立时递减，和 UI 可见性分开）：
```kotlin
LaunchedEffect(detail.id, shouldTickResumePromptCountdown, resumePromptDismissed) {
    if (resumePromptDismissed) return@LaunchedEffect
    while (resumePromptRemainingMs > 0L) {
        delay(50)
        if (!shouldTickResumePromptCountdown || resumePromptDismissed) break
        resumePromptRemainingMs = (resumePromptRemainingMs - 50L).coerceAtLeast(0L)
    }
    if (resumePromptRemainingMs <= 0L) {
        resumePromptDismissed = true
    }
}
```

**接卡 UI**（在播放器外层 Box 内）：
```kotlin
TvResumePromptCard(
    lastPositionMs = resumePromptLastPositionMs,
    visible = shouldShowResumePromptCard(
        ResumePromptGuardInput(
            hasResumeSeekTriggered = resumedFromHistoryVideoId == detail.id && resumePromptLastPositionMs > 0L,
            promptPermanentlyDismissed = resumePromptDismissed,
            isPlayerError = playerErrorMessage != null,
            isBackConfirmVisible = showBackConfirmPrompt,
            isEpisodeSelectorVisible = false,  // 电影/18+ 无电视剧选集面板；字幕/音轨夜台面板不进入本任务状态机
            isEndOverlayVisible = false,  // 电影/18+ 无连播覆盖层
            isAutoplayPromptVisible = false,  // 电影/18+ 无连播卡
            isPausedByUser = playbackSession.isPausedByUser,
            remainingMs = resumePromptRemainingMs,
        ),
    ),
    remainingSeconds = resumePromptCountdownTickRemaining(resumePromptRemainingMs),
    onContinue = { resumePromptDismissed = true },
    onStartFromBeginning = {
        exoPlayer.seekTo(0L)
        resumePromptDismissed = true
    },
    modifier = Modifier
        .align(Alignment.BottomStart)
        .padding(start = TvResumePromptTokens.HorizontalPaddingDp, bottom = TvResumePromptTokens.BottomPaddingDp),
)
```

#### `android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt`

与电影/`18+` 套路一致，**附加**：

- 切集时（`LaunchedEffect(uiState.currentVideoId)` 已有逻辑）顺手重置：
  ```kotlin
  resumePromptLastPositionMs = 0L
  resumePromptRemainingMs = 0L
  resumePromptDismissed = false
  ```
- 守卫 `isEndOverlayVisible` = `uiState.pendingEndOverlayKind != null`
- 守卫 `isAutoplayPromptVisible` = `shouldShowPromptCard`（已有的计算结果）
- 守卫 `isEpisodeSelectorVisible` = `uiState.selectorVisible`
- 触发点接 `currentEpisode?.watchSeconds`，复用现有 `initialResumePositionMs` 计算

#### `android-tv-app/tv-app/build.gradle.kts`

```diff
- versionCode = 62
- versionName = "0.1.61"
+ versionCode = 63
+ versionName = "0.1.62"
```

#### `CONTEXT.md`

在「TV 播放术语」段尾追加 6 条新术语（`prd.md` 第 4 节列出，原文照搬）。**不**改动既有任何条目；**不**改动其它段落。

#### `plan.md`

按 CLAUDE.md「`plan.md` 是 append-only」要求，**在文件顶部**追加倒序条目（日期/时间、摘要、受影响文件、验证状态）：

```markdown
## 2026-05-23 · TV 播放续播提醒

- 摘要：TV 电影 / `18+` / 电视剧播放器在按上次位置 seek 历史进度时，左下角弹「续播提示卡」5 秒，给用户「从头播放」逃生出口；与右下角连播卡左右镜像、状态守卫互斥。
- 受影响文件：
  - 新增：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvResumePromptCard.kt`、`.../TvResumePrompt.kt`、`.../test/.../TvResumePromptTest.kt`、`.../test/.../TvResumePromptCardSpecTest.kt`
  - 修改：`android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt`、`.../TvSeriesPlayerScreen.kt`、`android-tv-app/tv-app/build.gradle.kts`、`CONTEXT.md`
- 验证状态：单测全绿、`./gradlew :tv-app:assembleDebug` 通过、18 条手测脚本走完
```

## 2. 关键 LaunchedEffect 时序

```
T = 0           player Composable mount，watchSeconds * 1000 >= 10_000
                exoPlayer.seekTo(watchSeconds * 1000)
                resumePromptLastPositionMs = watchSeconds * 1000
                resumePromptRemainingMs = 5_000
                resumePromptDismissed = false
                continueFocusRequester.tryRequestFocus()
                卡 fadeIn (240ms)

T = 1_000       LaunchedEffect 递减 remainingMs → 4_000，label 显示 (4)
T = 2_000       remainingMs → 3_000，label (3)
T = 3_000       remainingMs → 2_000，label (2)
T = 4_000       remainingMs → 1_000，label (1)
T = 5_000       remainingMs → 0
                resumePromptDismissed = true
                卡 fadeOut (240ms)

—— 暂停分支 ——
T = 2_000       用户 PAUSE
                isPausedByUser = true
                守卫返回 false → 卡 hide（AnimatedVisibility fadeOut）
                LaunchedEffect 跳出 while，不再递减 remainingMs（停在 3_000）
T = ?           用户 PLAY
                isPausedByUser = false
                LaunchedEffect 重启（key 变化），继续从 remainingMs=3_000 递减
                守卫真值条件再次满足 → 卡 show

—— 永久 dismiss 分支 ——
任意时刻       playerError / episodeSelector / autoplay 任一为 true
                守卫返回 false → 卡 hide
                同时 LaunchedEffect 检测到对应 state（新增一个监听 LaunchedEffect）
                resumePromptDismissed = true → 卡永久不再出现

—— 退出确认临时隐藏分支 ——
任意时刻       backConfirm = true
                守卫返回 false → 卡 hide
                LaunchedEffect 跳出 while，不再递减 remainingMs
                不设置 resumePromptDismissed；退出确认消失且 remainingMs > 0 后按冻结值继续显示
```

## 3. 与连播卡的视觉 / 互斥关系

- **共享视觉 token**：`AppChrome.SurfaceMuted` / `AppChrome.SurfaceShape` / `AppChrome.ChipShape` / `AppChrome.TextPrimary` / `tvFocusableGlow` / `TvMotionTokens.DurationStandardMs` / `TvMotionTokens.EasingStandard`
- **不共享代码**：两个独立 Composable + 各自私有按钮副本，避免 API 膨胀
- **互斥**：续播卡在 player prepare 后 5 秒内有效，连播卡在 `remainingMs <= 10_000` 才有效——时间线天然无交集；为防御未来改动，续播卡守卫显式叠 `!isAutoplayPromptVisible` 与 `!isEndOverlayVisible`

## 4. 实施顺序（建议 TDD）

1. 写 `TvResumePromptTokens` + 纯函数 + `TvResumePromptTest`（先红后绿）
2. 写 `TvResumePromptCard` UI
3. 写 `TvResumePromptCardSpecTest` 源文 audit（先红，调用点接入后绿）
4. 接入 `TvLongFormPlayerScreen`，跑 H1–H7 手测
5. 接入 `TvSeriesPlayerScreen`，跑 H8–H12 手测
6. 边界 / 互斥手测 H13–H18
7. 改版本号、追加 `CONTEXT.md` 6 条术语、追加 `plan.md` 顶部条目
8. 一次性提交本实现阶段的代码 / 测试 / 版本号 / `plan.md` 改动（中文 commit message）；grill 审查阶段的任务文档 / `CONTEXT.md` / `plan.md` 提交可独立存在

## 5. 风险与回退

- **风险点 1**：`resumedFromHistoryVideoId` 是既有节流状态，本期复用它判断「是否走过续播 seek」——如未来有人改这个状态的语义，续播卡会跟着错。**对策**：源文 audit 在 `TvResumePromptCardSpecTest` 里同时锁定 `resumedFromHistoryVideoId == detail.id` / `resumedFromHistoryVideoId == uiState.currentVideoId` 字样必须出现在两个 player 文件中。
- **风险点 2**：`LaunchedEffect` 用 `delay(50)` 循环递减是简单的「土法」实现；如未来需要更精细的可见态计时语义可换 `withFrameNanos` 推进。**当前**：50ms 粒度足够支撑「秒级倒计时显示」，不引入额外抽象。
- **风险点 3**：暂停或退出确认临时隐藏时，`LaunchedEffect` 通过可见态 key 变化「跳出 → 重启」，重启后从冻结的 `remainingMs` 继续，保留剩余秒数。验证单测：`resumePromptCountdownTickRemaining` 用边界值确保「3_001 ms 显示 (4)、3_000 ms 显示 (3)」一致。
- **风险点 4**：焦点从「继续观看」移动到「从头播放」不冻结倒计时；续播卡是非阻塞提示，倒计时归零前未按 OK 即默认继续观看并 dismiss，与连播卡焦点移动不冻结倒计时的交互保持一致。
- **回退**：仅删除新增文件 + 撤销两个 player 中的接卡代码 + 还原 `tv-app/build.gradle.kts` 与 `CONTEXT.md` 增量即可，纯加法改动；不影响既有续播 seek 行为。
