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

    @Test
    fun `av detail page orders playback actions metadata actors and overview`() {
        val method = Class
            .forName("com.chee.videos.feature.detail.DetailScreenKt")
            .getDeclaredMethod("buildAvDetailContentOrder")
            .apply { isAccessible = true }

        val order = method.invoke(null) as List<*>

        assertEquals(
            listOf("player", "actions", "work_info", "actors", "overview", "tags"),
            order,
        )
    }

    @Test
    fun `av detail actor rail uses horizontal scrolling and larger avatars`() {
        val method = Class
            .forName("com.chee.videos.feature.detail.DetailScreenKt")
            .getDeclaredMethod("buildAvDetailActorRailSpec")
            .apply { isAccessible = true }

        val spec = method.invoke(null)
        val specClass = spec.javaClass
        val avatarSizeDp = specClass.getDeclaredMethod("getAvatarSizeDp").invoke(spec) as Int
        val horizontalScroll = specClass.getDeclaredMethod("getHorizontalScroll").invoke(spec) as Boolean

        assertEquals(76, avatarSizeDp)
        assertTrue(horizontalScroll)
    }

    @Test
    fun `av detail auto starts only playable av detail pages`() {
        val method = Class
            .forName("com.chee.videos.feature.detail.DetailScreenKt")
            .getDeclaredMethod("shouldAutoStartAvPlayback", Boolean::class.javaPrimitiveType, Boolean::class.javaPrimitiveType)
            .apply { isAccessible = true }

        assertTrue(method.invoke(null, true, true) as Boolean)
        assertEquals(false, method.invoke(null, true, false))
        assertEquals(false, method.invoke(null, false, true))
    }
}
