package com.chee.videos.feature.tv

import android.view.KeyEvent as AndroidKeyEvent
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormPlaybackSessionTest {
    @Test
    fun dedicatedMediaKeysResolveToExplicitPlaybackCommands() {
        assertEquals(
            TvLongFormMediaPlaybackCommand.Play,
            resolveTvLongFormMediaPlaybackCommand(AndroidKeyEvent.KEYCODE_MEDIA_PLAY),
        )
        assertEquals(
            TvLongFormMediaPlaybackCommand.Pause,
            resolveTvLongFormMediaPlaybackCommand(AndroidKeyEvent.KEYCODE_MEDIA_PAUSE),
        )
        assertEquals(
            TvLongFormMediaPlaybackCommand.Toggle,
            resolveTvLongFormMediaPlaybackCommand(AndroidKeyEvent.KEYCODE_MEDIA_PLAY_PAUSE),
        )
        assertNull(resolveTvLongFormMediaPlaybackCommand(AndroidKeyEvent.KEYCODE_DPAD_CENTER))
    }

    @Test
    fun capabilitiesExposeOnlyRealControls() {
        val actions = buildTvLongFormVisibleActions(
            TvLongFormPlaybackCapabilities(
                subtitleTrackCount = 0,
                audioTrackCount = 1,
                hasEpisodePicker = false,
                hasNextEpisode = false,
            ),
        )

        assertEquals(
            listOf(
                TvLongFormControlAction.Rewind,
                TvLongFormControlAction.PlayPause,
                TvLongFormControlAction.Forward,
            ),
            actions,
        )
    }

    @Test
    fun seriesCapabilitiesAddNextTracksAndEpisodePanel() {
        val actions = buildTvLongFormVisibleActions(
            TvLongFormPlaybackCapabilities(
                subtitleTrackCount = 2,
                audioTrackCount = 3,
                hasEpisodePicker = true,
                hasNextEpisode = true,
            ),
        )

        assertEquals(
            listOf(
                TvLongFormControlAction.Rewind,
                TvLongFormControlAction.PlayPause,
                TvLongFormControlAction.Forward,
                TvLongFormControlAction.NextEpisode,
                TvLongFormControlAction.Subtitles,
                TvLongFormControlAction.AudioTracks,
                TvLongFormControlAction.Episodes,
            ),
            actions,
        )
    }

    @Test
    fun backActionClosesDeepestLayerBeforeExit() {
        assertEquals(
            TvLongFormBackAction.ClosePanel,
            resolveTvLongFormBackAction(TvLongFormInteractionMode.EpisodePanel),
        )
        assertEquals(
            TvLongFormBackAction.CancelPrecisionSeek,
            resolveTvLongFormBackAction(TvLongFormInteractionMode.PrecisionSeek),
        )
        assertEquals(
            TvLongFormBackAction.HideControls,
            resolveTvLongFormBackAction(TvLongFormInteractionMode.Controls),
        )
        assertEquals(
            TvLongFormBackAction.ExitPlayback,
            resolveTvLongFormBackAction(TvLongFormInteractionMode.Hidden),
        )
    }

    @Test
    fun controlsAutoHideOnlyDuringPlainPlayback() {
        assertTrue(
            shouldAutoHideTvLongFormControls(
                isPlaying = true,
                mode = TvLongFormInteractionMode.Controls,
                idleMs = 4_000L,
            ),
        )
        assertFalse(
            shouldAutoHideTvLongFormControls(
                isPlaying = false,
                mode = TvLongFormInteractionMode.Controls,
                idleMs = 30_000L,
            ),
        )
        assertFalse(
            shouldAutoHideTvLongFormControls(
                isPlaying = true,
                mode = TvLongFormInteractionMode.PrecisionSeek,
                idleMs = 30_000L,
            ),
        )
    }

    @Test
    fun keepScreenOnFollowsPlayIntentAndActivePlayerState() {
        assertTrue(shouldKeepTvLongFormScreenOn(true, TvLongFormPlaybackStatus.Playing))
        assertTrue(shouldKeepTvLongFormScreenOn(true, TvLongFormPlaybackStatus.Buffering))
        assertTrue(shouldKeepTvLongFormScreenOn(true, TvLongFormPlaybackStatus.Preparing))
        assertFalse(shouldKeepTvLongFormScreenOn(false, TvLongFormPlaybackStatus.Paused))
        assertFalse(shouldKeepTvLongFormScreenOn(true, TvLongFormPlaybackStatus.Ended))
        assertFalse(shouldKeepTvLongFormScreenOn(true, TvLongFormPlaybackStatus.Error))
    }

    @Test
    fun delayedLoadingFeedbackAvoidsSpinnerFlashAndTimesOut() {
        assertEquals(
            TvLongFormLoadingFeedback.Hidden,
            resolveTvLongFormLoadingFeedback(hasRenderedFirstFrame = false, waitingMs = 299L),
        )
        assertEquals(
            TvLongFormLoadingFeedback.Startup,
            resolveTvLongFormLoadingFeedback(hasRenderedFirstFrame = false, waitingMs = 300L),
        )
        assertEquals(
            TvLongFormLoadingFeedback.Hidden,
            resolveTvLongFormLoadingFeedback(hasRenderedFirstFrame = true, waitingMs = 499L),
        )
        assertEquals(
            TvLongFormLoadingFeedback.Buffering,
            resolveTvLongFormLoadingFeedback(hasRenderedFirstFrame = true, waitingMs = 500L),
        )
        assertEquals(
            TvLongFormLoadingFeedback.TimedOut,
            resolveTvLongFormLoadingFeedback(hasRenderedFirstFrame = true, waitingMs = 15_000L),
        )
    }

    @Test
    fun automaticRetryUsesTwoBoundedDelays() {
        assertEquals(1_000L, tvLongFormAutomaticRetryDelayMs(attempt = 1))
        assertEquals(3_000L, tvLongFormAutomaticRetryDelayMs(attempt = 2))
        assertNull(tvLongFormAutomaticRetryDelayMs(attempt = 3))
    }

    @Test
    fun temporaryFailureSchedulesAtMostTwoAutomaticRetries() {
        assertEquals(
            TvLongFormErrorAction.ScheduleRetry(attempt = 1, delayMs = 1_000L),
            resolveTvLongFormErrorAction(retryable = true, completedAutomaticRetries = 0),
        )
        assertEquals(
            TvLongFormErrorAction.ScheduleRetry(attempt = 2, delayMs = 3_000L),
            resolveTvLongFormErrorAction(retryable = true, completedAutomaticRetries = 1),
        )
        assertEquals(
            TvLongFormErrorAction.ShowFinalError,
            resolveTvLongFormErrorAction(retryable = true, completedAutomaticRetries = 2),
        )
    }

    @Test
    fun deterministicFailureSkipsAutomaticRetry() {
        assertEquals(
            TvLongFormErrorAction.ShowFinalError,
            resolveTvLongFormErrorAction(retryable = false, completedAutomaticRetries = 0),
        )
    }

    @Test
    fun horizontalFocusStopsAtEdges() {
        val actions = listOf(
            TvLongFormControlAction.Rewind,
            TvLongFormControlAction.PlayPause,
            TvLongFormControlAction.Forward,
        )

        assertEquals(
            TvLongFormControlAction.Rewind,
            resolveTvLongFormHorizontalFocus(
                current = TvLongFormControlAction.Rewind,
                move = TvLongFormFocusMove.Left,
                actions = actions,
            ),
        )
        assertEquals(
            TvLongFormControlAction.Forward,
            resolveTvLongFormHorizontalFocus(
                current = TvLongFormControlAction.Forward,
                move = TvLongFormFocusMove.Right,
                actions = actions,
            ),
        )
        assertEquals(
            TvLongFormControlAction.PlayPause,
            resolveTvLongFormHorizontalFocus(
                current = TvLongFormControlAction.Rewind,
                move = TvLongFormFocusMove.Right,
                actions = actions,
            ),
        )
    }

    @Test
    fun chromeRestoresFocusAfterBlockingOverlayDisappears() {
        val source = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvLongFormPlaybackChrome.kt",
        ).toFile().readText()

        assertTrue(source.contains("LaunchedEffect(mode, blockingUiVisible)"))
        assertTrue(source.contains("if (blockingUiVisible) return@LaunchedEffect"))
        assertTrue(source.contains("if (blockingUiVisible && mode != TvLongFormInteractionMode.Hidden)"))
        assertTrue(source.contains("updateMode(TvLongFormInteractionMode.Hidden)"))
    }
}
