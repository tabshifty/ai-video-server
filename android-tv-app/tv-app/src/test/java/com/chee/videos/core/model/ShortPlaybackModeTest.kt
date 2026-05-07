package com.chee.videos.core.model

import androidx.media3.common.Player
import org.junit.Assert.assertEquals
import org.junit.Test

class ShortPlaybackModeTest {
    @Test
    fun mapsToPlayerRepeatMode() {
        assertEquals(Player.REPEAT_MODE_ONE, ShortPlaybackMode.LOOP_ONE.toPlayerRepeatMode())
        assertEquals(Player.REPEAT_MODE_OFF, ShortPlaybackMode.AUTO_NEXT.toPlayerRepeatMode())
    }
}
