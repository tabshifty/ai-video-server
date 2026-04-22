package com.chee.videos.feature.imagecollections

import androidx.compose.ui.geometry.Offset
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ImageCollectionViewerZoomMathTest {

    @Test
    fun smallImage_doesNotSupportZoomIn() {
        val maxScale = imageViewerMaxScale(
            imageWidthPx = 600f,
            imageHeightPx = 500f,
            containerWidthPx = 1080f,
            containerHeightPx = 1920f,
        )

        assertEquals(1f, maxScale, 0.0001f)
    }

    @Test
    fun largeImage_supportsZoomInWithDynamicCap() {
        val maxScale = imageViewerMaxScale(
            imageWidthPx = 3240f,
            imageHeightPx = 4320f,
            containerWidthPx = 1080f,
            containerHeightPx = 1920f,
        )

        assertEquals(3f, maxScale, 0.0001f)
    }

    @Test
    fun invalidImageSize_fallsBackToOne() {
        val maxScale = imageViewerMaxScale(
            imageWidthPx = 0f,
            imageHeightPx = 0f,
            containerWidthPx = 1080f,
            containerHeightPx = 1920f,
        )

        assertEquals(1f, maxScale, 0.0001f)
    }

    @Test
    fun clampOffset_returnsCenterWhenScaleAtOrBelowOne() {
        val clamped = imageViewerClampOffset(
            offset = Offset(240f, -180f),
            scale = 1f,
            imageWidthPx = 2400f,
            imageHeightPx = 3200f,
            containerWidthPx = 1080f,
            containerHeightPx = 1920f,
        )

        assertEquals(0f, clamped.x, 0.0001f)
        assertEquals(0f, clamped.y, 0.0001f)
    }

    @Test
    fun clampOffset_limitsPanWithinScaledBounds() {
        val clamped = imageViewerClampOffset(
            offset = Offset(900f, 900f),
            scale = 2f,
            imageWidthPx = 1080f,
            imageHeightPx = 1920f,
            containerWidthPx = 1080f,
            containerHeightPx = 1920f,
        )

        assertTrue(clamped.x <= 540.0001f)
        assertTrue(clamped.y <= 960.0001f)
    }
}
