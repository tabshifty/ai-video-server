package com.chee.videos.feature.detail

internal data class LongFormPlaybackSession(
    val hasStartedPlayback: Boolean = false,
    val isPausedByUser: Boolean = false,
) {
    fun requestPlay(canPlay: Boolean): LongFormPlaybackSession {
        if (!canPlay) {
            return this
        }
        return copy(
            hasStartedPlayback = true,
            isPausedByUser = false,
        )
    }

    fun togglePlayPause(canPlay: Boolean): LongFormPlaybackSession {
        if (!hasStartedPlayback) {
            return requestPlay(canPlay = canPlay)
        }
        if (!canPlay) {
            return copy(isPausedByUser = true)
        }
        return copy(isPausedByUser = !isPausedByUser)
    }

    fun shouldShowPlayer(canPlay: Boolean): Boolean {
        return hasStartedPlayback && canPlay
    }

    fun shouldResumeOnLifecycle(): Boolean {
        return hasStartedPlayback && !isPausedByUser
    }
}
