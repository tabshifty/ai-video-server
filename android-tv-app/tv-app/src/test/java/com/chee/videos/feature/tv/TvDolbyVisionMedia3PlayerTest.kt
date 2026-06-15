package com.chee.videos.feature.tv

import androidx.media3.common.Player
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvDolbyVisionMedia3PlayerTest {
    @Test
    fun seriesMedia3Route_autoStartsWhenNewSourceIsReady() {
        assertTrue(
            shouldAutoStartTvDolbyVisionMedia3Playback(
                currentSourceUrl = "https://example.test/source?profile=dv_source",
                isMedia3Route = true,
                autoStartedSourceUrl = "",
            ),
        )
        assertFalse(
            shouldAutoStartTvDolbyVisionMedia3Playback(
                currentSourceUrl = "https://example.test/source?profile=dv_source",
                isMedia3Route = true,
                autoStartedSourceUrl = "https://example.test/source?profile=dv_source",
            ),
        )
        assertFalse(
            shouldAutoStartTvDolbyVisionMedia3Playback(
                currentSourceUrl = "",
                isMedia3Route = true,
                autoStartedSourceUrl = "",
            ),
        )
        assertFalse(
            shouldAutoStartTvDolbyVisionMedia3Playback(
                currentSourceUrl = "https://example.test/source?profile=dv_source",
                isMedia3Route = false,
                autoStartedSourceUrl = "",
            ),
        )
    }

    @Test
    fun startupTimeout_onlyTriggersWhenPreparedAndStuckNotPlaying() {
        assertTrue(
            shouldReportTvDolbyVisionMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = true,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvDolbyVisionMedia3StartupTimeout(
                preparedSourceKey = "",
                shouldPlay = true,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvDolbyVisionMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = false,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvDolbyVisionMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = true,
                playbackState = Player.STATE_READY,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvDolbyVisionMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = true,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = true,
            ),
        )
    }
}
