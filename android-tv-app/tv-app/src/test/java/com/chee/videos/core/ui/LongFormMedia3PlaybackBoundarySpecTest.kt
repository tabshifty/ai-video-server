package com.chee.videos.core.ui

import java.nio.file.Path
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class LongFormMedia3PlaybackBoundarySpecTest {

    private val longFormScreen: String by lazy {
        Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").toFile().readText()
    }

    private val seriesScreen: String by lazy {
        Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()
    }

    private val subtitleSupport: String by lazy {
        Path.of("src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt").toFile().readText()
    }

    @Test
    fun tvLongFormScreensDoNotKeepVlcPlayingGateOrSubtitleSlaveState() {
        listOf(longFormScreen, seriesScreen).forEach { source ->
            assertFalse(source.contains("isVlcPlaying"))
            assertFalse(source.contains("appliedSubtitleSlaveUrl"))
            assertFalse(source.contains("addSlave("))
            assertFalse(source.contains("IMedia.Slave"))
        }
    }

    @Test
    fun tvLongFormScreensDoNotUseAccessTokenQueryForPlaybackAssets() {
        listOf(longFormScreen, seriesScreen).forEach { source ->
            assertFalse(source.contains("appendAccessTokenQuery"))
            assertFalse(source.contains("access_token"))
        }
    }

    @Test
    fun tvLongFormScreensUseMedia3SnapshotForHistory() {
        assertTrue(longFormScreen.contains("reportTvLongFormMedia3History("))
        assertTrue(seriesScreen.contains("reportTvSeriesMedia3History("))
        assertFalse(longFormScreen.contains("reportTvLongFormHistory("))
        assertFalse(seriesScreen.contains("reportTvSeriesHistory("))
    }

    @Test
    fun tvLongFormScreensReportMedia3HistoryOnLifecyclePause() {
        assertTrue(longFormScreen.contains("onLifecyclePauseSnapshot = { snapshot ->"))
        assertTrue(seriesScreen.contains("onLifecyclePauseSnapshot = { snapshot ->"))
        assertTrue(
            longFormScreen.contains("reportTvLongFormMedia3History(viewModel, detail.id, snapshot)"),
        )
        assertTrue(
            seriesScreen.contains("reportTvSeriesMedia3History(viewModel, uiState.currentVideoId, snapshot)"),
        )
    }

    @Test
    fun typeOnlyFallbackStillExistsInTrackAndSubtitleResolvers() {
        val trackSelectionSource = Path.of(
            "src/main/java/com/chee/videos/core/player/TvLongFormTrackSelection.kt",
        ).toFile().readText()
        assertTrue(
            "resolveLongFormTrackByLanguage 必须在 language.isBlank 时检查 type 并按 type 匹配",
            trackSelectionSource.contains("type-only fallback"),
        )
        assertTrue(
            "resolveSelectedSubtitleTrackByPreference 必须在 language.isBlank 时按 type 匹配",
            subtitleSupport.contains("type-only fallback"),
        )
    }
}
