package com.chee.videos.feature.shorts

import androidx.compose.ui.layout.ContentScale
import com.chee.videos.core.model.VideoFitMode
import org.junit.Assert.assertEquals
import org.junit.Test

class ShortFeedPosterScaleTest {

    @Test
    fun `fill mode uses crop poster scale`() {
        assertEquals(ContentScale.Crop, shortPosterContentScale(VideoFitMode.FILL))
    }

    @Test
    fun `fit mode uses fit poster scale`() {
        assertEquals(ContentScale.Fit, shortPosterContentScale(VideoFitMode.FIT))
    }
}
