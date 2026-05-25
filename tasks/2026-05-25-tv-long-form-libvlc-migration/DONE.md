# DONE：TV 长视频 LibVLC 迁移

- 完成日期：2026-05-25
- 相关提交：
  - `51750977`：新增 TV 长视频 LibVLC 迁移任务文档与 ADR-0004
  - `c82967e7`：完成 TV 长视频 LibVLC 迁移
  - `2c944963`：记录 TV 长视频 LibVLC 模拟器验证
  - `b0bce451`：记录 TV 长视频 LibVLC 自动化审计
- 用户验收：2026-05-25 用户确认任务实测通过。

## 完成范围

- TV 长视频播放器（电影 / `18+` / 电视剧）从 Media3 ExoPlayer 迁移到 LibVLC `MediaPlayer`。
- ASS / SSA 字幕在后端按原文存储，TV 端通过 LibVLC / libass 自渲染。
- TV 长视频字幕与音轨偏好从不稳定 track id 改为 `language + type` 持久化，旧 DataStore track id key 在读写时清除。
- TV 长视频保留遥控器 controls、续播提示、连播提示、手动下一集、快进/快退步长、连按合并跳转、长按 2x、触屏手势与退出确认等交互契约。
- `CONTEXT.md` 与 ADR-0004 已同步长期技术约定。

## 验证摘要

- `go test ./internal/services ./pkg/ffmpeg -count=1` 通过。
- `go test ./... -count=1` 通过。
- `go test ./internal/services -run TestSubtitle -count=1` 通过。
- `go test ./pkg/ffmpeg -run 'Test.*Subtitle|TestBuildExtractSubtitleToAssArgs' -count=1` 通过。
- `cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:testDebugUnitTest` 通过。
- `cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过。
- `cd android-tv-app && ./gradlew --no-daemon --init-script /tmp/force-maven-central.gradle :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过。
- `adb -s emulator-5554 shell am instrument -w com.chee.videos.tv.test/androidx.test.runner.AndroidJUnitRunner` 通过，共 3 个 androidTest。
- Media3 残留、版本号、CONTEXT 术语、ADR、admin 字幕上传文案、`git diff --check` 与乱码扫描均已检查。
- `review.md` §1 的 ASS 复杂样式、TV 长视频交互回归、字幕/音轨切换、续播/自动连播、性能稳定性与服务端 ASS 原文存储场景已由用户实测确认通过。
