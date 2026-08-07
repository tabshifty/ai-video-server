# TV App 下线 IPTV 客户端能力

> 状态：已接受（2026-08-07）

TV App 决定完整移除 IPTV 的首页入口、内部路由、频道获取与直播播放能力，以收缩客厅端产品表面和 IPTV 专属维护成本；后端播放列表、管理端维护能力、频道数据及 `GET /api/v1/tv/iptv/channels` 继续保留，已安装的旧版 TV App 和历史 APK 不做服务端封禁。新版不保留下线提示页，原 `tv/iptv` 仅是内部路由且没有外部 deep link，删除目的地后统一回到正常首页导航边界。

## 后果

- TV 首页固定为“电视剧、电影、18+、短视频、搜索、设置”，短视频继续是 action-only 入口。
- IPTV 专属页面、状态、客户端网络契约、资源和正向测试删除，并以负向源码契约防止入口或客户端 API 被误接回。
- `libvlc-all` 与 ARM 双 ABI APK 暂时保留，因为现有分发服务通过 APK 内单一 ARM 原生库识别 `armeabi-v7a` / `arm64-v8a` 产物；遗留 VLC 编译代码和分发模型迁移属于独立变更，不与本次客户端下线捆绑。

## 取代

- 取代 `0013-tv-long-form-exoplayer-unification.md` 与 `0018-tv-long-form-ott-playback-session.md` 中“TV IPTV 客户端继续使用 LibVLC”的当前能力前提；两份 ADR 的长视频 Media3 决策继续有效。
