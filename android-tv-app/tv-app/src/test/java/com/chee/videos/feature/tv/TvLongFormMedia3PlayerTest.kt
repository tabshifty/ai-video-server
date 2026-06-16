package com.chee.videos.feature.tv

import androidx.media3.common.Player
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormMedia3PlayerTest {
    @Test
    fun seriesMedia3RouteAutoStartsWhenNewSourceIsReady() {
        assertTrue(
            shouldAutoStartTvLongFormMedia3Playback(
                currentSourceUrl = "https://example.test/source?profile=dv_source",
                isMedia3Route = true,
                autoStartedSourceUrl = "",
            ),
        )
        assertFalse(
            shouldAutoStartTvLongFormMedia3Playback(
                currentSourceUrl = "https://example.test/source?profile=dv_source",
                isMedia3Route = true,
                autoStartedSourceUrl = "https://example.test/source?profile=dv_source",
            ),
        )
        assertFalse(
            shouldAutoStartTvLongFormMedia3Playback(
                currentSourceUrl = "",
                isMedia3Route = true,
                autoStartedSourceUrl = "",
            ),
        )
        assertFalse(
            shouldAutoStartTvLongFormMedia3Playback(
                currentSourceUrl = "https://example.test/source?profile=dv_source",
                isMedia3Route = false,
                autoStartedSourceUrl = "",
            ),
        )
    }

    @Test
    fun startupTimeoutOnlyTriggersWhenPreparedAndStuckNotPlaying() {
        assertTrue(
            shouldReportTvLongFormMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = true,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvLongFormMedia3StartupTimeout(
                preparedSourceKey = "",
                shouldPlay = true,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvLongFormMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = false,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvLongFormMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = true,
                playbackState = Player.STATE_READY,
                isPlaying = false,
            ),
        )
        assertFalse(
            shouldReportTvLongFormMedia3StartupTimeout(
                preparedSourceKey = "video-1|https://example.test/source?profile=dv_source",
                shouldPlay = true,
                playbackState = Player.STATE_BUFFERING,
                isPlaying = true,
            ),
        )
    }

    @Test
    fun startupTimeoutMessageUsesUnifiedLongFormWording() {
        assertEquals("长视频 ExoPlayer 播放链路启动超时，请重试", TvLongFormMedia3StartupTimeoutMessage)
    }

    @Test
    fun preparingSourceAlsoDependsOnCurrentPlayerInstance() {
        val sourceKey = "video-1|https://example.test/source.m3u8"

        assertFalse(
            shouldPrepareTvLongFormMedia3Source(
                preparedSourceKey = sourceKey,
                currentSourceKey = sourceKey,
                isPreparedPlayerCurrent = true,
            ),
        )
        assertTrue(
            shouldPrepareTvLongFormMedia3Source(
                preparedSourceKey = sourceKey,
                currentSourceKey = sourceKey,
                isPreparedPlayerCurrent = false,
            ),
        )
        assertTrue(
            shouldPrepareTvLongFormMedia3Source(
                preparedSourceKey = "video-1|https://example.test/old.m3u8",
                currentSourceKey = sourceKey,
                isPreparedPlayerCurrent = true,
            ),
        )
        assertFalse(
            shouldPrepareTvLongFormMedia3Source(
                preparedSourceKey = "",
                currentSourceKey = "",
                isPreparedPlayerCurrent = false,
            ),
        )
    }
}
