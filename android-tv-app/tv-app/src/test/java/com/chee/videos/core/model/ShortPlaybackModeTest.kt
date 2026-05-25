package com.chee.videos.core.model

import org.junit.Assert.assertEquals
import org.junit.Test

class ShortPlaybackModeTest {
    @Test
    fun mapsToPlayerRepeatMode() {
        assertEquals(1, ShortPlaybackMode.LOOP_ONE.toPlayerRepeatMode())
        assertEquals(0, ShortPlaybackMode.AUTO_NEXT.toPlayerRepeatMode())
    }
}
