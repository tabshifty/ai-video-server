# AGENTS.md

## Scope
- 本文件适用于 `android-app/` 整个 Android 手机端工程。

## Commands
- 构建：`cd android-app && ./gradlew :app:assembleDebug`
- 单测：`cd android-app && ./gradlew :app:testDebugUnitTest`

## Rules
- 本工程仅承载手机端能力，禁止在此重新引入 `:tv-app` module 或 TV APK 构建入口。
- 手机端保留 TV 扫码授权确认能力，但不承载 TV 首页、TV 导航、TV 播放器运行时代码。
- Android TV 独立工程位于同级目录 `../android-tv-app`；两端通过后端 API 协作，不通过源码目录引用协作。
- 短视频播放页若同时存在多个覆盖操作按钮，必须使用显式纵向/横向动作栏布局与间距约束，禁止把多个按钮直接放进同一 `Box` 默认位置导致重合。
