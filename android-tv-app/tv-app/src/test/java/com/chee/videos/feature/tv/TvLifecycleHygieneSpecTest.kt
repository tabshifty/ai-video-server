package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * TV 端生命周期卫生审计：锁住内存泄漏修复的关键形态，防止后续改动回退。
 */
class TvLifecycleHygieneSpecTest {

    @Test
    fun pairingPollIsGatedByForegroundLifecycle() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvPairingScreen.kt").readText()

        assertTrue(
            "配对轮询必须暴露 setPollingAllowed 前台门控入口",
            source.contains("fun setPollingAllowed(allowed: Boolean)"),
        )
        assertTrue(
            "配对轮询循环必须在请求前挂起等待前台放行（pollingAllowed.first { it }）",
            source.contains("pollingAllowed.first { it }"),
        )
        assertTrue(
            "配对页必须监听 ON_START/ON_STOP 切换轮询放行，与 TvShellApp 的协调器门控方式一致",
            source.contains("Lifecycle.Event.ON_START -> viewModel.setPollingAllowed(true)") &&
                source.contains("Lifecycle.Event.ON_STOP -> viewModel.setPollingAllowed(false)"),
        )
        assertTrue(
            "配对页离开组合时必须恢复放行并移除生命周期观察者，避免观察者泄漏",
            source.contains("lifecycleOwner.lifecycle.removeObserver(observer)"),
        )
    }

}
