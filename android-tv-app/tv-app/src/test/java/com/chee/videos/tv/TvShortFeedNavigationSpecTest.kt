package com.chee.videos.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvShortFeedNavigationSpecTest {
    @Test
    fun tvShortFeedRouteDisablesTransitionsAndReturnsFocusToShorts() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvShellApp.kt").readText()

        listOf(
            "enterTransition = { EnterTransition.None }",
            "exitTransition = { ExitTransition.None }",
            "popEnterTransition = { EnterTransition.None }",
            "popExitTransition = { ExitTransition.None }",
        ).forEach { line ->
            assertTrue("短视频页路由必须禁用 $line", source.contains(line))
        }
        assertTrue("短视频页路由必须挂载到 TvShortFeedRoute", source.contains("route = TvShortFeedRoute"))
        assertTrue("短视频页路由必须进入 TvShortFeedScreen", source.contains("TvShortFeedScreen("))
        assertTrue("返回短视频页时必须恢复到短视频入口焦点", source.contains("homeMenuFocusTarget = TvHomeMenuItem.Shorts"))
    }
}
