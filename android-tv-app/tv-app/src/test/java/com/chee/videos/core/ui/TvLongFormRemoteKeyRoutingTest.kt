package com.chee.videos.core.ui

import android.view.KeyEvent
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvLongFormRemoteKeyRoutingTest {
    @Test
    fun hiddenControls_leftAndRightSeekAndDoNotEnterControls() {
        assertEquals(
            TvRemoteKeyAction.Seek(-15_000L),
            resolveTvRemoteKeyAction(
                visible = false,
                focusLayer = TvPlayerFocusLayer.Root,
                episodeRailEnabled = false,
                keyCode = KeyEvent.KEYCODE_DPAD_LEFT,
                repeatCount = 0,
                seekStepSec = 15,
            ),
        )
        assertEquals(
            TvRemoteKeyAction.Seek(20_000L),
            resolveTvRemoteKeyAction(
                visible = false,
                focusLayer = TvPlayerFocusLayer.Root,
                episodeRailEnabled = false,
                keyCode = KeyEvent.KEYCODE_DPAD_RIGHT,
                repeatCount = 0,
                seekStepSec = 20,
            ),
        )
    }

    @Test
    fun repeatedSeekUsesTripleStep() {
        assertEquals(
            TvRemoteKeyAction.Seek(-45_000L),
            resolveTvRemoteKeyAction(
                visible = false,
                focusLayer = TvPlayerFocusLayer.Root,
                episodeRailEnabled = false,
                keyCode = KeyEvent.KEYCODE_DPAD_LEFT,
                repeatCount = 1,
                seekStepSec = 15,
            ),
        )
        assertEquals(
            TvRemoteKeyAction.Seek(60_000L),
            resolveTvRemoteKeyAction(
                visible = true,
                focusLayer = TvPlayerFocusLayer.Root,
                episodeRailEnabled = false,
                keyCode = KeyEvent.KEYCODE_DPAD_RIGHT,
                repeatCount = 2,
                seekStepSec = 20,
            ),
        )
    }

    @Test
    fun downEntersControlsWhenHiddenOrVisibleOutsideControls() {
        assertEquals(
            TvRemoteKeyAction.EnterFocus,
            resolveTvRemoteKeyAction(false, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_DPAD_DOWN, 0, 10),
        )
        assertEquals(
            TvRemoteKeyAction.EnterFocus,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_DPAD_DOWN, 0, 10),
        )
        assertNull(resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Controls, false, KeyEvent.KEYCODE_DPAD_DOWN, 0, 10))
    }

    @Test
    fun controlsDownEntersEpisodeRailOnlyWhenEnabled() {
        assertEquals(
            TvRemoteKeyAction.EnterEpisodeRail,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Controls, true, KeyEvent.KEYCODE_DPAD_DOWN, 0, 10),
        )
        assertNull(resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.EpisodeRail, true, KeyEvent.KEYCODE_DPAD_DOWN, 0, 10))
    }

    @Test
    fun visibleControls_leftAndRightPassThroughOnlyWhenFocusIsInsideControls() {
        assertEquals(
            TvRemoteKeyAction.Seek(-10_000L),
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_DPAD_LEFT, 0, 10),
        )
        assertEquals(
            TvRemoteKeyAction.PassThrough,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Controls, false, KeyEvent.KEYCODE_DPAD_LEFT, 0, 10),
        )
        assertEquals(
            TvRemoteKeyAction.PassThrough,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Controls, false, KeyEvent.KEYCODE_DPAD_RIGHT, 0, 10),
        )
        assertEquals(
            TvRemoteKeyAction.PassThrough,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.EpisodeRail, true, KeyEvent.KEYCODE_DPAD_RIGHT, 0, 10),
        )
    }

    @Test
    fun upMovesBackThroughFocusLayers() {
        assertNull(resolveTvRemoteKeyAction(false, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_DPAD_UP, 0, 10))
        assertNull(resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_DPAD_UP, 0, 10))
        assertEquals(
            TvRemoteKeyAction.ExitFocus,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Controls, false, KeyEvent.KEYCODE_DPAD_UP, 0, 10),
        )
        assertEquals(
            TvRemoteKeyAction.ExitEpisodeRail,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.EpisodeRail, true, KeyEvent.KEYCODE_DPAD_UP, 0, 10),
        )
    }

    @Test
    fun okTogglesPlaybackOutsideControlsAndPassesThroughInsideControls() {
        listOf(
            KeyEvent.KEYCODE_DPAD_CENTER,
            KeyEvent.KEYCODE_ENTER,
            KeyEvent.KEYCODE_NUMPAD_ENTER,
            KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
            KeyEvent.KEYCODE_MEDIA_PLAY,
            KeyEvent.KEYCODE_MEDIA_PAUSE,
        ).forEach { keyCode ->
            assertEquals(
                TvRemoteKeyAction.TogglePlayPause,
                resolveTvRemoteKeyAction(false, TvPlayerFocusLayer.Root, false, keyCode, 0, 10),
            )
            assertEquals(
                TvRemoteKeyAction.PassThrough,
                resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Controls, false, keyCode, 0, 10),
            )
            assertEquals(
                TvRemoteKeyAction.PassThrough,
                resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.EpisodeRail, true, keyCode, 0, 10),
            )
        }
    }

    @Test
    fun backDismissesVisibleUiAndPassesThroughWhenHidden() {
        assertEquals(
            TvRemoteKeyAction.DismissUi,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_BACK, 0, 10),
        )
        assertEquals(
            TvRemoteKeyAction.DismissUi,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.Controls, false, KeyEvent.KEYCODE_ESCAPE, 0, 10),
        )
        assertEquals(
            TvRemoteKeyAction.DismissUi,
            resolveTvRemoteKeyAction(true, TvPlayerFocusLayer.EpisodeRail, true, KeyEvent.KEYCODE_BACK, 0, 10),
        )
        assertNull(resolveTvRemoteKeyAction(false, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_BACK, 0, 10))
    }

    @Test
    fun unknownKeysPassThrough() {
        assertNull(resolveTvRemoteKeyAction(false, TvPlayerFocusLayer.Root, false, KeyEvent.KEYCODE_MENU, 0, 10))
    }
}
