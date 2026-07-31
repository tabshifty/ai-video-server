package com.chee.videos.feature.tv

import java.nio.file.Path
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesMixedPlaybackControlsSpecTest {
    @Test
    fun seriesExoPlayerRouteUsesSharedOttChromeWithTrackAndEpisodeCapabilities() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()

        assertTrue(source.contains("TvLongFormPlaybackChrome("))
        assertTrue(source.contains("subtitleTracks = currentEpisode.subtitleTracks"))
        assertTrue(source.contains("audioTracks = media3AudioTracks"))
        assertTrue(source.contains("seasons = playbackSeasons"))
        assertTrue(source.contains("onPlayNextEpisode = if (hasNextEpisode)"))
        assertFalse(source.contains("TvSeriesCorePlaybackOverlay("))
        assertFalse(source.contains("TvMedia3TrackPickerLayer("))
    }

    @Test
    fun rightPanelsMoveFocusToTheCurrentOptionAndContainHorizontalNavigation() {
        val chrome = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlaybackChrome.kt").toFile().readText()

        assertTrue(chrome.contains("initialFocusRequester.tryRequestFocus()"))
        assertTrue(chrome.contains("Modifier.focusRequester(initialFocusRequester)"))
        assertTrue(chrome.contains("seasonFocusRequesters[seasons[targetIndex].number]?.tryRequestFocus()"))
        assertTrue(chrome.contains("AndroidKeyEvent.KEYCODE_DPAD_LEFT,"))
        assertTrue(chrome.contains("AndroidKeyEvent.KEYCODE_DPAD_RIGHT,"))
    }

    @Test
    fun media3FinalFailureUsesRetryableErrorLayerWithoutLegacyFallback() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()

        assertTrue(source.contains("resolveTvLongFormErrorAction("))
        assertTrue(source.contains("TvLongFormErrorAction.ShowFinalError"))
        assertTrue(source.contains("requestPlaybackRetry()"))
        assertTrue(source.contains("secondaryActionLabel = \"返回详情\""))
        assertFalse(source.contains("强行播放"))
        assertFalse(source.contains("LibVLC"))
    }

    @Test
    fun seriesScreenUsesSingleLongFormMedia3ComponentAndSnapshotPath() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()

        assertTrue(source.contains("TvLongFormMedia3Player("))
        assertTrue(source.contains("reportTvSeriesMedia3History("))
        assertFalse(source.contains("TvDolbyVisionMedia3Player("))
        assertFalse(source.contains("shouldUseMedia3SnapshotForTvSeriesHistory"))
        assertFalse(source.contains("isVlcRoute"))
        assertFalse(source.contains("LongFormVideoPlayer("))
        assertFalse(source.contains("org.videolan.libvlc"))
    }
}
