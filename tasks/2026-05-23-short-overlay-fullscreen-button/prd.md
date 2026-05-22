# PRD：手机端短视频浮层「全屏播放」按钮

- 日期：2026-05-23
- 目标端：`android-app/`（手机 App）
- 范围：短视频播放浮层

## 1. 用户故事

作为手机端用户，我在搜索结果（以及其他三处短视频列表）里点开一个短视频后，能够在右侧操作栏看到一个「全屏播放」按钮；点一下后，屏幕**强制旋转到横屏**，进入一个带「进度条 + 快进/快退 + 亮度 + 音量」控制条的长视频风格播放体验；该状态下视频**单循环重播**，不会被自动切到下一条；再次点击按钮或按系统返回键，可以退出全屏，回到原来的短视频浮层并继续播放原视频。

## 2. 作用域

本次改动覆盖手机端**所有四处**短视频播放浮层；全部使用同一个共享的「短视频全屏播放」组件，避免在四个文件里各 copy 一份。

| # | 入口 | 文件 |
|---|------|------|
| 1 | 短视频搜索结果浮层 | `feature/shortsearch/ShortSearchScreen.kt`（`ShortSearchPlayerOverlay`） |
| 2 | 短视频发现页（标签/合集） | `feature/shortdiscover/ShortDiscoverScreen.kt` |
| 3 | 主页短视频信息流 | `feature/shorts/ShortFeedScreen.kt` |
| 4 | 个人 history/like/favorite 列表的短视频播放 | `feature/player/UnifiedPlayerScreen.kt`（仅短视频分支） |

不在本次范围：TV 端（`android-tv-app/`）、电影/电视剧/`18+` 详情页的全屏入口（这些已有自己的全屏路径）、后端接口、admin-web。

## 3. 验收标准

### 3.1 入口

- A1：四处短视频浮层右侧操作栏底部新增「全屏播放」按钮（图标 `Icons.Filled.Fullscreen`，contentDescription = "全屏播放"），位于现有「播放模式（顺序/单循环）」按钮之下。
- A2：按钮视觉风格沿用现有 `ShortVideoOverlayActionButton` 的圆形/icon-only 风格，焦点态/按压态与其他按钮一致。
- A3：按钮可用性始终为 true（不依赖详情接口加载状态）。

### 3.2 进入全屏

- B1：点击全屏按钮后，屏幕**立即强制旋转到 landscape**（90° 顺时针，`SCREEN_ORIENTATION_LANDSCAPE`），无论系统「自动旋转」开关状态如何。
- B2：横屏画面铺满整个屏幕，状态栏与导航栏隐藏（`WindowInsetsController.hide(systemBars())`）。
- B3：播放位置**无缝接续** —— 进入全屏的那一刻视频不会从头开始、不会卡顿，从竖屏短视频当前的播放位置继续播放。
- B4：控制条来自现有共享组件 `core/ui/LongFormVideoPlayer.kt`，包含：
  - 顶部返回 icon + 视频标题
  - 中央播放/暂停手势 + 反馈
  - 底部进度条（可拖动 seek）+ 当前时长/总时长
  - 左半屏垂直拖拽调节屏幕亮度
  - 右半屏垂直拖拽调节媒体音量
  - 水平拖拽预览快进/快退
  - 长按倍速
  - 右下「退出全屏」icon（`Icons.Filled.FullscreenExit`）
- B5：视频画面以 `RESIZE_MODE_FIT` 显示（保持原始比例，可能上下/左右黑边，不裁切）—— 与 `LongFormVideoPlayer` 默认一致。

### 3.3 全屏期间的播放策略

- C1：进入全屏的同时，`ExoPlayer.repeatMode` 被强制改为 `REPEAT_MODE_ONE`（**单视频循环重播**），覆盖用户在浮层里设置的 `LOOP_ONE / AUTO_NEXT` 偏好。
- C2：全屏期间**不会自动切到下一条短视频** —— 即原 `onPlaybackStateChanged → STATE_ENDED → pagerState.animateScrollToPage(next)` 分支被屏蔽。
- C3：全屏期间用户无法上下滑动切视频（手势被 `LongFormVideoPlayer` 的 fillMaxSize 覆盖层吃掉，且 VerticalPager 不在视觉栈上）。
- C4：用户原 `playbackMode` 偏好**不被写入 DataStore**，全屏只是临时覆盖，退出全屏后恢复。

### 3.4 退出全屏

- D1：用户可通过两种方式退出全屏：
  - **系统返回键** —— 第一次按返回退出全屏，回到竖屏短视频浮层；再按一次返回才会关闭整个搜索/发现/信息流浮层。
  - **控制条右下「退出全屏」icon** —— 等价于上面的第一种返回。
- D2：退出全屏后：
  - 屏幕方向恢复 `SCREEN_ORIENTATION_UNSPECIFIED`（跟随系统）
  - 系统栏（状态栏 + 导航栏）恢复显示
  - `ExoPlayer.repeatMode` 恢复为用户在浮层设置的 `playbackMode.toPlayerRepeatMode()`
  - 视频继续播放（不暂停）
  - 短视频浮层 UI 重新出现，停留在原视频的 pager 位置
- D3：退出全屏不会重新加载视频，播放位置接续。

### 3.5 失败/边界

- E1：进入全屏期间 App 切换到后台 → `ON_PAUSE` 时仍按现有 `sharedPlayer.pause()` 处理；回到前台 `ON_RESUME` 时仍处于全屏（不自动退出）、继续播放。
- E2：进入全屏期间发生认证过期（401 → AuthExpiredException）→ 仍走现有 `handleAuthError` 路径，整个浮层会关闭，全屏状态一并清理。
- E3：长视频 / 电影 / `18+` 等长视频已有的全屏入口（DetailScreen 全屏按钮、`LongFormVideoPlayer` 内部 `onToggleFullscreen`）不受本次改动影响。

## 4. 非功能要求

- F1：进入全屏的方向旋转延迟 < 300ms（含 PlayerView 重新 attach）。
- F2：全屏 ↔ 浮层切换时画面**不黑屏**（PlayerView surface 不重新创建，复用同一 ExoPlayer 实例）。
- F3：方向、系统栏、`repeatMode` 三项状态的进入/退出操作必须**对称且幂等** —— 若 Composable 异常退出（崩溃、配置变化），下次启动时不应残留横屏锁定或系统栏隐藏。
- F4：本次改动仅影响手机 App，不波及 TV 工程、后端、admin-web。
- F5：版本号必须按约定 +1：`android-app/app/build.gradle.kts` 的 `versionCode +1`、`versionName` 末位 +1。

## 5. 不做的事

- 不持久化全屏偏好（每次打开默认非全屏）
- 不在全屏画面里再嵌套上下滑切视频（C3 已禁用）
- 不改变长视频/电视剧 / `18+` 已有的全屏路径
- 不引入新的方向控制设置项（强制 LANDSCAPE 是硬决策）
- 不对短视频字幕做适配（短视频通常无字幕轨道，`LongFormVideoPlayer` 的字幕入口在 `subtitleTracks` 为空时优雅隐藏）

## 6. 术语收口（写入 CONTEXT.md 候选）

- **短视频全屏播放**：手机端短视频浮层中通过右下「全屏播放」按钮触发的横屏长视频风格播放体验。期间强制 `REPEAT_MODE_ONE`、屏蔽 VerticalPager 顺序切换，退出后恢复原状态。作用域仅限手机端四处短视频浮层，TV 端与长视频路径不套用。
