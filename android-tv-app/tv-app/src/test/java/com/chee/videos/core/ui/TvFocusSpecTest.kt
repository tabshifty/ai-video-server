package com.chee.videos.core.ui

import org.junit.Assert.assertTrue
import org.junit.Test

class TvFocusSpecTest {
    @Test
    fun `poster focus safe space covers scale overflow and border`() {
        assertTrue(
            TvFocusSafeSpec.posterFocusSafeSpaceDp >=
                TvFocusSafeSpec.requiredSafeSpaceDp(
                    baseSizeDp = TvFocusSafeSpec.posterCardWidthDp,
                    focusedScale = TvFocusSafeSpec.posterFocusedScale,
                    focusedBorderWidthDp = TvFocusSafeSpec.focusedBorderWidthDp,
                ),
        )
    }
}
