package com.chee.videos.feature.player

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class UnifiedPlayerShortFitToggleTest {

    @Test
    fun `shows short fit toggle only for short videos`() {
        assertTrue(shouldShowUnifiedPlayerShortFitToggle("short"))
        assertFalse(shouldShowUnifiedPlayerShortFitToggle("SHORT"))
        assertFalse(shouldShowUnifiedPlayerShortFitToggle(" short "))
        assertFalse(shouldShowUnifiedPlayerShortFitToggle("movie"))
        assertFalse(shouldShowUnifiedPlayerShortFitToggle("episode"))
        assertFalse(shouldShowUnifiedPlayerShortFitToggle("av"))
        assertFalse(shouldShowUnifiedPlayerShortFitToggle(""))
    }
}
