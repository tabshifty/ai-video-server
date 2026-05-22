# IPTV 播放页依赖 LibVLC 原生桥接和事件回调，Release R8 时保留 VLC API 面。
-keep class org.videolan.** { *; }
-dontwarn org.videolan.**

# API 模型只通过 Retrofit suspend 签名与 Gson 反射使用。
# Release R8 若 shrink 掉 envelope/payload/dto/request 类，设备上会退化为错误类型或空数据并触发加载失败。
-keep class com.chee.videos.core.model.** { *; }

# TV 授权配对模型已有线上 Release 崩溃记录，保留显式规则作为回归提示。
-keep class com.chee.videos.core.model.TvAuth* { *; }
