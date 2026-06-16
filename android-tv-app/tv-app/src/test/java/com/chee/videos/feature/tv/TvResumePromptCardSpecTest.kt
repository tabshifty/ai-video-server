package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvResumePromptCardSpecTest {
    @Test
    fun `long form and series players render resume prompt at bottom start`() {
        val longForm = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()

        listOf(longForm, series).forEach { source ->
            assertTrue(source.contains("TvResumePromptCard("))
            assertTrue(source.contains("Alignment.BottomStart"))
            assertTrue(source.contains("TvResumePromptTokens.HorizontalPaddingDp"))
            assertTrue(source.contains("TvResumePromptTokens.BottomPaddingDp"))
            assertTrue(source.contains("shouldTriggerResumePrompt("))
            assertTrue(source.contains("shouldShowResumePromptCard("))
            assertTrue(source.contains("LaunchedEffect") && source.contains("shouldTickResumePromptCountdown"))
            assertTrue(source.contains("isTrackSheetVisible = isTrackSheetVisible"))
            assertTrue(source.contains("TvSeriesCorePlaybackOverlay("))
            assertTrue(source.contains("resumePromptSlot = {"))
            assertTrue(source.contains("TvMedia3TrackPickerLayer("))
            assertTrue(source.contains("showBackConfirmPrompt"))
            assertTrue(source.contains("withFrameNanos"))
        }
        assertTrue(longForm.contains("resumedFromHistoryVideoId == detail.id"))
        assertTrue(series.contains("resumedFromHistoryVideoId == uiState.currentVideoId"))
    }

    @Test
    fun `resume prompt card uses tv focus shape and motion tokens`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvResumePromptCard.kt").readText()

        assertTrue(source.contains("LaunchedTvInitialFocus(visible)"))
        assertTrue(source.contains(".tryRequestFocus()"))
        assertTrue(source.contains("AppChrome.SurfaceShape"))
        assertTrue(source.contains("AppChrome.ChipShape"))
        assertTrue(source.contains("tvFocusableGlow("))
        assertTrue(source.contains("TvMotionTokens.DurationStandardMs"))
        assertTrue(source.contains("TvMotionTokens.EasingStandard"))
        assertFalse(source.contains("RoundedCornerShape("))
        assertFalse(source.contains("tween("))
        assertFalse(source.contains("LaunchedTvInitialFocus(visible, lastPositionMs)"))
    }

    @Test
    fun `resume prompt call sites avoid naked timing and position literals`() {
        val longForm = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()
        val card = Path.of("src/main/java/com/chee/videos/feature/tv/TvResumePromptCard.kt").readText()

        listOf(longForm, series, card).forEach { source ->
            assertFalse(source.contains("5_000L"))
            assertFalse(source.contains("5000L"))
            assertFalse(source.contains("10_000L"))
            assertFalse(source.contains("10000L"))
        }
        listOf(longForm, series).forEach { source ->
            assertFalse(source.contains("padding(start = 48.dp"))
            assertFalse(source.contains("bottom = 156.dp"))
        }
    }

    @Test
    fun `resume prompt does not expose focus freeze behavior`() {
        val card = Path.of("src/main/java/com/chee/videos/feature/tv/TvResumePromptCard.kt").readText()
        val longForm = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()

        listOf(card, longForm, series).forEach { source ->
            assertFalse(source.contains("onFocusChanged"))
            assertFalse(source.contains("isResumePromptFocused"))
            assertFalse(source.contains("freezeResumePrompt"))
        }
    }
}
