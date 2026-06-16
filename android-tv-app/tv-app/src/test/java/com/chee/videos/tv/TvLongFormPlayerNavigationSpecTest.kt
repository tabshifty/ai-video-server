package com.chee.videos.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormPlayerNavigationSpecTest {
    @Test
    fun tvLongFormPlayerRoutesDisableNavTransitionsForSurfaceRelease() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvShellApp.kt").readText()

        assertTrue(source.contains("import androidx.compose.animation.EnterTransition"))
        assertTrue(source.contains("import androidx.compose.animation.ExitTransition"))
        assertPlayerRouteTransitionsDisabled(source, "TvLongFormPlayerRoutePattern", "TvLongFormPlayerScreen(")
        assertPlayerRouteTransitionsDisabled(source, "TvPlayerRoutePattern", "TvSeriesPlayerScreen(")
    }

    private fun assertPlayerRouteTransitionsDisabled(source: String, routePattern: String, screenCall: String) {
        val routeBlock = source.substringAfter("route = $routePattern,")
            .substringBefore(screenCall)

        listOf(
            "enterTransition = { EnterTransition.None }",
            "exitTransition = { ExitTransition.None }",
            "popEnterTransition = { EnterTransition.None }",
            "popExitTransition = { ExitTransition.None }",
        ).forEach { line ->
            assertTrue(
                "$routePattern should disable $line to avoid retaining SurfaceView playback pages behind detail screens",
                routeBlock.contains(line),
            )
        }
    }
}
