package com.chee.videos.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRemoteCoordinatorLifecycleSpecTest {
    @Test
    fun shellAppStartsAndStopsCoordinatorPollingWithLifecycle() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvShellApp.kt").readText()

        listOf(
            "Lifecycle.Event.ON_START -> {",
            "remoteCoordinatorViewModel.setPollingEnabled(true)",
            "remoteCoordinatorViewModel.refreshNow()",
            "Lifecycle.Event.ON_STOP -> remoteCoordinatorViewModel.setPollingEnabled(false)",
            "onSessionEnded = remoteCoordinatorViewModel::dismissSession",
        ).forEach { line ->
            assertTrue("TvShellApp 必须包含 $line", source.contains(line))
        }
    }

    @Test
    fun coordinatorDismissSessionKeepsRetryingEndRequestUntilSuccess() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvRemoteCoordinatorViewModel.kt").readText()

        listOf(
            "dismissedSessionIds += normalizedSessionId",
            "ensureDismissJob(normalizedSessionId)",
            "dismissJobs[sessionId] = viewModelScope.launch {",
            "while (true) {",
            "val ended = repository.endTvRemoteSession(sessionId).isSuccess",
            "dismissedSessionIds.remove(sessionId)",
            "dismissJobs.remove(sessionId)",
            "delay(1_000L)",
        ).forEach { line ->
            assertTrue("TvRemoteCoordinatorViewModel 必须包含 $line", source.contains(line))
        }
    }
}
