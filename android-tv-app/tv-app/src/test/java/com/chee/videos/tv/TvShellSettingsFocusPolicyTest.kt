package com.chee.videos.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvShellSettingsFocusPolicyTest {
    @Test
    fun settingsButtonReturnsToHomeContentFromLeftOrDown() {
        assertEquals(
            TvShellSettingsFocusTarget.HomeContent,
            resolveTvShellSettingsFocusTarget(TvShellSettingsFocusDirection.Left),
        )
        assertEquals(
            TvShellSettingsFocusTarget.HomeContent,
            resolveTvShellSettingsFocusTarget(TvShellSettingsFocusDirection.Down),
        )
    }

    @Test
    fun settingsButtonBlocksOutOfBoundsRightOrUp() {
        assertEquals(
            TvShellSettingsFocusTarget.Boundary,
            resolveTvShellSettingsFocusTarget(TvShellSettingsFocusDirection.Right),
        )
        assertEquals(
            TvShellSettingsFocusTarget.Boundary,
            resolveTvShellSettingsFocusTarget(TvShellSettingsFocusDirection.Up),
        )
    }
}
