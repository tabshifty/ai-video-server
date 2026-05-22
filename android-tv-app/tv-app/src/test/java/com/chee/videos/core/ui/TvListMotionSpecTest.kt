package com.chee.videos.core.ui

import androidx.compose.animation.core.CubicBezierEasing
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test

class TvListMotionSpecTest {

    @Test
    fun `stagger tokens stay within tv perceived range`() {
        assertTrue(
            "TV 端单项 stagger 间隔必须落在 25–50ms（低于 25ms 看不见、高于 50ms 累计延迟过长）",
            TvListMotionTokens.StaggerPerItemMs in 25L..50L,
        )
        assertTrue(
            "TV 端单项淡入时长必须落在 200–280ms：低于 200ms 看不到动效、超过 280ms 会被感知为迟滞",
            TvListMotionTokens.StaggerEntryDurationMs in 200..280,
        )
        assertTrue(
            "TV 端 stagger 入场上移距离应控制在 8–20dp，避免大幅位移引起的视线追踪疲劳",
            TvListMotionTokens.StaggerEntryDistanceDp.value in 8f..20f,
        )
        assertTrue(
            "TV 端 stagger 阶梯上限必须收口（避免深处滚动 item 等待 1s+），合理范围 8–16",
            TvListMotionTokens.StaggerMaxSteps in 8..16,
        )
        assertNotNull(
            "stagger 入场必须使用 cubic 缓动而非 linear，参见原计划 A3 要求",
            TvListMotionTokens.StaggerEntryEasing as? CubicBezierEasing,
        )
    }

    @Test
    fun `stagger delay helper clamps deep indices`() {
        val perItem = TvListMotionTokens.StaggerPerItemMs
        val maxSteps = TvListMotionTokens.StaggerMaxSteps

        assertEquals(0L, tvStaggerEntryDelayMs(0))
        assertEquals(perItem, tvStaggerEntryDelayMs(1))
        assertEquals(3 * perItem, tvStaggerEntryDelayMs(3))
        assertEquals(
            "深处滚动入场的 item 不应等待超过 maxSteps * perItem，否则用户感觉懒加载卡顿",
            maxSteps * perItem,
            tvStaggerEntryDelayMs(maxSteps),
        )
        assertEquals(
            "超过 maxSteps 的 index 必须被夹紧到 maxSteps",
            maxSteps * perItem,
            tvStaggerEntryDelayMs(maxSteps + 50),
        )
        assertEquals(
            "负 index 不应产生负延迟",
            0L,
            tvStaggerEntryDelayMs(-3),
        )
    }

    @Test
    fun `tv list motion source declares modifier and tokens`() {
        val path = Path.of("src/main/java/com/chee/videos/core/ui/TvListMotion.kt")
        assertTrue("TvListMotion.kt 必须存在于 core/ui 下", path.exists())

        val source = path.readText()
        assertTrue(
            "必须导出 object TvListMotionTokens 集中暴露 stagger 入场 token",
            source.contains("object TvListMotionTokens"),
        )
        assertTrue(
            "必须暴露 fun Modifier.tvStaggerEntry(index) 作为唯一 stagger 接入点",
            source.contains("fun Modifier.tvStaggerEntry"),
        )
        assertTrue(
            "stagger 入场必须使用 graphicsLayer 完成 alpha + translation，避免触发布局回流",
            source.contains("graphicsLayer"),
        )
        assertTrue(
            "stagger 必须通过 LaunchedEffect 调度延时启动，单元测试已固定该结构",
            source.contains("LaunchedEffect"),
        )
        assertTrue(
            "stagger 必须复用 tvStaggerEntryDelayMs(...) 计算延时，避免 modifier 内联硬编码",
            source.contains("tvStaggerEntryDelayMs"),
        )
    }

    @Test
    fun `tv poster wall opts in to stagger entry`() {
        val path = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        assertTrue("TV 海报墙必须存在", path.exists())

        val source = path.readText()
        assertTrue(
            "TvPosterWallScreen 必须 import tvStaggerEntry，参见 A3 第一批接入点",
            source.contains("import com.chee.videos.core.ui.tvStaggerEntry"),
        )
        assertTrue(
            "TvPosterWallScreen 必须在 LazyVerticalGrid 内调用 .tvStaggerEntry(",
            source.contains(".tvStaggerEntry("),
        )
        assertTrue(
            "TvPosterWallScreen 必须改用 itemsIndexed 才能拿到 index 传给 stagger",
            source.contains("itemsIndexed"),
        )
    }
}
