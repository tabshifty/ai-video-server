package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvPlayerBackConfirmTest {
    @Test
    fun firstBackPressShowsPrompt() {
        assertEquals(
            TvPlayerBackAction.ShowPrompt,
            resolveTvPlayerBackAction(previousPromptUptimeMillis = null, nowUptimeMillis = 1_000L),
        )
    }

    @Test
    fun secondBackPressWithinConfirmWindowExits() {
        assertEquals(
            TvPlayerBackAction.Exit,
            resolveTvPlayerBackAction(previousPromptUptimeMillis = 1_000L, nowUptimeMillis = 2_999L),
        )
    }

    @Test
    fun backPressAfterConfirmWindowShowsPromptAgain() {
        assertEquals(
            TvPlayerBackAction.ShowPrompt,
            resolveTvPlayerBackAction(previousPromptUptimeMillis = 1_000L, nowUptimeMillis = 3_001L),
        )
    }
}
