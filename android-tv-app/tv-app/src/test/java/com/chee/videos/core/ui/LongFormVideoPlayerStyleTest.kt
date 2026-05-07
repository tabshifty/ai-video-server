package com.chee.videos.core.ui

import androidx.media3.ui.CaptionStyleCompat
import org.junit.Assert.assertEquals
import org.junit.Test

class LongFormVideoPlayerStyleTest {
    @Test
    fun buildLongFormSubtitleStyle_usesTransparentBackgroundAndOutline() {
        val method = Class
            .forName("com.chee.videos.core.ui.LongFormVideoPlayerKt")
            .getDeclaredMethod("buildLongFormSubtitleStyle")
            .apply { isAccessible = true }

        val style = method.invoke(null) as CaptionStyleCompat

        assertEquals(0xFFFFFFFF.toInt(), style.foregroundColor)
        assertEquals(0x00000000, style.backgroundColor)
        assertEquals(0x00000000, style.windowColor)
        assertEquals(CaptionStyleCompat.EDGE_TYPE_OUTLINE, style.edgeType)
        assertEquals(0xB3000000.toInt(), style.edgeColor)
    }
}
