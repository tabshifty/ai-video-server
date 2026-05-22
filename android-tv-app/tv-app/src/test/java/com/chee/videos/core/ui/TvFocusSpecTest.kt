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

    @Test
    fun `tv focus modifiers replace tween easing with spring physics`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvFocus.kt").toFile().readText()

        assertTrue(
            "焦点缩放/光晕动画不应再使用 tween(durationMillis = 140) 这类线性近线性曲线，应改 spring 物理",
            !source.contains("tween(durationMillis = 140)"),
        )
        assertTrue(
            "tvFocusableGlow / tvFocusableScaleOnly 应使用 spring(...) 动效",
            source.contains("spring("),
        )
        assertTrue(
            "spring 动效参数必须来自共享 TvFocusMotionTokens，不能硬编码裸字面量",
            source.contains("TvFocusMotionTokens.ScaleDampingRatio") &&
                source.contains("TvFocusMotionTokens.ScaleStiffness"),
        )
    }

    @Test
    fun `tv focus motion tokens are defined and reusable`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvFocus.kt").toFile().readText()

        assertTrue(
            "应在 TvFocus.kt 暴露共享焦点动效 token 对象 TvFocusMotionTokens",
            source.contains("object TvFocusMotionTokens"),
        )
        assertTrue(
            "TvFocusMotionTokens 必须暴露 ScaleDampingRatio（焦点放大 spring 阻尼）",
            source.contains("ScaleDampingRatio"),
        )
        assertTrue(
            "TvFocusMotionTokens 必须暴露 ScaleStiffness（焦点放大 spring 刚度）",
            source.contains("ScaleStiffness"),
        )
        assertTrue(
            "TvFocusMotionTokens 必须暴露 SurfaceDampingRatio（光晕/提亮淡入 spring 阻尼）",
            source.contains("SurfaceDampingRatio"),
        )
        assertTrue(
            "TvFocusMotionTokens 必须暴露 SurfaceStiffness（光晕/提亮淡入 spring 刚度）",
            source.contains("SurfaceStiffness"),
        )
    }

    @Test
    fun `scale damping ratio is slightly bouncy and stiffness is medium-fast`() {
        assertTrue(
            "TV 焦点缩放阻尼应在 0.7-0.9 区间，保留细微回弹但不过冲",
            TvFocusMotionTokens.ScaleDampingRatio in 0.7f..0.9f,
        )
        assertTrue(
            "TV 焦点缩放刚度应在 320-440 区间，避免动效迟滞或过快丢失分量感",
            TvFocusMotionTokens.ScaleStiffness in 320f..440f,
        )
    }

    @Test
    fun `surface alpha spring is critically damped and faster than scale`() {
        assertTrue(
            "光晕/提亮淡入阻尼应等于或大于 1，避免视觉上的 alpha 抖动",
            TvFocusMotionTokens.SurfaceDampingRatio >= 1f,
        )
        assertTrue(
            "光晕/提亮淡入刚度应高于焦点缩放刚度，让背景反馈追上 scale 起步",
            TvFocusMotionTokens.SurfaceStiffness > TvFocusMotionTokens.ScaleStiffness,
        )
    }
}
