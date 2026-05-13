package com.chee.videos.core.ui

import android.view.KeyEvent
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class LongFormVideoPlayerTransportKeyTest {
    @Test
    fun hiddenControls_leftAndRightSeekTenSecondsAndBoostToThirtyOnRepeat() {
        assertEquals(
            TvHiddenTransportKeyAction.Seek(-10_000L),
            resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_DPAD_LEFT, repeatCount = 0),
        )
        assertEquals(
            TvHiddenTransportKeyAction.Seek(-30_000L),
            resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_DPAD_LEFT, repeatCount = 1),
        )
        assertEquals(
            TvHiddenTransportKeyAction.Seek(10_000L),
            resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_DPAD_RIGHT, repeatCount = 0),
        )
        assertEquals(
            TvHiddenTransportKeyAction.Seek(30_000L),
            resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_DPAD_RIGHT, repeatCount = 2),
        )
    }

    @Test
    fun hiddenControls_centerAndPlayKeysTogglePlayback() {
        listOf(
            KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
            KeyEvent.KEYCODE_MEDIA_PLAY,
            KeyEvent.KEYCODE_MEDIA_PAUSE,
            KeyEvent.KEYCODE_DPAD_CENTER,
            KeyEvent.KEYCODE_ENTER,
            KeyEvent.KEYCODE_NUMPAD_ENTER,
        ).forEach { keyCode ->
            assertEquals(
                TvHiddenTransportKeyAction.TogglePlayPause,
                resolveTvHiddenTransportKeyAction(keyCode, repeatCount = 0),
            )
        }
    }

    @Test
    fun hiddenControls_upAndDownRevealControls() {
        assertEquals(
            TvHiddenTransportKeyAction.ShowControlsAndFocusPlayPause,
            resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_DPAD_UP, repeatCount = 0),
        )
        assertEquals(
            TvHiddenTransportKeyAction.ShowControlsAndFocusPlayPause,
            resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_DPAD_DOWN, repeatCount = 0),
        )
    }

    @Test
    fun hiddenControls_menuOpensSubtitleSheet() {
        assertEquals(
            TvHiddenTransportKeyAction.OpenSubtitleSheet,
            resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_MENU, repeatCount = 0),
        )
    }

    @Test
    fun unknownKeysReturnNull() {
        assertNull(resolveTvHiddenTransportKeyAction(KeyEvent.KEYCODE_BACK, repeatCount = 0))
    }
}
