package com.chee.videos.feature.tv

const val TvPeriodicHistoryReportIntervalMillis: Long = 15_000L
const val TvMinimumResumeWatchSeconds: Int = 30
private const val TvCompletionPercent: Long = 95L
private const val TvMaximumCompletionWindowMs: Long = 5 * 60 * 1_000L

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
    val completionWindowMs = minOf(
        safeDurationMs * (100L - TvCompletionPercent) / 100L,
        TvMaximumCompletionWindowMs,
    )
    return TvPlaybackHistorySnapshot(
        watchSeconds = (safePositionMs / 1000L).toInt(),
        completed = safeDurationMs > 0L &&
            completionWindowMs > 0L &&
            safePositionMs >= safeDurationMs - completionWindowMs,
    )
}

internal fun resolveTvMedia3ResumePositionMs(
    historyPositionMs: Long,
    currentSnapshotPositionMs: Long,
    hasCurrentPlaybackSnapshot: Boolean,
): Long {
    val safeHistoryPositionMs = historyPositionMs.coerceAtLeast(0L)
    val safeCurrentSnapshotPositionMs = currentSnapshotPositionMs.coerceAtLeast(0L)
    if (hasCurrentPlaybackSnapshot && safeCurrentSnapshotPositionMs > 0L) {
        return safeCurrentSnapshotPositionMs
    }
    return safeHistoryPositionMs
}

fun shouldStartPeriodicHistoryReport(
    videoId: String,
    canPlay: Boolean,
    hasStartedPlayback: Boolean,
    isPausedByUser: Boolean,
): Boolean =
    videoId.isNotBlank() && canPlay && hasStartedPlayback && !isPausedByUser

fun shouldReportHistory(videoId: String, watchSeconds: Int): Boolean =
    shouldSubmitTvPlaybackHistory(
        videoId = videoId,
        watchSeconds = watchSeconds,
        completed = false,
    )

fun shouldSubmitTvPlaybackHistory(
    videoId: String,
    watchSeconds: Int,
    completed: Boolean,
): Boolean =
    videoId.isNotBlank() && (completed || watchSeconds >= TvMinimumResumeWatchSeconds)

fun shouldOfferTvLongFormContinuePlayback(
    watchSeconds: Int,
    durationSeconds: Int,
): Boolean {
    if (watchSeconds < TvMinimumResumeWatchSeconds || durationSeconds <= 0) {
        return false
    }
    return !tvPlaybackHistorySnapshot(
        positionMs = watchSeconds.toLong() * 1_000L,
        durationMs = durationSeconds.toLong() * 1_000L,
    ).completed
}
