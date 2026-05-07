package com.chee.videos.feature.shorts

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortFeedActionRailVisibilityTest {

    @Test
    fun `hides action rail when detail sheet is open`() {
        assertFalse(
            shouldShowShortFeedActionRail(
                detailSheetOpen = true,
            ),
        )
    }

    @Test
    fun `shows action rail when detail sheet is closed`() {
        assertTrue(
            shouldShowShortFeedActionRail(
                detailSheetOpen = false,
            ),
        )
    }
}
