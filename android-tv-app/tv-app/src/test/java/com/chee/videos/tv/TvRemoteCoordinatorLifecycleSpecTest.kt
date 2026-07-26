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
    fun coordinatorDismissSessionReportsEndRequestWithBoundedRetries() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvRemoteCoordinatorViewModel.kt").readText()

        listOf(
            "dismissedSessionIds += normalizedSessionId",
            "ensureDismissJob(normalizedSessionId)",
            "dismissJobs[sessionId] = viewModelScope.launch {",
            "repeat(MAX_DISMISS_REPORT_ATTEMPTS)",
            "val ended = repository.endTvRemoteSession(sessionId).isSuccess",
            "dismissedSessionIds.remove(sessionId)",
            "dismissJobs.remove(sessionId)",
            "DISMISS_RETRY_BASE_DELAY_MS",
            "DISMISS_RETRY_MAX_DELAY_MS",
            "MAX_DISMISSED_SESSION_IDS",
        ).forEach { line ->
            assertTrue("TvRemoteCoordinatorViewModel 必须包含 $line", source.contains(line))
        }

        val dismissJobBody = source.substringAfter("private fun ensureDismissJob")
        assertTrue(
            "dismiss 上报必须使用有界重试而非无限 while(true) 循环：单轮达到上限后由 5s 会话轮询在会话仍 active 时重新拉起，宏观上仍保持持续上报直到成功的语义",
            !dismissJobBody.contains("while (true)"),
        )
        assertTrue(
            "dismissedSessionIds 必须有容量上界并按插入序淘汰，防止长时间运行下无界增长",
            source.contains("while (dismissedSessionIds.size > MAX_DISMISSED_SESSION_IDS)"),
        )
    }
}
