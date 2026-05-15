package com.chee.videos.feature.home

import com.chee.videos.core.ui.TvFocusSafeSpec
import org.junit.Assert.assertTrue
import org.junit.Test

class HomeFocusLayoutSpecTest {
    @Test
    fun `home poster grids keep focus safe padding and spacing`() {
        assertTrue(HomeFocusLayoutSpec.gridHorizontalPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(HomeFocusLayoutSpec.gridTopPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(HomeFocusLayoutSpec.gridItemSpacingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp * 2)
    }

    @Test
    fun `home cards use focus safe outer containers`() {
        assertTrue(HomeFocusLayoutSpec.avPosterCardUsesFocusSafeContainer)
        assertTrue(HomeFocusLayoutSpec.videoCardUsesFocusSafeContainer)
    }
}
