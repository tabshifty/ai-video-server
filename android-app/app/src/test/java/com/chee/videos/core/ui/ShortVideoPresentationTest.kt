package com.chee.videos.core.ui

import androidx.media3.ui.AspectRatioFrameLayout
import androidx.compose.ui.layout.ContentScale
import com.chee.videos.core.model.VideoFitMode
import org.junit.Assert.assertEquals
import org.junit.Test

class ShortVideoPresentationTest {

    @Test
    fun `fill mode uses zoom player resize and crop poster scale`() {
        assertEquals(AspectRatioFrameLayout.RESIZE_MODE_ZOOM, shortVideoResizeMode(VideoFitMode.FILL))
        assertEquals(ContentScale.Crop, shortPosterContentScale(VideoFitMode.FILL))
    }

    @Test
    fun `fit mode uses fit player resize and fit poster scale`() {
        assertEquals(AspectRatioFrameLayout.RESIZE_MODE_FIT, shortVideoResizeMode(VideoFitMode.FIT))
        assertEquals(ContentScale.Fit, shortPosterContentScale(VideoFitMode.FIT))
    }
}
