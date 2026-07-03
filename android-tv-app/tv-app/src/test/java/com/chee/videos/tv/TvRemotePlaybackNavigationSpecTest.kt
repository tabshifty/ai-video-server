package com.chee.videos.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRemotePlaybackNavigationSpecTest {
    @Test
    fun tvRemotePlaybackRouteDisablesTransitionsAndMountsDedicatedScreen() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvShellApp.kt").readText()

        listOf(
            "route = TvRemotePlaybackRoutePattern",
            "enterTransition = { EnterTransition.None }",
            "exitTransition = { ExitTransition.None }",
            "popEnterTransition = { EnterTransition.None }",
            "popExitTransition = { ExitTransition.None }",
            "TvRemotePlaybackScreen(",
        ).forEach { line ->
            assertTrue("远程投放页必须包含 $line", source.contains(line))
        }
    }
}
