package com.chee.videos.core.ui

import androidx.media3.common.Player
import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortOverlayFullscreenSpecTest {

    @Test
    fun `while fullscreen repeat mode is always forced to one`() {
        assertEquals(Player.REPEAT_MODE_ONE, shortOverlayRepeatModeWhileFullscreen(Player.REPEAT_MODE_OFF))
        assertEquals(Player.REPEAT_MODE_ONE, shortOverlayRepeatModeWhileFullscreen(Player.REPEAT_MODE_ONE))
        assertEquals(Player.REPEAT_MODE_ONE, shortOverlayRepeatModeWhileFullscreen(Player.REPEAT_MODE_ALL))
    }

    @Test
    fun `after fullscreen repeat mode restores fallback`() {
        assertEquals(Player.REPEAT_MODE_OFF, shortOverlayRepeatModeAfterFullscreen(Player.REPEAT_MODE_OFF))
        assertEquals(Player.REPEAT_MODE_ONE, shortOverlayRepeatModeAfterFullscreen(Player.REPEAT_MODE_ONE))
        assertEquals(Player.REPEAT_MODE_ALL, shortOverlayRepeatModeAfterFullscreen(Player.REPEAT_MODE_ALL))
    }

    @Test
    fun `short overlay files use the shared fullscreen host and button`() {
        assertSourceUsesFullscreenHost("src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt")
        assertSourceUsesFullscreenHost("src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt")
        assertSourceUsesFullscreenHost("src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt")
        assertSourceUsesFullscreenHost("src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt")
    }

    @Test
    fun `short overlay files do not hardcode fullscreen orientation or icon`() {
        val searchSource = Path.of("src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt").readText()
        val discoverSource = Path.of("src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt").readText()
        val feedSource = Path.of("src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt").readText()
        val playerSource = Path.of("src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt").readText()

        listOf(searchSource, discoverSource, feedSource, playerSource).forEach { source ->
            assertTrue("全屏按钮图标必须复用共享组件，不得裸写 Icons.Filled.Fullscreen", !source.contains("Icons.Filled.Fullscreen"))
            assertTrue("横屏锁定必须由共享组件负责，不得裸写 SCREEN_ORIENTATION_LANDSCAPE", !source.contains("SCREEN_ORIENTATION_LANDSCAPE"))
        }
    }

    @Test
    fun `shared fullscreen host uses fullscreen dialog to escape parent scaffold`() {
        val source = Path.of("src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt").readText()

        assertTrue("短视频全屏必须用 Dialog 跨过首页头部和底部导航所在的外层 Scaffold", source.contains("Dialog("))
        assertTrue("短视频全屏 Dialog 必须显式配置 DialogProperties", source.contains("DialogProperties("))
        assertTrue("短视频全屏 Dialog 不能使用平台默认窄宽度", source.contains("usePlatformDefaultWidth = false"))
        assertTrue("短视频全屏 Dialog 必须关闭 decorFitsSystemWindows 才能铺满系统栏区域", source.contains("decorFitsSystemWindows = false"))
    }

    private fun assertSourceUsesFullscreenHost(path: String) {
        val source = Path.of(path).readText()
        assertTrue("$path 必须导入 ShortOverlayFullscreenHost", source.contains("ShortOverlayFullscreenHost"))
        assertTrue("$path 必须导入 ShortOverlayFullscreenButton", source.contains("ShortOverlayFullscreenButton"))
    }
}
