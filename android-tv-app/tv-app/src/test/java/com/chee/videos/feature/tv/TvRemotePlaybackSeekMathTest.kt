package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvRemotePlaybackSeekMathTest {

    @Test
    fun singleForwardStepUsesNormalizedBaseStep() {
        val target = calculateTvRemoteSeekTarget(
            currentPositionMs = 10_000L,
            durationMs = 60_000L,
            stepSeconds = 10,
            repeatCount = 0,
            forward = true,
        )

        assertEquals(20_000L, target)
    }

    @Test
    fun repeatForwardStepUsesAcceleratedDelta() {
        val target = calculateTvRemoteSeekTarget(
            currentPositionMs = 10_000L,
            durationMs = 60_000L,
            stepSeconds = 10,
            repeatCount = 2,
            forward = true,
        )

        assertEquals(40_000L, target)
    }

    @Test
    fun backwardStepClampsToStart() {
        val target = calculateTvRemoteSeekTarget(
            currentPositionMs = 2_000L,
            durationMs = 60_000L,
            stepSeconds = 10,
            repeatCount = 0,
            forward = false,
        )

        assertEquals(0L, target)
    }

    @Test
    fun invalidDurationReturnsNull() {
        val target = calculateTvRemoteSeekTarget(
            currentPositionMs = 2_000L,
            durationMs = 0L,
            stepSeconds = 10,
            repeatCount = 0,
            forward = false,
        )

        assertNull(target)
    }
}
