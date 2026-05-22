package com.chee.videos.feature.tv

import org.junit.Assert.assertTrue
import org.junit.Test

class TvFeaturedHeroMotionSpecTest {

    private val catalogSource: String by lazy {
        java.nio.file.Path
            .of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")
            .toFile()
            .readText()
    }

    private val accessibilitySource: String by lazy {
        java.nio.file.Path
            .of("src/main/java/com/chee/videos/core/ui/TvAccessibilityMotion.kt")
            .toFile()
            .readText()
    }

    private val heroBody: String by lazy {
        val start = catalogSource.indexOf("private fun TvFeaturedHero(")
        require(start >= 0) { "TvCatalogScreen.kt 未找到 private fun TvFeaturedHero(" }
        val end = catalogSource.indexOf("private fun TvFeaturedPoster(", start)
        require(end > start) { "TvCatalogScreen.kt 未找到 TvFeaturedHero 的结束锚点 TvFeaturedPoster" }
        catalogSource.substring(start, end)
    }

    @Test
    fun `hero declares infinite transition with ramp duration token`() {
        assertTrue(
            "TvFeaturedHero 必须使用 rememberInfiniteTransition 驱动 ken-burns 环境动效",
            heroBody.contains("rememberInfiniteTransition("),
        )
        assertTrue(
            "TvFeaturedHero 必须使用 infiniteRepeatable 包裹 tween",
            heroBody.contains("infiniteRepeatable("),
        )
        assertTrue(
            "TvFeaturedHero 必须使用 RepeatMode.Reverse 做双向往返而非 Restart",
            heroBody.contains("RepeatMode.Reverse"),
        )
        assertTrue(
            "TvFeaturedHero 的 tween duration 必须取自 TvHeroMotionTokens.RampDurationMs 而非字面量",
            heroBody.contains("TvHeroMotionTokens.RampDurationMs"),
        )
        assertTrue(
            "TvFeaturedHero 的 tween easing 必须复用 TvMotionTokens.EasingStandard",
            heroBody.contains("TvMotionTokens.EasingStandard"),
        )
    }

    @Test
    fun `hero references scale and pan tokens not literals`() {
        assertTrue(heroBody.contains("TvHeroMotionTokens.ScaleStart"))
        assertTrue(heroBody.contains("TvHeroMotionTokens.ScaleEnd"))
        assertTrue(heroBody.contains("TvHeroMotionTokens.ScaleStaticTarget"))
        assertTrue(heroBody.contains("TvHeroMotionTokens.PanOffsetXDp"))
        assertTrue(heroBody.contains("TvHeroMotionTokens.PanOffsetYDp"))
    }

    @Test
    fun `hero body forbids bare literals that should be tokenized`() {
        assertTrue(
            "TvFeaturedHero 函数体禁止裸 120_000——必须走 TvHeroMotionTokens.RampDurationMs",
            !heroBody.contains("120_000"),
        )
        assertTrue(
            "TvFeaturedHero 函数体禁止裸 1.05f——必须走 TvHeroMotionTokens.ScaleStart",
            !heroBody.contains("1.05f"),
        )
        assertTrue(
            "TvFeaturedHero 函数体禁止裸 1.10f——必须走 TvHeroMotionTokens.ScaleEnd",
            !heroBody.contains("1.10f"),
        )
        assertTrue(
            "TvFeaturedHero 函数体禁止裸 1.075f——必须走 TvHeroMotionTokens.ScaleStaticTarget",
            !heroBody.contains("1.075f"),
        )
    }

    @Test
    fun `hero applies graphics layer and consults reduce motion helper`() {
        assertTrue(
            "TvFeaturedHero 必须用 graphicsLayer 把 scale + translation 挂到 backdrop AsyncImage",
            heroBody.contains("graphicsLayer"),
        )
        assertTrue(
            "TvFeaturedHero 必须读 rememberTvReduceMotionEnabled 才能在系统关动画时冻结",
            heroBody.contains("rememberTvReduceMotionEnabled"),
        )
    }

    @Test
    fun `accessibility helper reads animator duration scale from settings global`() {
        assertTrue(
            "TvAccessibilityMotion.kt 必须读 Settings.Global.ANIMATOR_DURATION_SCALE 探测 reduce-motion",
            accessibilitySource.contains("Settings.Global.ANIMATOR_DURATION_SCALE"),
        )
        assertTrue(
            "rememberTvReduceMotionEnabled 必须是 @Composable",
            accessibilitySource.contains("@Composable"),
        )
        assertTrue(
            "rememberTvReduceMotionEnabled 必须经 LocalContext.current 拿 contentResolver",
            accessibilitySource.contains("LocalContext.current"),
        )
        assertTrue(
            "必须用 remember 缓存结果，避免每次 recomposition 都打 ContentResolver",
            accessibilitySource.contains("remember("),
        )
    }
}
