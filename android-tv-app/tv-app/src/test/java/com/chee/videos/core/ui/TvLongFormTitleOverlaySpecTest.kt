package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormTitleOverlaySpecTest {
    @Test
    fun longFormPlayerUsesTvTitleOverlayOnlyInTvMode() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")

        assertTrue(source.contains("TvLongFormTitleOverlay("))
        assertTrue(source.contains("buildTvLongFormTitleOverlayData("))
        assertTrue(source.contains("if (tvMode)"))
    }

    @Test
    fun titleOverlayUsesTextShadowAndTokenizedValues() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/TvLongFormTitleOverlay.kt")

        assertTrue(source.contains("object TvLongFormTitleOverlayTokens"))
        assertTrue(source.contains("shadow = titleOverlayShadow()"))
        assertTrue(source.contains("return Shadow("))
        assertFalse(source.contains("Modifier.shadow("))
        assertFalse(source.contains("22.sp"))
        assertFalse(source.contains("16.sp"))
        assertFalse(source.contains("Offset(0f, 2f)"))
        assertFalse(source.contains("4f"))
        assertFalse(source.contains("Color(0xCC000000)"))
        assertFalse(source.contains("RoundedCornerShape"))
        assertTrue(source.contains("maxLines = 1"))
        assertTrue(source.contains("overflow = TextOverflow.Ellipsis"))
    }

    @Test
    fun longFormPlayerRoutesTvKeysThroughSingleHelperAndRemovesOldHiddenTransportRouter() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        val keyHandler = source.substringAfter(".onPreviewKeyEvent").substringBefore(".onSizeChanged")

        assertTrue(source.contains("resolveTvRemoteKeyAction("))
        assertFalse(source.contains("resolveTvHiddenTransportKeyAction"))
        assertFalse(source.contains("TvHiddenTransportKeyAction"))
        assertFalse(source.contains("handleTvTransportKey"))
        assertFalse(keyHandler.contains("KEYCODE_"))
    }

    @Test
    fun fadeAnimationsUseTvMotionTokens() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        assertFalse(source.contains("fadeIn()"))
        assertFalse(source.contains("fadeOut()"))
        assertTrue(source.contains("fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard))"))
        assertTrue(source.contains("fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard))"))
    }

    @Test
    fun tvSeriesCallSitePassesSeriesAndEpisodeMetadata() {
        val source = readSource("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt")
        val call = source.substringAfter("LongFormVideoPlayer(").substringBefore("TvAutoplayPromptCard(")

        assertTrue(call.contains("title = series.title.ifBlank"))
        assertTrue(call.contains("seriesTitleForOverlay = series.title"))
        assertTrue(call.contains("seasonNumber = uiState.selectedSeasonNumber"))
        assertTrue(call.contains("episodeNumber = uiState.selectedEpisodeNumber"))
        assertTrue(call.contains("episodeTitle = currentEpisode?.title"))
    }

    private fun readSource(relative: String): String {
        val path = Path.of(relative)
        assertTrue("$relative 必须存在", path.exists())
        return path.readText()
    }
}
