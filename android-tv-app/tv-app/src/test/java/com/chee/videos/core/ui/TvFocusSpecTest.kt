package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
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

    @Test
    fun `press feedback tokens are defined within tv press range`() {
        assertTrue(
            "TV DPad 按下 scale 目标应落在 0.94-0.99 之间，给到可感知的下沉但不会过度形变",
            TvFocusMotionTokens.PressedScale in 0.94f..0.99f,
        )
        assertTrue(
            "TV DPad 按下阻尼应落在 0.55-0.85，按下的回弹比焦点放大略紧凑",
            TvFocusMotionTokens.PressDampingRatio in 0.55f..0.85f,
        )
        assertTrue(
            "TV DPad 按下刚度必须高于焦点放大刚度，让按下与回弹明显比悬停焦点快",
            TvFocusMotionTokens.PressStiffness > TvFocusMotionTokens.ScaleStiffness,
        )
    }

    @Test
    fun `resolveTvFocusableScaleTarget collapses disabled to one regardless of focus or press`() {
        assertEquals(
            1f,
            resolveTvFocusableScaleTarget(
                focused = true,
                pressed = true,
                enabled = false,
                focusedScale = 1.04f,
            ),
            0f,
        )
    }

    @Test
    fun `resolveTvFocusableScaleTarget collapses pressed to pressed scale ahead of focus`() {
        assertEquals(
            TvFocusMotionTokens.PressedScale,
            resolveTvFocusableScaleTarget(
                focused = true,
                pressed = true,
                enabled = true,
                focusedScale = 1.04f,
            ),
            0f,
        )
    }

    @Test
    fun `resolveTvFocusableScaleTarget returns focused scale when focused not pressed`() {
        assertEquals(
            1.04f,
            resolveTvFocusableScaleTarget(
                focused = true,
                pressed = false,
                enabled = true,
                focusedScale = 1.04f,
            ),
            0f,
        )
    }

    @Test
    fun `resolveTvFocusableScaleTarget returns one when neither focused nor pressed`() {
        assertEquals(
            1f,
            resolveTvFocusableScaleTarget(
                focused = false,
                pressed = false,
                enabled = true,
                focusedScale = 1.04f,
            ),
            0f,
        )
    }

    @Test
    fun `tv focus modifiers handle dpad center key event with press feedback`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvFocus.kt").toFile().readText()

        assertTrue(
            "tvFocusable* 必须接 onPreviewKeyEvent 才能拦截 DPad center / Enter 按下反馈",
            source.contains("onPreviewKeyEvent"),
        )
        assertTrue(
            "DPad center 必须显式接入 Key.DirectionCenter",
            source.contains("Key.DirectionCenter"),
        )
        assertTrue(
            "遥控/键盘 Enter 必须显式接入 Key.Enter",
            source.contains("Key.Enter"),
        )
        assertTrue(
            "部分遥控/键盘 NumPad Enter 必须显式接入 Key.NumPadEnter",
            source.contains("Key.NumPadEnter"),
        )
        assertTrue(
            "必须区分 KeyDown 与 KeyUp，KeyDown 进入按下态，KeyUp 回弹并触发触觉",
            source.contains("KeyEventType.KeyDown") && source.contains("KeyEventType.KeyUp"),
        )
    }

    @Test
    fun `tv focus modifiers route scale and press through shared helpers`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvFocus.kt").toFile().readText()

        assertTrue(
            "tvFocusable* 内 scale targetValue 必须通过 resolveTvFocusableScaleTarget 派生，不允许内联三元",
            source.contains("resolveTvFocusableScaleTarget("),
        )
        assertTrue(
            "tvFocusable* 必须按 pressed 状态选择不同 spring，集中在 tvFocusableScaleSpring",
            source.contains("tvFocusableScaleSpring("),
        )
    }

    @Test
    fun `tv press feedback triggers haptic with api guarded fallback`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvFocus.kt").toFile().readText()

        assertTrue(
            "支持 CONFIRM 触觉的设备应使用 HapticFeedbackConstants.CONFIRM",
            source.contains("HapticFeedbackConstants.CONFIRM"),
        )
        assertTrue(
            "低于 API 30 的设备必须回退到 HapticFeedbackConstants.VIRTUAL_KEY，避免触觉直接失败",
            source.contains("HapticFeedbackConstants.VIRTUAL_KEY"),
        )
        assertTrue(
            "触觉 API 必须按 SDK_INT 守门到 Build.VERSION_CODES.R",
            source.contains("Build.VERSION.SDK_INT") && source.contains("Build.VERSION_CODES.R"),
        )
    }
}
