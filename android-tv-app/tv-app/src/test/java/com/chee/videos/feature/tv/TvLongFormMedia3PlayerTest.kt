package com.chee.videos.feature.tv

import androidx.media3.common.Player
import androidx.media3.common.PlaybackException
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormMedia3PlayerTest {
    @Test
    fun delayedMedia3EventsOnlyBelongToThePreparedIdentity() {
        val prepared = TvLongFormMedia3EventIdentity(mediaId = "video-2", retryKey = 0)

        assertTrue(isCurrentTvLongFormMedia3Event(prepared, prepared))
        assertFalse(
            isCurrentTvLongFormMedia3Event(
                eventIdentity = TvLongFormMedia3EventIdentity(mediaId = "video-1", retryKey = 0),
                preparedIdentity = prepared,
            ),
        )

        val source = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt",
        ).toFile().readText()
        val errorHandler = source.substringAfter("override fun onPlayerError(").substringBefore("player.addListener")
        assertTrue(errorHandler.contains("!isCurrentTvLongFormMedia3Event(eventIdentity, preparedIdentity)"))
        assertTrue(source.contains("onEnded: (TvLongFormMedia3EventIdentity) -> Unit"))
        assertTrue(source.contains("latestOnEnded(eventIdentity)"))
    }

    @Test
    fun mediaSessionUserRequestsAreReportedAsExplicitPlaybackIntent() {
        val source = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt",
        ).toFile().readText()

        assertTrue(source.contains("onPlaybackIntentChanged: (Boolean) -> Unit"))
        assertTrue(source.contains("Player.PLAY_WHEN_READY_CHANGE_REASON_USER_REQUEST"))
        assertTrue(source.contains("latestOnPlaybackIntentChanged(playWhenReady)"))
        assertTrue(source.contains("pendingInternalPlaybackIntent == playWhenReady"))
        assertTrue(source.contains("applyInternalPlaybackIntent(shouldPlay)"))
    }

    @Test
    fun playbackHostsFilterEndedEventsAgainstTheirCurrentIdentity() {
        val single = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt",
        ).toFile().readText()
        val series = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt",
        ).toFile().readText()

        listOf(single, series).forEach { source ->
            val endedHandler = source.substringAfter("onEnded = { eventIdentity ->").substringBefore("onSnapshotChanged")
            assertTrue(endedHandler.contains("eventIdentity == currentPlaybackIdentity"))
        }
    }

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
