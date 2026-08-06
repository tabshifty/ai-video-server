package com.chee.videos.core.model

import java.io.File
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class PhoneContentRemovalSpecTest {
    private val sourceRoot = File("src/main/java/com/chee/videos")

    @Test
    fun `phone app removes movie and series navigation and dedicated series module`() {
        val navSource = File(sourceRoot, "VideoHomeApp.kt").readText()
        val homeSource = File(sourceRoot, "feature/home/HomeScreen.kt").readText()

        assertFalse(File(sourceRoot, "feature/tv").exists())
        assertFalse(File(sourceRoot, "core/di/TvRepositoryModule.kt").exists())
        listOf("TvSeriesRoutePattern", "TvPlayerRoutePattern", "TvSeriesDetailScreen", "TvSeriesPlayerScreen").forEach { symbol ->
            assertFalse("手机导航仍包含 $symbol", navSource.contains(symbol))
        }
        listOf("TvCatalogScreen", "onOpenTvSeries", "onOpenTvContinueWatching", "\"movie\" ->", "\"episode\" ->").forEach { symbol ->
            assertFalse("手机首页仍包含 $symbol", homeSource.contains(symbol))
        }
    }

    @Test
    fun `phone app keeps tv authorization and short casting capabilities`() {
        assertTrue(File(sourceRoot, "feature/tvauth/TvAuthApprovalScreen.kt").isFile)
        assertTrue(File(sourceRoot, "feature/tvauth/TvAuthDeepLink.kt").isFile)
        assertTrue(File(sourceRoot, "core/ui/cast/TvCast.kt").isFile)

        val navSource = File(sourceRoot, "VideoHomeApp.kt").readText()
        assertTrue(navSource.contains("TvAuthApprovalScreen"))
        assertTrue(navSource.contains("TvAuthDeepLinkParser"))
        assertTrue(navSource.contains("ShortSearchRemoteControlScreen"))
    }

    @Test
    fun `mixed phone surfaces share the same content policy`() {
        val requiredSources = listOf(
            "feature/mine/MineViewModel.kt",
            "feature/actor/ActorDetailViewModel.kt",
            "feature/player/UnifiedPlayerViewModel.kt",
        )
        requiredSources.forEach { relativePath ->
            val source = File(sourceRoot, relativePath).readText()
            assertTrue("$relativePath 未使用统一分页过滤", source.contains("loadPhoneContentBatch"))
        }

        val detailSource = File(sourceRoot, "feature/detail/DetailViewModel.kt").readText()
        assertTrue(detailSource.contains("isPhoneSupportedVideoType"))

        val actorScreenSource = File(sourceRoot, "feature/actor/ActorDetailScreen.kt").readText()
        assertFalse(actorScreenSource.contains("totalCount"))
        assertFalse(actorScreenSource.contains("部作品"))
    }

    @Test
    fun `unsupported detail and player routes use the shared navigation fallback`() {
        val navSource = File(sourceRoot, "VideoHomeApp.kt").readText()
        val detailScreenSource = File(sourceRoot, "feature/detail/DetailScreen.kt").readText()
        val playerScreenSource = File(sourceRoot, "feature/player/UnifiedPlayerScreen.kt").readText()

        assertTrue(navSource.contains("该内容不在手机端提供"))
        assertTrue(navSource.contains("returnFromUnavailableContent"))
        assertTrue(detailScreenSource.contains("onUnsupportedContent"))
        assertTrue(playerScreenSource.contains("onUnsupportedContent"))
    }
}
