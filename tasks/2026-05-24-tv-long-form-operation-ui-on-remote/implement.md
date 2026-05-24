# Implement：TV 长视频播放器操作 UI 跟随遥控器互动

- 日期：2026-05-24
- 关联 PRD：`./prd.md`

## 1. 总体方案

把 `LongFormVideoPlayer` 在 `tvMode = true` 路径下的"操作 UI 层"切成两个独立组件 + 一套统一的键路由 + 焦点环绕 wiring：

1. **领域层（纯函数）**：`core/ui/TvLongFormRemoteKeyRouting.kt` 把"按键 + 当前可见性 + 焦点是否在 controls + 按键 repeat" 映射成 `TvRemoteKeyAction` 密封类；stateless、单测覆盖所有 6 键 × 3 状态 = 18 组合 + repeat 变体。
2. **视图层（Composable）**：
   - `core/ui/TvLongFormTitleOverlay.kt` 新增——左上信息层，纯文字 + Shadow，受 `controlsVisible` 单一状态驱动。
   - `core/ui/LongFormVideoPlayer.kt` 修改——
     - 顶部信息区分 `tvMode` 走两条独立 `AnimatedVisibility` 渲染：tvMode 走 `TvLongFormTitleOverlay`，非 tvMode 沿用既有玻璃顶栏。
     - 控制条按钮链中每个按钮挂上 `Modifier.focusProperties { left = prev; right = next }` 实现首尾环绕。
     - Slider 上挂 `Modifier.focusProperties { canFocus = false }`。
     - `onPreviewKeyEvent` 重写为单一路由器：调用 `resolveTvRemoteKeyAction` 拿到 action 后分发。
     - 新增 `focusInControls: Boolean` 状态（通过控制条容器 `onFocusChanged` 与每个按钮的 `onFocusChanged` 累加判断）。
3. **调用方（无视图改动，仅参数迁移）**：
   - `TvLongFormPlayerScreen.kt` 与 `TvSeriesPlayerScreen.kt` 把 `title` 与可选的 `seasonNumber` / `episodeNumber` / `episodeTitle` 一并传入；通过 `buildTvLongFormTitleOverlayData(...)` 构造 `TvLongFormTitleOverlayData`。

`LongFormVideoPlayer` 非 TV 模式（`UnifiedPlayerScreen` 等手机短视频全屏）零改动——所有新行为都在 `if (tvMode) { ... }` 分支内，新增 props 用默认值兜底。

## 2. 文件结构

### 2.1 新增

| 路径 | 用途 |
|---|---|
| `core/ui/TvLongFormRemoteKeyRouting.kt` | 纯函数：`resolveTvRemoteKeyAction(visible, focusInControls, keyCode, repeatCount, seekStepSec): TvRemoteKeyAction`；`TvRemoteKeyAction` 密封类；`buildTvLongFormTitleOverlayData(...)`；`shouldResetAutoHideTimer(prev, next): Boolean` |
| `core/ui/TvLongFormTitleOverlay.kt` | `@Composable TvLongFormTitleOverlay(data, visible)`；`object TvLongFormTitleOverlayTokens`（字号 / alpha / shadow offset / blur radius / 左上 padding） |
| `tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormRemoteKeyRoutingTest.kt` | 18 + 状态组合 + repeat 变体单测 |
| `tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlayDataTest.kt` | `buildTvLongFormTitleOverlayData` 各种 fallback 路径 |
| `tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormControlsAutoHideTest.kt` | `shouldResetAutoHideTimer` 矩阵 |
| `tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlaySpecTest.kt` | 源文 audit：调用点 import、Token-only、Shadow API、`fadeIn/fadeOut` 必带 `tween(TvMotionTokens...)` |
| `tv-app/src/androidTest/java/com/chee/videos/core/ui/LongFormVideoPlayerFocusWrapTest.kt` | 焦点环绕集成测试 |

### 2.2 修改

| 路径 | 改动 |
|---|---|
| `core/ui/LongFormVideoPlayer.kt` | 新增 props：`seasonNumber: Int? = null` / `episodeNumber: Int? = null` / `episodeTitle: String? = null`；tvMode 顶部信息层切到 `TvLongFormTitleOverlay`；控制条按钮链挂焦点环绕；Slider `canFocus = false`；`onPreviewKeyEvent` 改为单一 `resolveTvRemoteKeyAction` 路由；删除 `resolveTvHiddenTransportKeyAction` / `TvHiddenTransportKeyAction` 内部封装（被新路由取代） |
| `feature/tv/TvSeriesPlayerScreen.kt` | line 528 调用：把 `title` 从 `currentEpisode?.title ?: series.title` 改为 `series.title`；新增传入 `seasonNumber = uiState.selectedSeasonNumber` / `episodeNumber = uiState.selectedEpisodeNumber` / `episodeTitle = currentEpisode?.title`（保留 fallback 到 series.title 的兜底逻辑——仅在 series.title 为空时用 episode.title 顶上） |
| `feature/tv/TvLongFormPlayerScreen.kt` | 仅确认 `title = detail.title`，不传 season/episode props（保持默认 null） |
| `feature/player/UnifiedPlayerScreen.kt` | **零改动**（新增 props 都有默认值，非 TV 模式不渲染左上信息层） |
| `CONTEXT.md` | 「TV 播放术语」区追加 §4 列出的 7 条新术语；line 130 [[TV 播放器退出确认]] 条目末尾追加"[[BACK 优先收 UI]] 是其前置例外，仅在 [[TV 操作 UI 层]] 可见时介入"一句 |
| `android-tv-app/tv-app/build.gradle.kts` | `versionCode +1`、`versionName` 末位 +1 |
| `plan.md` | 追加 reverse-chronological 进度条目 |

### 2.3 顺手不修

- 不删除现有 `resolveTvHiddenTransportKeyAction` 的测试文件 `LongFormVideoPlayerTransportKeyTest.kt`——把测试用例迁移到新的 `TvLongFormRemoteKeyRoutingTest.kt` 后**删除该文件**或改名为新文件（实施时按测试用例迁移完成度决定，不强制单一动作）
- 不动 `LongFormVideoPlayerFocusTest.kt`（androidTest 中既有的焦点测试），仅在 wrap 测试中**新增**用例，不重写既有
- 不动 [[续播提示卡]] / [[连播提示卡]] / [[TV 播放器退出确认]] 现有调用方代码

## 3. 关键函数 / 类型签名草稿

### 3.1 `TvLongFormRemoteKeyRouting.kt`

```kotlin
package com.chee.videos.core.ui

import android.view.KeyEvent as AndroidKeyEvent

internal sealed interface TvRemoteKeyAction {
    /** seek + 唤起 UI（焦点不进入），用于 controls 隐藏时的 ←/→，或可见+焦点未进入时的 ←/→ */
    data class Seek(val deltaMs: Long) : TvRemoteKeyAction
    /** 唤起 UI 并把焦点送到「播放/暂停」按钮，用于 ↓ 键 */
    data object EnterFocus : TvRemoteKeyAction
    /** 焦点退出 controls 回到播放器根，UI 不收起，5s 计时不动，用于焦点在 controls 内时按 ↑ */
    data object ExitFocus : TvRemoteKeyAction
    /** 切换播放/暂停 + 中央反馈 toast，用于焦点不在 controls 时的 OK */
    data object TogglePlayPause : TvRemoteKeyAction
    /** UI 可见时立即收 UI，不进退出确认（[[BACK 优先收 UI]]） */
    data object DismissUi : TvRemoteKeyAction
    /** 焦点在 controls 内按 ←/→ 时：由 Compose `focusProperties` 完成，路由层只需 PassThrough 给系统 */
    data object PassThrough : TvRemoteKeyAction
}

/**
 * TV 长视频播放器统一遥控器键路由。
 *
 * @param visible 当前 [[TV 操作 UI 层]] 是否可见
 * @param focusInControls 当前焦点是否在底部 controls 内（任意按钮 focused）
 * @param keyCode `KeyEvent.keyCode` 原始值
 * @param repeatCount `KeyEvent.repeatCount` 原始值（长按 / 连按时 > 0）
 * @param seekStepSec [[快进/快退步长]] 当前秒数（5/10/15/20/30）
 * @return 应执行的动作；`null` 表示"未识别 / 让系统继续处理"
 */
internal fun resolveTvRemoteKeyAction(
    visible: Boolean,
    focusInControls: Boolean,
    keyCode: Int,
    repeatCount: Int,
    seekStepSec: Int,
): TvRemoteKeyAction? {
    val stepMs = normalizeTvSeekStepSeconds(seekStepSec) * 1_000L
    val seekDelta = if (repeatCount > 0) stepMs * 3 else stepMs
    return when (keyCode) {
        AndroidKeyEvent.KEYCODE_DPAD_LEFT,
        AndroidKeyEvent.KEYCODE_MEDIA_REWIND,
        -> when {
            !visible -> TvRemoteKeyAction.Seek(-seekDelta) // 隐藏：seek + 由 caller 唤起 UI
            visible && !focusInControls -> TvRemoteKeyAction.Seek(-seekDelta) // 可见+焦点未进入：继续 seek + 重置计时
            else -> TvRemoteKeyAction.PassThrough // 可见+焦点已进入：Compose focusProperties 处理环绕
        }

        AndroidKeyEvent.KEYCODE_DPAD_RIGHT,
        AndroidKeyEvent.KEYCODE_MEDIA_FAST_FORWARD,
        -> when {
            !visible -> TvRemoteKeyAction.Seek(seekDelta)
            visible && !focusInControls -> TvRemoteKeyAction.Seek(seekDelta)
            else -> TvRemoteKeyAction.PassThrough
        }

        AndroidKeyEvent.KEYCODE_DPAD_DOWN -> when {
            !visible -> TvRemoteKeyAction.EnterFocus
            visible && !focusInControls -> TvRemoteKeyAction.EnterFocus
            else -> null // 已在 controls 内，吞掉
        }

        AndroidKeyEvent.KEYCODE_DPAD_UP -> when {
            focusInControls -> TvRemoteKeyAction.ExitFocus
            else -> null // 隐藏或可见但焦点在根，吞掉（不唤起 UI）
        }

        AndroidKeyEvent.KEYCODE_DPAD_CENTER,
        AndroidKeyEvent.KEYCODE_ENTER,
        AndroidKeyEvent.KEYCODE_NUMPAD_ENTER,
        AndroidKeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
        AndroidKeyEvent.KEYCODE_MEDIA_PLAY,
        AndroidKeyEvent.KEYCODE_MEDIA_PAUSE,
        -> when {
            focusInControls -> TvRemoteKeyAction.PassThrough // Compose 默认触发焦点按钮 onClick
            else -> TvRemoteKeyAction.TogglePlayPause
        }

        AndroidKeyEvent.KEYCODE_BACK,
        AndroidKeyEvent.KEYCODE_ESCAPE,
        -> when {
            visible -> TvRemoteKeyAction.DismissUi
            else -> null // UI 已收：让外层 BackHandler 接管 → [[TV 播放器退出确认]]
        }

        else -> null
    }
}

/** PRD §6.3 C2 决议：`Seek` / `EnterFocus` / `TogglePlayPause` 触发计时重置；`ExitFocus` / `DismissUi` / `PassThrough` 不重置 */
internal fun shouldResetAutoHideTimer(action: TvRemoteKeyAction): Boolean = when (action) {
    is TvRemoteKeyAction.Seek,
    TvRemoteKeyAction.EnterFocus,
    TvRemoteKeyAction.TogglePlayPause -> true
    TvRemoteKeyAction.ExitFocus,
    TvRemoteKeyAction.DismissUi,
    TvRemoteKeyAction.PassThrough -> false
}
```

### 3.2 `TvLongFormTitleOverlay.kt`

```kotlin
package com.chee.videos.core.ui

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.Immutable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shadow
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

internal object TvLongFormTitleOverlayTokens {
    val LeftPaddingDp = 24
    val TopPaddingDp = 24
    val PrimaryFontSizeSp = 22
    val SecondaryFontSizeSp = 16
    val PrimaryColor = Color.White
    val SecondaryColor = Color.White.copy(alpha = 0.72f)
    val ShadowColor = Color(0xCC000000)
    val ShadowOffset = Offset(0f, 2f)
    val ShadowBlurRadius = 4f
}

@Immutable
internal data class TvLongFormTitleOverlayData(
    val primary: String,
    val secondary: String?, // null = 电影 / 18+，仅渲染主行
)

internal fun buildTvLongFormTitleOverlayData(
    primaryFallback: String,
    seriesTitle: String?,
    seasonNumber: Int?,
    episodeNumber: Int?,
    episodeTitle: String?,
): TvLongFormTitleOverlayData {
    val isSeries = seasonNumber != null && episodeNumber != null
    val primary = (seriesTitle?.takeIf { it.isNotBlank() } ?: primaryFallback).ifBlank { primaryFallback }
    val secondary = if (isSeries) {
        buildString {
            append("第 ")
            append(seasonNumber)
            append(" 季 · 第 ")
            append(episodeNumber)
            append(" 集")
            val trailing = episodeTitle?.trim().orEmpty()
            if (trailing.isNotEmpty()) {
                append(' ')
                append(trailing)
            }
        }
    } else null
    return TvLongFormTitleOverlayData(primary = primary, secondary = secondary)
}

@Composable
internal fun TvLongFormTitleOverlay(
    data: TvLongFormTitleOverlayData,
    visible: Boolean,
    showStatusBarPadding: Boolean,
    modifier: Modifier = Modifier,
) {
    val tokens = TvLongFormTitleOverlayTokens
    AnimatedVisibility(
        visible = visible,
        enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
        exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
        modifier = modifier.then(if (showStatusBarPadding) Modifier.statusBarsPadding() else Modifier),
    ) {
        Column(
            modifier = Modifier.padding(start = tokens.LeftPaddingDp.dp, top = tokens.TopPaddingDp.dp),
        ) {
            Text(
                text = data.primary,
                color = tokens.PrimaryColor,
                style = MaterialTheme.typography.titleLarge.copy(
                    fontSize = tokens.PrimaryFontSizeSp.sp,
                    fontWeight = FontWeight.SemiBold,
                    shadow = Shadow(
                        color = tokens.ShadowColor,
                        offset = tokens.ShadowOffset,
                        blurRadius = tokens.ShadowBlurRadius,
                    ),
                ),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            data.secondary?.let { secondary ->
                Text(
                    text = secondary,
                    color = tokens.SecondaryColor,
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontSize = tokens.SecondaryFontSizeSp.sp,
                        shadow = Shadow(
                            color = tokens.ShadowColor,
                            offset = tokens.ShadowOffset,
                            blurRadius = tokens.ShadowBlurRadius,
                        ),
                    ),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}
```

### 3.3 `LongFormVideoPlayer.kt` 改动要点

```kotlin
// 新增 props（可空，默认 null，非 TV 模式与电影 / 18+ 走默认）
fun LongFormVideoPlayer(
    title: String,
    // ... existing params ...
    seasonNumber: Int? = null,
    episodeNumber: Int? = null,
    episodeTitle: String? = null,
    seriesTitleForOverlay: String? = null, // 调用方明确传剧名，避免 title=单集标题歧义
    // ... existing params ...
) {
    // ... existing state ...
    var focusInControls by remember { mutableStateOf(false) }
    val controlsContainerFocusRequester = remember { FocusRequester() }
    val controlButtonFocusRequesters = remember { mutableMapOf<String, FocusRequester>() }

    // ... existing scheduleAutoHideControls / showControlsTemporarily ...

    // 替换原有 handleTvTransportKey 的实现：
    fun handleTvKey(keyCode: Int, repeatCount: Int): Boolean {
        val action = resolveTvRemoteKeyAction(
            visible = controlsVisible,
            focusInControls = focusInControls,
            keyCode = keyCode,
            repeatCount = repeatCount,
            seekStepSec = tvSeekStepSeconds,
        ) ?: return false
        when (action) {
            is TvRemoteKeyAction.Seek -> {
                performDebouncedStepSeek(action.deltaMs, showControls = false /* 我们手动调用 */)
                if (!controlsVisible) controlsVisible = true
                scheduleAutoHideControls()
            }
            TvRemoteKeyAction.EnterFocus -> {
                controlsVisible = true
                pendingPlayPauseFocusRequest = true
                scheduleAutoHideControls()
            }
            TvRemoteKeyAction.ExitFocus -> {
                rootFocusRequester.tryRequestFocus()
                focusInControls = false
                // 不重置计时器（PRD §1.5 ↑ 决议）
            }
            TvRemoteKeyAction.TogglePlayPause -> {
                togglePlaybackWithFeedback(showControls = false)
                // OK 不唤起 UI（PRD §6.1 A4 决议），但中央 toast 走 showTransientFeedback
            }
            TvRemoteKeyAction.DismissUi -> {
                controlsVisible = false
                hideControlsJob?.cancel()
                if (focusInControls) {
                    rootFocusRequester.tryRequestFocus()
                    focusInControls = false
                }
            }
            TvRemoteKeyAction.PassThrough -> return false // Compose 默认处理（焦点环绕 / 按钮 onClick）
        }
        return true
    }

    Box(
        modifier = modifier
            // ... existing modifiers ...
            .onPreviewKeyEvent { event ->
                if (!tvMode || event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                handleTvKey(event.nativeKeyEvent.keyCode, event.nativeKeyEvent.repeatCount)
            }
            // ... existing pointerInput ...
    ) {
        AndroidView(/* ... */)

        // tvMode 走新左上信息层；非 tvMode 沿用既有顶部玻璃栏
        if (tvMode) {
            TvLongFormTitleOverlay(
                data = buildTvLongFormTitleOverlayData(
                    primaryFallback = title,
                    seriesTitle = seriesTitleForOverlay,
                    seasonNumber = seasonNumber,
                    episodeNumber = episodeNumber,
                    episodeTitle = episodeTitle,
                ),
                visible = controlsVisible,
                showStatusBarPadding = showStatusBarPadding,
                modifier = Modifier.align(Alignment.TopStart),
            )
        } else {
            // 既有顶部玻璃顶栏渲染（line 670–718 完整保留），仅包在 else 分支
        }

        // ... 既有 SeekPreview / CenterFeedback ...

        // 底部控制条：tvMode 路径增加 focusInControls 状态联动 + focusProperties wiring
        AnimatedVisibility(
            visible = controlsVisible,
            // ... existing fade ...
        ) {
            Row(
                modifier = Modifier
                    .onFocusChanged { /* 任一子按钮聚焦时 focusInControls = true */ }
                    // ...
            ) {
                // 每个按钮分配独立 FocusRequester；用 controlButtonFocusRequesters 字典追踪
                // 第一个按钮（播放/暂停）：focusProperties { left = lastButtonRequester }
                // 最后一个按钮（动态：退出播放 / 返回详情 / 全屏，按可见性挑）：focusProperties { right = firstButtonRequester }
                // Slider：focusProperties { canFocus = false }
                // 中间按钮：focusProperties { left = prevButton; right = nextButton } 显式指定，避免 Slider 被绕过后的 2D focus search 不稳定
            }
        }
    }
}
```

### 3.4 `TvSeriesPlayerScreen.kt` 调用点改动

```kotlin
LongFormVideoPlayer(
    title = series.title.ifBlank { currentEpisode?.title.orEmpty() }, // primary fallback
    seriesTitleForOverlay = series.title,
    seasonNumber = uiState.selectedSeasonNumber,
    episodeNumber = uiState.selectedEpisodeNumber,
    episodeTitle = currentEpisode?.title,
    // ... existing params unchanged ...
)
```

> 注：原 `title = currentEpisode?.title ?: series.title` 在 tvMode 顶部信息栏显示的是单集标题，现在新 overlay 主行需要剧名。`title` 参数仍保留用于非 tvMode 老顶栏 + 部分内部使用（如 `AsyncImage contentDescription`），保持兼容性兜底。

### 3.5 `TvLongFormPlayerScreen.kt` 调用点改动

```kotlin
LongFormVideoPlayer(
    title = detail.title,
    // 不传 seasonNumber / episodeNumber / episodeTitle，使用默认 null
    // 不传 seriesTitleForOverlay，使用默认 null（buildTvLongFormTitleOverlayData 会 fallback 到 title）
    // ... existing params unchanged ...
)
```

## 4. CONTEXT.md 更新草稿

在 line 134（`快进/快退步长` 之后、`连按合并跳转` 之前不强求位置；插入到「TV 播放术语」区合适语义位置即可）追加：

```markdown
- `TV 操作 UI 层`：TV 长视频播放器（`LongFormVideoPlayer` 在 `tvMode = true`）的"操作可见性单元"，由底部 controls 玻璃条与 [[左上信息层]] 两部分组成；两者由单一 `controlsVisible` 状态驱动、同一组 5 秒计时控制、同一组 `TvMotionTokens.DurationStandardMs` + `EasingStandard` 240ms tween fade 同步出场入场。"层"是可见性 / 动效收口的概念，不是组件 hierarchy；底部 controls 与 [[左上信息层]] 在 `LongFormVideoPlayer` 中仍是平行 `AnimatedVisibility` 节点，仅共享 visibility flag。
- `左上信息层`：[[TV 操作 UI 层]] 的左上角无背板纯文字标题层。主行 = 剧名（电视剧）/ 电影标题（电影）/ AV 标题（`18+`），白色 SemiBold 约 22sp；电视剧追加副行「第 X 季 · 第 Y 集 单集标题」，72% alpha 约 16sp。文字阴影使用 `TextStyle.shadow = Shadow(Color(0xCC000000), Offset(0, 2), 4f)`，跨亮 / 暗背景自适应；不引入新的 `Surface` / 背板 / `Shape`。受 [[TV 文本溢出保护]] 约束 `maxLines = 1 + ellipsis`。所有视觉参数收口在 `core/ui/TvLongFormTitleOverlay.kt` 的 `TvLongFormTitleOverlayTokens`，调用点不允许出现 `22.sp` / `0xCC000000` / `Offset(0f, 2f)` 等字面量。
- `操作 UI 互动唤起`：[[TV 操作 UI 层]] 隐藏时，任意遥控器键事件（除 ↑ 在焦点已在根态外）都唤起整个 [[TV 操作 UI 层]] 并启动 5 秒自动隐藏计时。← / → 同时承担"seek + 唤起"双语义；↓ 同时承担"唤起 + 送焦点到「播放/暂停」"；OK 不唤起整套 UI，仅触发播放暂停切换 + 中央反馈 toast（由 `showTransientFeedback` 承载）。[[TV 操作 UI 层]] 已可见时，可见性互动（按钮 onClick / 焦点切换 / 控制条内 ← / → / 焦点未进入时的 ← / → seek）均重置 5 秒计时。
- `controls 焦点入口`：TV 模式下，**仅** DPad ↓ 键能把焦点从播放器根送入底部 controls；其它键（← / → / OK / ↑）即使唤起 [[TV 操作 UI 层]] 也保持焦点在播放器根。这是与 phone-era "controls 一出现就自动抢焦点" 模式的显式分离——目标是让焦点位置可预测、不被自动化打破，避免用户"想 seek 却按下去触发了按钮"的误触。
- `controls 焦点环绕`：[[TV 操作 UI 层]] 底部控制条焦点链按当前可见按钮的顺序首尾相连：在最左侧按钮按 ← 跳到最右侧按钮，反之亦然。实现使用 Compose 原生 `Modifier.focusProperties { left = prevReq; right = nextReq }` 配 `FocusRequester`，不在 `onPreviewKeyEvent` 写硬编码 if-else 焦点路由。Slider 通过 `Modifier.focusProperties { canFocus = false }` 排除在焦点链外（控制条已有快退 / 快进按钮覆盖 ±10s seek，Slider 退化为只读进度条）。条件渲染的按钮（如 [[手动下一集按钮]] / 退出播放）出入链时需重建首尾 FocusRequester 绑定。
- `BACK 优先收 UI`：TV 长视频播放器中 BACK 键的两段式语义。[[TV 操作 UI 层]] 可见时按 BACK 立即收 UI（不等 5 秒、不进入 [[TV 播放器退出确认]]）；UI 已收起时按 BACK 才进入 [[TV 播放器退出确认]] 的"第一次提示、第二次确认退出"流程。这是对 [[TV 播放器退出确认]] 的**前置例外**——不破坏其主流程的"两段式提示"承诺，只是在 UI 可见时插入一段"先收 UI"的前置步骤。实现位点在 `LongFormVideoPlayer` 的 `onPreviewKeyEvent` 内，UI 可见时拦截 BACK 走 `controlsVisible = false`；UI 已收时 `return false` 让外层 `BackHandler` 接管。
- `controls 焦点退出键`：焦点在 [[TV 操作 UI 层]] 内时，↑ 键把焦点退回播放器根（不收 UI，5 秒计时不重置——延续余下时间继续倒计）。其它"焦点离开 controls"的方式：BACK（收 UI 同时退焦点）/ 5 秒未互动自动隐藏（隐藏 UI 同时退焦点）。↑ 在焦点已在根 / UI 已隐藏时一律吞掉，**不**唤起 UI——避免用户拍 ↑ 时频繁误触亮 UI（仅 ↓ 是 [[controls 焦点入口]]）。
```

并在 line 130 的 `TV 播放器退出确认` 条目末尾追加：

```markdown
[[BACK 优先收 UI]] 是其前置例外：[[TV 操作 UI 层]] 可见时按 BACK 优先收 UI，不进入本流程；UI 已收起时按 BACK 才进入本流程的第一次提示。该例外**不**适用于控制条上的「返回详情」/「退出播放」按钮（焦点 + OK 触发），那两个按钮仍然直接进入本流程。
```

## 5. 步骤

1. **先写测试**（PRD §8.1 / §8.2 / §8.3 列出的 6 个文件），全部红
2. **领域层**：实现 `TvLongFormRemoteKeyRouting.kt`、`TvLongFormTitleOverlay.kt` 的数据 / 纯函数部分（不含 Composable），跑单测转绿
3. **视图层**：
   - 写 `TvLongFormTitleOverlay` Composable，写好 token、shadow API、AnimatedVisibility wiring
   - 改 `LongFormVideoPlayer`：新增 props、新增 `focusInControls` state、重写 `onPreviewKeyEvent` 为 `handleTvKey`、tvMode 顶部分支切到新 overlay
   - 加焦点环绕 wiring：为每个可见按钮分配 `FocusRequester`，按可见列表动态计算 prev / next 绑定 `Modifier.focusProperties`
   - 给 Slider 加 `Modifier.focusProperties { canFocus = false }`
4. **调用方**：改 `TvSeriesPlayerScreen.kt` 传入剧名 / 季号 / 集号 / 单集标题
5. **审计测试**：跑 `TvLongFormTitleOverlaySpecTest` 等源文 audit 测试转绿
6. **集成测试**：写 `LongFormVideoPlayerFocusWrapTest`（androidTest），跑通环绕场景
7. **手测覆盖 PRD §6** 所有 A–G
8. **更新 `CONTEXT.md`**（§4 草稿）
9. **`android-tv-app/tv-app/build.gradle.kts` 版本号 +1**
10. **`plan.md` 追加 reverse-chronological 进度条目**
11. **`git diff` 自检**：仅触及 §2 文件清单
12. **build + 单测 + androidTest 全绿后 commit**（commit message 中文）

## 6. 注意事项 / 已知陷阱

- **焦点环绕 wiring 的 FocusRequester 生命周期**：每个按钮的 `FocusRequester` 必须与按钮的 Composable 节点同 attach；`focusProperties { left = otherReq }` 的 `otherReq` 必须指向当前**实际 composed** 的按钮节点（遵循 `CONTEXT.md` 的「焦点请求安全调用 tryRequestFocus」式约束）。控制条中按钮条件渲染（电视剧才有选集/下一集；非系列才有全屏；series 才有退出播放）→ 链长度动态。建议：先把可见按钮按顺序放进 `remember(...)` 的 `List<Pair<String, FocusRequester>>`，再按列表 index 推导 prev/next。
- **`focusInControls` 状态同步**：用 `Modifier.onFocusChanged` 挂在 controls 的 `Row` 容器上不可靠（Compose 中 Row 不天然 focusable）；改为给每个按钮的 `Modifier.onFocusChanged { isFocused -> if (isFocused) focusInControls = true }`，再用一个根容器的 `onFocusChanged { if (!isFocused) focusInControls = false }` 兜底——或更稳：维护一个 `focusedButtonId: String?` 状态，`focusInControls = focusedButtonId != null`。
- **`pendingPlayPauseFocusRequest` 与 `pendingRootFocusRequest` 现有逻辑保留**：现有 `LaunchedTvInitialFocus`（line 405–414, 416–425）保留不动；新的 `ExitFocus` 直接调 `rootFocusRequester.tryRequestFocus()`；新的 `EnterFocus` 直接 set `pendingPlayPauseFocusRequest = true`（沿用现有路径）。
- **`LaunchedEffect(tvMode, controlsVisible)` 的现有 effect**（line 391–403）今天会在 `controlsVisible = true` 时自动 set `pendingPlayPauseFocusRequest = true`——这与 PRD Q7 决议（首次进入不抢焦点）冲突，必须改写：把 `requestPlayPauseFocusWhenReady()` 移到**仅由 EnterFocus 触发**，自动 `LaunchedEffect(tvMode)` 入场只 set `controlsVisible = true` + `scheduleAutoHideControls()`，**不** set focus request；同理 `controlsVisible -> false` 时也不自动把焦点送回根（焦点本来就没离开根）。
- **OK 键不唤起 UI**（PRD §6.1 A4）：`TogglePlayPause` 分支不调 `showControlsTemporarily`，仅调 `togglePlaybackWithFeedback(showControls = false)`。这把 `togglePlaybackWithFeedback` 的现有"默认 showControls = true"行为改为按需——保留参数默认值即可。
- **测试迁移**：现有 `LongFormVideoPlayerTransportKeyTest.kt` 测试 `resolveTvHiddenTransportKeyAction`，新文件 `TvLongFormRemoteKeyRoutingTest.kt` 取代后**删除老测试文件**，并在源文 audit 中验证 `resolveTvHiddenTransportKeyAction` 已从 `LongFormVideoPlayer.kt` 移除。
- **不要在 `onPreviewKeyEvent` 内同步调 `tryRequestFocus`**：`ExitFocus` 走 `rootFocusRequester.tryRequestFocus()` 应该是安全的（rootFocusRequester 在 Box 上已 attach），但避免重复请求建议加 `if (focusInControls)` 守卫。
