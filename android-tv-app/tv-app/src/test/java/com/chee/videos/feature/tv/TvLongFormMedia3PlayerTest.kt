package com.chee.videos.feature.tv

import androidx.media3.common.Player
import androidx.media3.common.PlaybackException
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormMedia3PlayerTest {
    @Test
    fun mediaSessionSeekCommandsUseTheConfiguredTvStep() {
        val source = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt",
        ).toFile().readText()

        assertTrue(source.contains(".setSeekBackIncrementMs(seekIncrementMs)"))
        assertTrue(source.contains(".setSeekForwardIncrementMs(seekIncrementMs)"))
    }

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
        assertEquals("视频加载超时，请重试", TvLongFormMedia3StartupTimeoutMessage)
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

    @Test
    fun onlyTemporaryMedia3FailuresAreRetried() {
        assertTrue(isTvLongFormMedia3ErrorCodeRetryable(PlaybackException.ERROR_CODE_TIMEOUT))
        assertTrue(isTvLongFormMedia3ErrorCodeRetryable(PlaybackException.ERROR_CODE_IO_UNSPECIFIED))
        assertTrue(isTvLongFormMedia3ErrorCodeRetryable(PlaybackException.ERROR_CODE_IO_NETWORK_CONNECTION_FAILED))
        assertTrue(isTvLongFormMedia3ErrorCodeRetryable(PlaybackException.ERROR_CODE_IO_NETWORK_CONNECTION_TIMEOUT))
        assertFalse(isTvLongFormMedia3ErrorCodeRetryable(PlaybackException.ERROR_CODE_DECODING_FAILED))
        assertFalse(isTvLongFormMedia3ErrorCodeRetryable(PlaybackException.ERROR_CODE_IO_FILE_NOT_FOUND))
        assertFalse(isTvLongFormMedia3ErrorCodeRetryable(PlaybackException.ERROR_CODE_IO_NO_PERMISSION))
    }

    @Test
    fun httpRetryPolicyAcceptsThrottlingTimeoutAndServerFailuresOnly() {
        assertTrue(isTvLongFormHttpStatusRetryable(408))
        assertTrue(isTvLongFormHttpStatusRetryable(429))
        assertTrue(isTvLongFormHttpStatusRetryable(503))
        assertFalse(isTvLongFormHttpStatusRetryable(401))
        assertFalse(isTvLongFormHttpStatusRetryable(404))
    }
}
