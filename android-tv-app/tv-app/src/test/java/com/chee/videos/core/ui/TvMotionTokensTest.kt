package com.chee.videos.core.ui

import androidx.compose.animation.core.CubicBezierEasing
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvMotionTokensTest {

    @Test
    fun `tv motion duration tokens fall inside 200 to 260 ms budget`() {
        assertTrue(
            "TV Fast duration 应落在 200-260ms 之间（不允许快于 200，避免动画感丢失）",
            TvMotionTokens.DurationFastMs in 200..260,
        )
        assertTrue(
            "TV Standard duration 应落在 200-260ms 之间（A5 budget 中段）",
            TvMotionTokens.DurationStandardMs in 200..260,
        )
        assertTrue(
            "TV Emphasized duration 应落在 200-260ms 之间（不允许超过 260，TV 端 300+ms 即迟滞）",
            TvMotionTokens.DurationEmphasizedMs in 200..260,
        )
    }

    @Test
    fun `tv motion duration tokens are ordered fast lt standard le emphasized`() {
        assertTrue(
            "Fast 应严格快于 Standard，否则 token 失去区分度",
            TvMotionTokens.DurationFastMs < TvMotionTokens.DurationStandardMs,
        )
        assertTrue(
            "Standard 应不超过 Emphasized，强调时长不应短于默认",
            TvMotionTokens.DurationStandardMs <= TvMotionTokens.DurationEmphasizedMs,
        )
    }

    @Test
    fun `tv motion standard easing is the material standard cubic bezier`() {
        assertTrue(
            "EasingStandard 必须是 CubicBezierEasing（强制非 Linear / FastOutSlowIn）",
            TvMotionTokens.EasingStandard is CubicBezierEasing,
        )
    }

    @Test
    fun `tv list motion tokens reuse tv motion tokens`() {
        assertEquals(
            "TvListMotionTokens.StaggerEntryDurationMs 必须复用 TvMotionTokens.DurationEmphasizedMs",
            TvMotionTokens.DurationEmphasizedMs,
            TvListMotionTokens.StaggerEntryDurationMs,
        )
        assertTrue(
            "TvListMotionTokens.StaggerEntryEasing 必须复用 TvMotionTokens.EasingStandard 同一实例",
            TvListMotionTokens.StaggerEntryEasing === TvMotionTokens.EasingStandard,
        )
    }

    @Test
    fun `tv motion tokens file declares object and shared exports`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvMotion.kt").toFile().readText()
        assertTrue("TvMotion.kt 必须声明共享 object TvMotionTokens", source.contains("object TvMotionTokens"))
        assertTrue("TvMotion.kt 必须暴露 DurationFastMs", source.contains("DurationFastMs"))
        assertTrue("TvMotion.kt 必须暴露 DurationStandardMs", source.contains("DurationStandardMs"))
        assertTrue("TvMotion.kt 必须暴露 DurationEmphasizedMs", source.contains("DurationEmphasizedMs"))
        assertTrue("TvMotion.kt 必须暴露 EasingStandard", source.contains("EasingStandard"))
    }

    @Test
    fun `tv list motion source references tv motion tokens`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvListMotion.kt").toFile().readText()
        assertTrue(
            "TvListMotion 应通过引用 TvMotionTokens 而非硬编码 cubic-bezier 字面量来获取 easing",
            source.contains("TvMotionTokens.EasingStandard"),
        )
        assertTrue(
            "TvListMotion 应通过引用 TvMotionTokens 而非硬编码数字来获取 duration",
            source.contains("TvMotionTokens.DurationEmphasizedMs"),
        )
        assertTrue(
            "TvListMotion 不应再写裸 CubicBezierEasing 字面量",
            !source.contains("CubicBezierEasing(0.2f, 0f, 0f, 1f)"),
        )
    }

    @Test
    fun `long form video player fades use shared tv motion tokens`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt").toFile().readText()
        assertTrue(
            "LongFormVideoPlayer 内 fadeIn / fadeOut 应通过 tween + TvMotionTokens 配置，禁止裸 fadeIn()/fadeOut()",
            !source.contains("fadeIn()") && !source.contains("fadeOut()"),
        )
        assertTrue(
            "LongFormVideoPlayer 必须引用 TvMotionTokens.DurationStandardMs 控制浮层 fade 时长",
            source.contains("TvMotionTokens.DurationStandardMs"),
        )
        assertTrue(
            "LongFormVideoPlayer 必须引用 TvMotionTokens.EasingStandard 控制浮层 fade 缓动",
            source.contains("TvMotionTokens.EasingStandard"),
        )
    }
}
