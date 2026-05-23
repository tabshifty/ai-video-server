package com.chee.videos.feature.tv

import androidx.compose.ui.unit.dp
import kotlin.math.ceil

object TvResumePromptTokens {
    const val CountdownDurationMs: Long = 5_000L
    const val MinResumeMs: Long = 10_000L
    val CardMinWidthDp = 320.dp
    val CardMaxWidthDp = 380.dp
    val HorizontalPaddingDp = 48.dp
    val BottomPaddingDp = 156.dp
}

data class ResumePromptGuardInput(
    val hasResumeSeekTriggered: Boolean,
    val promptPermanentlyDismissed: Boolean,
    val isPlayerError: Boolean,
    val isBackConfirmVisible: Boolean,
    val isEpisodeSelectorVisible: Boolean,
    val isEndOverlayVisible: Boolean,
    val isAutoplayPromptVisible: Boolean,
    val isPausedByUser: Boolean,
    val remainingMs: Long,
)

fun shouldShowResumePromptCard(input: ResumePromptGuardInput): Boolean =
    shouldTickResumePromptCountdown(input) && input.remainingMs > 0L

fun shouldTickResumePromptCountdown(input: ResumePromptGuardInput): Boolean =
    input.hasResumeSeekTriggered &&
        !input.promptPermanentlyDismissed &&
        !input.isPlayerError &&
        !input.isBackConfirmVisible &&
        !input.isEpisodeSelectorVisible &&
        !input.isEndOverlayVisible &&
        !input.isAutoplayPromptVisible &&
        !input.isPausedByUser

fun resumePromptCountdownTickRemaining(remainingMs: Long): Int {
    if (remainingMs <= 0L) return 0
    val seconds = ceil(remainingMs / 1000.0).toInt()
    return seconds.coerceIn(0, 5)
}

fun shouldTriggerResumePrompt(initialResumePositionMs: Long): Boolean =
    initialResumePositionMs >= TvResumePromptTokens.MinResumeMs

fun formatResumePromptTimestamp(positionMs: Long): String {
    val totalSeconds = (positionMs.coerceAtLeast(0L) / 1000L).toInt()
    val hours = totalSeconds / 3600
    val minutes = (totalSeconds % 3600) / 60
    val seconds = totalSeconds % 60
    return if (hours > 0) {
        "%d:%02d:%02d".format(hours, minutes, seconds)
    } else {
        "%d:%02d".format(minutes, seconds)
    }
}
