package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertSame
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * TV 端动画 defer-read 审计：动画 State 的读取必须延后到 graphicsLayer 作用域，
 * 动画收敛期间只触发 layer 重绘、不逐帧重组挂载节点。禁止「简化」回 by 解包形态。
 */
class TvDeferredAnimationReadSpecTest {

    private val focusSource: String by lazy {
        Path.of("src/main/java/com/chee/videos/core/ui/TvFocus.kt").readText()
    }

    private val listMotionSource: String by lazy {
        Path.of("src/main/java/com/chee/videos/core/ui/TvListMotion.kt").readText()
    }

    private val catalogSource: String by lazy {
        Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt").readText()
    }

    private val heroBody: String by lazy {
        val start = catalogSource.indexOf("private fun TvFeaturedHero(")
        require(start >= 0) { "TvCatalogScreen.kt 未找到 private fun TvFeaturedHero(" }
        val end = catalogSource.indexOf("private fun TvFeaturedPoster(", start)
        require(end > start) { "TvCatalogScreen.kt 未找到结束锚点 TvFeaturedPoster" }
        catalogSource.substring(start, end)
    }

    @Test
    fun focusScaleReadIsDeferredToGraphicsLayer() {
        assertTrue(
            "tvFocusableScaleOnly 必须保留 State 句柄（val scaleState = animateFloatAsState）而非 by 解包，" +
                "否则每次焦点移动都会让失焦/得焦两个节点在弹簧收敛期逐帧重组",
            focusSource.contains("val scaleState = animateFloatAsState"),
        )
        assertTrue(
            "tvFocusableScaleOnly 禁止回退到组合期解包（val scale by animateFloatAsState）",
            !focusSource.contains("val scale by animateFloatAsState"),
        )
        val layerIndex = focusSource.indexOf("graphicsLayer {")
        val readIndex = focusSource.indexOf("scaleState.value")
        assertTrue(
            "scaleState.value 必须在 graphicsLayer 块内读取",
            layerIndex in 0 until readIndex,
        )
    }

    @Test
    fun staggerEntryReadIsDeferredToGraphicsLayer() {
        assertTrue(
            "tvStaggerEntry 必须保留 State 句柄（val progressState = animateFloatAsState）",
            listMotionSource.contains("val progressState = animateFloatAsState"),
        )
        assertTrue(
            "tvStaggerEntry 禁止回退到组合期解包（val progress by animateFloatAsState）",
            !listMotionSource.contains("val progress by animateFloatAsState"),
        )
        assertTrue(
            "progressState.value 必须在 graphicsLayer 块内读取",
            listMotionSource.indexOf("graphicsLayer {") in 0 until listMotionSource.indexOf("progressState.value"),
        )
    }

    @Test
    fun heroKenBurnsReadIsDeferredToGraphicsLayer() {
        assertTrue(
            "TvFeaturedHero 必须保留 State 句柄（val progressState = transition.animateFloat）",
            heroBody.contains("val progressState = transition.animateFloat"),
        )
        assertTrue(
            "TvFeaturedHero 禁止回退到组合期解包（val progress by transition.animateFloat），" +
                "那会让 120s Ken Burns 逐帧重组整个 hero 子树",
            !heroBody.contains("val progress by transition.animateFloat"),
        )
        assertTrue(
            "progressState.value 必须在 graphicsLayer 块内读取",
            heroBody.indexOf("graphicsLayer {") in 0 until heroBody.indexOf("progressState.value"),
        )
        assertTrue(
            "旧的组合期派生值（val heroScale = / val heroTranslationX =）不得残留",
            !heroBody.contains("val heroScale = ") && !heroBody.contains("val heroTranslationX = "),
        )
    }

    @Test
    fun appChromeGradientsAreSingletonValues() {
        val chromeSource = Path.of("src/main/java/com/chee/videos/core/ui/AppChrome.kt").readText()
        assertTrue(
            "PageGradient/HeroGradient 必须是 val 单例赋值，禁止 get() 属性——每次读取新建 Brush 会使 background modifier 相等性失效",
            !Regex("""val\s+(PageGradient|HeroGradient)\s*:\s*Brush\s*\n\s*get\(\)""").containsMatchIn(chromeSource),
        )
        assertTrue(
            chromeSource.contains("val PageGradient: Brush = Brush.verticalGradient") &&
                chromeSource.contains("val HeroGradient: Brush = Brush.verticalGradient"),
        )
        assertTrue(
            "PageGradient 依赖 CanvasRaised/Canvas，颜色常量必须声明在渐变之前（object 初始化顺序）",
            chromeSource.indexOf("val CanvasRaised") in 0 until chromeSource.indexOf("val PageGradient"),
        )
        // 运行时同一实例：直接证明不再每次分配
        assertSame(AppChrome.PageGradient, AppChrome.PageGradient)
        assertSame(AppChrome.HeroGradient, AppChrome.HeroGradient)
    }
}
