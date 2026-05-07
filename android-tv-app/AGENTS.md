# AGENTS.md

## Scope
- 本文件适用于 `android-tv-app/` 整个独立 Android TV 工程。

## Commands
- 构建：`cd android-tv-app && ./gradlew :tv-app:assembleDebug`
- 单测：`cd android-tv-app && ./gradlew :tv-app:testDebugUnitTest`

## Rules
- TV 与手机端目录隔离，禁止通过 `sourceSets`、相对路径源码引用或 Gradle project dependency 依赖 `../android-app` 源码。
- TV 工程仅实现长视频相关能力；不要在此工程引入短视频、图片合集、上传或手机端互动流。
- 若 TV 与手机端都需要同一逻辑，默认在 TV 工程内独立维护，除非任务明确要求抽共享层。
