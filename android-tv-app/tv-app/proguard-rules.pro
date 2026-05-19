# IPTV 播放页依赖 LibVLC 原生桥接和事件回调，Release R8 时保留 VLC API 面。
-keep class org.videolan.** { *; }
-dontwarn org.videolan.**
