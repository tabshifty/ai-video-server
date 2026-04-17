# android-app

家用视频 Android 客户端（Kotlin + Jetpack Compose）。

## 当前已实现

- 首次启动局域网嗅探（`/healthz`）+ 手动输入 IP/端口兜底
- 服务器地址多条保存与切换
- 登录（含 token 存储）
- 首页四分类：短视频 / 电影 / 电视剧 / AV
- 短视频分类：抖音式全屏竖滑播放
- 详情占位页：基础信息 + 点赞/收藏/不喜欢

## 本地运行

```bash
cd android-app
./gradlew :app:assembleDebug
```

在 Android Studio 中打开 `android-app` 目录即可。
