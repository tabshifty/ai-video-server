package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvPlaybackHistoryPolicyTest {
    @Test
    fun watchSnapshot_convertsPositionAndDetectsCompletionInsideFinalFivePercent() {
        val snapshot = tvPlaybackHistorySnapshot(positionMs = 114_000L, durationMs = 120_000L)

        assertEquals(114, snapshot.watchSeconds)
        assertTrue(snapshot.completed)
    }

    @Test
    fun watchSnapshot_doesNotCompleteBeforeFinalFivePercent() {
        val snapshot = tvPlaybackHistorySnapshot(positionMs = 113_999L, durationMs = 120_000L)

        assertEquals(113, snapshot.watchSeconds)
        assertFalse(snapshot.completed)
    }

    @Test
    fun watchSnapshot_capsCompletionWindowAtFiveMinutesForLongContent() {
        assertFalse(tvPlaybackHistorySnapshot(positionMs = 10_260_000L, durationMs = 10_800_000L).completed)
        assertTrue(tvPlaybackHistorySnapshot(positionMs = 10_500_000L, durationMs = 10_800_000L).completed)
    }

    @Test
    fun retryResumePrefersCurrentMedia3SnapshotWhenAvailable() {
        val resumePositionMs = resolveTvMedia3ResumePositionMs(
            historyPositionMs = 48_000L,
            currentSnapshotPositionMs = 91_000L,
            hasCurrentPlaybackSnapshot = true,
        )

        assertEquals(91_000L, resumePositionMs)
    }

    @Test
    fun retryResumeFallsBackToHistoryWhenNoCurrentSnapshotExists() {
        val resumePositionMs = resolveTvMedia3ResumePositionMs(
            historyPositionMs = 48_000L,
            currentSnapshotPositionMs = 91_000L,
            hasCurrentPlaybackSnapshot = false,
        )

        assertEquals(48_000L, resumePositionMs)
    }

    @Test
    fun shouldStartPeriodicHistoryReport_requiresPlayableStartedUnpausedVideo() {
        assertTrue(
            shouldStartPeriodicHistoryReport(
                videoId = "video-1",
                canPlay = true,
                hasStartedPlayback = true,
                isPausedByUser = false,
            ),
        )
        assertFalse(
            shouldStartPeriodicHistoryReport(
                videoId = "",
                canPlay = true,
                hasStartedPlayback = true,
                isPausedByUser = false,
            ),
        )
        assertFalse(
            shouldStartPeriodicHistoryReport(
                videoId = "video-1",
                canPlay = false,
                hasStartedPlayback = true,
                isPausedByUser = false,
            ),
        )
        assertFalse(
            shouldStartPeriodicHistoryReport(
                videoId = "video-1",
                canPlay = true,
                hasStartedPlayback = false,
                isPausedByUser = false,
            ),
        )
        assertFalse(
            shouldStartPeriodicHistoryReport(
                videoId = "video-1",
                canPlay = true,
                hasStartedPlayback = true,
                isPausedByUser = true,
            ),
        )
    }

    @Test
    fun shouldReportHistory_ignoresBlankVideoAndShortTrialPlayback() {
        assertFalse(shouldReportHistory(videoId = "", watchSeconds = 12))
        assertFalse(shouldReportHistory(videoId = "video-1", watchSeconds = 0))
        assertFalse(shouldReportHistory(videoId = "video-1", watchSeconds = 29))
        assertTrue(shouldReportHistory(videoId = "video-1", watchSeconds = 30))
    }

    @Test
    fun completedPlaybackIsReportedEvenWhenContentIsShorterThanResumeThreshold() {
        assertFalse(shouldSubmitTvPlaybackHistory(videoId = "video-1", watchSeconds = 29, completed = false))
        assertTrue(shouldSubmitTvPlaybackHistory(videoId = "video-1", watchSeconds = 29, completed = true))
        assertFalse(shouldSubmitTvPlaybackHistory(videoId = "", watchSeconds = 29, completed = true))
    }

    @Test
    fun periodicHistoryReportInterval_isFifteenSeconds() {
        assertEquals(15_000L, TvPeriodicHistoryReportIntervalMillis)
    }

    @Test
    fun continuePlaybackIsOfferedOnlyInsideEffectiveResumeWindow() {
        assertFalse(shouldOfferTvLongFormContinuePlayback(watchSeconds = 29, durationSeconds = 3_600))
        assertTrue(shouldOfferTvLongFormContinuePlayback(watchSeconds = 30, durationSeconds = 3_600))
        assertFalse(shouldOfferTvLongFormContinuePlayback(watchSeconds = 3_420, durationSeconds = 3_600))
    }
}
