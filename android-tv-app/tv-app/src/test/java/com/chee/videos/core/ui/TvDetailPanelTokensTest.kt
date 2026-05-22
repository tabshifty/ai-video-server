package com.chee.videos.core.ui

import org.junit.Assert.assertTrue
import org.junit.Test

class TvDetailPanelTokensTest {

    @Test
    fun `blur radius falls inside 12 to 32 dp budget`() {
        assertTrue(
            "BlurRadiusDp 应 ≥ 12dp 才能在 10-foot 视距下看出「玻璃」感",
            TvDetailPanelTokens.BlurRadiusDp.value >= 12f,
        )
        assertTrue(
            "BlurRadiusDp 应 ≤ 32dp，否则面板边缘过度晕开干扰文字可读性",
            TvDetailPanelTokens.BlurRadiusDp.value <= 32f,
        )
    }

    @Test
    fun `upper gradient height inside 16 to 40 dp budget`() {
        assertTrue(
            "UpperGradientHeightDp 应 ≥ 16dp 才能看出渐入效果",
            TvDetailPanelTokens.UpperGradientHeightDp.value >= 16f,
        )
        assertTrue(
            "UpperGradientHeightDp 应 ≤ 40dp，否则渐变占据过多视觉空间稀释面板内容",
            TvDetailPanelTokens.UpperGradientHeightDp.value <= 40f,
        )
    }

    @Test
    fun `top corner radius is at least 16 dp`() {
        assertTrue(
            "TopCornerRadiusDp 应 ≥ 16dp 与 TV 端整体圆角语言一致",
            TvDetailPanelTokens.TopCornerRadiusDp.value >= 16f,
        )
    }

    @Test
    fun `content padding has reasonable minimums`() {
        assertTrue(
            "ContentPaddingHorizontalDp 应 ≥ 24dp 保证文字两侧呼吸空间",
            TvDetailPanelTokens.ContentPaddingHorizontalDp.value >= 24f,
        )
        assertTrue(
            "ContentPaddingVerticalDp 应 ≥ 16dp 保证面板上下不挤压内容",
            TvDetailPanelTokens.ContentPaddingVerticalDp.value >= 16f,
        )
    }

    @Test
    fun `fallback scrim is strictly more opaque than blurred scrim`() {
        val blurredAlpha = TvDetailPanelTokens.ScrimColorBlurred.alpha
        val fallbackAlpha = TvDetailPanelTokens.ScrimColorFallback.alpha
        assertTrue(
            "ScrimColorFallback 的 alpha ($fallbackAlpha) 必须严格大于 ScrimColorBlurred 的 alpha ($blurredAlpha)——" +
                "API < 31 不模糊时需要更不透明的 scrim 保证文字可读",
            fallbackAlpha > blurredAlpha,
        )
    }

    @Test
    fun `tv detail panel file declares object and tokens`() {
        val source = java.nio.file.Path
            .of("src/main/java/com/chee/videos/core/ui/TvDetailPanel.kt")
            .toFile()
            .readText()
        assertTrue("TvDetailPanel.kt 必须声明 object TvDetailPanelTokens", source.contains("object TvDetailPanelTokens"))
        assertTrue("必须暴露 BlurRadiusDp", source.contains("BlurRadiusDp"))
        assertTrue("必须暴露 ScrimColorBlurred", source.contains("ScrimColorBlurred"))
        assertTrue("必须暴露 ScrimColorFallback", source.contains("ScrimColorFallback"))
        assertTrue("必须暴露 UpperGradientHeightDp", source.contains("UpperGradientHeightDp"))
        assertTrue("必须暴露 ContentPaddingHorizontalDp", source.contains("ContentPaddingHorizontalDp"))
        assertTrue("必须暴露 ContentPaddingVerticalDp", source.contains("ContentPaddingVerticalDp"))
        assertTrue("必须暴露 TopCornerRadiusDp", source.contains("TopCornerRadiusDp"))
    }
}
