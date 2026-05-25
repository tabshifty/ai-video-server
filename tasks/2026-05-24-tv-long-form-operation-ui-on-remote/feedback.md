# Feedback：TV 长视频遥控操作 UI 真机回归修复

- 日期：2026-05-25
- 关联 PRD：`./prd.md`
- 关联 Implement：`./implement.md`
- 关联 Review：`./review.md`
- 初版完成提交：`1897c475`

## 1. 真机实测发现的两个故障

1. 进入视频播放后，等待 5 秒 UI 自动消失，**遥控器完全失效**，约 30 秒后才能再次响应。
2. 执行一次 ←/→ 快进快退后会唤起 [[TV 操作 UI 层]]，但**继续按 ←/→ 无法继续 seek**（焦点似乎被环绕劫持）。

## 2. 根因分析

`core/ui/LongFormVideoPlayer.kt` 根 Box 的 modifier 顺序错了：

```kotlin
.background(Color.Black)
.focusable()                          // ← focusable 在前
.focusRequester(rootFocusRequester)   // ← focusRequester 在后
.onPreviewKeyEvent { ... }
```

Compose 语义：`Modifier.focusRequester(fr)` 把 `fr` 绑定到链中**之后**出现的 focusable 节点。当前顺序里 `focusRequester` 之后没有任何 focusable（`onPreviewKeyEvent` / `onSizeChanged` / `pointerInput` 都不是 focusable 节点），所以 `rootFocusRequester.tryRequestFocus()` **静默失败**。

衍生结果：

- 5 秒 hide 触发 `scheduleAutoHideControls` 内 `requestRootFocusWhenReady` → `tryRequestFocus` 是 no-op；fadeOut 完成、按钮从树移除后焦点丢失 → 无焦点持有者 → 遥控器按键无人接收（Bug 1）。
- 一次 ← 唤起 controls 后，按钮入场被 Compose 2D 焦点搜索兜底获焦，`focusInControls` 经按钮 `onFocusChanged` 翻 true；下一次 ← 在 `resolveTvRemoteKeyAction` 中走 `PassThrough` → `focusProperties.left` 焦点环绕，而非 Seek（Bug 2）。

对比同期已知正确实现：`feature/tv/TvIptvScreen.kt:219-220` / `TvPosterWallScreen.kt:247` 全部是 `.focusRequester(...).focusable()` 顺序。本任务在 `LongFormVideoPlayer` 引入 modifier 时与项目惯例反了。

## 3. 修复

### 3.1 红灯测试

`tv-app/src/test/java/com/chee/videos/core/ui/TvLongFormTitleOverlaySpecTest.kt` 追加两条源文审计：

- `rootBoxBindsFocusRequesterBeforeFocusable`：断言根 Box modifier 链中 `focusRequester(rootFocusRequester)` 必须排在 `focusable()` 之前。
- `seekBranchReanchorsFocusToRootAfterShowingControls`：断言 `handleTvRemoteKeyAction` 的 Seek 分支必须调用 `requestRootFocusWhenReady()`。

### 3.2 代码修复

1. `LongFormVideoPlayer.kt:540-541`：交换两个 modifier 顺序为 `.focusRequester(rootFocusRequester).focusable()`，让 `tryRequestFocus()` 真正能锚回根节点。
2. `LongFormVideoPlayer.kt:308-318`：Seek 分支追加 `focusInControls = false` + `requestRootFocusWhenReady()`，防御按钮入场抢焦——即使一帧内按钮短暂抢了焦点，下一帧 `LaunchedTvInitialFocus` 也会把焦点拉回根 Box。

### 3.3 版本号

`android-tv-app/tv-app/build.gradle.kts`：`versionCode = 68`、`versionName = "0.1.67"`。

## 4. 验证

- `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest'`：先红（两条新审计失败）后绿。
- `cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest`：全量 TV 单测通过。
- `cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug`：构建通过。
- 真机回归（必跑）：
  - 进入电影 / `18+` / 电视剧任一播放器，等 5 秒 UI 隐藏，按 ← 应立即 seek + 唤起 UI（不再 30 秒卡顿）。
  - 重复按 ← / → 应连续 seek（不再被焦点环绕劫持）。
  - 按 ↓ 才进入 controls 焦点；UP 退回根；BACK 优先收 UI。
