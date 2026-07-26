package com.chee.videos

import android.app.Application
import coil.ImageLoader
import coil.ImageLoaderFactory
import coil.disk.DiskCache
import coil.memory.MemoryCache
import dagger.hilt.android.HiltAndroidApp

@HiltAndroidApp
class VideoApp : Application(), ImageLoaderFactory {

    // TV 机顶盒典型 1–2GB 物理内存、单应用 heap 常被限到 96–192MB。
    // Coil 默认内存缓存约占 heap 20%（低内存设备 15%），在小 heap 上装不下
    // 首页海报 + hero backdrop + 相邻货架的预取；TV 是单前台应用、无多任务内存压力，
    // 抬到 25% 换取「横向划回已看过的海报即时命中」。用百分比而非字节数：不同盒子 heap 差数倍。
    override fun newImageLoader(): ImageLoader =
        ImageLoader.Builder(this)
            .memoryCache {
                MemoryCache.Builder(this)
                    .maxSizePercent(0.25)
                    .build()
            }
            .diskCache {
                DiskCache.Builder()
                    // 沿用 Coil 默认目录名以复用既有缓存目录，仅收紧上限：
                    // TV 盒子内置存储紧张且被视频缓存挤占，128MB 足以容纳数千张海报缩略图，
                    // 远小于 Coil 默认「可用磁盘 2%」在大存储设备上可能膨胀出的 GB 级占用。
                    .directory(cacheDir.resolve("image_cache"))
                    .maxSizeBytes(128L * 1024 * 1024)
                    .build()
            }
            // 全局关闭 crossfade（显式声明默认值以固化决策）：Coil 仅豁免内存缓存命中，
            // 磁盘命中仍会重新淡入，货架横向划出再划回时海报会反复闪淡入；需要淡入的位点
            // （如短视频封面）已由调用方用 AnimatedVisibility 自行控制，全局开启会叠成双重淡入。
            // 注意：海报缩略图 URL（/api/v1/videos/{id}/thumbnail）不带版本参数，
            // 必须尊重服务端缓存响应头（Coil 默认行为），否则封面重刮削后 TV 端会永久展示旧图。
            .crossfade(false)
            .build()
}
