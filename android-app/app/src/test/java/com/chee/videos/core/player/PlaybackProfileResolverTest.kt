package com.chee.videos.core.player

import org.junit.Assert.assertEquals
import org.junit.Test

class PlaybackProfileResolverTest {
    @Test
    fun resolvePreferredLongFormPlaybackProfile_usesCompatWhenHevcSupported() {
        val profile = resolvePreferredLongFormPlaybackProfile(
            deviceSupportsHevc = true,
            isProbablyEmulator = false,
        )

        assertEquals(PlaybackProfile.COMPAT, profile)
    }

    @Test
    fun resolvePreferredLongFormPlaybackProfile_fallsBackToCompatWhenHevcUnsupported() {
        val profile = resolvePreferredLongFormPlaybackProfile(
            deviceSupportsHevc = false,
            isProbablyEmulator = false,
        )

        assertEquals(PlaybackProfile.COMPAT, profile)
    }

    @Test
    fun resolvePreferredLongFormPlaybackProfile_fallsBackToCompatOnEmulator() {
        val profile = resolvePreferredLongFormPlaybackProfile(
            deviceSupportsHevc = true,
            isProbablyEmulator = true,
        )

        assertEquals(PlaybackProfile.COMPAT, profile)
    }
}
