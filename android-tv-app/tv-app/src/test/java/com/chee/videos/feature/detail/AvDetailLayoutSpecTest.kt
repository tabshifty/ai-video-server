package com.chee.videos.feature.detail

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class AvDetailLayoutSpecTest {
    @Test
    fun `av detail media layout uses 16 by 9 player ratio and top safe padding`() {
        val spec = buildAvDetailMediaLayoutSpec()

        assertEquals(16f / 9f, spec.aspectRatio, 0.0001f)
        assertTrue(spec.applyStatusBarPadding)
    }
}
