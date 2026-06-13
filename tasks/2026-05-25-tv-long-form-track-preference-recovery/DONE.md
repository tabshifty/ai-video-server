# DONE：TV 长视频字幕/音轨偏好恢复

- 完成日期：2026-06-14
- 原实现提交：`9ee66d1`（修复TV长视频字幕音轨偏好不记忆）
- 恢复到当前主线提交：`3a4446b`（Revert "Revert "修复TV长视频字幕音轨偏好不记忆""）
- 收口基线：`9505659`（收敛 DV 重试快照语义）之后的当前工作区

## 收口说明

本次按现有代码收口，不回填旧 review 中的 `0.1.70` 历史版本断言；当前 TV 端版本已推进到 `0.1.89`。任务目标对应的 VLC Playing gate、type-only preference fallback、音轨状态回灌、测试与 `CONTEXT.md` 术语仍在当前主线中。

## 验证摘要

- 自动化验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过。
- 静态核对：`addSlave(IMedia.Slave.Type.Subtitle` 同时出现在 `TvLongFormPlayerScreen.kt` 与 `TvSeriesPlayerScreen.kt`。
- 静态核对：`LongFormVideoPlayer.kt` 的 `onSelectAudioTrack` 同时存在 `isUserAction=false` 的自动回灌路径和 `isUserAction=true` 的用户选择路径。
- 文档核对：`CONTEXT.md` 已包含 `VLC Playing gate`、`Type-only preference fallback`、`Audio LaunchedEffect 状态回灌`。

## 人工验收

未在本次收口中重新执行 `review.md §1` 的 A1~A7 真机/模拟器手测脚本；本标记表示代码与自动化准入已按当前主线收口。若后续重新发现字幕或音轨偏好恢复问题，应作为新问题复现并另开任务。
