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

    @Test
    fun rootBoxBindsFocusRequesterBeforeFocusable() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        val rootBoxModifierChain = source
            .substringAfter("Box(\n        modifier = modifier")
            .substringBefore(".onPreviewKeyEvent")

        val focusRequesterIdx = rootBoxModifierChain.indexOf(".focusRequester(rootFocusRequester)")
        val focusableIdx = rootBoxModifierChain.indexOf(".focusable()")

        assertTrue(
            "根 Box 必须挂 Modifier.focusRequester(rootFocusRequester)",
            focusRequesterIdx >= 0,
        )
        assertTrue(
            "根 Box 必须挂 Modifier.focusable()",
            focusableIdx >= 0,
        )
        assertTrue(
            "Modifier.focusRequester(rootFocusRequester) 必须排在 Modifier.focusable() 之前；反向顺序会让 tryRequestFocus() 无法绑定到根 focusable 节点，导致 5 秒自动隐藏后焦点丢失、遥控器失效",
            focusRequesterIdx < focusableIdx,
        )
    }

    @Test
    fun seekBranchReanchorsFocusToRootAfterShowingControls() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        val seekBranch = source
            .substringAfter("is TvRemoteKeyAction.Seek ->")
            .substringBefore("TvRemoteKeyAction.EnterFocus ->")

        assertTrue(
            "Seek 分支必须显式 requestRootFocusWhenReady()，防御按钮入场抢焦",
            seekBranch.contains("requestRootFocusWhenReady()"),
        )
    }

    private fun readSource(relative: String): String {
        val path = Path.of(relative)
        assertTrue("$relative 必须存在", path.exists())
        return path.readText()
    }
}
