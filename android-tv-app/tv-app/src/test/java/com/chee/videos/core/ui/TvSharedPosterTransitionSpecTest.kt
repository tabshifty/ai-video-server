package com.chee.videos.core.ui

import java.nio.file.Path
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSharedPosterTransitionSpecTest {

    private fun readMain(relative: String): String =
        Path.of("src/main/java", relative).toFile().readText()

    @Test
    fun `shared poster helper declares both composition locals and modifier extension`() {
        val source = readMain("com/chee/videos/core/ui/TvSharedPoster.kt")

        assertTrue(
            "TvSharedPoster.kt 必须暴露 LocalTvSharedTransitionScope CompositionLocal，供 TvShellApp 在 SharedTransitionLayout 内提供",
            source.contains("LocalTvSharedTransitionScope"),
        )
        assertTrue(
            "TvSharedPoster.kt 必须暴露 LocalTvAnimatedContentScope CompositionLocal，供每个 composable() 目的页面提供",
            source.contains("LocalTvAnimatedContentScope"),
        )
        assertTrue(
            "必须暴露统一的 Modifier 扩展 tvSharedSeriesPoster(seriesId)，集中处理 shared-element 绑定与作用域缺失回退",
            source.contains("fun Modifier.tvSharedSeriesPoster"),
        )
        assertTrue(
            "shared-content key 必须使用稳定命名空间前缀 tv-series-poster-，避免与其他 shared key 撞键",
            source.contains("tv-series-poster-"),
        )
        assertTrue(
            "shared-element 调用使用实验 API，必须标注 @OptIn(ExperimentalSharedTransitionApi::class)",
            source.contains("ExperimentalSharedTransitionApi"),
        )
        assertTrue(
            "scope 为 null 时 helper 应原样返回 Modifier，避免非 SharedTransitionLayout 环境崩溃",
            source.contains("return this") || source.contains("?: return this"),
        )
    }

    @Test
    fun `tv shell app wraps navhost in shared transition layout`() {
        val source = readMain("com/chee/videos/tv/TvShellApp.kt")

        assertTrue(
            "TvShellApp 必须 import SharedTransitionLayout",
            source.contains("import androidx.compose.animation.SharedTransitionLayout"),
        )
        assertTrue(
            "TvShellApp 必须在 NavHost 外层用 SharedTransitionLayout 包裹，否则跨 destination shared-element 不会生效",
            source.contains("SharedTransitionLayout"),
        )
        assertTrue(
            "TvShellApp 必须把 SharedTransitionLayout 作用域注入 LocalTvSharedTransitionScope CompositionLocal",
            source.contains("LocalTvSharedTransitionScope provides"),
        )
        assertTrue(
            "TvShellApp 必须把每个 composable() 块的 AnimatedContentScope 注入 LocalTvAnimatedContentScope CompositionLocal",
            source.contains("LocalTvAnimatedContentScope provides"),
        )
    }

    @Test
    fun `poster wall applies shared element only for tv series posters`() {
        val source = readMain("com/chee/videos/feature/tv/TvPosterWallScreen.kt")

        assertTrue(
            "TvPosterWallScreen 必须导入 tvSharedSeriesPoster modifier",
            source.contains("import com.chee.videos.core.ui.tvSharedSeriesPoster"),
        )
        assertTrue(
            "TvPosterWallScreen 必须在卡片图片层调用 tvSharedSeriesPoster(...)",
            source.contains("tvSharedSeriesPoster("),
        )
        assertTrue(
            "TvPosterWallScreen 必须仅在电视剧类型应用 shared-element：源文里必须能看到 type / kind 对 tv 的条件分支",
            source.contains("\"tv\""),
        )
    }

    @Test
    fun `series detail poster declares matching shared element`() {
        val source = readMain("com/chee/videos/feature/tv/TvSeriesDetailScreen.kt")

        assertTrue(
            "TvSeriesDetailScreen 必须导入 tvSharedSeriesPoster modifier",
            source.contains("import com.chee.videos.core.ui.tvSharedSeriesPoster"),
        )
        assertTrue(
            "TvSeriesDetailScreen 必须在详情小海报上调用 tvSharedSeriesPoster(...)，使 key 与海报墙对齐",
            source.contains("tvSharedSeriesPoster("),
        )
    }
}
