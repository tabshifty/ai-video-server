package com.chee.videos.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvRemoteCoordinatorViewModelStateTest {
    @Test
    fun clearStateClearsMatchingActiveSession() {
        val state = TvRemoteCoordinatorUiState(activeSessionId = "session-1")

        val cleared = clearTvRemoteCoordinatorState(state, "session-1")

        assertEquals(null, cleared.activeSessionId)
    }

    @Test
    fun clearStateKeepsOtherActiveSession() {
        val state = TvRemoteCoordinatorUiState(activeSessionId = "session-1")

        val kept = clearTvRemoteCoordinatorState(state, "session-2")

        assertEquals("session-1", kept.activeSessionId)
    }
}
