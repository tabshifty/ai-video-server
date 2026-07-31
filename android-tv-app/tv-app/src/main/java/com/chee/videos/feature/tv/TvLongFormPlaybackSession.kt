package com.chee.videos.feature.tv

internal const val TvLongFormControlsAutoHideMillis: Long = 4_000L
internal const val TvLongFormStartupFeedbackDelayMillis: Long = 300L
internal const val TvLongFormBufferingFeedbackDelayMillis: Long = 500L
internal const val TvLongFormLoadingTimeoutMillis: Long = 15_000L

internal data class TvLongFormPlaybackCapabilities(
    val subtitleTrackCount: Int = 0,
    val audioTrackCount: Int = 0,
    val hasEpisodePicker: Boolean = false,
    val hasNextEpisode: Boolean = false,
)

internal enum class TvLongFormControlAction {
    Rewind,
    PlayPause,
    Forward,
    NextEpisode,
    Subtitles,
    AudioTracks,
    Episodes,
}

internal fun buildTvLongFormVisibleActions(
    capabilities: TvLongFormPlaybackCapabilities,
): List<TvLongFormControlAction> = buildList {
    add(TvLongFormControlAction.Rewind)
    add(TvLongFormControlAction.PlayPause)
    add(TvLongFormControlAction.Forward)
    if (capabilities.hasNextEpisode) add(TvLongFormControlAction.NextEpisode)
    if (capabilities.subtitleTrackCount > 0) add(TvLongFormControlAction.Subtitles)
    if (capabilities.audioTrackCount > 1) add(TvLongFormControlAction.AudioTracks)
    if (capabilities.hasEpisodePicker) add(TvLongFormControlAction.Episodes)
}

internal enum class TvLongFormInteractionMode {
    Hidden,
    Controls,
    PrecisionSeek,
    SubtitlePanel,
    AudioPanel,
    EpisodePanel,
    Completed,
    Error,
}

internal enum class TvLongFormBackAction {
    ClosePanel,
    CancelPrecisionSeek,
    HideControls,
    ExitPlayback,
}

internal fun resolveTvLongFormBackAction(
    mode: TvLongFormInteractionMode,
): TvLongFormBackAction = when (mode) {
    TvLongFormInteractionMode.SubtitlePanel,
    TvLongFormInteractionMode.AudioPanel,
    TvLongFormInteractionMode.EpisodePanel,
    -> TvLongFormBackAction.ClosePanel

    TvLongFormInteractionMode.PrecisionSeek -> TvLongFormBackAction.CancelPrecisionSeek
    TvLongFormInteractionMode.Controls -> TvLongFormBackAction.HideControls
    TvLongFormInteractionMode.Hidden,
    TvLongFormInteractionMode.Completed,
    TvLongFormInteractionMode.Error,
    -> TvLongFormBackAction.ExitPlayback
}

internal fun shouldAutoHideTvLongFormControls(
    isPlaying: Boolean,
    mode: TvLongFormInteractionMode,
    idleMs: Long,
): Boolean =
    isPlaying &&
        mode == TvLongFormInteractionMode.Controls &&
        idleMs >= TvLongFormControlsAutoHideMillis

internal enum class TvLongFormPlaybackStatus {
    Idle,
    Preparing,
    Playing,
    Paused,
    Buffering,
    Ended,
    Error,
}

internal fun shouldKeepTvLongFormScreenOn(
    playIntent: Boolean,
    status: TvLongFormPlaybackStatus,
): Boolean = playIntent && status in setOf(
    TvLongFormPlaybackStatus.Preparing,
    TvLongFormPlaybackStatus.Playing,
    TvLongFormPlaybackStatus.Buffering,
)

internal enum class TvLongFormLoadingFeedback {
    Hidden,
    Startup,
    Buffering,
    TimedOut,
}

internal fun resolveTvLongFormLoadingFeedback(
    hasRenderedFirstFrame: Boolean,
    waitingMs: Long,
): TvLongFormLoadingFeedback {
    val safeWaitingMs = waitingMs.coerceAtLeast(0L)
    if (safeWaitingMs >= TvLongFormLoadingTimeoutMillis) {
        return TvLongFormLoadingFeedback.TimedOut
    }
    val threshold = if (hasRenderedFirstFrame) {
        TvLongFormBufferingFeedbackDelayMillis
    } else {
        TvLongFormStartupFeedbackDelayMillis
    }
    if (safeWaitingMs < threshold) {
        return TvLongFormLoadingFeedback.Hidden
    }
    return if (hasRenderedFirstFrame) {
        TvLongFormLoadingFeedback.Buffering
    } else {
        TvLongFormLoadingFeedback.Startup
    }
}

internal fun tvLongFormAutomaticRetryDelayMs(attempt: Int): Long? = when (attempt) {
    1 -> 1_000L
    2 -> 3_000L
    else -> null
}

internal sealed interface TvLongFormErrorAction {
    data class ScheduleRetry(
        val attempt: Int,
        val delayMs: Long,
    ) : TvLongFormErrorAction

    data object ShowFinalError : TvLongFormErrorAction
}

internal fun resolveTvLongFormErrorAction(
    retryable: Boolean,
    completedAutomaticRetries: Int,
): TvLongFormErrorAction {
    if (!retryable) {
        return TvLongFormErrorAction.ShowFinalError
    }
    val nextAttempt = completedAutomaticRetries.coerceAtLeast(0) + 1
    val delayMs = tvLongFormAutomaticRetryDelayMs(nextAttempt)
        ?: return TvLongFormErrorAction.ShowFinalError
    return TvLongFormErrorAction.ScheduleRetry(nextAttempt, delayMs)
}

internal enum class TvLongFormFocusMove {
    Left,
    Right,
}

internal fun resolveTvLongFormHorizontalFocus(
    current: TvLongFormControlAction,
    move: TvLongFormFocusMove,
    actions: List<TvLongFormControlAction>,
): TvLongFormControlAction {
    if (actions.isEmpty()) {
        return TvLongFormControlAction.PlayPause
    }
    val index = actions.indexOf(current)
    if (index < 0) {
        return TvLongFormControlAction.PlayPause.takeIf(actions::contains) ?: actions.first()
    }
    val nextIndex = when (move) {
        TvLongFormFocusMove.Left -> (index - 1).coerceAtLeast(0)
        TvLongFormFocusMove.Right -> (index + 1).coerceAtMost(actions.lastIndex)
    }
    return actions[nextIndex]
}
