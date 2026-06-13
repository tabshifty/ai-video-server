# DONE：TV 长视频播放器焦点兜底

- 完成日期：2026-06-14
- 原实现提交：`059bed4`（修复TV长视频焦点真空导致遥控器键失灵）
- 恢复到当前主线提交：`a3a69fe`（Revert "Revert "修复TV长视频焦点真空导致遥控器键失灵""）
- 收口基线：`9505659`（收敛 DV 重试快照语义）之后的当前工作区

## 收口说明

本次按现有代码收口，不回填旧 review 中的 `0.1.69` 历史版本断言；当前 TV 端版本已推进到 `0.1.89`。任务目标对应的焦点兜底代码、测试与 `CONTEXT.md` 术语仍在当前主线中。

## 验证摘要

- 自动化验证：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest :tv-app:assembleDebug :tv-app:assembleDebugAndroidTest` 通过。
- 静态核对：`TvResumePromptCard(` 仅出现在 `TvResumePromptCard.kt` 定义和 `TvLongFormPlayerScreen.kt` / `TvSeriesPlayerScreen.kt` 的 `resumePromptSlot` 调用内。
- 文档核对：`CONTEXT.md` 已包含 `TV 长视频焦点真空`、`LongFormVideoPlayer focus 兜底`、`续播提示卡内嵌位置`。

## 人工验收

未在本次收口中重新执行 `review.md §1` 的 R1~R10 真机/模拟器手测脚本；本标记表示代码与自动化准入已按当前主线收口。若后续重新发现遥控器焦点问题，应作为新问题复现并另开任务。
