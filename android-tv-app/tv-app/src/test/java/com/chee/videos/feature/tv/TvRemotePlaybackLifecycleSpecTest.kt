package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRemotePlaybackLifecycleSpecTest {
    @Test
    fun remotePlaybackScreenGatesPollingAndClearsCoordinatorOnExit() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvRemotePlaybackScreen.kt").readText()

        listOf(
            "Lifecycle.Event.ON_START -> {",
            "viewModel.setPollingEnabled(true)",
            "viewModel.refreshNow()",
            "Lifecycle.Event.ON_STOP -> viewModel.setPollingEnabled(false)",
            "onSessionEnded(viewModel.remoteSessionId)",
            "onSessionEnded(session.sessionId)",
        ).forEach { line ->
            assertTrue("TvRemotePlaybackScreen 必须包含 $line", source.contains(line))
        }
    }
}
