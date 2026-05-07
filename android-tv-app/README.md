# android-tv-app

家用视频 Android TV 独立客户端（Kotlin + Jetpack Compose）。

## 当前定位

- 独立 TV 工程目录，不依赖 `../android-app` 的源码目录
- 首启链路：服务器选择 -> TV 配对登录 -> 首页 -> 搜索 -> 详情 -> 播放器
- 仅包含长视频能力：电视剧、电影、AV

## 本地运行

```bash
cd android-tv-app
./gradlew :tv-app:assembleDebug
```

在 Android Studio 中打开 `android-tv-app` 目录即可。
