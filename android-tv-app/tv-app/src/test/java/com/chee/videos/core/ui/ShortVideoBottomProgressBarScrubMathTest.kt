package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Test

class ShortVideoBottomProgressBarScrubMathTest {

    @Test
    fun deltaZero_keepsAnchorPosition() {
        val target = shortScrubTargetFromDelta(
            anchorMs = 3_000L,
            durationMs = 10_000L,
            deltaFraction = 0f,
        )

        assertEquals(3_000L, target)
    }

    @Test
    fun positiveDelta_movesForwardFromAnchor() {
        val target = shortScrubTargetFromDelta(
            anchorMs = 3_000L,
            durationMs = 10_000L,
            deltaFraction = 0.1f,
        )

        assertEquals(4_000L, target)
    }

    @Test
    fun negativeDelta_movesBackwardAndClampsToStart() {
        val target = shortScrubTargetFromDelta(
            anchorMs = 3_000L,
            durationMs = 10_000L,
            deltaFraction = -0.5f,
        )

        assertEquals(0L, target)
    }

    @Test
    fun positiveDelta_clampsToDurationEnd() {
        val target = shortScrubTargetFromDelta(
            anchorMs = 9_500L,
            durationMs = 10_000L,
            deltaFraction = 0.2f,
        )

        assertEquals(10_000L, target)
    }
}
