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
    fun `shared fullscreen host stays in compose tree and avoids dialog surface overlap`() {
        val source = Path.of("src/main/java/com/chee/videos/core/ui/ShortOverlayFullscreenHost.kt").readText()

        assertTrue("短视频全屏不能再用 Dialog 叠加，否则会和竖屏 PlayerView 争用 surface 并露出底层 UI", !source.contains("Dialog("))
        assertTrue("短视频全屏不能再依赖 DialogProperties，外层壳应通过 isShortFullscreen 隐藏", !source.contains("DialogProperties("))
    }

    @Test
    fun `home short feed hides vertical pager while fullscreen`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/shorts/ShortFeedScreen.kt").readText()

        assertTrue("主页短视频全屏时不能继续渲染竖屏 VerticalPager，否则会和全屏播放器叠层", source.contains("if (isFullscreen)"))
        assertTrue("主页短视频全屏时必须只渲染 ShortOverlayFullscreenHost", source.contains("ShortOverlayFullscreenHost("))
        assertTrue("主页短视频竖屏内容必须放在 else 分支，避免 PlayerView surface 争用", source.contains("} else {"))
    }

    @Test
    fun `all non home short overlays hide vertical pager while fullscreen`() {
        assertShortFullscreenBranches(
            "src/main/java/com/chee/videos/feature/shortsearch/ShortSearchScreen.kt",
            "isFullscreen",
        )
        assertShortFullscreenBranches(
            "src/main/java/com/chee/videos/feature/shortdiscover/ShortDiscoverScreen.kt",
            "isFullscreen",
        )
        assertShortFullscreenBranches(
            "src/main/java/com/chee/videos/feature/player/UnifiedPlayerScreen.kt",
            "isShortFullscreen",
        )
    }

    @Test
    fun `short fullscreen state must hide the app shell chrome`() {
        val appSource = Path.of("src/main/java/com/chee/videos/VideoHomeApp.kt").readText()
        val homeSource = Path.of("src/main/java/com/chee/videos/feature/home/HomeScreen.kt").readText()

        assertTrue("应用壳必须感知短视频全屏状态，才能隐藏底部 tabbar", appSource.contains("isShortFullscreen"))
        assertTrue("首页必须感知短视频全屏状态，才能隐藏头部导航", homeSource.contains("isShortFullscreen"))
        assertTrue("首页头部必须在短视频全屏时隐藏", homeSource.contains("if (!isShortFullscreen)"))
    }

    private fun assertSourceUsesFullscreenHost(path: String) {
        val source = Path.of(path).readText()
        assertTrue("$path 必须导入 ShortOverlayFullscreenHost", source.contains("ShortOverlayFullscreenHost"))
        assertTrue("$path 必须导入 ShortOverlayFullscreenButton", source.contains("ShortOverlayFullscreenButton"))
    }

    private fun assertShortFullscreenBranches(path: String, stateName: String) {
        val source = Path.of(path).readText()
        assertTrue("$path 全屏时必须进入独立分支", source.contains("if ($stateName)"))
        assertTrue("$path 全屏分支必须渲染 ShortOverlayFullscreenHost", source.contains("ShortOverlayFullscreenHost("))
        assertTrue("$path 竖屏 VerticalPager 必须放在非全屏 else 分支", source.contains("} else {"))
    }
}
