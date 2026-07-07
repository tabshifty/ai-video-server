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

    @Test
    fun remoteControlScreenUsesMediaRemoteLayoutAndAutoplayToggle() {
        val source = Path.of("src/main/java/com/chee/videos/feature/shortsearch/ShortSearchRemoteControlScreen.kt").readText()

        listOf(
            "自动播放下一条",
            "Switch(",
            "viewModel::setAutoplayNextEnabled",
            "Brush.verticalGradient",
            "TvRemoteControlStatusPill(",
            "TvRemoteControlActionButton(",
            "当前条结束后自动切到下一条",
        ).forEach { line ->
            assertTrue("手机端投放控制页必须包含 $line", source.contains(line))
        }
    }
}
