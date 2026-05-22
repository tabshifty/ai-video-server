package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvHeroMotionTokensTest {

    @Test
    fun `ramp duration falls inside 60s to 300s budget`() {
        assertTrue(
            "TV hero ken-burns 半周期应 ≥ 60s——更快会让人察觉运动方向，破坏「环境动效」语义",
            TvHeroMotionTokens.RampDurationMs >= 60_000,
        )
        assertTrue(
            "TV hero ken-burns 半周期应 ≤ 300s——更慢则用户在一次会话中根本感知不到呼吸",
            TvHeroMotionTokens.RampDurationMs <= 300_000,
        )
    }

    @Test
    fun `scale range is strictly increasing and within safe area`() {
        assertTrue(
            "ScaleStart 必须严格小于 ScaleEnd，否则 lerp 无意义",
            TvHeroMotionTokens.ScaleStart < TvHeroMotionTokens.ScaleEnd,
        )
        assertTrue(
            "ScaleStart 应 ≥ 1.02f 才能为 translation pan 预留安全余量",
            TvHeroMotionTokens.ScaleStart >= 1.02f,
        )
        assertTrue(
            "ScaleStart 应 ≤ 1.10f，否则起手就过度放大、画质损失",
            TvHeroMotionTokens.ScaleStart <= 1.10f,
        )
        assertTrue(
            "ScaleEnd 应 ≥ 1.05f 才能让 10-foot 视距下能感觉到呼吸",
            TvHeroMotionTokens.ScaleEnd >= 1.05f,
        )
        assertTrue(
            "ScaleEnd 应 ≤ 1.15f，否则放大过度产生「在动」的察觉感",
            TvHeroMotionTokens.ScaleEnd <= 1.15f,
        )
    }

    @Test
    fun `static target sits at scale midpoint within tolerance`() {
        val midpoint = (TvHeroMotionTokens.ScaleStart + TvHeroMotionTokens.ScaleEnd) / 2f
        assertEquals(
            "ScaleStaticTarget 必须落在 (ScaleStart + ScaleEnd) / 2 的中点，避免 reduce-motion 时冻结端点造成视觉突变",
            midpoint,
            TvHeroMotionTokens.ScaleStaticTarget,
            0.001f,
        )
    }

    @Test
    fun `pan offsets are within safe range and y not exceeding x`() {
        assertTrue(
            "PanOffsetXDp 应 ≥ 4dp 才有可感知的水平漂移",
            TvHeroMotionTokens.PanOffsetXDp.value >= 4f,
        )
        assertTrue(
            "PanOffsetXDp 应 ≤ 16dp 否则被 ScaleStart=1.05 的安全余量挡不住、可能漏出 backdrop 边缘",
            TvHeroMotionTokens.PanOffsetXDp.value <= 16f,
        )
        assertTrue(
            "PanOffsetYDp 应 ≥ 2dp 才能形成菱形漂移而非单轴线性",
            TvHeroMotionTokens.PanOffsetYDp.value >= 2f,
        )
        assertTrue(
            "PanOffsetYDp 应 ≤ 8dp——hero 高度 324dp，垂直漂移过大会与上层渐变 / 文案 Row 视觉错位",
            TvHeroMotionTokens.PanOffsetYDp.value <= 8f,
        )
        assertTrue(
            "TV hero 横向更长，垂直振幅应不大于水平振幅以保持视觉自然",
            TvHeroMotionTokens.PanOffsetYDp.value <= TvHeroMotionTokens.PanOffsetXDp.value,
        )
    }

    @Test
    fun `tv hero motion file declares object and tokens`() {
        val source = java.nio.file.Path
            .of("src/main/java/com/chee/videos/core/ui/TvHeroMotion.kt")
            .toFile()
            .readText()
        assertTrue("TvHeroMotion.kt 必须声明共享 object TvHeroMotionTokens", source.contains("object TvHeroMotionTokens"))
        assertTrue("必须暴露 RampDurationMs", source.contains("RampDurationMs"))
        assertTrue("必须暴露 ScaleStart", source.contains("ScaleStart"))
        assertTrue("必须暴露 ScaleEnd", source.contains("ScaleEnd"))
        assertTrue("必须暴露 ScaleStaticTarget", source.contains("ScaleStaticTarget"))
        assertTrue("必须暴露 PanOffsetXDp", source.contains("PanOffsetXDp"))
        assertTrue("必须暴露 PanOffsetYDp", source.contains("PanOffsetYDp"))
    }
}
