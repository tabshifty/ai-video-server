# Implement：TV 长视频播放器焦点兜底

- 日期：2026-05-25
- 关联 PRD：`./prd.md`

## 1. 总体方案

把"防焦点真空"作为 `LongFormVideoPlayer` 的内部不变量，由 player 自身负责：

1. **续播卡内嵌**：`TvResumePromptCard` 从屏幕级 sibling 位置移到 `LongFormVideoPlayer` 内部，通过 `resumePromptSlot: @Composable () -> Unit` 槽位提供。这样卡片 dispose 时 Compose 焦点回收到 ancestor（root Box）而不是清空。
2. **overlay 聚合的兜底 LaunchedEffect**：在 player 内部聚合所有可能持焦的叠加层可见性（控制条 / 字幕 picker / 音轨 picker / 续播卡 / 返回二次确认 / playerError），任一从 true→false 跃迁且没有其他 overlay 仍在显示时，显式调用 `rootFocusRequester.tryRequestFocus()`。
3. **root Box `onFocusChanged` 兜底**：root Box 在自己丢焦点（`!isFocused && !hasFocus`）且没有 overlay 显示时，schedule 一次 `tryRequestFocus()`。这是最后一道关，专门接 Dialog dismiss 这种 Compose 不通过 LaunchedEffect 通知的路径。

实施顺序（强制按这个走，避免半完成态卡住）：

```
1. 续播卡内嵌（槽位 API + 调用方迁移）
2. overlay 聚合 LaunchedEffect
3. root onFocusChanged 兜底
4. 单测：聚合判定纯函数 + 源文 audit
5. 手测脚本（review.md §1）
6. CONTEXT.md sync + 版本号 + 提交
```

## 2. 文件结构

### 2.1 新增

| 路径 | 用途 |
|---|---|
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormPlayerFocusGuard.kt` | 纯函数 `shouldReclaimRootFocus(input: PlayerFocusGuardInput): Boolean`：根据六个 overlay 可见性聚合判定是否需要把焦点收回 root Box；以及 `PlayerFocusGuardInput` data class |
| `tv-app/src/test/java/com/chee/videos/core/ui/LongFormPlayerFocusGuardTest.kt` | 聚合判定纯函数全分支单测 |
| `tv-app/src/test/java/com/chee/videos/core/ui/LongFormVideoPlayerSpecTest.kt` | 源文 audit：扫描 `LongFormVideoPlayer.kt` 必须 import `LongFormPlayerFocusGuard`、必须存在 `resumePromptSlot` 参数、`TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt` 不再把 `TvResumePromptCard` 平铺在外层 Box |

### 2.2 修改

| 路径 | 改动 |
|---|---|
| `tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt` | 新增可选参数 `resumePromptSlot: (@Composable () -> Unit)? = null` 与 `resumePromptVisible: Boolean = false`；在 player 根 Box 内部把 slot 渲染在控制条之上、按当前续播卡布局位置（`Modifier.align(Alignment.BottomStart)` + 同等 padding）；新增 `LaunchedEffect` 聚合六 overlay 可见性，跃迁触发 `requestRootFocusWhenReady()`；root Box 加 `Modifier.onFocusChanged { ... }` 在 root 自身丢焦点且无 overlay 时 `requestRootFocusWhenReady()` |
| `tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt` | 删除外层 Box 里平铺的 `TvResumePromptCard(...)` 调用；改为给 `LongFormVideoPlayer(...)` 传 `resumePromptVisible = shouldShowResumePromptCard` + `resumePromptSlot = { TvResumePromptCard(lastPositionMs = ..., visible = shouldShowResumePromptCard, remainingSeconds = resumePromptCountdownTickRemaining(...), onContinue = ..., onStartFromBeginning = ...) }` |
| `tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt` | 同上路径处理。同时审查电视剧的连播提示卡（如挂在外层 Box）一并迁移到 LongFormVideoPlayer slot（若架构不同则保留并在 LaunchedEffect 聚合处加入其可见性 state） |
| `android-tv-app/tv-app/build.gradle.kts` | `versionCode = 69`、`versionName = "0.1.69"`（功能性变更，按仓库规则版本号 +1） |

## 3. 关键代码片段（参考实现，最终以代码 review 为准）

### 3.1 `LongFormPlayerFocusGuard.kt`

```kotlin
package com.chee.videos.core.ui

internal data class PlayerFocusGuardInput(
    val controlsVisible: Boolean,
    val subtitleSheetVisible: Boolean,
    val audioTrackSheetVisible: Boolean,
    val resumePromptVisible: Boolean,
    val backConfirmPromptVisible: Boolean,
    val playerErrorVisible: Boolean,
) {
    fun anyOverlayVisible(): Boolean =
        subtitleSheetVisible ||
            audioTrackSheetVisible ||
            resumePromptVisible ||
            backConfirmPromptVisible ||
            playerErrorVisible
}

/**
 * 上一次 input 与当前 input 之间，是否需要把焦点回收到 root Box：
 * - 任一 overlay 从 true 跃迁到 false
 * - 且当前没有其他 overlay 仍在显示
 * - 且控制条不可见或不需要焦点（focusInControls 由调用方在 effect 里另行判断）
 */
internal fun shouldReclaimRootFocus(
    previous: PlayerFocusGuardInput,
    current: PlayerFocusGuardInput,
): Boolean {
    val transitionedFalse =
        (previous.subtitleSheetVisible && !current.subtitleSheetVisible) ||
            (previous.audioTrackSheetVisible && !current.audioTrackSheetVisible) ||
            (previous.resumePromptVisible && !current.resumePromptVisible) ||
            (previous.backConfirmPromptVisible && !current.backConfirmPromptVisible) ||
            (previous.playerErrorVisible && !current.playerErrorVisible) ||
            (previous.controlsVisible && !current.controlsVisible)
    return transitionedFalse && !current.anyOverlayVisible()
}
```

### 3.2 `LongFormVideoPlayer.kt`（关键插入点）

```kotlin
@Composable
fun LongFormVideoPlayer(
    // 既有参数...
    resumePromptVisible: Boolean = false,
    resumePromptSlot: (@Composable () -> Unit)? = null,
) {
    // 既有 state...

    var previousFocusGuardInput by remember {
        mutableStateOf(
            PlayerFocusGuardInput(
                controlsVisible = controlsVisible,
                subtitleSheetVisible = subtitleSheetVisible,
                audioTrackSheetVisible = audioTrackSheetVisible,
                resumePromptVisible = resumePromptVisible,
                backConfirmPromptVisible = false,  // 由父级状态映射，见 §3.3
                playerErrorVisible = false,
            ),
        )
    }
    val currentFocusGuardInput = PlayerFocusGuardInput(
        controlsVisible = controlsVisible,
        subtitleSheetVisible = subtitleSheetVisible,
        audioTrackSheetVisible = audioTrackSheetVisible,
        resumePromptVisible = resumePromptVisible,
        backConfirmPromptVisible = false,
        playerErrorVisible = false,
    )
    LaunchedEffect(currentFocusGuardInput) {
        if (tvMode && shouldReclaimRootFocus(previousFocusGuardInput, currentFocusGuardInput)) {
            requestRootFocusWhenReady()
        }
        previousFocusGuardInput = currentFocusGuardInput
    }

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black)
            .focusRequester(rootFocusRequester)
            .focusable()
            .onFocusChanged { focusState ->
                if (tvMode &&
                    !focusState.hasFocus &&  // 整个子树都不持焦
                    !currentFocusGuardInput.anyOverlayVisible() &&
                    !focusInControls
                ) {
                    requestRootFocusWhenReady()
                }
            }
            .onPreviewKeyEvent { event -> /* 不动 */ }
            // 既有 pointerInput...
    ) {
        // VlcLongFormSurface / Poster / SeekPreview / CenterFeedback / Controls 既有内容

        // 新增：续播卡 slot 放在 controls 之上、与既有 BottomStart 位置一致
        if (resumePromptSlot != null) {
            Box(
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .padding(
                        start = TvResumePromptTokens.HorizontalPaddingDp,
                        bottom = TvResumePromptTokens.BottomPaddingDp,
                    ),
            ) {
                resumePromptSlot()
            }
        }
    }
}
```

### 3.3 `TvLongFormPlayerScreen.kt` 调用方

```kotlin
LongFormVideoPlayer(
    // 既有参数...
    resumePromptVisible = shouldShowResumePromptCard,
    resumePromptSlot = {
        TvResumePromptCard(
            lastPositionMs = resumePromptLastPositionMs,
            visible = shouldShowResumePromptCard,
            remainingSeconds = resumePromptCountdownTickRemaining(resumePromptRemainingMs),
            onContinue = { resumePromptDismissed = true },
            onStartFromBeginning = {
                mediaPlayer.time = 0L
                resumePromptDismissed = true
            },
        )
    },
)
// 删除原外层 Box 里的 TvResumePromptCard(...) 调用
```

`TvSeriesPlayerScreen.kt` 同样路径。

### 3.4 backConfirmPromptVisible / playerErrorVisible 的接入

PRD §1 列出的 B4 路径（返回二次确认）实际上叠在外层 Box，**不在** LongFormVideoPlayer 子树内。两种处理方式：

- **方案 a**：在 LongFormVideoPlayer 新增对应可见性参数 `backConfirmPromptVisible: Boolean = false` 与 `playerErrorVisible: Boolean = false`，由父级把 `showBackConfirmPrompt` / `playerErrorMessage != null` 透传进来。这样 LaunchedEffect 聚合能感知到 B4 / B5 路径的跃迁，焦点兜底完整。
- **方案 b**：把 BackConfirmPrompt / Error Surface 也内嵌到 LongFormVideoPlayer 子树。Compose 焦点自然回收。

**采用方案 a**：因为这两个 overlay 的 dispose 由父级状态控制，槽位化反而把状态机搬来搬去，更乱。透传 bool 是最小改动。`TvSeriesPlayerScreen.kt` 同步透传。

## 4. 单测（必须）

### 4.1 `LongFormPlayerFocusGuardTest.kt`

覆盖 `shouldReclaimRootFocus` 的所有重要分支：

- T1：续播卡 true→false 且无其他 overlay → 返回 true
- T2：续播卡 true→false 但字幕 picker 还在 → 返回 false（不抢字幕 picker 的焦点）
- T3：字幕 picker true→false 且无其他 overlay → 返回 true
- T4：音轨 picker true→false 且续播卡还在 → 返回 false
- T5：返回二次确认 true→false 且无其他 overlay → 返回 true
- T6：playerError true→false 且无其他 overlay → 返回 true
- T7：controls 5 秒 auto‑hide true→false 且无其他 overlay → 返回 true
- T8：input 无变化 → 返回 false
- T9：input 反向变化（false→true，新打开 overlay）→ 返回 false
- T10：多 overlay 同时 false→true→false 复杂场景

### 4.2 `LongFormVideoPlayerSpecTest.kt`

源文 audit（用 `Files.lines` 读取 .kt 文件做字符串扫描）：

- S1：`LongFormVideoPlayer.kt` 必须 `import com.chee.videos.core.ui.PlayerFocusGuardInput` 和 `shouldReclaimRootFocus`
- S2：`LongFormVideoPlayer.kt` 必须包含 `resumePromptSlot` 参数定义
- S3：`LongFormVideoPlayer.kt` 必须存在 `onFocusChanged` 调用
- S4：`TvLongFormPlayerScreen.kt` **不**直接出现 `TvResumePromptCard(...)` 在 `Box { ... LongFormVideoPlayer(...) ... TvResumePromptCard(...) ... }` 这样的兄弟位置（粗匹配：`TvResumePromptCard` 只能出现在 `resumePromptSlot = {` 块里）
- S5：`TvSeriesPlayerScreen.kt` 同 S4

## 5. 验证脚本

### 5.1 静态

```bash
cd android-tv-app
./gradlew --no-daemon :tv-app:compileDebugKotlin
./gradlew --no-daemon :tv-app:testDebugUnitTest
./gradlew --no-daemon :tv-app:assembleDebug
./gradlew --no-daemon :tv-app:assembleDebugAndroidTest
```

源文/版本/CONTEXT 扫描（在 repo 根）：

```bash
rg -n 'media3-' android-tv-app/tv-app/build.gradle.kts           # 应无（沿用 LibVLC 迁移结论）
rg -n 'versionCode = 69' android-tv-app/tv-app/build.gradle.kts  # 必须有
rg -n 'TvResumePromptCard\(' android-tv-app/tv-app/src/main      # 期望仅出现在 TvResumePromptCard.kt 定义与三处 resumePromptSlot 调用内
git diff --check
rg -n $'�' android-tv-app/tv-app/src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt \
    android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt \
    android-tv-app/tv-app/src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt
```

### 5.2 手测

详见 `review.md §1`。**未跑通真机/模拟器之前不允许 mark DONE.md**。

## 6. CONTEXT.md sync（实施完成时）

在 CONTEXT.md TV 段（紧邻 [[TV 长视频 LibVLC 内核]] 条目）插入 PRD §5 的三条术语。注意 `[[兄弟节点焦点抢占]]` 与 `[[续播提示卡]]` 这两个反向链接对应的术语：若尚未沉淀，本任务**不**新增；只在新增的三条里用 `[[...]]` 标记，等真正用到时再补。

## 7. 提交

提交信息（Chinese-first）：

```
修复TV长视频焦点真空导致遥控器键失灵

- 续播卡从屏幕级兄弟节点内嵌到 LongFormVideoPlayer 子树，dispose 时焦点自然回收
- LongFormVideoPlayer 内部聚合六个 overlay 可见性，任一 true→false 跃迁触发 root 焦点请求
- root Box 加 onFocusChanged 兜底，覆盖 Dialog dismiss 等 LaunchedEffect 不通知的路径
- 版本号 0.1.69 / 69
```
