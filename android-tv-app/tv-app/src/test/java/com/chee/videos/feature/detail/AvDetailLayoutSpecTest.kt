package com.chee.videos.feature.detail

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class AvDetailLayoutSpecTest {
    @Test
    fun `av detail media layout uses 16 by 9 player ratio and top safe padding`() {
        val method = Class
            .forName("com.chee.videos.feature.detail.DetailScreenKt")
            .getDeclaredMethod("buildAvDetailMediaLayoutSpec")
            .apply { isAccessible = true }

        val spec = method.invoke(null)
        val specClass = spec.javaClass
        val aspectRatio = specClass.getDeclaredMethod("getAspectRatio").invoke(spec) as Float
        val applyStatusBarPadding = specClass.getDeclaredMethod("getApplyStatusBarPadding").invoke(spec) as Boolean

        assertEquals(16f / 9f, aspectRatio, 0.0001f)
        assertTrue(applyStatusBarPadding)
    }
}
