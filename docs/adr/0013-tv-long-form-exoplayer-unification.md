# TV 长视频统一迁移到 ExoPlayer，IPTV 继续使用 LibVLC

TV 长视频（电影、`18+`、电视剧分集）决定从 LibVLC 一次性迁回统一的 ExoPlayer/Media3 内核，既有 Dolby Vision Media3 分支并入同一长视频内核；IPTV 直播继续保留 LibVLC。这个决策接受 ASS/SSA 字幕不再承诺 libass 级样式和特效等价，换取长视频播放路径、鉴权方式、Dolby Vision 门控和控制层适配的统一，避免继续维护普通 LibVLC、DV Media3 和 IPTV LibVLC 三套播放语义。

## 考虑过的替代方案

- **继续保留长视频 LibVLC**：能保留 libass 渲染能力，但普通长视频、DV 专用 Media3 和 IPTV LibVLC 的边界越来越复杂，播放器返回、字幕音轨、历史上报和错误处理需要持续在多套内核间对齐。
- **长视频双轨灰度或旧播放器开关**：短期回退方便，但会把遥控器控制层、续播、选集、连播、字幕音轨和诊断长期绑定到两套实现；本项目选择通过小提交和版本回滚处理严重兼容问题。
- **为了 ASS 保留长视频 LibVLC 分支**：保真字幕价值明确，但当前目标优先统一 TV 长视频播放内核；高保真 ASS 以后应作为独立能力重新设计，而不是让长视频长期保留第二播放内核。

## 后果

- TV 长视频失败页只提供重试、返回或诊断，不回退 LibVLC，也不提供旧播放器开关。
- `org.videolan.android:libvlc-all` 仍保留在 TV 工程中，但归属限定为 IPTV 直播路径。
- TV 长视频视频源和外挂字幕请求使用 ExoPlayer/Media3 HTTP data source 的 `Authorization: Bearer <token>` header，不再依赖 `access_token` query。
- 手机端播放器不纳入本决策，不抽共享播放层。

## 取代

- 取代 `0004-tv-long-form-libvlc-for-ass-rendering.md` 中“TV 长视频使用 LibVLC 以获得 libass 渲染”的决策。
- 更新 `0012-tv-series-shared-controls-for-mixed-playback-engines.md` 中“电视剧混合 LibVLC / Media3 播放内核”的前提；迁移后控制层仍保留，但长视频内核不再混合。
