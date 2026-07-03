package com.chee.videos.feature.shortsearch

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortSearchRemoteControlLifecycleSpecTest {
    @Test
    fun remoteControlScreenOnlyPollsWhileForegrounded() {
        val source = Path.of("src/main/java/com/chee/videos/feature/shortsearch/ShortSearchRemoteControlScreen.kt").readText()

        listOf(
            "Lifecycle.Event.ON_START -> {",
            "viewModel.setPollingEnabled(true)",
            "viewModel.refreshNow()",
            "Lifecycle.Event.ON_STOP -> viewModel.setPollingEnabled(false)",
        ).forEach { line ->
            assertTrue("ShortSearchRemoteControlScreen 必须包含 $line", source.contains(line))
        }
    }
}
