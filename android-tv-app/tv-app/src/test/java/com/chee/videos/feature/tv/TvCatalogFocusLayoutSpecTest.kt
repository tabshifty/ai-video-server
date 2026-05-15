package com.chee.videos.feature.tv

import com.chee.videos.core.ui.TvFocusSafeSpec
import org.junit.Assert.assertTrue
import org.junit.Test

class TvCatalogFocusLayoutSpecTest {
    @Test
    fun `horizontal shelves keep edge and item focus safe space`() {
        assertTrue(TvCatalogFocusLayoutSpec.shelfHorizontalPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvCatalogFocusLayoutSpec.shelfVerticalPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvCatalogFocusLayoutSpec.shelfItemSpacingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp * 2)
    }

    @Test
    fun `poster shelf cards use focus safe outer containers`() {
        assertTrue(TvCatalogFocusLayoutSpec.posterCardsUseFocusSafeContainer)
    }
}
