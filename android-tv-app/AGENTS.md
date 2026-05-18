# AGENTS.md

## Scope
- 本文件适用于 `android-tv-app/` 整个独立 Android TV 工程。

## Commands
- 构建：`cd android-tv-app && ./gradlew :tv-app:assembleDebug`
- 单测：`cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest`

## 版本管理
- 版本文件：`tv-app/build.gradle.kts`。
- TV 端功能修改必须同步更新 `versionCode` 和 `versionName`。
- 默认递增规则：`versionCode +1`，`versionName` 的 patch 位 `+1`，例如 `0.1.0` -> `0.1.1`。
- 功能更新后若形成长期有效的术语、决策、接口或踩坑经验，必须追加到根级 `CONTEXT.md`。

## Rules
- TV 与手机端目录隔离，禁止通过 `sourceSets`、相对路径源码引用或 Gradle project dependency 依赖 `../android-app` 源码。
- TV 工程仅实现长视频相关能力；不要在此工程引入短视频、图片合集、上传或手机端互动流。
- 若 TV 与手机端都需要同一逻辑，默认在 TV 工程内独立维护，除非任务明确要求抽共享层。
