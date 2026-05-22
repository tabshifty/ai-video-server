package com.chee.videos.core.ui

import androidx.compose.ui.graphics.Color
import kotlin.math.pow
import org.junit.Assert.assertTrue
import org.junit.Test

class TvColorContrastTest {
    @Test
    fun `text muted on surface elevated meets wcag aaa 7 to 1`() {
        val contrast = wcagContrastRatio(AppChrome.TextMuted, AppChrome.SurfaceElevated)
        assertTrue(
            "TextMuted on SurfaceElevated 必须 ≥ 7.0:1（WCAG AAA），实测 $contrast",
            contrast >= 7.0,
        )
    }

    @Test
    fun `text muted stays above 7 to 1 on every dark surface token`() {
        val muted = AppChrome.TextMuted
        val surfaces = listOf(
            "Surface" to AppChrome.Surface,
            "SurfaceElevated" to AppChrome.SurfaceElevated,
            "SurfaceMuted" to AppChrome.SurfaceMuted,
            "SurfaceStrong" to AppChrome.SurfaceStrong,
            "CanvasRaised" to AppChrome.CanvasRaised,
            "Canvas" to AppChrome.Canvas,
        )
        for ((name, surface) in surfaces) {
            val contrast = wcagContrastRatio(muted, surface)
            assertTrue(
                "TextMuted on $name 必须 ≥ 7.0:1，实测 $contrast",
                contrast >= 7.0,
            )
        }
    }

    @Test
    fun `text secondary on surface elevated meets wcag aaa 7 to 1`() {
        val contrast = wcagContrastRatio(AppChrome.TextSecondary, AppChrome.SurfaceElevated)
        assertTrue(
            "TextSecondary on SurfaceElevated 必须 ≥ 7.0:1（WCAG AAA），实测 $contrast",
            contrast >= 7.0,
        )
    }

    @Test
    fun `text primary on surface elevated meets wcag aaa 7 to 1`() {
        val contrast = wcagContrastRatio(AppChrome.TextPrimary, AppChrome.SurfaceElevated)
        assertTrue(
            "TextPrimary on SurfaceElevated 必须 ≥ 7.0:1（WCAG AAA），实测 $contrast",
            contrast >= 7.0,
        )
    }

    @Test
    fun `wcag relative luminance follows sRGB linearization formula`() {
        assertTrue(
            "白色相对亮度应 ≈ 1.0",
            kotlin.math.abs(wcagRelativeLuminance(Color.White) - 1.0) < 1e-3,
        )
        assertTrue(
            "黑色相对亮度应 ≈ 0.0",
            kotlin.math.abs(wcagRelativeLuminance(Color.Black)) < 1e-3,
        )
    }

    @Test
    fun `wcag contrast ratio respects ordering`() {
        val bw = wcagContrastRatio(Color.White, Color.Black)
        assertTrue("白底黑字应 ≈ 21:1，实测 $bw", kotlin.math.abs(bw - 21.0) < 0.05)
        val wb = wcagContrastRatio(Color.Black, Color.White)
        assertTrue("对比函数应对前后景对称，实测 $wb", kotlin.math.abs(bw - wb) < 1e-6)
    }
}

private fun wcagChannelLinear(srgb: Float): Double {
    val s = srgb.toDouble()
    return if (s <= 0.03928) s / 12.92 else ((s + 0.055) / 1.055).pow(2.4)
}

internal fun wcagRelativeLuminance(color: Color): Double {
    val r = wcagChannelLinear(color.red)
    val g = wcagChannelLinear(color.green)
    val b = wcagChannelLinear(color.blue)
    return 0.2126 * r + 0.7152 * g + 0.0722 * b
}

internal fun wcagContrastRatio(fg: Color, bg: Color): Double {
    val l1 = wcagRelativeLuminance(fg)
    val l2 = wcagRelativeLuminance(bg)
    val lighter = kotlin.math.max(l1, l2)
    val darker = kotlin.math.min(l1, l2)
    return (lighter + 0.05) / (darker + 0.05)
}
