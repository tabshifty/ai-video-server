package com.chee.videos.feature.shorts

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortFeedProgressVisibilityTest {

    @Test
    fun `hides progress bar while pager is transitioning`() {
        assertFalse(
            shouldShowShortFeedProgressBar(
                currentVideoId = "video-1",
                detailSheetOpen = false,
                pagerSettled = false,
            ),
        )
    }

    @Test
    fun `hides progress bar when detail sheet is open`() {
        assertFalse(
            shouldShowShortFeedProgressBar(
                currentVideoId = "video-1",
                detailSheetOpen = true,
                pagerSettled = true,
            ),
        )
    }

    @Test
    fun `shows progress bar only when current page is stable`() {
        assertTrue(
            shouldShowShortFeedProgressBar(
                currentVideoId = "video-1",
                detailSheetOpen = false,
                pagerSettled = true,
            ),
        )
    }
}
