package com.chee.videos.feature.tv

import com.chee.videos.core.ui.TvFocusSafeSpec
import com.chee.videos.core.ui.TvLayoutSpec
import org.junit.Assert.assertTrue
import org.junit.Test

class TvPosterWallFocusLayoutSpecTest {
    @Test
    fun `poster wall grid keeps focus safe padding and spacing`() {
        assertTrue(TvPosterWallFocusLayoutSpec.gridHorizontalPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridTopPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridBottomPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridBottomPaddingDp >= TvLayoutSpec.scrollBottomSafePaddingDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridItemSpacingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp * 2)
    }

    @Test
    fun `poster wall cards use focus safe outer containers`() {
        assertTrue(TvPosterWallFocusLayoutSpec.posterCardsUseFocusSafeContainer)
    }
}
