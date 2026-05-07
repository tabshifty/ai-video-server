package com.chee.videos.core.ui

import com.chee.videos.core.model.ShortPlaybackMode
import androidx.compose.ui.unit.dp
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortVideoPlaybackChromeTest {

    @Test
    fun `loop playback mode keeps loop label and description`() {
        assertEquals("循环单视频", shortPlaybackModeLabel(ShortPlaybackMode.LOOP_ONE))
        assertEquals("播放模式：循环单视频", shortPlaybackModeContentDescription(ShortPlaybackMode.LOOP_ONE))
    }

    @Test
    fun `auto next playback mode keeps auto next label and description`() {
        assertEquals("自动播放下一个", shortPlaybackModeLabel(ShortPlaybackMode.AUTO_NEXT))
        assertEquals("播放模式：自动播放下一个", shortPlaybackModeContentDescription(ShortPlaybackMode.AUTO_NEXT))
    }

    @Test
    fun `non home short progress bar keeps extra bottom spacing`() {
        assertEquals(12.dp, ShortNonHomeProgressBarBottomSpacing)
    }

    @Test
    fun `overlay progress bar only shows when current video exists`() {
        assertTrue(shouldShowShortOverlayProgressBar("video-1"))
        assertFalse(shouldShowShortOverlayProgressBar(null))
        assertFalse(shouldShowShortOverlayProgressBar(""))
        assertFalse(shouldShowShortOverlayProgressBar("   "))
    }
}
