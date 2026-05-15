package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class LongFormGestureSpecTest {
    @Test
    fun gestureArea_fullscreenLeftSideControlsBrightness() {
        assertEquals(
            LongFormVerticalAdjustmentTarget.Brightness,
            longFormVerticalAdjustmentTarget(
                isFullscreen = true,
                startX = 120f,
                widthPx = 400f,
            ),
        )
    }

    @Test
    fun gestureArea_fullscreenRightSideControlsVolume() {
        assertEquals(
            LongFormVerticalAdjustmentTarget.Volume,
            longFormVerticalAdjustmentTarget(
                isFullscreen = true,
                startX = 260f,
                widthPx = 400f,
            ),
        )
    }

    @Test
    fun gestureArea_notAvailableOutsideFullscreen() {
        assertNull(
            longFormVerticalAdjustmentTarget(
                isFullscreen = false,
                startX = 120f,
                widthPx = 400f,
            ),
        )
    }

    @Test
    fun verticalDrag_upIncreasesAndDownDecreasesWithinBounds() {
        assertEquals(
            0.75f,
            longFormAdjustedPercent(
                startPercent = 0.5f,
                dragDistanceY = -100f,
                heightPx = 400f,
            ),
            0.0001f,
        )
        assertEquals(
            0.25f,
            longFormAdjustedPercent(
                startPercent = 0.5f,
                dragDistanceY = 100f,
                heightPx = 400f,
            ),
            0.0001f,
        )
        assertEquals(
            1f,
            longFormAdjustedPercent(
                startPercent = 0.9f,
                dragDistanceY = -300f,
                heightPx = 400f,
            ),
            0.0001f,
        )
        assertEquals(
            0f,
            longFormAdjustedPercent(
                startPercent = 0.1f,
                dragDistanceY = 300f,
                heightPx = 400f,
            ),
            0.0001f,
        )
    }

    @Test
    fun brightnessPercent_defaultWindowValueStartsFromMiddle() {
        assertEquals(0.5f, longFormBrightnessPercent(-1f), 0.0001f)
        assertEquals(0.01f, longFormBrightnessPercent(0f), 0.0001f)
        assertEquals(1f, longFormBrightnessPercent(2f), 0.0001f)
    }

    @Test
    fun volumePercent_convertsCurrentAndMaxVolume() {
        assertEquals(0.4f, longFormVolumePercent(currentVolume = 4, maxVolume = 10), 0.0001f)
        assertEquals(0f, longFormVolumePercent(currentVolume = 4, maxVolume = 0), 0.0001f)
        assertEquals(1f, longFormVolumePercent(currentVolume = 12, maxVolume = 10), 0.0001f)
    }

    @Test
    fun percentFeedbackText_isRoundedAndClamped() {
        assertEquals("亮度 65%", longFormAdjustmentFeedbackText(LongFormVerticalAdjustmentTarget.Brightness, 0.654f))
        assertEquals("音量 0%", longFormAdjustmentFeedbackText(LongFormVerticalAdjustmentTarget.Volume, -1f))
        assertEquals("音量 100%", longFormAdjustmentFeedbackText(LongFormVerticalAdjustmentTarget.Volume, 2f))
    }
}
