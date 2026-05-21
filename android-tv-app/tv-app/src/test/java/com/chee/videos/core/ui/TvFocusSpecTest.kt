package com.chee.videos.core.ui

import org.junit.Assert.assertTrue
import org.junit.Test

class TvFocusSpecTest {
    @Test
    fun `poster focus safe space covers scale overflow and halo`() {
        assertTrue(
            TvFocusSafeSpec.posterFocusSafeSpaceDp >=
                TvFocusSafeSpec.requiredSafeSpaceDp(
                    baseSizeDp = TvFocusSafeSpec.posterCardWidthDp,
                    focusedScale = TvFocusSafeSpec.posterFocusedScale,
                    focusedHaloPaddingDp = TvFocusSafeSpec.focusedHaloPaddingDp,
                ),
        )
    }

    @Test
    fun `global tv focus glow uses cyan glow instead of hard pink border`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvFocus.kt").toFile().readText()

        assertTrue("TV 焦点语言不应再使用旧的粉红硬描边", !source.contains("0xFFFF5A7A"))
        assertTrue("TV 焦点语言应定义蓝青焦点色", source.contains("TvFocusGlowColor"))
        assertTrue("TV 焦点语言应使用柔和背景提亮", source.contains(".background("))
        assertTrue("TV 焦点语言不应依赖 border 作为默认焦点反馈", !source.contains(".border("))
    }
}
