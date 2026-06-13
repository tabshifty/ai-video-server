package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvPlaybackHistoryPolicyTest {
    @Test
    fun watchSnapshot_convertsPositionAndDetectsCompletionNearEnd() {
        val snapshot = tvPlaybackHistorySnapshot(positionMs = 117_900L, durationMs = 120_000L)

        assertEquals(117, snapshot.watchSeconds)
        assertTrue(snapshot.completed)
    }

    @Test
    fun watchSnapshot_doesNotCompleteWhenMoreThanThreeSecondsRemain() {
        val snapshot = tvPlaybackHistorySnapshot(positionMs = 116_900L, durationMs = 120_000L)

        assertEquals(116, snapshot.watchSeconds)
        assertFalse(snapshot.completed)
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
    fun shouldReportHistory_ignoresBlankVideoAndZeroProgress() {
        assertFalse(shouldReportHistory(videoId = "", watchSeconds = 12))
        assertFalse(shouldReportHistory(videoId = "video-1", watchSeconds = 0))
        assertTrue(shouldReportHistory(videoId = "video-1", watchSeconds = 1))
    }

    @Test
    fun periodicHistoryReportInterval_isFifteenSeconds() {
        assertEquals(15_000L, TvPeriodicHistoryReportIntervalMillis)
    }
}
