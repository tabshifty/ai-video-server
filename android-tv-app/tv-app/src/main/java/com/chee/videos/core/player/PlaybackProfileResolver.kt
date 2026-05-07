package com.chee.videos.core.player

import javax.inject.Inject
import javax.inject.Singleton

enum class PlaybackProfile(val wireValue: String) {
    PRIMARY("primary"),
    COMPAT("compat"),
}

internal fun resolvePreferredLongFormPlaybackProfile(
    deviceSupportsHevc: Boolean,
    isProbablyEmulator: Boolean,
): PlaybackProfile {
    return PlaybackProfile.COMPAT
}

@Singleton
class PlaybackProfileResolver @Inject constructor() {
    fun preferredLongFormProfile(): PlaybackProfile {
        return resolvePreferredLongFormPlaybackProfile(
            deviceSupportsHevc = false,
            isProbablyEmulator = false,
        )
    }
}
