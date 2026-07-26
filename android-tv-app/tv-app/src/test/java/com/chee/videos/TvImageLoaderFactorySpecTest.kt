package com.chee.videos

import java.io.File
import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * TV 端全局图片加载配置审计：Coil 必须由应用级 ImageLoaderFactory 显式配置，
 * 缓存尺寸与淡入策略是针对 TV 机顶盒（小 heap、紧张存储、单前台应用）的既定决策。
 */
class TvImageLoaderFactorySpecTest {

    private val appSource: String by lazy {
        Path.of("src/main/java/com/chee/videos/VideoApp.kt").readText()
    }

    @Test
    fun videoAppProvidesExplicitImageLoader() {
        assertTrue(
            "VideoApp 必须实现 ImageLoaderFactory 并覆写 newImageLoader()",
            appSource.contains("ImageLoaderFactory") && appSource.contains("override fun newImageLoader()"),
        )
        assertTrue(
            "内存缓存必须显式按 heap 百分比设定（TV 盒子 heap 差异数倍，禁止字节数硬编码）",
            appSource.contains("MemoryCache.Builder") && appSource.contains("maxSizePercent("),
        )
        assertTrue(
            "磁盘缓存必须显式设定字节上限（禁止 Coil 默认按可用磁盘 2% 膨胀）",
            appSource.contains("DiskCache.Builder") && appSource.contains("maxSizeBytes("),
        )
        assertTrue(
            "全局必须关闭 crossfade：Coil 只豁免内存命中、磁盘命中仍会重复淡入，" +
                "货架划回会反复闪；需要淡入的位点由调用方 AnimatedVisibility 自行控制",
            appSource.contains(".crossfade(false)"),
        )
        assertTrue(
            "自建服务端缩略图 URL 含资源 ID，必须忽略缓存响应头以最大化磁盘命中",
            appSource.contains(".respectCacheHeaders(false)"),
        )
    }

    @Test
    fun liveScreensDoNotEnablePerRequestCrossfade() {
        val offenders = listOf(
            File("src/main/java/com/chee/videos/feature/tv"),
            File("src/main/java/com/chee/videos/core/ui"),
        )
            .flatMap { dir -> dir.walkTopDown().filter { it.isFile && it.extension == "kt" }.toList() }
            .filter { it.readText().contains("crossfade(true)") }
            .map { it.name }
        assertTrue(
            "TV 实编屏幕禁止逐请求开启 crossfade(true)（与全局关闭决策冲突），命中：$offenders",
            offenders.isEmpty(),
        )
    }
}
