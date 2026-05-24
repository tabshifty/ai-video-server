package com.chee.videos.core.ui

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormControlsAutoHideTest {
    @Test
    fun interactiveActionsResetAutoHideTimer() {
        assertTrue(shouldResetAutoHideTimer(TvRemoteKeyAction.Seek(10_000L)))
        assertTrue(shouldResetAutoHideTimer(TvRemoteKeyAction.EnterFocus))
        assertTrue(shouldResetAutoHideTimer(TvRemoteKeyAction.TogglePlayPause))
    }

    @Test
    fun nonInteractiveOrDelegatedActionsDoNotResetAutoHideTimer() {
        assertFalse(shouldResetAutoHideTimer(TvRemoteKeyAction.ExitFocus))
        assertFalse(shouldResetAutoHideTimer(TvRemoteKeyAction.DismissUi))
        assertFalse(shouldResetAutoHideTimer(TvRemoteKeyAction.PassThrough))
    }
}
