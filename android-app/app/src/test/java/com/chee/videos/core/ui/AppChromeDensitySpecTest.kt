package com.chee.videos.core.ui

import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.ui.unit.Density
import androidx.compose.ui.unit.LayoutDirection
import androidx.compose.ui.unit.dp
import org.junit.Assert.assertEquals
import org.junit.Assert.fail
import org.junit.Test

class AppChromeDensitySpecTest {
    private val density = Density(1f)

    @Test
    fun `app chrome uses slightly tighter shared shapes`() {
        assertCornerRadius(18, AppChrome.CardShape)
        assertCornerRadius(14, AppChrome.SectionShape)
    }

    @Test
    fun `long form player overlays use compact chrome metrics`() {
        val spec = longFormPlayerChromeSpec()

        assertEquals(14, spec.seekPreviewCornerDp)
        assertEquals(18, spec.centerFeedbackCornerDp)
        assertEquals(16, spec.controlBarCornerDp)
        assertEquals(10, spec.controlBarOuterPaddingDp)
        assertEquals(6, spec.controlBarHorizontalPaddingDp)
        assertEquals(5, spec.controlBarVerticalPaddingDp)
        assertEquals(32, spec.controlButtonSizeDp)
    }

    private fun assertCornerRadius(expectedDp: Int, shape: RoundedCornerShape) {
        val outline = shape.createOutline(
            size = androidx.compose.ui.geometry.Size(100f, 100f),
            layoutDirection = LayoutDirection.Ltr,
            density = density,
        )
        val roundRect = (outline as androidx.compose.ui.graphics.Outline.Rounded).roundRect
        assertEquals(expectedDp.dp.value, roundRect.topLeftCornerRadius.x, 0.0001f)
        assertEquals(expectedDp.dp.value, roundRect.topRightCornerRadius.x, 0.0001f)
        assertEquals(expectedDp.dp.value, roundRect.bottomLeftCornerRadius.x, 0.0001f)
        assertEquals(expectedDp.dp.value, roundRect.bottomRightCornerRadius.x, 0.0001f)
    }

    private fun longFormPlayerChromeSpec(): Any {
        return try {
            Class
                .forName("com.chee.videos.core.ui.LongFormVideoPlayerKt")
                .getDeclaredMethod("buildLongFormPlayerChromeSpec")
                .apply { isAccessible = true }
                .invoke(null)
        } catch (exception: NoSuchMethodException) {
            fail("LongFormVideoPlayer must expose compact chrome metrics through buildLongFormPlayerChromeSpec")
        }
    }

    private val Any.seekPreviewCornerDp: Int
        get() = readInt("getSeekPreviewCornerDp")

    private val Any.centerFeedbackCornerDp: Int
        get() = readInt("getCenterFeedbackCornerDp")

    private val Any.controlBarCornerDp: Int
        get() = readInt("getControlBarCornerDp")

    private val Any.controlBarOuterPaddingDp: Int
        get() = readInt("getControlBarOuterPaddingDp")

    private val Any.controlBarHorizontalPaddingDp: Int
        get() = readInt("getControlBarHorizontalPaddingDp")

    private val Any.controlBarVerticalPaddingDp: Int
        get() = readInt("getControlBarVerticalPaddingDp")

    private val Any.controlButtonSizeDp: Int
        get() = readInt("getControlButtonSizeDp")

    private fun Any.readInt(methodName: String): Int {
        return javaClass.getDeclaredMethod(methodName).invoke(this) as Int
    }
}
