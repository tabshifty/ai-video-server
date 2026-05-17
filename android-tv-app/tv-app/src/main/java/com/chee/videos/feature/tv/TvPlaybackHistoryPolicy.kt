package com.chee.videos.feature.tv

const val TvPeriodicHistoryReportIntervalMillis: Long = 15_000L

data class TvPlaybackHistorySnapshot(
    val watchSeconds: Int,
    val completed: Boolean,
)

fun tvPlaybackHistorySnapshot(
    positionMs: Long,
    durationMs: Long,
): TvPlaybackHistorySnapshot {
    val safePositionMs = positionMs.coerceAtLeast(0L)
    val safeDurationMs = durationMs.coerceAtLeast(0L)
    return TvPlaybackHistorySnapshot(
        watchSeconds = (safePositionMs / 1000L).toInt(),
        completed = safeDurationMs > 0L && safePositionMs >= safeDurationMs - 3_000L,
    )
}

fun shouldStartPeriodicHistoryReport(
    videoId: String,
    canPlay: Boolean,
    hasStartedPlayback: Boolean,
    isPausedByUser: Boolean,
): Boolean =
    videoId.isNotBlank() && canPlay && hasStartedPlayback && !isPausedByUser

fun shouldReportHistory(videoId: String, watchSeconds: Int): Boolean =
    videoId.isNotBlank() && watchSeconds > 0
