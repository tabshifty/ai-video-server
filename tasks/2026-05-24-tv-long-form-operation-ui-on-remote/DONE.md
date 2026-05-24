# 完成记录：TV 长视频播放器操作 UI 跟随遥控器互动

- 完成日期：2026-05-24
- 关联提交：`8826107e`
- 任务目录：`tasks/2026-05-24-tv-long-form-operation-ui-on-remote`

## 完成摘要

- TV 长视频播放器 `tvMode=true` 已接入统一遥控键路由：LEFT/RIGHT seek 并唤起操作 UI，DOWN 进入控制条，UP 退出控制条焦点，BACK 在 UI 可见时优先收起 UI，UI 收起后继续走既有退出确认。
- 新增左上信息层：电影 / `18+` 显示主标题；电视剧显示剧名 +「第 X 季 · 第 Y 集 单集标题」副行，样式和阴影参数收口到 token。
- 控制条按钮使用 `focusProperties` 实现左右首尾环绕，Slider 不参与 TV 焦点链。
- 首次进入视频时操作 UI 自动亮起 5 秒，但焦点停留在播放器根，只有按 DOWN 才进入控制条。
- TV 版本号递增到 `versionCode = 67`、`versionName = "0.1.66"`，`CONTEXT.md` 已沉淀本次 TV 操作 UI 术语。

## 验证摘要

- 红灯：新增 TV 遥控键路由 / 标题叠层 / 自动隐藏 / 源文审计测试后，定向单测因 `TvRemoteKeyAction` 与 `buildTvLongFormTitleOverlayData` 未实现失败。
- 绿灯：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest --tests 'com.chee.videos.core.ui.TvLongFormRemoteKeyRoutingTest' --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlayDataTest' --tests 'com.chee.videos.core.ui.TvLongFormControlsAutoHideTest' --tests 'com.chee.videos.core.ui.TvLongFormTitleOverlaySpecTest' --tests 'com.chee.videos.core.ui.LongFormVideoPlayerTransportKeyTest'` 通过。
- 全量 TV 单测：`cd android-tv-app && ./gradlew --no-daemon :tv-app:testDebugUnitTest` 通过。
- TV 构建：`cd android-tv-app && ./gradlew --no-daemon :tv-app:assembleDebug` 通过。
- 连接设备 instrumentation：`cd android-tv-app && ./gradlew --no-daemon :tv-app:connectedDebugAndroidTest` 已完成 androidTest 编译与测试 APK 打包，但当前本机无连接设备，最终失败于 `DeviceException: No connected devices!`，未能执行真机/模拟器用例。
- 提交前扫描：`git diff --check` 通过；乱码扫描无输出。
