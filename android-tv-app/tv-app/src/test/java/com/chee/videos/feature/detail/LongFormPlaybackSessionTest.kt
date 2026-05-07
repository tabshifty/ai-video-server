package com.chee.videos.feature.detail

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class LongFormPlaybackSessionTest {

    @Test
    fun `pause after playback start keeps player mounted`() {
        val pausedSession = LongFormPlaybackSession()
            .requestPlay(canPlay = true)
            .togglePlayPause(canPlay = true)

        assertTrue(pausedSession.shouldShowPlayer(canPlay = true))
        assertTrue(pausedSession.isPausedByUser)
        assertFalse(pausedSession.shouldResumeOnLifecycle())
    }

    @Test
    fun `resume after pause keeps player mounted and can auto resume`() {
        val resumedSession = LongFormPlaybackSession()
            .requestPlay(canPlay = true)
            .togglePlayPause(canPlay = true)
            .togglePlayPause(canPlay = true)

        assertTrue(resumedSession.shouldShowPlayer(canPlay = true))
        assertFalse(resumedSession.isPausedByUser)
        assertTrue(resumedSession.shouldResumeOnLifecycle())
    }

    @Test
    fun `request play without source keeps player hidden`() {
        val session = LongFormPlaybackSession().requestPlay(canPlay = false)

        assertFalse(session.hasStartedPlayback)
        assertFalse(session.shouldShowPlayer(canPlay = false))
        assertFalse(session.shouldResumeOnLifecycle())
    }
}
