# IPTV 播放页依赖 LibVLC 原生桥接和事件回调，Release R8 时保留 VLC API 面。
-keep class org.videolan.** { *; }
-dontwarn org.videolan.**

# TV 授权配对模型只通过 Retrofit suspend 签名与 Gson 反射使用。
# Release R8 若 shrink 掉这些 envelope/payload 类，设备上会退化为错误类型并触发 ClassCastException。
-keep class com.chee.videos.core.model.TvAuth* { *; }
