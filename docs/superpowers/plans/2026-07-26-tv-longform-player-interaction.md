# TV 长视频播放器交互模型重构 · 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Do not free-style: each task below is self-contained, ends in a green test run and a git commit.

**规格来源**：`docs/superpowers/specs/2026-07-26-tv-longform-player-interaction-design.md`（已批准，commit `468a177`）

---

## Goal

把 TV 端长视频（单片 + 剧集）播放器的交互模型从「手机播放器移植 + 补丁」重做为主流 TV 播放器模型：

- 左右键进入 **scrub 模式**（ghost 游标 + 目标时间，OK 提交 / BACK 取消 / 1.5s 自动提交），视频不暂停，长按三档加速。
- UI 状态由**单一状态机** `TvPlayerUiMode { Hidden, Chrome, Scrubbing, EpisodeRail }` 表达，替换现状四个互相打架的布尔/可空变量。
- 控制条重构为**双层**（顶部信息层 + 底部操作层），进度轨只展示不进焦点链。
- 新增**缓冲态指示**（350ms 延迟显示 / 8s 慢速提示 / 20s 转软重试）。
- 接入 **MediaSession**，让系统级播放控制与「继续观看」可用。

## Architecture

```
feature/tv/TvLongFormPlayerScreen.kt   单片宿主（瘦身：只做数据 + 播放内核 + 接线）
feature/tv/TvSeriesPlayerScreen.kt     剧集宿主（同上 + 分集）
feature/tv/TvLongFormMedia3Player.kt   播放内核（新增 onBufferingChanged + MediaSession effect 挂载）
feature/tv/TvLongFormMediaSession.kt   新：MediaSession 的 DisposableEffect 封装

core/ui/player/                        新包：TV 播放器共享控制层
  TvPlayerTokens.kt          视觉 token + 焦点目标枚举 + 时间格式化 + 图标按钮
  TvPlayerScrubMath.kt       纯 Kotlin：锚点 / 累加 / 加速档 / 钳位 / 偏移标签
  TvPlayerInteraction.kt     纯 Kotlin reducer：按键 → (状态, 副作用)
  TvPlayerBufferingPolicy.kt 纯 Kotlin：缓冲时长 → 展示阶段
  TvPlayerTopInfo.kt         Chrome 顶层信息
  TvPlayerScrubTrack.kt      进度轨 + ghost 游标（不可聚焦）
  TvPlayerActionRow.kt       Chrome 底层操作行（唯一焦点入口）
  TvPlayerCenterFeedback.kt  中心提示（scrub 偏移 / 不可 seek / 分集切换反馈）
  TvPlayerBufferingIndicator.kt 缓冲圈 + 慢速文案
  TvPlayerEpisodeRail.kt     分集横向轨（从 LongFormVideoPlayer.kt 迁出并重命名）
  TvPlayerChrome.kt          组装以上组件，替代 TvSeriesCorePlaybackOverlay
```

三条边界规则（规格 §4.1）：

1. `TvPlayerInteraction.kt` / `TvPlayerScrubMath.kt` / `TvPlayerBufferingPolicy.kt` **不得** import 任何 `androidx.compose.*`。
2. `TvPlayerChrome.kt` 及其子组件**不得** import 任何 `androidx.media3.*`——播放内核只通过参数/回调交互。
3. MediaSession 生命周期用 `DisposableEffect` 创建 / 释放，不进 Chrome。

## Tech Stack

- Kotlin + Jetpack Compose（Compose BOM 2024.10.01，compiler ext 1.5.15），minSdk 26 / targetSdk 35，JVM 17
- Media3 / ExoPlayer 1.4.1（本次新增 `androidx.media3:media3-session:1.4.1`）
- JUnit4 JVM 单测（`:tv-app:testDebugUnitTest`），含项目既有的「源码形态审计」测试风格

## Global Constraints

- **中文优先**：所有用户可见字符串、Markdown、提交信息一律中文，不得乱码。
- **TV 焦点视觉语言**：焦点用柔和蓝青光晕；圆形小图标按钮必须走 `TvIconActionButton`；同一节点不得叠加 `tvFocusableGlow()` 与额外 `.focusable()`；`FocusRequester.requestFocus()` 只能指向当时已组合的节点。
- **每个任务结束都要 commit**，提交信息中文，结尾带 `Co-Authored-By: Claude Opus 5 <noreply@anthropic.com>`。
- **不要触碰** `.superpowers/`、`.codex/skills/*`、`android-app/`（手机端）、IPTV / 短视频 / 投放相关文件。
- 每个任务的验证命令：
  ```bash
  cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest
  ```
  最后一个任务追加 `:tv-app:assembleDebug`。

## 对规格的两处有意偏离（必须遵守本计划的版本）

**偏离 1：`core/ui/LongFormVideoPlayer.kt` 将被排除出 TV 编译图。**
规格 §4.2 假设该文件已不在 TV 编译图内。实测**它在**（只有它的调用方 `feature/player/**` 与 `DetailScreen.kt` 被排除，它自己没有）。且它内部 7 个 TV 要用的工具（`TvSeriesControlsPage` / `TvPlaybackProgressBar` / `TvEpisodeRail` / `CompactPlayerControlButton` / `formatPlaybackTime` / `normalizeTvSeekStepSeconds` / `TvControlFocusTarget` 等）**全都被它自己的 body 使用**——移出去就必须给它加 import，那反而改动了手机端文件。
因此：新包 `core.ui.player` 用**重命名后的符号**独立重建这些工具（避免与手机端文件重复声明），并在 Task 8 把 `com/chee/videos/core/ui/LongFormVideoPlayer.kt` 加入 `tvMainSourceExcludes`。这同时满足 CLAUDE.md「TV 编译图刻意收窄」，并从 TV 构建里移除 1664 行死代码。

**偏离 2：`TvLongFormMediaSessionEffect` 从 `TvLongFormMedia3Player` 内部调用，而非宿主 Screen。**
规格 §4.1 规则三字面写「宿主层的 effect」，但 `ExoPlayer` 实例是 `TvLongFormMedia3Player.kt` 的私有 `remember`，宿主拿不到 `Player` 引用。折中：effect 本体放独立文件 `feature/tv/TvLongFormMediaSession.kt`，调用点在 `TvLongFormMedia3Player` 内部。仍满足「独立文件 + `DisposableEffect` 创建/释放 + 不进 Chrome」。

---

## File Structure

新增：

```
android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/player/
  TvPlayerTokens.kt
  TvPlayerScrubMath.kt
  TvPlayerInteraction.kt
  TvPlayerBufferingPolicy.kt
  TvPlayerTopInfo.kt
  TvPlayerScrubTrack.kt
  TvPlayerActionRow.kt
  TvPlayerCenterFeedback.kt
  TvPlayerBufferingIndicator.kt
  TvPlayerEpisodeRail.kt
  TvPlayerChrome.kt
android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/
  TvLongFormMediaSession.kt
android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/player/
  TvPlayerTokensTest.kt
  TvPlayerScrubMathTest.kt
  TvPlayerInteractionTest.kt
  TvPlayerBufferingPolicyTest.kt
  TvPlayerChromeSpecTest.kt
android-tv-app/tv-app/src/test/java/com/chee/videos/feature/tv/
  TvLongFormMediaSessionSpecTest.kt
```

修改：

```
android-tv-app/tv-app/build.gradle.kts                 版本号 + media3-session + 排除清单
.../feature/tv/TvLongFormMedia3Player.kt               onBufferingChanged + MediaSession 挂载
.../feature/tv/TvLongFormPlayerScreen.kt               接 TvPlayerChrome
.../feature/tv/TvSeriesPlayerScreen.kt                 接 TvPlayerChrome
.../core/ui/TvLongFormTitleOverlay.kt                  接收迁入的 buildTvLongFormTitleOverlayData
plan.md / CONTEXT.md                                   追加记录
```

删除：

```
.../core/ui/TvSeriesCorePlaybackOverlay.kt             被 TvPlayerChrome 取代
.../core/ui/TvLongFormRemoteKeyRouting.kt              被 TvPlayerInteraction 取代
```

测试迁移/删除见 Task 8。

---

## Task 1：建立 `core/ui/player` 基础层（tokens / 时间 / 焦点目标 / 图标按钮）

**Files:**
- 新增 `android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/player/TvPlayerTokens.kt`
- 新增 `android-tv-app/tv-app/src/test/java/com/chee/videos/core/ui/player/TvPlayerTokensTest.kt`

**Interfaces:**

```kotlin
internal val TvPlayerGlassSurface: Color
internal val TvPlayerGlassSurfaceStrong: Color
internal val TvPlayerSubtleSurface: Color
internal val TvPlayerProgressTrack: Color
internal val TvPlayerProgressActiveTrack: Color
internal val TvPlayerScrubGhostTrack: Color

internal enum class TvPlayerActionTarget { PlayPause, Subtitle, AudioTrack, EpisodeList }
internal enum class TvPlayerActionDirection { Left, Right }
internal fun resolveTvPlayerHorizontalActionTarget(
    current: TvPlayerActionTarget,
    direction: TvPlayerActionDirection,
    targets: List<TvPlayerActionTarget>,
): TvPlayerActionTarget

internal fun normalizeTvPlayerSeekStepSeconds(seconds: Int): Int
internal fun formatTvPlayerTime(millis: Long): String

internal object TvPlayerEpisodeRailLayoutTokens {
    val ItemSlotWidthDp = 120.dp
    val TooltipHeightDp = 44.dp
    val TooltipMaxWidthDp = 220.dp
}

@Composable
internal fun TvPlayerIconButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    focusRequester: FocusRequester? = null,
    enabled: Boolean = true,
)
```

**Steps:**

- [ ] 写 `TvPlayerTokensTest.kt`，覆盖三组纯逻辑：
  ```kotlin
  package com.chee.videos.core.ui.player

  import org.junit.Assert.assertEquals
  import org.junit.Test

  class TvPlayerTokensTest {
      private val targets = listOf(
          TvPlayerActionTarget.PlayPause,
          TvPlayerActionTarget.Subtitle,
          TvPlayerActionTarget.AudioTrack,
      )

      @Test
      fun rightMovesToNextTarget() {
          assertEquals(
              TvPlayerActionTarget.Subtitle,
              resolveTvPlayerHorizontalActionTarget(
                  TvPlayerActionTarget.PlayPause,
                  TvPlayerActionDirection.Right,
                  targets,
              ),
          )
      }

      @Test
      fun rightWrapsAroundToFirstTarget() {
          assertEquals(
              TvPlayerActionTarget.PlayPause,
              resolveTvPlayerHorizontalActionTarget(
                  TvPlayerActionTarget.AudioTrack,
                  TvPlayerActionDirection.Right,
                  targets,
              ),
          )
      }

      @Test
      fun leftWrapsAroundToLastTarget() {
          assertEquals(
              TvPlayerActionTarget.AudioTrack,
              resolveTvPlayerHorizontalActionTarget(
                  TvPlayerActionTarget.PlayPause,
                  TvPlayerActionDirection.Left,
                  targets,
              ),
          )
      }

      @Test
      fun unavailableCurrentTargetFallsBackToFirst() {
          assertEquals(
              TvPlayerActionTarget.PlayPause,
              resolveTvPlayerHorizontalActionTarget(
                  TvPlayerActionTarget.EpisodeList,
                  TvPlayerActionDirection.Right,
                  targets,
              ),
          )
      }

      @Test
      fun seekStepFallsBackToTenForUnsupportedValues() {
          assertEquals(10, normalizeTvPlayerSeekStepSeconds(7))
          assertEquals(10, normalizeTvPlayerSeekStepSeconds(0))
          assertEquals(10, normalizeTvPlayerSeekStepSeconds(-30))
      }

      @Test
      fun seekStepKeepsWhitelistedValues() {
          listOf(5, 10, 15, 20, 30).forEach {
              assertEquals(it, normalizeTvPlayerSeekStepSeconds(it))
          }
      }

      @Test
      fun timeFormatsBelowOneHourAsMinutesSeconds() {
          assertEquals("00:00", formatTvPlayerTime(0L))
          assertEquals("01:05", formatTvPlayerTime(65_000L))
          assertEquals("59:59", formatTvPlayerTime(3_599_000L))
      }

      @Test
      fun timeFormatsAboveOneHourWithHours() {
          assertEquals("1:00:00", formatTvPlayerTime(3_600_000L))
          assertEquals("2:03:04", formatTvPlayerTime(7_384_000L))
      }

      @Test
      fun negativeTimeClampsToZero() {
          assertEquals("00:00", formatTvPlayerTime(-5_000L))
      }
  }
  ```
- [ ] 跑 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests '*TvPlayerTokensTest'`，确认**编译失败**（符号不存在）。
- [ ] 写 `TvPlayerTokens.kt`。颜色沿用 `AppChrome` 既有配色，`TvPlayerScrubGhostTrack` 用 `AppChrome.TextPrimary.copy(alpha = 0.92f)`（ghost 游标要比实心条更亮以便区分）：
  ```kotlin
  package com.chee.videos.core.ui.player

  import androidx.compose.foundation.layout.size
  import androidx.compose.runtime.Composable
  import androidx.compose.ui.Modifier
  import androidx.compose.ui.focus.FocusRequester
  import androidx.compose.ui.focus.focusRequester
  import androidx.compose.ui.graphics.Color
  import androidx.compose.ui.graphics.vector.ImageVector
  import androidx.compose.ui.unit.dp
  import com.chee.videos.core.ui.AppChrome
  import com.chee.videos.core.ui.TvIconActionButton

  internal val TvPlayerGlassSurface: Color = AppChrome.Surface.copy(alpha = 0.90f)
  internal val TvPlayerGlassSurfaceStrong: Color = AppChrome.SurfaceMuted.copy(alpha = 0.92f)
  internal val TvPlayerSubtleSurface: Color = AppChrome.Surface.copy(alpha = 0.72f)
  internal val TvPlayerProgressTrack: Color = AppChrome.TextMuted.copy(alpha = 0.24f)
  internal val TvPlayerProgressActiveTrack: Color = AppChrome.Accent.copy(alpha = 0.96f)
  internal val TvPlayerScrubGhostTrack: Color = AppChrome.TextPrimary.copy(alpha = 0.92f)

  internal enum class TvPlayerActionTarget { PlayPause, Subtitle, AudioTrack, EpisodeList }

  internal enum class TvPlayerActionDirection { Left, Right }

  internal fun resolveTvPlayerHorizontalActionTarget(
      current: TvPlayerActionTarget,
      direction: TvPlayerActionDirection,
      targets: List<TvPlayerActionTarget>,
  ): TvPlayerActionTarget {
      if (targets.isEmpty()) return current
      val index = targets.indexOf(current)
      if (index < 0) return targets.first()
      val next = when (direction) {
          TvPlayerActionDirection.Right -> (index + 1) % targets.size
          TvPlayerActionDirection.Left -> (index - 1 + targets.size) % targets.size
      }
      return targets[next]
  }

  private val TvPlayerSeekStepWhitelist = listOf(5, 10, 15, 20, 30)

  internal fun normalizeTvPlayerSeekStepSeconds(seconds: Int): Int =
      if (seconds in TvPlayerSeekStepWhitelist) seconds else 10

  internal fun formatTvPlayerTime(millis: Long): String {
      val totalSeconds = (millis.coerceAtLeast(0L) / 1000L)
      val hours = totalSeconds / 3600L
      val minutes = (totalSeconds % 3600L) / 60L
      val seconds = totalSeconds % 60L
      return if (hours > 0L) {
          String.format("%d:%02d:%02d", hours, minutes, seconds)
      } else {
          String.format("%02d:%02d", minutes, seconds)
      }
  }

  internal object TvPlayerEpisodeRailLayoutTokens {
      val ItemSlotWidthDp = 120.dp
      val TooltipHeightDp = 44.dp
      val TooltipMaxWidthDp = 220.dp
  }

  @Composable
  internal fun TvPlayerIconButton(
      icon: ImageVector,
      contentDescription: String,
      onClick: () -> Unit,
      modifier: Modifier = Modifier,
      focusRequester: FocusRequester? = null,
  ) {
      val focusModifier = focusRequester?.let { Modifier.focusRequester(it) } ?: Modifier
      TvIconActionButton(
          icon = icon,
          contentDescription = contentDescription,
          onClick = onClick,
          modifier = modifier.then(focusModifier),
          size = 42.dp,
          iconSize = 24.dp,
          focusedScale = 1.12f,
      )
  }
  ```
- [ ] `TvIconActionButton`（`core/ui/TvIconAction.kt:19`）的真实签名是 `(icon, contentDescription, onClick, modifier, iconModifier, size, iconSize, shape, containerColor, contentColor, focusedScale)` —— **没有 `enabled` 参数**，且它自己内部就 `.size(size)` 并挂 `tvFocusableScaleOnly`。不要在外层再叠加 `.size()` 或 `.focusable()`。
- [ ] 跑 `./gradlew --no-daemon :tv-app:testDebugUnitTest --tests '*TvPlayerTokensTest'`，确认通过。
- [ ] commit：`重构：新增 TV 播放器控制层基础 token 与工具`

---

## Task 2：scrub 数学（`TvPlayerScrubMath.kt`）

**Files:**
- 新增 `.../core/ui/player/TvPlayerScrubMath.kt`
- 新增 `.../test/java/com/chee/videos/core/ui/player/TvPlayerScrubMathTest.kt`

**Interfaces:**

```kotlin
internal const val TvPlayerScrubIdleCommitMillis = 1_500L

internal data class TvPlayerScrubState(
    val anchorMs: Long,
    val targetMs: Long,
    val durationMs: Long,
) {
    val offsetMs: Long get() = targetMs - anchorMs
}

internal fun canTvPlayerScrub(durationMs: Long): Boolean
internal fun resolveTvPlayerScrubSpeedFactor(repeatCount: Int): Int
internal fun resolveTvPlayerScrubStepMs(seekStepSeconds: Int, repeatCount: Int): Long
internal fun beginTvPlayerScrub(currentPositionMs: Long, durationMs: Long): TvPlayerScrubState
internal fun advanceTvPlayerScrub(state: TvPlayerScrubState, deltaMs: Long): TvPlayerScrubState
internal fun formatTvPlayerScrubOffsetLabel(offsetMs: Long): String
```

**Steps:**

- [ ] 写 `TvPlayerScrubMathTest.kt`：
  ```kotlin
  package com.chee.videos.core.ui.player

  import org.junit.Assert.assertEquals
  import org.junit.Assert.assertFalse
  import org.junit.Assert.assertTrue
  import org.junit.Test

  class TvPlayerScrubMathTest {
      @Test
      fun scrubRequiresKnownDuration() {
          assertFalse(canTvPlayerScrub(0L))
          assertFalse(canTvPlayerScrub(-1L))
          assertTrue(canTvPlayerScrub(1L))
      }

      @Test
      fun speedFactorHasThreeTiers() {
          listOf(0, 1, 2).forEach { assertEquals(1, resolveTvPlayerScrubSpeedFactor(it)) }
          listOf(3, 5, 7).forEach { assertEquals(3, resolveTvPlayerScrubSpeedFactor(it)) }
          listOf(8, 20).forEach { assertEquals(6, resolveTvPlayerScrubSpeedFactor(it)) }
      }

      @Test
      fun stepCombinesNormalizedSecondsAndSpeedFactor() {
          assertEquals(15_000L, resolveTvPlayerScrubStepMs(15, repeatCount = 0))
          assertEquals(45_000L, resolveTvPlayerScrubStepMs(15, repeatCount = 3))
          assertEquals(90_000L, resolveTvPlayerScrubStepMs(15, repeatCount = 8))
          // 非白名单步长退化为 10 秒
          assertEquals(10_000L, resolveTvPlayerScrubStepMs(7, repeatCount = 0))
      }

      @Test
      fun beginCapturesAnchorOnceAndClampsPosition() {
          val state = beginTvPlayerScrub(currentPositionMs = 30_000L, durationMs = 600_000L)
          assertEquals(30_000L, state.anchorMs)
          assertEquals(30_000L, state.targetMs)
          assertEquals(600_000L, state.durationMs)
          assertEquals(0L, state.offsetMs)

          val clamped = beginTvPlayerScrub(currentPositionMs = -5_000L, durationMs = 600_000L)
          assertEquals(0L, clamped.anchorMs)

          val overshoot = beginTvPlayerScrub(currentPositionMs = 900_000L, durationMs = 600_000L)
          assertEquals(600_000L, overshoot.anchorMs)
      }

      @Test
      fun advanceAccumulatesFromTargetNotAnchor() {
          var state = beginTvPlayerScrub(30_000L, 600_000L)
          state = advanceTvPlayerScrub(state, 15_000L)
          state = advanceTvPlayerScrub(state, 15_000L)
          assertEquals(30_000L, state.anchorMs)
          assertEquals(60_000L, state.targetMs)
          assertEquals(30_000L, state.offsetMs)
      }

      @Test
      fun advanceClampsToBothEnds() {
          val head = advanceTvPlayerScrub(beginTvPlayerScrub(5_000L, 600_000L), -60_000L)
          assertEquals(0L, head.targetMs)

          val tail = advanceTvPlayerScrub(beginTvPlayerScrub(590_000L, 600_000L), 60_000L)
          assertEquals(600_000L, tail.targetMs)
      }

      @Test
      fun offsetLabelCarriesSignAndPadding() {
          assertEquals("+00:00", formatTvPlayerScrubOffsetLabel(0L))
          assertEquals("+01:30", formatTvPlayerScrubOffsetLabel(90_000L))
          assertEquals("-00:45", formatTvPlayerScrubOffsetLabel(-45_000L))
          assertEquals("-1:00:00", formatTvPlayerScrubOffsetLabel(-3_600_000L))
      }
  }
  ```
- [ ] 跑 `--tests '*TvPlayerScrubMathTest'`，确认编译失败。
- [ ] 写 `TvPlayerScrubMath.kt`（**零 Compose import**）：
  ```kotlin
  package com.chee.videos.core.ui.player

  /** scrub 停手后自动提交跳转的等待时长。 */
  internal const val TvPlayerScrubIdleCommitMillis = 1_500L

  /**
   * scrub 会话状态。
   * [anchorMs] 是进入 scrub 时的播放位置，只在进入时取一次；BACK 取消时回到它（且不执行 seek）。
   */
  internal data class TvPlayerScrubState(
      val anchorMs: Long,
      val targetMs: Long,
      val durationMs: Long,
  ) {
      val offsetMs: Long get() = targetMs - anchorMs
  }

  internal fun canTvPlayerScrub(durationMs: Long): Boolean = durationMs > 0L

  internal fun resolveTvPlayerScrubSpeedFactor(repeatCount: Int): Int = when {
      repeatCount <= 2 -> 1
      repeatCount <= 7 -> 3
      else -> 6
  }

  internal fun resolveTvPlayerScrubStepMs(seekStepSeconds: Int, repeatCount: Int): Long =
      normalizeTvPlayerSeekStepSeconds(seekStepSeconds) * 1_000L *
          resolveTvPlayerScrubSpeedFactor(repeatCount)

  internal fun beginTvPlayerScrub(currentPositionMs: Long, durationMs: Long): TvPlayerScrubState {
      val anchor = currentPositionMs.coerceIn(0L, durationMs.coerceAtLeast(0L))
      return TvPlayerScrubState(anchorMs = anchor, targetMs = anchor, durationMs = durationMs)
  }

  internal fun advanceTvPlayerScrub(
      state: TvPlayerScrubState,
      deltaMs: Long,
  ): TvPlayerScrubState = state.copy(
      targetMs = (state.targetMs + deltaMs).coerceIn(0L, state.durationMs.coerceAtLeast(0L)),
  )

  internal fun formatTvPlayerScrubOffsetLabel(offsetMs: Long): String {
      val sign = if (offsetMs < 0L) "-" else "+"
      return sign + formatTvPlayerTime(kotlin.math.abs(offsetMs))
  }
  ```
- [ ] 跑 `--tests '*TvPlayerScrubMathTest'`，确认通过。
- [ ] commit：`重构：新增 TV 播放器 scrub 数学模型`

---

## Task 3：交互状态机 reducer（`TvPlayerInteraction.kt`）

**Files:**
- 新增 `.../core/ui/player/TvPlayerInteraction.kt`
- 新增 `.../test/java/com/chee/videos/core/ui/player/TvPlayerInteractionTest.kt`

**Interfaces:**

```kotlin
internal enum class TvPlayerUiMode { Hidden, Chrome, Scrubbing, EpisodeRail }
internal enum class TvPlayerFocusLayer { Root, Controls, EpisodeRail }

internal const val TvPlayerChromeAutoHideMillis = 5_000L

internal data class TvPlayerInteractionState(
    val mode: TvPlayerUiMode = TvPlayerUiMode.Hidden,
    val focusLayer: TvPlayerFocusLayer = TvPlayerFocusLayer.Root,
    val scrub: TvPlayerScrubState? = null,
)

internal data class TvPlayerKeyInput(
    val keyCode: Int,
    val repeatCount: Int = 0,
    val positionMs: Long = 0L,
    val durationMs: Long = 0L,
    val seekStepSeconds: Int = 10,
    val episodeRailEnabled: Boolean = false,
    val overlayVisible: Boolean = false,
)

internal sealed interface TvPlayerEffect {
    data class CommitSeek(val targetMs: Long) : TvPlayerEffect
    data object TogglePlayPause : TvPlayerEffect
    data class RequestFocus(val layer: TvPlayerFocusLayer) : TvPlayerEffect
    data object OpenEpisodeRail : TvPlayerEffect
    data object CloseEpisodeRail : TvPlayerEffect
    data object AnnounceSeekUnavailable : TvPlayerEffect
}

internal data class TvPlayerInteractionResult(
    val state: TvPlayerInteractionState,
    val effects: List<TvPlayerEffect> = emptyList(),
    val handled: Boolean = true,
)

internal fun reduceTvPlayerKey(
    state: TvPlayerInteractionState,
    input: TvPlayerKeyInput,
): TvPlayerInteractionResult

internal fun reduceTvPlayerScrubIdleCommit(
    state: TvPlayerInteractionState,
): TvPlayerInteractionResult

internal fun shouldFreezeTvPlayerAutoHide(
    state: TvPlayerInteractionState,
    isPlaying: Boolean,
    overlayVisible: Boolean,
): Boolean
```

**语义合同（规格 §3.1 / §3.2 的可执行版本）：**

| 当前 | 按键 | 结果 |
|---|---|---|
| 任意 | 任意（`overlayVisible = true`） | `handled = false` 透传给上层弹层 |
| Root 层 + 任意 mode | LEFT / RIGHT / REWIND / FF | 可 seek → `Scrubbing` + 累加；不可 seek → `Chrome` + `AnnounceSeekUnavailable` |
| Controls / EpisodeRail 层 | LEFT / RIGHT | `handled = false`（交给控制条自己切焦点） |
| Root，任意 mode | DOWN | `Chrome` + `focusLayer = Controls` + `RequestFocus(Controls)`；若在 `Scrubbing` 先 `CommitSeek` |
| Controls | DOWN（`episodeRailEnabled`） | `EpisodeRail` + `OpenEpisodeRail` + `RequestFocus(EpisodeRail)` |
| Controls | DOWN（无分集） | `handled = false` |
| EpisodeRail | UP | `Chrome` + `focusLayer = Controls` + `CloseEpisodeRail` + `RequestFocus(Controls)` |
| Controls | UP | `Chrome` + `focusLayer = Root` + `RequestFocus(Root)` |
| Root | UP | `Chrome`（仅显示，不发焦点副作用） |
| Root + `Scrubbing` | OK / ENTER / PLAY_PAUSE | `CommitSeek(scrub.targetMs)`，`scrub = null`，转 `Chrome` |
| Root + 非 Scrubbing | OK / ENTER / PLAY_PAUSE | `TogglePlayPause` + `Chrome` |
| Controls / EpisodeRail | OK / ENTER | `handled = false`（按钮自己处理） |
| 任意层 + `Scrubbing` | BACK / ESCAPE | 丢弃 `scrub`（**不发 CommitSeek**），转 `Chrome` |
| `EpisodeRail` | BACK | `Chrome` + `focusLayer = Controls` + `CloseEpisodeRail` + `RequestFocus(Controls)` |
| `Chrome` | BACK | `Hidden` + `focusLayer = Root` + `RequestFocus(Root)` |
| `Hidden` | BACK | `handled = false`（交给宿主既有的双击退出确认） |

**Steps:**

- [ ] 写 `TvPlayerInteractionTest.kt`（JVM 单测里可直接用 `android.view.KeyEvent.KEYCODE_*`，它们是编译期内联的 `static final int`，既有 `TvLongFormRemoteKeyRoutingTest` 已验证）：
  ```kotlin
  package com.chee.videos.core.ui.player

  import android.view.KeyEvent
  import org.junit.Assert.assertEquals
  import org.junit.Assert.assertFalse
  import org.junit.Assert.assertNull
  import org.junit.Assert.assertTrue
  import org.junit.Test

  class TvPlayerInteractionTest {
      private val hidden = TvPlayerInteractionState()

      private fun input(
          keyCode: Int,
          repeatCount: Int = 0,
          positionMs: Long = 30_000L,
          durationMs: Long = 600_000L,
          seekStepSeconds: Int = 10,
          episodeRailEnabled: Boolean = false,
          overlayVisible: Boolean = false,
      ) = TvPlayerKeyInput(
          keyCode = keyCode,
          repeatCount = repeatCount,
          positionMs = positionMs,
          durationMs = durationMs,
          seekStepSeconds = seekStepSeconds,
          episodeRailEnabled = episodeRailEnabled,
          overlayVisible = overlayVisible,
      )

      @Test
      fun overlayVisiblePassesEveryKeyThrough() {
          val keys = listOf(
              KeyEvent.KEYCODE_DPAD_LEFT,
              KeyEvent.KEYCODE_DPAD_DOWN,
              KeyEvent.KEYCODE_DPAD_CENTER,
              KeyEvent.KEYCODE_BACK,
          )
          keys.forEach { key ->
              val result = reduceTvPlayerKey(hidden, input(key, overlayVisible = true))
              assertFalse(result.handled)
              assertEquals(hidden, result.state)
              assertTrue(result.effects.isEmpty())
          }
      }

      @Test
      fun rightFromHiddenEntersScrubbingDirectly() {
          val result = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT))
          assertEquals(TvPlayerUiMode.Scrubbing, result.state.mode)
          assertEquals(TvPlayerFocusLayer.Root, result.state.focusLayer)
          assertEquals(30_000L, result.state.scrub?.anchorMs)
          assertEquals(40_000L, result.state.scrub?.targetMs)
          assertTrue(result.effects.isEmpty())
      }

      @Test
      fun leftFromHiddenEntersScrubbingBackwards() {
          val result = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_LEFT))
          assertEquals(20_000L, result.state.scrub?.targetMs)
      }

      @Test
      fun mediaTransportKeysAlsoDriveScrubbing() {
          val forward = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_MEDIA_FAST_FORWARD))
          assertEquals(40_000L, forward.state.scrub?.targetMs)
          val backward = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_MEDIA_REWIND))
          assertEquals(20_000L, backward.state.scrub?.targetMs)
      }

      @Test
      fun scrubbingAccumulatesAndKeepsAnchor() {
          var state = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT)).state
          state = reduceTvPlayerKey(state, input(KeyEvent.KEYCODE_DPAD_RIGHT)).state
          state = reduceTvPlayerKey(state, input(KeyEvent.KEYCODE_DPAD_RIGHT)).state
          assertEquals(30_000L, state.scrub?.anchorMs)
          assertEquals(60_000L, state.scrub?.targetMs)
      }

      @Test
      fun longPressAcceleratesScrubStep() {
          val fast = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT, repeatCount = 8))
          assertEquals(30_000L + 60_000L, fast.state.scrub?.targetMs)
      }

      @Test
      fun scrubIsRejectedWhenDurationUnknown() {
          val result = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT, durationMs = 0L))
          assertEquals(TvPlayerUiMode.Chrome, result.state.mode)
          assertNull(result.state.scrub)
          assertEquals(listOf(TvPlayerEffect.AnnounceSeekUnavailable), result.effects)
      }

      @Test
      fun horizontalKeysArePassedThroughOutsideRootLayer() {
          val controls = TvPlayerInteractionState(
              mode = TvPlayerUiMode.Chrome,
              focusLayer = TvPlayerFocusLayer.Controls,
          )
          val result = reduceTvPlayerKey(controls, input(KeyEvent.KEYCODE_DPAD_RIGHT))
          assertFalse(result.handled)
          assertEquals(controls, result.state)
      }

      @Test
      fun okCommitsScrubTarget() {
          val scrubbing = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT)).state
          val result = reduceTvPlayerKey(scrubbing, input(KeyEvent.KEYCODE_DPAD_CENTER))
          assertEquals(TvPlayerUiMode.Chrome, result.state.mode)
          assertNull(result.state.scrub)
          assertEquals(listOf(TvPlayerEffect.CommitSeek(40_000L)), result.effects)
      }

      @Test
      fun backCancelsScrubWithoutSeeking() {
          val scrubbing = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT)).state
          val result = reduceTvPlayerKey(scrubbing, input(KeyEvent.KEYCODE_BACK))
          assertEquals(TvPlayerUiMode.Chrome, result.state.mode)
          assertNull(result.state.scrub)
          assertTrue(result.effects.none { it is TvPlayerEffect.CommitSeek })
          assertTrue(result.handled)
      }

      @Test
      fun idleCommitSeeksToScrubTarget() {
          val scrubbing = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT)).state
          val result = reduceTvPlayerScrubIdleCommit(scrubbing)
          assertEquals(TvPlayerUiMode.Chrome, result.state.mode)
          assertNull(result.state.scrub)
          assertEquals(listOf(TvPlayerEffect.CommitSeek(40_000L)), result.effects)
      }

      @Test
      fun idleCommitIsNoOpOutsideScrubbing() {
          val result = reduceTvPlayerScrubIdleCommit(hidden)
          assertEquals(hidden, result.state)
          assertTrue(result.effects.isEmpty())
      }

      @Test
      fun okTogglesPlaybackWhenNotScrubbing() {
          val result = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_CENTER))
          assertEquals(TvPlayerUiMode.Chrome, result.state.mode)
          assertEquals(listOf(TvPlayerEffect.TogglePlayPause), result.effects)
      }

      @Test
      fun okIsPassedThroughWhenControlsHoldFocus() {
          val controls = TvPlayerInteractionState(
              mode = TvPlayerUiMode.Chrome,
              focusLayer = TvPlayerFocusLayer.Controls,
          )
          assertFalse(reduceTvPlayerKey(controls, input(KeyEvent.KEYCODE_DPAD_CENTER)).handled)
      }

      @Test
      fun downFromRootMovesFocusIntoControls() {
          val result = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_DOWN))
          assertEquals(TvPlayerUiMode.Chrome, result.state.mode)
          assertEquals(TvPlayerFocusLayer.Controls, result.state.focusLayer)
          assertEquals(
              listOf(TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Controls)),
              result.effects,
          )
      }

      @Test
      fun downFromScrubbingCommitsBeforeEnteringControls() {
          val scrubbing = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_RIGHT)).state
          val result = reduceTvPlayerKey(scrubbing, input(KeyEvent.KEYCODE_DPAD_DOWN))
          assertEquals(
              listOf(
                  TvPlayerEffect.CommitSeek(40_000L),
                  TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Controls),
              ),
              result.effects,
          )
          assertNull(result.state.scrub)
          assertEquals(TvPlayerFocusLayer.Controls, result.state.focusLayer)
      }

      @Test
      fun downFromControlsOpensEpisodeRailOnlyForSeries() {
          val controls = TvPlayerInteractionState(
              mode = TvPlayerUiMode.Chrome,
              focusLayer = TvPlayerFocusLayer.Controls,
          )
          val opened = reduceTvPlayerKey(
              controls,
              input(KeyEvent.KEYCODE_DPAD_DOWN, episodeRailEnabled = true),
          )
          assertEquals(TvPlayerUiMode.EpisodeRail, opened.state.mode)
          assertEquals(TvPlayerFocusLayer.EpisodeRail, opened.state.focusLayer)
          assertEquals(
              listOf(
                  TvPlayerEffect.OpenEpisodeRail,
                  TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.EpisodeRail),
              ),
              opened.effects,
          )

          val ignored = reduceTvPlayerKey(controls, input(KeyEvent.KEYCODE_DPAD_DOWN))
          assertFalse(ignored.handled)
      }

      @Test
      fun upWalksBackThroughFocusLayers() {
          val rail = TvPlayerInteractionState(
              mode = TvPlayerUiMode.EpisodeRail,
              focusLayer = TvPlayerFocusLayer.EpisodeRail,
          )
          val toControls = reduceTvPlayerKey(rail, input(KeyEvent.KEYCODE_DPAD_UP))
          assertEquals(TvPlayerUiMode.Chrome, toControls.state.mode)
          assertEquals(TvPlayerFocusLayer.Controls, toControls.state.focusLayer)
          assertEquals(
              listOf(
                  TvPlayerEffect.CloseEpisodeRail,
                  TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Controls),
              ),
              toControls.effects,
          )

          val toRoot = reduceTvPlayerKey(toControls.state, input(KeyEvent.KEYCODE_DPAD_UP))
          assertEquals(TvPlayerFocusLayer.Root, toRoot.state.focusLayer)
          assertEquals(
              listOf(TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Root)),
              toRoot.effects,
          )
      }

      @Test
      fun upFromRootRevealsChromeWithoutMovingFocus() {
          val result = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_DPAD_UP))
          assertEquals(TvPlayerUiMode.Chrome, result.state.mode)
          assertEquals(TvPlayerFocusLayer.Root, result.state.focusLayer)
          assertTrue(result.effects.isEmpty())
      }

      @Test
      fun backCollapsesLayerByLayerAndFinallyPassesThrough() {
          val rail = TvPlayerInteractionState(
              mode = TvPlayerUiMode.EpisodeRail,
              focusLayer = TvPlayerFocusLayer.EpisodeRail,
          )
          val chrome = reduceTvPlayerKey(rail, input(KeyEvent.KEYCODE_BACK))
          assertEquals(TvPlayerUiMode.Chrome, chrome.state.mode)
          assertTrue(chrome.effects.contains(TvPlayerEffect.CloseEpisodeRail))

          val hiddenAgain = reduceTvPlayerKey(chrome.state, input(KeyEvent.KEYCODE_BACK))
          assertEquals(TvPlayerUiMode.Hidden, hiddenAgain.state.mode)
          assertEquals(TvPlayerFocusLayer.Root, hiddenAgain.state.focusLayer)

          val exit = reduceTvPlayerKey(hiddenAgain.state, input(KeyEvent.KEYCODE_BACK))
          assertFalse(exit.handled)
      }

      @Test
      fun escapeBehavesLikeBack() {
          val chrome = TvPlayerInteractionState(mode = TvPlayerUiMode.Chrome)
          val result = reduceTvPlayerKey(chrome, input(KeyEvent.KEYCODE_ESCAPE))
          assertEquals(TvPlayerUiMode.Hidden, result.state.mode)
      }

      @Test
      fun unknownKeysArePassedThrough() {
          val result = reduceTvPlayerKey(hidden, input(KeyEvent.KEYCODE_A))
          assertFalse(result.handled)
          assertEquals(hidden, result.state)
      }

      @Test
      fun autoHideFreezesWhileScrubbingRailOverlayOrPaused() {
          val chrome = TvPlayerInteractionState(mode = TvPlayerUiMode.Chrome)
          assertFalse(shouldFreezeTvPlayerAutoHide(chrome, isPlaying = true, overlayVisible = false))
          assertTrue(shouldFreezeTvPlayerAutoHide(chrome, isPlaying = false, overlayVisible = false))
          assertTrue(shouldFreezeTvPlayerAutoHide(chrome, isPlaying = true, overlayVisible = true))
          assertTrue(
              shouldFreezeTvPlayerAutoHide(
                  chrome.copy(mode = TvPlayerUiMode.Scrubbing),
                  isPlaying = true,
                  overlayVisible = false,
              ),
          )
          assertTrue(
              shouldFreezeTvPlayerAutoHide(
                  chrome.copy(mode = TvPlayerUiMode.EpisodeRail),
                  isPlaying = true,
                  overlayVisible = false,
              ),
          )
      }
  }
  ```
- [ ] 跑 `--tests '*TvPlayerInteractionTest'`，确认编译失败。
- [ ] 写 `TvPlayerInteraction.kt`（**只允许 import `android.view.KeyEvent`，零 Compose**）：
  ```kotlin
  package com.chee.videos.core.ui.player

  import android.view.KeyEvent

  internal enum class TvPlayerUiMode { Hidden, Chrome, Scrubbing, EpisodeRail }

  internal enum class TvPlayerFocusLayer { Root, Controls, EpisodeRail }

  /** Chrome 无操作后自动隐藏的等待时长。 */
  internal const val TvPlayerChromeAutoHideMillis = 5_000L

  internal data class TvPlayerInteractionState(
      val mode: TvPlayerUiMode = TvPlayerUiMode.Hidden,
      val focusLayer: TvPlayerFocusLayer = TvPlayerFocusLayer.Root,
      val scrub: TvPlayerScrubState? = null,
  )

  internal data class TvPlayerKeyInput(
      val keyCode: Int,
      val repeatCount: Int = 0,
      val positionMs: Long = 0L,
      val durationMs: Long = 0L,
      val seekStepSeconds: Int = 10,
      val episodeRailEnabled: Boolean = false,
      val overlayVisible: Boolean = false,
  )

  internal sealed interface TvPlayerEffect {
      data class CommitSeek(val targetMs: Long) : TvPlayerEffect
      data object TogglePlayPause : TvPlayerEffect
      data class RequestFocus(val layer: TvPlayerFocusLayer) : TvPlayerEffect
      data object OpenEpisodeRail : TvPlayerEffect
      data object CloseEpisodeRail : TvPlayerEffect
      data object AnnounceSeekUnavailable : TvPlayerEffect
  }

  internal data class TvPlayerInteractionResult(
      val state: TvPlayerInteractionState,
      val effects: List<TvPlayerEffect> = emptyList(),
      val handled: Boolean = true,
  )

  private fun passThrough(state: TvPlayerInteractionState) =
      TvPlayerInteractionResult(state = state, handled = false)

  internal fun reduceTvPlayerKey(
      state: TvPlayerInteractionState,
      input: TvPlayerKeyInput,
  ): TvPlayerInteractionResult {
      if (input.overlayVisible) return passThrough(state)
      return when (input.keyCode) {
          KeyEvent.KEYCODE_DPAD_LEFT,
          KeyEvent.KEYCODE_MEDIA_REWIND,
          -> reduceHorizontal(state, input, forward = false)

          KeyEvent.KEYCODE_DPAD_RIGHT,
          KeyEvent.KEYCODE_MEDIA_FAST_FORWARD,
          -> reduceHorizontal(state, input, forward = true)

          KeyEvent.KEYCODE_DPAD_DOWN -> reduceDown(state, input)
          KeyEvent.KEYCODE_DPAD_UP -> reduceUp(state)

          KeyEvent.KEYCODE_DPAD_CENTER,
          KeyEvent.KEYCODE_ENTER,
          KeyEvent.KEYCODE_NUMPAD_ENTER,
          KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
          -> reduceConfirm(state)

          KeyEvent.KEYCODE_BACK,
          KeyEvent.KEYCODE_ESCAPE,
          -> reduceBack(state)

          else -> passThrough(state)
      }
  }

  private fun reduceHorizontal(
      state: TvPlayerInteractionState,
      input: TvPlayerKeyInput,
      forward: Boolean,
  ): TvPlayerInteractionResult {
      // 控制条/分集轨持焦时，左右键是它们自己的横向导航。
      if (state.focusLayer != TvPlayerFocusLayer.Root) return passThrough(state)
      if (!canTvPlayerScrub(input.durationMs)) {
          return TvPlayerInteractionResult(
              state = state.copy(mode = TvPlayerUiMode.Chrome, scrub = null),
              effects = listOf(TvPlayerEffect.AnnounceSeekUnavailable),
          )
      }
      val stepMs = resolveTvPlayerScrubStepMs(input.seekStepSeconds, input.repeatCount)
      val deltaMs = if (forward) stepMs else -stepMs
      val base = state.scrub?.takeIf { state.mode == TvPlayerUiMode.Scrubbing }
          ?: beginTvPlayerScrub(input.positionMs, input.durationMs)
      return TvPlayerInteractionResult(
          state = state.copy(
              mode = TvPlayerUiMode.Scrubbing,
              focusLayer = TvPlayerFocusLayer.Root,
              scrub = advanceTvPlayerScrub(base, deltaMs),
          ),
      )
  }

  private fun reduceDown(
      state: TvPlayerInteractionState,
      input: TvPlayerKeyInput,
  ): TvPlayerInteractionResult = when (state.focusLayer) {
      TvPlayerFocusLayer.Root -> {
          val commit = state.scrub
              ?.takeIf { state.mode == TvPlayerUiMode.Scrubbing }
              ?.let { listOf(TvPlayerEffect.CommitSeek(it.targetMs)) }
              .orEmpty()
          TvPlayerInteractionResult(
              state = state.copy(
                  mode = TvPlayerUiMode.Chrome,
                  focusLayer = TvPlayerFocusLayer.Controls,
                  scrub = null,
              ),
              effects = commit + TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Controls),
          )
      }

      TvPlayerFocusLayer.Controls -> if (input.episodeRailEnabled) {
          TvPlayerInteractionResult(
              state = state.copy(
                  mode = TvPlayerUiMode.EpisodeRail,
                  focusLayer = TvPlayerFocusLayer.EpisodeRail,
              ),
              effects = listOf(
                  TvPlayerEffect.OpenEpisodeRail,
                  TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.EpisodeRail),
              ),
          )
      } else {
          passThrough(state)
      }

      TvPlayerFocusLayer.EpisodeRail -> passThrough(state)
  }

  private fun reduceUp(state: TvPlayerInteractionState): TvPlayerInteractionResult =
      when (state.focusLayer) {
          TvPlayerFocusLayer.EpisodeRail -> TvPlayerInteractionResult(
              state = state.copy(
                  mode = TvPlayerUiMode.Chrome,
                  focusLayer = TvPlayerFocusLayer.Controls,
              ),
              effects = listOf(
                  TvPlayerEffect.CloseEpisodeRail,
                  TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Controls),
              ),
          )

          TvPlayerFocusLayer.Controls -> TvPlayerInteractionResult(
              state = state.copy(
                  mode = TvPlayerUiMode.Chrome,
                  focusLayer = TvPlayerFocusLayer.Root,
              ),
              effects = listOf(TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Root)),
          )

          // Root 层的 UP 只唤起 Chrome，不夺焦点，避免遥控器上键把焦点吸进控制条。
          TvPlayerFocusLayer.Root -> TvPlayerInteractionResult(
              state = state.copy(mode = TvPlayerUiMode.Chrome),
          )
      }

  private fun reduceConfirm(state: TvPlayerInteractionState): TvPlayerInteractionResult {
      if (state.focusLayer != TvPlayerFocusLayer.Root) return passThrough(state)
      val scrub = state.scrub?.takeIf { state.mode == TvPlayerUiMode.Scrubbing }
      return if (scrub != null) {
          TvPlayerInteractionResult(
              state = state.copy(mode = TvPlayerUiMode.Chrome, scrub = null),
              effects = listOf(TvPlayerEffect.CommitSeek(scrub.targetMs)),
          )
      } else {
          TvPlayerInteractionResult(
              state = state.copy(mode = TvPlayerUiMode.Chrome),
              effects = listOf(TvPlayerEffect.TogglePlayPause),
          )
      }
  }

  private fun reduceBack(state: TvPlayerInteractionState): TvPlayerInteractionResult {
      // scrub 取消：丢弃目标位置，不发 CommitSeek，播放位置留在锚点。
      if (state.mode == TvPlayerUiMode.Scrubbing) {
          return TvPlayerInteractionResult(
              state = state.copy(mode = TvPlayerUiMode.Chrome, scrub = null),
          )
      }
      return when (state.mode) {
          TvPlayerUiMode.EpisodeRail -> TvPlayerInteractionResult(
              state = state.copy(
                  mode = TvPlayerUiMode.Chrome,
                  focusLayer = TvPlayerFocusLayer.Controls,
              ),
              effects = listOf(
                  TvPlayerEffect.CloseEpisodeRail,
                  TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Controls),
              ),
          )

          TvPlayerUiMode.Chrome -> TvPlayerInteractionResult(
              state = state.copy(
                  mode = TvPlayerUiMode.Hidden,
                  focusLayer = TvPlayerFocusLayer.Root,
              ),
              effects = listOf(TvPlayerEffect.RequestFocus(TvPlayerFocusLayer.Root)),
          )

          // Hidden 时交回宿主，走既有的双击退出确认契约。
          else -> passThrough(state)
      }
  }

  internal fun reduceTvPlayerScrubIdleCommit(
      state: TvPlayerInteractionState,
  ): TvPlayerInteractionResult {
      val scrub = state.scrub?.takeIf { state.mode == TvPlayerUiMode.Scrubbing }
          ?: return TvPlayerInteractionResult(state = state)
      return TvPlayerInteractionResult(
          state = state.copy(mode = TvPlayerUiMode.Chrome, scrub = null),
          effects = listOf(TvPlayerEffect.CommitSeek(scrub.targetMs)),
      )
  }

  internal fun shouldFreezeTvPlayerAutoHide(
      state: TvPlayerInteractionState,
      isPlaying: Boolean,
      overlayVisible: Boolean,
  ): Boolean = state.mode == TvPlayerUiMode.Scrubbing ||
      state.mode == TvPlayerUiMode.EpisodeRail ||
      overlayVisible ||
      !isPlaying
  ```
- [ ] 跑 `--tests '*TvPlayerInteractionTest'`，确认 24 个用例全绿。
- [ ] commit：`重构：新增 TV 播放器交互状态机 reducer`

---

## Task 4：缓冲态策略 + 播放内核回调

**Files:**
- 新增 `.../core/ui/player/TvPlayerBufferingPolicy.kt`
- 新增 `.../test/java/com/chee/videos/core/ui/player/TvPlayerBufferingPolicyTest.kt`
- 修改 `.../feature/tv/TvLongFormMedia3Player.kt`

**Interfaces:**

```kotlin
internal const val TvPlayerBufferingIndicatorDelayMillis = 350L
internal const val TvPlayerBufferingSlowNoticeMillis = 8_000L
internal const val TvPlayerBufferingFailureMillis = 20_000L
internal const val TvPlayerBufferingSlowNoticeMessage = "网络较慢，正在缓冲"

internal enum class TvPlayerBufferingStage { None, Spinner, SlowNotice, Failed }

internal fun resolveTvPlayerBufferingStage(bufferingElapsedMs: Long?): TvPlayerBufferingStage
```

`TvLongFormMedia3Player` 新增参数：

```kotlin
onBufferingChanged: (Boolean, TvLongFormMedia3EventIdentity) -> Unit = { _, _ -> },
```

**Steps:**

- [ ] 写 `TvPlayerBufferingPolicyTest.kt`：
  ```kotlin
  package com.chee.videos.core.ui.player

  import org.junit.Assert.assertEquals
  import org.junit.Test

  class TvPlayerBufferingPolicyTest {
      @Test
      fun notBufferingShowsNothing() {
          assertEquals(TvPlayerBufferingStage.None, resolveTvPlayerBufferingStage(null))
      }

      @Test
      fun shortBufferingIsSuppressedToAvoidFlicker() {
          assertEquals(TvPlayerBufferingStage.None, resolveTvPlayerBufferingStage(0L))
          assertEquals(TvPlayerBufferingStage.None, resolveTvPlayerBufferingStage(349L))
      }

      @Test
      fun spinnerAppearsAfterDelay() {
          assertEquals(TvPlayerBufferingStage.Spinner, resolveTvPlayerBufferingStage(350L))
          assertEquals(TvPlayerBufferingStage.Spinner, resolveTvPlayerBufferingStage(7_999L))
      }

      @Test
      fun slowNoticeAppearsAtEightSeconds() {
          assertEquals(TvPlayerBufferingStage.SlowNotice, resolveTvPlayerBufferingStage(8_000L))
          assertEquals(TvPlayerBufferingStage.SlowNotice, resolveTvPlayerBufferingStage(19_999L))
      }

      @Test
      fun failureIsReportedAtTwentySeconds() {
          assertEquals(TvPlayerBufferingStage.Failed, resolveTvPlayerBufferingStage(20_000L))
          assertEquals(TvPlayerBufferingStage.Failed, resolveTvPlayerBufferingStage(60_000L))
      }

      @Test
      fun thresholdsAreOrdered() {
          assert(TvPlayerBufferingIndicatorDelayMillis < TvPlayerBufferingSlowNoticeMillis)
          assert(TvPlayerBufferingSlowNoticeMillis < TvPlayerBufferingFailureMillis)
      }
  }
  ```
- [ ] 跑 `--tests '*TvPlayerBufferingPolicyTest'`，确认编译失败。
- [ ] 写 `TvPlayerBufferingPolicy.kt`（零 Compose）：
  ```kotlin
  package com.chee.videos.core.ui.player

  /** 缓冲进入后延迟这么久才显示指示器，避免 seek 后一闪而过的抖动。 */
  internal const val TvPlayerBufferingIndicatorDelayMillis = 350L

  /** 缓冲超过这个时长追加「网络较慢」提示。 */
  internal const val TvPlayerBufferingSlowNoticeMillis = 8_000L

  /** 缓冲超过这个时长视为失败，转入既有软重试流程。 */
  internal const val TvPlayerBufferingFailureMillis = 20_000L

  internal const val TvPlayerBufferingSlowNoticeMessage = "网络较慢，正在缓冲"

  internal enum class TvPlayerBufferingStage { None, Spinner, SlowNotice, Failed }

  /** [bufferingElapsedMs] 为 null 表示当前不在缓冲。 */
  internal fun resolveTvPlayerBufferingStage(bufferingElapsedMs: Long?): TvPlayerBufferingStage {
      val elapsed = bufferingElapsedMs ?: return TvPlayerBufferingStage.None
      return when {
          elapsed >= TvPlayerBufferingFailureMillis -> TvPlayerBufferingStage.Failed
          elapsed >= TvPlayerBufferingSlowNoticeMillis -> TvPlayerBufferingStage.SlowNotice
          elapsed >= TvPlayerBufferingIndicatorDelayMillis -> TvPlayerBufferingStage.Spinner
          else -> TvPlayerBufferingStage.None
      }
  }
  ```
- [ ] 跑 `--tests '*TvPlayerBufferingPolicyTest'`，确认通过。
- [ ] 改 `TvLongFormMedia3Player.kt`：
  - 在参数表加 `onBufferingChanged: (Boolean, TvLongFormMedia3EventIdentity) -> Unit = { _, _ -> },`
  - 按文件既有写法为它建 `rememberUpdatedState` 包装（照抄邻近的 `latestOnEnded` / `latestOnError` 命名与形态，命名为 `latestOnBufferingChanged`）
  - 在 `AnalyticsListener.onPlaybackStateChanged` 里 `playbackState = nextPlaybackState` 之后、`STATE_ENDED` 判断附近追加：
    ```kotlin
    latestOnBufferingChanged(
        nextPlaybackState == Player.STATE_BUFFERING,
        eventIdentity,
    )
    ```
    `eventIdentity` 取该监听器里既有的身份变量（与 `latestOnEnded` 分流用的同一个），保证过期播放器事件不会污染新会话的缓冲态。
- [ ] 跑 `./gradlew --no-daemon :tv-app:compileDebugKotlin`，确认编译通过（此时还没有调用方传新参数，靠默认值兼容）。
- [ ] commit：`重构：新增 TV 播放器缓冲态策略与内核回调`

---

## Task 5：Chrome 子组件（信息层 / 进度轨 / 操作行 / 中心反馈 / 缓冲指示 / 分集轨）

**Files:**
- 新增 `.../core/ui/player/TvPlayerTopInfo.kt`
- 新增 `.../core/ui/player/TvPlayerScrubTrack.kt`
- 新增 `.../core/ui/player/TvPlayerActionRow.kt`
- 新增 `.../core/ui/player/TvPlayerCenterFeedback.kt`
- 新增 `.../core/ui/player/TvPlayerBufferingIndicator.kt`
- 新增 `.../core/ui/player/TvPlayerEpisodeRail.kt`

**Interfaces:**

```kotlin
@Composable
internal fun TvPlayerTopInfo(
    primaryText: String,
    secondaryText: String?,
    visible: Boolean,
    modifier: Modifier = Modifier,
)

@Composable
internal fun TvPlayerScrubTrack(
    positionMs: Long,
    durationMs: Long,
    scrub: TvPlayerScrubState?,
    modifier: Modifier = Modifier,
)

@Composable
internal fun TvPlayerActionRow(
    isPlaying: Boolean,
    focusedTarget: TvPlayerActionTarget,
    targets: List<TvPlayerActionTarget>,
    focusRequester: FocusRequester,
    onFocusedTargetChange: (TvPlayerActionTarget) -> Unit,
    onTogglePlayPause: () -> Unit,
    onOpenSubtitle: () -> Unit,
    onOpenAudioTrack: () -> Unit,
    onOpenEpisodeList: () -> Unit,
    modifier: Modifier = Modifier,
)

@Composable
internal fun TvPlayerCenterFeedback(
    scrub: TvPlayerScrubState?,
    transientMessage: String?,
    persistentMessage: String?,
    modifier: Modifier = Modifier,
)

@Composable
internal fun TvPlayerBufferingIndicator(
    stage: TvPlayerBufferingStage,
    modifier: Modifier = Modifier,
)

@Composable
internal fun TvPlayerEpisodeRail(
    items: List<TvEpisodeRailItem>,
    currentItemId: String?,
    visible: Boolean,
    focusRequester: FocusRequester,
    onSelectItem: (TvEpisodeRailItem) -> Unit,
    modifier: Modifier = Modifier,
)
```

**Steps:**

- [ ] 写 `TvPlayerTopInfo.kt`：直接复用 `core/ui/TvLongFormTitleOverlay.kt` 的排版参数（左/上 24dp、主标题 22sp、副标题 16sp、文字阴影），但改为接受已拼好的 `primaryText` / `secondaryText` 两个字符串（不再依赖 `TvLongFormTitleOverlayData`），并用 `AnimatedVisibility` 跟随 `visible`。
- [ ] 写 `TvPlayerScrubTrack.kt`：单一 Row（当前时间 + 轨道 + 总时长），轨道高 6dp、`AppChrome.PillShape`：
  - 底轨 `TvPlayerProgressTrack`
  - 实心已播段宽度按 `positionMs / durationMs`，色 `TvPlayerProgressActiveTrack`（**始终跟随真实播放位置，scrub 期间不改**）
  - `scrub != null` 时在 `scrub.targetMs / durationMs` 处画 ghost 游标（宽 3dp、高 14dp、色 `TvPlayerScrubGhostTrack`），并把右侧时间文字换成 `formatTvPlayerTime(scrub.targetMs)`
  - 整个组件挂 `.focusProperties { canFocus = false }`，**不进焦点链**
  - `durationMs <= 0` 时实心段宽度按 0 处理，时间文字显示 `--:--`
- [ ] 写 `TvPlayerActionRow.kt`：横向 Row，按 `targets` 顺序渲染 `TvPlayerIconButton`；播放/暂停图标按 `isPlaying` 切换；`focusRequester` 只挂在 `targets.first()` 对应的按钮上（保证 `RequestFocus(Controls)` 落到确实已组合的节点）；每个按钮 `onFocusChanged` 里回调 `onFocusedTargetChange`。**不要**在按钮上叠加额外 `.focusable()`。
- [ ] 写 `TvPlayerCenterFeedback.kt`：
  - `scrub != null` → 大字显示 `formatTvPlayerTime(scrub.targetMs)` + 次行 `formatTvPlayerScrubOffsetLabel(scrub.offsetMs)`
  - 否则 `persistentMessage` 优先于 `transientMessage`
  - 全部包在 `TvPlayerGlassSurfaceStrong` 圆角卡片里，居中
- [ ] 写 `TvPlayerBufferingIndicator.kt`：
  - `None` → 不渲染
  - `Spinner` → `CircularProgressIndicator`
  - `SlowNotice` → spinner + `TvPlayerBufferingSlowNoticeMessage`
  - `Failed` → 不渲染（失败由宿主的软重试 UI 接手，避免两层错误提示叠加）
- [ ] 写 `TvPlayerEpisodeRail.kt`：从 `core/ui/LongFormVideoPlayer.kt:1497` 的 `TvEpisodeRail` 迁移实现，改用 `TvPlayerEpisodeRailLayoutTokens`，继续复用 `core/ui/TvEpisodeRailPolicy.kt` 的 `resolveEpisodeRailInitialFirstVisibleItemIndex` / `resolveEpisodeRailFollowScrollFirstVisibleItemIndex` / `formatTvEpisodeRailLabel`，`focusRequester` 挂在当前集对应的 item 上。
- [ ] 跑 `./gradlew --no-daemon :tv-app:compileDebugKotlin`，确认全部编译通过。
- [ ] commit：`重构：新增 TV 播放器控制层子组件`

---

## Task 6：组装 `TvPlayerChrome.kt` + 形态审计测试

**Files:**
- 新增 `.../core/ui/player/TvPlayerChrome.kt`
- 新增 `.../test/java/com/chee/videos/core/ui/player/TvPlayerChromeSpecTest.kt`

**Interfaces:**

```kotlin
@Composable
internal fun TvPlayerChrome(
    // 播放态
    isPlaying: Boolean,
    positionMs: Long,
    durationMs: Long,
    isBuffering: Boolean,
    seekStepSeconds: Int,
    // 标题信息
    primaryTitle: String,
    secondaryTitle: String?,
    // 分集
    episodeRailItems: List<TvEpisodeRailItem>,
    currentEpisodeRailItemId: String?,
    openEpisodeRailRequestKey: Int,
    episodeSwitchState: TvEpisodeSwitchUiState?,
    // 弹层可见性（决定按键是否透传）
    focusGuardInput: PlayerFocusGuardInput,
    // 回调
    onSeekTo: (Long) -> Unit,
    onTogglePlayPause: () -> Unit,
    onOpenSubtitle: () -> Unit,
    onOpenAudioTrack: () -> Unit,
    onSelectEpisodeRailItem: (TvEpisodeRailItem) -> Unit,
    onEpisodeRailVisibilityChanged: (Boolean) -> Unit,
    onDismissEpisodeSwitchFeedback: () -> Unit,
    onControlsVisibilityChanged: (Boolean) -> Unit,
    onScrubbingChanged: (Boolean) -> Unit,
    onBufferingTimeout: () -> Unit,
    onRequestExitPlayback: () -> Unit,
    // 插槽
    resumePromptSlot: (@Composable BoxScope.() -> Unit)? = null,
    modifier: Modifier = Modifier,
    content: @Composable BoxScope.() -> Unit,
)
```

> 参数表刻意贴近旧 `TvSeriesCorePlaybackOverlay`，让 Task 8 的宿主接线尽量接近 drop-in。新增的三个是 `isBuffering`、`onScrubbingChanged`（剧集宿主用它在 scrub 期间抑制自动连播提示卡）、`onBufferingTimeout`（20s 转既有软重试）。
>
> 跨包类型：`TvEpisodeSwitchUiState` 是 `com.chee.videos.feature.tv`（定义在 `TvSeriesPlayerViewModel.kt:47`，成员为 `Preparing` / `Succeeded` / `Canceled` / `Failed`）；`PlayerFocusGuardInput` / `anyOverlayVisible()` / `shouldReclaimRootFocus` 是 `com.chee.videos.core.ui`（`LongFormPlayerFocusGuard.kt`）；`TvEpisodeRailItem` 是 `com.chee.videos.core.ui`（`TvEpisodeRailPolicy.kt:3`）。旧覆盖层就是这么跨包引用的，沿用即可，不要为此再造一层映射。

**Steps:**

- [ ] 写 `TvPlayerChromeSpecTest.kt`（沿用项目既有的「读源码文本断言架构约束」风格）：
  ```kotlin
  package com.chee.videos.core.ui.player

  import java.nio.file.Path
  import org.junit.Assert.assertFalse
  import org.junit.Assert.assertTrue
  import org.junit.Test

  class TvPlayerChromeSpecTest {
      private fun source(name: String): String =
          Path.of("src/main/java/com/chee/videos/core/ui/player/$name").toFile().readText()

      @Test
      fun reducerLayerStaysFreeOfCompose() {
          listOf(
              "TvPlayerInteraction.kt",
              "TvPlayerScrubMath.kt",
              "TvPlayerBufferingPolicy.kt",
          ).forEach { name ->
              val text = source(name)
              assertFalse("$name 不应依赖 Compose", text.contains("import androidx.compose"))
              assertFalse("$name 不应依赖 Media3", text.contains("import androidx.media3"))
          }
      }

      @Test
      fun chromeStaysFreeOfPlaybackEngine() {
          listOf(
              "TvPlayerChrome.kt",
              "TvPlayerScrubTrack.kt",
              "TvPlayerActionRow.kt",
              "TvPlayerTopInfo.kt",
              "TvPlayerCenterFeedback.kt",
              "TvPlayerBufferingIndicator.kt",
              "TvPlayerEpisodeRail.kt",
          ).forEach { name ->
              assertFalse(
                  "$name 不应直接依赖播放内核",
                  source(name).contains("import androidx.media3"),
              )
          }
      }

      @Test
      fun scrubTrackIsNotFocusable() {
          assertTrue(source("TvPlayerScrubTrack.kt").contains("canFocus = false"))
      }

      @Test
      fun chromeRoutesKeysThroughReducer() {
          val text = source("TvPlayerChrome.kt")
          assertTrue(text.contains("onPreviewKeyEvent"))
          assertTrue(text.contains("reduceTvPlayerKey("))
          assertTrue(text.contains("reduceTvPlayerScrubIdleCommit("))
          assertTrue(text.contains("TvPlayerScrubIdleCommitMillis"))
          assertTrue(text.contains("TvPlayerChromeAutoHideMillis"))
          assertTrue(text.contains("shouldFreezeTvPlayerAutoHide("))
      }

      @Test
      fun chromeHandlesEveryReducerEffect() {
          val text = source("TvPlayerChrome.kt")
          listOf(
              "TvPlayerEffect.CommitSeek",
              "TvPlayerEffect.TogglePlayPause",
              "TvPlayerEffect.RequestFocus",
              "TvPlayerEffect.OpenEpisodeRail",
              "TvPlayerEffect.CloseEpisodeRail",
              "TvPlayerEffect.AnnounceSeekUnavailable",
          ).forEach { effect ->
              assertTrue("$effect 未被处理", text.contains(effect))
          }
      }

      @Test
      fun chromeDropsLegacyDebounceContract() {
          val text = source("TvPlayerChrome.kt")
          assertFalse(text.contains("TvStepSeekDebounceMillis"))
          assertFalse(text.contains("pendingStepSeek"))
      }

      @Test
      fun chromeKeepsEpisodeSwitchFeedbackStates() {
          val text = source("TvPlayerChrome.kt")
          listOf(
              "TvEpisodeSwitchUiState.Preparing",
              "TvEpisodeSwitchUiState.Succeeded",
              "TvEpisodeSwitchUiState.Canceled",
              "TvEpisodeSwitchUiState.Failed",
              "persistentCenterFailureMessage",
              "text = \"留在当前分集\"",
          ).forEach { assertTrue(it, text.contains(it)) }
      }

      @Test
      fun chromeReportsBufferingTimeout() {
          val text = source("TvPlayerChrome.kt")
          assertTrue(text.contains("TvPlayerBufferingStage.Failed"))
          assertTrue(text.contains("onBufferingTimeout"))
      }
  }
  ```
- [ ] 跑 `--tests '*TvPlayerChromeSpecTest'`，确认失败（文件不存在）。
- [ ] 写 `TvPlayerChrome.kt`，结构如下（实现要点，不是伪代码占位——按此结构写完整实现）：
  1. 单一状态源：`var interaction by remember { mutableStateOf(TvPlayerInteractionState()) }`
  2. 焦点句柄：`val rootFocusRequester = remember { FocusRequester() }`、`controlsFocusRequester`、`episodeRailFocusRequester`
  3. 效果分发函数：
     ```kotlin
     fun dispatch(effects: List<TvPlayerEffect>) {
         effects.forEach { effect ->
             when (effect) {
                 is TvPlayerEffect.CommitSeek -> onSeekTo(effect.targetMs)
                 TvPlayerEffect.TogglePlayPause -> onTogglePlayPause()
                 is TvPlayerEffect.RequestFocus -> when (effect.layer) {
                     TvPlayerFocusLayer.Root -> rootFocusRequester.tryRequestFocus()
                     TvPlayerFocusLayer.Controls -> controlsFocusRequester.tryRequestFocus()
                     TvPlayerFocusLayer.EpisodeRail -> episodeRailFocusRequester.tryRequestFocus()
                 }
                 TvPlayerEffect.OpenEpisodeRail -> onEpisodeRailVisibilityChanged(true)
                 TvPlayerEffect.CloseEpisodeRail -> onEpisodeRailVisibilityChanged(false)
                 TvPlayerEffect.AnnounceSeekUnavailable ->
                     transientMessage = "该内容暂不支持快进"
             }
         }
     }
     ```
     注意 `RequestFocus` 必须放到 `LaunchedEffect` 里等一帧再请求（目标节点在同一帧才刚被 `AnimatedVisibility` 挂载）：用 `var pendingFocusLayer by remember { mutableStateOf<TvPlayerFocusLayer?>(null) }` + `LaunchedEffect(pendingFocusLayer)`，遵守「只对已组合节点 requestFocus」。
  4. 根 `Box` 的 `onPreviewKeyEvent`：首行仍保留旧契约 `if (focusGuardInput.anyOverlayVisible()) return@onPreviewKeyEvent false`，然后只处理 `KeyEventType.KeyDown`，构造 `TvPlayerKeyInput` 交给 `reduceTvPlayerKey`，`dispatch(result.effects)`，`interaction = result.state`，返回 `result.handled`。
  5. scrub 自动提交：`LaunchedEffect(interaction.scrub) { if (interaction.mode == Scrubbing) { delay(TvPlayerScrubIdleCommitMillis); val r = reduceTvPlayerScrubIdleCommit(interaction); dispatch(r.effects); interaction = r.state } }`
  6. Chrome 自动隐藏：`LaunchedEffect(interaction, isPlaying, focusGuardInput) { if (interaction.mode == Chrome && !shouldFreezeTvPlayerAutoHide(...)) { delay(TvPlayerChromeAutoHideMillis); interaction = interaction.copy(mode = Hidden, focusLayer = Root) } }`
  7. 缓冲计时：`var bufferingStartedElapsed by remember { mutableStateOf<Long?>(null) }` + `LaunchedEffect(isBuffering)`，用 `withFrameMillis` 或 `delay` 轮询推进 `bufferingElapsedMs`，喂给 `resolveTvPlayerBufferingStage`；命中 `Failed` 调一次 `onBufferingTimeout()`（用 `var timeoutReported` 去重）。
  8. 对外状态回调：`LaunchedEffect(interaction.mode) { onControlsVisibilityChanged(interaction.mode != Hidden); onScrubbingChanged(interaction.mode == Scrubbing) }`
  9. `openEpisodeRailRequestKey` 变化时（宿主要求打开分集轨）：`LaunchedEffect(openEpisodeRailRequestKey) { if (openEpisodeRailRequestKey > 0) { interaction = interaction.copy(mode = EpisodeRail, focusLayer = EpisodeRail); pendingFocusLayer = EpisodeRail } }`
  10. 复用 `shouldReclaimRootFocus(previous, current)`（`core/ui/LongFormPlayerFocusGuard.kt`）在弹层关闭后把焦点收回 Root——沿用旧文件的写法。
  11. 布局：`Box { content(); TvPlayerBufferingIndicator(...); TvPlayerCenterFeedback(...); resumePromptSlot?.invoke(this); AnimatedVisibility(interaction.mode != Hidden) { Column(顶部 TvPlayerTopInfo, 底部 TvPlayerScrubTrack + TvPlayerActionRow) }; TvPlayerEpisodeRail(visible = interaction.mode == EpisodeRail, ...) }`
  12. `targets` 按内容动态构造：`buildList { add(PlayPause); add(Subtitle); add(AudioTrack); if (episodeRailItems.isNotEmpty()) add(EpisodeList) }`；`focusedTarget` 用 `remember` 保存并靠 `resolveTvPlayerHorizontalActionTarget` 在 Controls 层左右键时更新。
- [ ] 跑 `--tests '*TvPlayerChromeSpecTest'`，确认全绿。
- [ ] 跑 `./gradlew --no-daemon :tv-app:compileDebugKotlin`，确认通过。
- [ ] commit：`重构：新增 TvPlayerChrome 双层控制条`

---

## Task 7：MediaSession 接入

**Files:**
- 修改 `android-tv-app/tv-app/build.gradle.kts`（加依赖）
- 新增 `.../feature/tv/TvLongFormMediaSession.kt`
- 修改 `.../feature/tv/TvLongFormMedia3Player.kt`（挂载 effect）
- 新增 `.../test/java/com/chee/videos/feature/tv/TvLongFormMediaSessionSpecTest.kt`

**Interfaces:**

```kotlin
@Composable
internal fun TvLongFormMediaSessionEffect(player: Player, sessionId: String)
```

**Steps:**

- [ ] 写 `TvLongFormMediaSessionSpecTest.kt`：
  ```kotlin
  package com.chee.videos.feature.tv

  import java.nio.file.Path
  import org.junit.Assert.assertTrue
  import org.junit.Test

  class TvLongFormMediaSessionSpecTest {
      private fun source(name: String): String =
          Path.of("src/main/java/com/chee/videos/feature/tv/$name").toFile().readText()

      @Test
      fun sessionIsReleasedWithDisposableEffect() {
          val text = source("TvLongFormMediaSession.kt")
          assertTrue(text.contains("DisposableEffect"))
          assertTrue(text.contains("MediaSession.Builder"))
          assertTrue(text.contains("onDispose"))
          assertTrue(text.contains(".release()"))
      }

      @Test
      fun sessionCreationIsFaultTolerant() {
          // 导航过渡期可能短暂存在两个 session，id 冲突不能崩。
          assertTrue(source("TvLongFormMediaSession.kt").contains("runCatching"))
      }

      @Test
      fun playbackKernelMountsSession() {
          assertTrue(
              source("TvLongFormMedia3Player.kt").contains("TvLongFormMediaSessionEffect("),
          )
      }

      @Test
      fun chromeDoesNotOwnMediaSession() {
          val chrome = Path
              .of("src/main/java/com/chee/videos/core/ui/player/TvPlayerChrome.kt")
              .toFile()
              .readText()
          assertTrue(!chrome.contains("MediaSession"))
      }
  }
  ```
- [ ] 跑 `--tests '*TvLongFormMediaSessionSpecTest'`，确认失败。
- [ ] 在 `build.gradle.kts` 的 dependencies 块里，紧邻既有 media3 两行之后追加：
  ```kotlin
  implementation("androidx.media3:media3-session:1.4.1")
  ```
- [ ] 写 `TvLongFormMediaSession.kt`：
  ```kotlin
  package com.chee.videos.feature.tv

  import androidx.compose.runtime.Composable
  import androidx.compose.runtime.DisposableEffect
  import androidx.compose.ui.platform.LocalContext
  import androidx.media3.common.Player
  import androidx.media3.session.MediaSession

  /**
   * 把 TV 长视频播放器接入系统级媒体会话，让遥控器媒体键、系统播放控件与
   * Android TV「继续观看」频道可用。
   *
   * 会话生命周期严格跟随调用点的组合：进入时创建，离开时 release。
   * 导航过渡期两个播放器可能短暂共存导致 id 冲突，这里用 runCatching 兜底，
   * 宁可这一次没有系统会话，也不能让播放页崩溃。
   */
  @Composable
  internal fun TvLongFormMediaSessionEffect(player: Player, sessionId: String) {
      val context = LocalContext.current
      DisposableEffect(player, sessionId) {
          val session = runCatching {
              MediaSession.Builder(context, player).setId(sessionId).build()
          }.getOrNull()
          onDispose { session?.release() }
      }
  }
  ```
- [ ] 在 `TvLongFormMedia3Player.kt` 里，`player` 已创建之后调用：
  ```kotlin
  TvLongFormMediaSessionEffect(player = player, sessionId = "tv-longform-$mediaId")
  ```
  `mediaId` 用该 composable 参数表里既有的媒体标识（与 `buildTvLongFormMedia3ItemId` 用的同一个），保证不同影片/分集不会撞 id。
- [ ] 跑 `--tests '*TvLongFormMediaSessionSpecTest'` 与 `./gradlew --no-daemon :tv-app:compileDebugKotlin`，确认通过。
- [ ] commit：`功能：TV 长视频播放器接入 MediaSession`

---

## Task 8：宿主接线 + 拆除旧控制层（本计划最大、最容易踩坑的一步）

**Files:**
- 修改 `.../feature/tv/TvLongFormPlayerScreen.kt`
- 修改 `.../feature/tv/TvSeriesPlayerScreen.kt`
- 修改 `.../core/ui/TvLongFormTitleOverlay.kt`（接收迁入的 builder）
- 修改 `android-tv-app/tv-app/build.gradle.kts`（排除清单）
- 删除 `.../core/ui/TvSeriesCorePlaybackOverlay.kt`
- 删除 `.../core/ui/TvLongFormRemoteKeyRouting.kt`
- 测试迁移/删除见下

**必须一次做完的原因：** `TvSeriesCorePlaybackOverlay.kt` 依赖 `LongFormVideoPlayer.kt` 的 `TvSeriesControlsPage` / `TvPendingStepSeekUpdate` / `TvSeriesBottomPanelPage` / `TvPlaybackProgressBar`。只有它被删掉之后，`LongFormVideoPlayer.kt` 才能安全地从 TV 编译图里排除。两件事必须在同一次提交里落地，中间态编译不过。

**Steps:**

- [ ] 先把仍在用的符号从待删文件里搬走：
  - 把 `TvLongFormRemoteKeyRouting.kt` 里的 `TvLongFormTitleOverlayData` 与 `buildTvLongFormTitleOverlayData(...)` **原样**移入 `core/ui/TvLongFormTitleOverlay.kt`（它们是标题覆盖层的数据，本来就该住那儿；`TvLongFormTitleOverlaySpecTest` 与 `TvLongFormTitleOverlayDataTest` 断言的是行为不是文件，迁移后仍应通过）。
  - `TvLongFormRemoteKeyRouting.kt` 里的 `resolveTvRemoteKeyAction` / `shouldResetAutoHideTimer` / `TvRemoteKeyAction` / 旧 `TvPlayerFocusLayer` 全部废弃（已被 `TvPlayerInteraction.kt` 取代），随文件删除。
- [ ] **`core/data/AppPreferencesStore.kt` 不需要改动**（已核实）：该文件在 `com.chee.videos.core.data` 包，没有 import `com.chee.videos.core.ui.normalizeTvSeekStepSeconds`，它第 74/141 行调用的是自己第 317 行那个 `private fun normalizeTvSeekStepSeconds(seconds: Int?)`。排除 `LongFormVideoPlayer.kt` 不影响它。
- [ ] 改 `TvLongFormPlayerScreen.kt`：
  - `TvSeriesCorePlaybackOverlay(...)` 换成 `TvPlayerChrome(...)`，按新参数表逐项对应；`episodeRailItems = emptyList()`、`currentEpisodeRailItemId = null`、`episodeSwitchState = null`、`openEpisodeRailRequestKey = 0`、`onSelectEpisodeRailItem = {}`、`onEpisodeRailVisibilityChanged = {}`、`onDismissEpisodeSwitchFeedback = {}`、`onScrubbingChanged = {}`
  - `onSeekTo` 保持既有实现：`{ targetMs -> media3SeekPositionMs = targetMs; media3SeekRequestKey += 1 }`
  - 新增 `var isBuffering by remember { mutableStateOf(false) }`，在 `TvLongFormMedia3Player(...)` 里接上：
    ```kotlin
    onBufferingChanged = { buffering, identity ->
        if (identity == currentEventIdentity) isBuffering = buffering
    },
    ```
    身份比较照抄该文件里 `onSnapshotChanged` / `onError` 已有的过期事件丢弃写法。
  - `onBufferingTimeout` 接到该文件既有的软重试入口（与 `TvLongFormMedia3StartupTimeoutMillis` 超时走同一条路径），保证 20s 缓冲转入既有 `Failed` 态而不是另起一套错误 UI。
  - `PlayerGlassSurface` / `PlayerGlassSurfaceStrong` 的引用改为 `TvPlayerGlassSurface` / `TvPlayerGlassSurfaceStrong`（新包）。
- [ ] 改 `TvSeriesPlayerScreen.kt`：同上，另外补齐分集相关参数（`seriesTitleForOverlay` / `seasonNumber` / `episodeNumber` / `episodeTitle` 拼成 `primaryTitle` + `secondaryTitle`；`episodeRailItems`、`currentEpisodeRailItemId`、`episodeSwitchState = uiState.episodeSwitchState`、`onSelectEpisodeRailItem`、`onEpisodeRailVisibilityChanged = viewModel::setSelectorVisible`、`onDismissEpisodeSwitchFeedback = viewModel::clearEpisodeSwitchFeedback`、`openEpisodeRailRequestKey`）。
  - 新增 `var isScrubbing by remember { mutableStateOf(false) }`，`onScrubbingChanged = { isScrubbing = it }`，并把 `shouldShowAutoplayPromptCard`（约 236 行）的条件补上 `&& !isScrubbing`——scrub 期间不弹自动连播卡。
- [ ] 删除 `core/ui/TvSeriesCorePlaybackOverlay.kt`。
- [ ] 在 `build.gradle.kts` 的 `tvMainSourceExcludes` 列表里追加一行（注意该清单在 `kotlin.sourceSets`、`tasks.withType<KotlinCompile>`、`tasks.withType<KaptGenerateStubsTask>` 三处生效，只改这一个列表即可）：
  ```kotlin
  "com/chee/videos/core/ui/LongFormVideoPlayer.kt",
  ```
- [ ] 处理受影响的测试：
  - `TvLongFormRemoteKeyRoutingTest.kt` → **删除**（其断言的旧 seek/焦点语义已被规格 §3.3 废止，新语义由 `TvPlayerInteractionTest` 全量覆盖）。
  - `TvLongFormControlsAutoHideTest.kt` → **删除**（`shouldResetAutoHideTimer` 已被 `shouldFreezeTvPlayerAutoHide` 取代，后者已有测试）。
  - `LongFormVideoPlayerTransportKeyTest.kt` → **删除**（引用 `resolveTvPendingStepSeek` / `TvStepSeekDebounceMillis`，这两个符号随文件排除而消失；其 300ms 防抖契约已被规格废止）。
  - `LongFormVideoPlayerControlsFocusPolicyTest.kt` → **删除**（`resolveTvControlHorizontalFocusTarget` 的环绕语义已由 `TvPlayerTokensTest` 对 `resolveTvPlayerHorizontalActionTarget` 覆盖）。
  - `TvSeriesCorePlaybackOverlaySpecTest.kt` → **删除**（目标文件已删，其四条断言已由 `TvPlayerChromeSpecTest.chromeKeepsEpisodeSwitchFeedbackStates` 覆盖）。
  - 其余读取 `LongFormVideoPlayer.kt` 源码文本的 spec 测试（`LongFormVideoPlayerSpecTest`、`TvSeriesMixedPlaybackControlsSpecTest`、`TvSeriesEpisodeRailSpecTest`、`TvAutoplayPromptCardSpecTest`、`TvMotionTokensTest`、`TvIconActionSpecTest`、`TvLongFormExoPlayerSpecTest`、`TvPlayerFunctionReferenceStyleSpecTest`）：文件仍在磁盘上，只是不参与 TV 编译，**它们从磁盘读文本，仍会通过，不要动**。跑完测试逐个确认；若某个确实断言了已删除文件的内容（如 `TvSeriesCorePlaybackOverlay.kt`），改为断言 `core/ui/player/` 下的对应新文件。
  - `android-tv-app/tv-app/src/test/.../TvLongFormTitleOverlayDataTest`：只需把 import 从 `TvLongFormRemoteKeyRouting` 隐含的包路径调整到实际所在包（同包 `com.chee.videos.core.ui`，通常无需改）。
- [ ] 跑 `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`，全绿。
- [ ] 跑 `./gradlew --no-daemon :tv-app:assembleDebug`，编译通过。
- [ ] commit：`重构：TV 长视频宿主接入新播放器控制层并拆除旧覆盖层`

---

## Task 9：版本号、知识沉淀与全量验证

**Files:**
- 修改 `android-tv-app/tv-app/build.gradle.kts`
- 修改 `plan.md`
- 修改 `CONTEXT.md`

**Steps:**

- [ ] `build.gradle.kts`：`versionCode = 141` → `142`，`versionName = "0.1.141"` → `"0.1.142"`。
- [ ] `plan.md` **顶部追加**（反向时间序，不得覆盖任何既有条目）一条记录，含：日期时间（2026-07-26）、摘要（TV 长视频播放器交互模型重构：scrub 模式 + 状态机 + 双层控制条 + 缓冲态 + MediaSession）、影响文件清单、验证状态（`:tv-app:testDebugUnitTest` 与 `:tv-app:assembleDebug` 结果）。
- [ ] `CONTEXT.md` 追加长期知识，**必须显式写明被替换的四条旧契约**，避免后来者按旧契约"修回去"：
  - 新术语：`scrub 模式`、`ghost 游标`、`锚点（anchor）`、`Chrome（控制层）`、`加速档`
  - 新契约：TV 长视频 UI 状态由 `TvPlayerUiMode` 单一状态机表达；左右键在 Root 层直接进 scrub 且**视频不暂停**；OK 提交 / BACK 取消（取消不 seek）/ 1.5s 自动提交；进度轨永不进焦点链；暂停时 Chrome 常驻不自动隐藏
  - **已废止**：`seek 进度显示防抖`、`连按合并跳转（300ms）`
  - **已修订**：`电视剧进度条只展示不交互` → 进度轨仍不可聚焦，但由 Root 层 scrub 驱动；`controls 焦点入口` → 保留 DOWN 分层，新增 Root 层 UP 只唤起不夺焦
  - **仍有效**：`操作 UI 互动唤起`、`controls 焦点环绕`、`controls 左右键切焦点`、`controls 持焦横向导航`、`overlay 可见时按键透传`
  - 架构决策：`core/ui/player` 三层边界（reducer 无 Compose / Chrome 无 Media3 / MediaSession 只在 `DisposableEffect`）
  - 兼容策略：`LongFormVideoPlayer.kt` 已排除出 TV 编译图（TV 侧工具在 `core/ui/player` 以 `TvPlayer*` 前缀重建），手机端该文件保持原样
  - 坑：MediaSession id 冲突需 `runCatching` 兜底；缓冲指示 350ms 延迟是为了避免 seek 后闪烁；`RequestFocus` 必须延后一帧，否则目标节点尚未组合
- [ ] 全量验证：
  ```bash
  cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug
  ```
- [ ] commit：`版本：TV 端 0.1.142 并沉淀播放器重构知识`

---

## 完成判据

- [ ] `:tv-app:testDebugUnitTest` 与 `:tv-app:assembleDebug` 全绿
- [ ] `TvSeriesCorePlaybackOverlay.kt` 与 `TvLongFormRemoteKeyRouting.kt` 已删除，`LongFormVideoPlayer.kt` 已移出 TV 编译图
- [ ] 规格 §3.1 状态机的每条边都有对应 reducer 测试
- [ ] 规格 §3.3 的四条被替换契约已在 `CONTEXT.md` 标注
- [ ] TV 版本号已 bump 到 141/0.1.141 → 142/0.1.142
- [ ] 工作区干净，所有改动均已提交
- [ ] **按 CLAUDE.md 要求，编码完成后必须再让子代理独立评审，修复问题并复审，直到没有阻塞问题才可宣布完成**

## 已知风险（用户已知悉并接受）

1. **1.5s 自动提交比现状 300ms 防抖长很多**，真机手感可能需要调；阈值集中在 `TvPlayerScrubIdleCommitMillis` 一处，便于调整。
2. **暂停时 Chrome 常驻**是新增行为，与旧版「暂停也会 5s 自动隐藏」不同。
3. **MediaSession 会让观看记录出现在 Android TV 系统「继续观看」频道**，这是接入的预期副作用。
