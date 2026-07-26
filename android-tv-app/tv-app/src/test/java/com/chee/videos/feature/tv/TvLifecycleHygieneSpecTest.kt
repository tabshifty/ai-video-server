package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * TV 端生命周期卫生审计：锁住内存泄漏修复的关键形态，防止后续改动回退。
 */
class TvLifecycleHygieneSpecTest {

    @Test
    fun iptvHandlerPostsAreGuardedAndClearedOnDispose() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt").readText()

        assertTrue(
            "IPTV 退出页面必须清空主线程 Handler 消息队列，防止迟到的 LibVLC 事件回调触碰已释放播放器",
            source.contains("mainHandler.removeCallbacksAndMessages(null)"),
        )
        assertTrue(
            "IPTV 事件回调必须带 released 守卫：post 到主线程的 runnable 在 release() 后必须直接返回",
            source.contains("var released = false") && source.contains("released = true"),
        )
        val guardIndex = source.indexOf("if (released)")
        val listenerIndex = source.indexOf("MediaPlayer.EventListener")
        assertTrue(
            "released 守卫必须位于事件回调 post 的 runnable 内部（listener 声明之后）",
            listenerIndex in 0 until guardIndex,
        )
        val releaseIndex = source.indexOf("vlcPlayer.release()")
        val clearIndex = source.indexOf("mainHandler.removeCallbacksAndMessages(null)")
        assertTrue(
            "Handler 清空必须先于 vlcPlayer.release()，保证消息队列里的残留 runnable 不会命中已释放实例",
            clearIndex in 0 until releaseIndex,
        )
    }

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

    @Test
    fun iptvKeepsSingleReleaseOwnerForVlcPlayer() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt").readText()
        assertEquals(
            "vlcPlayer.release() 必须只有一个属主（外层 DisposableEffect），避免多处释放引入次序歧义",
            1,
            Regex("vlcPlayer\\.release\\(\\)").findAll(source).count(),
        )
        assertTrue(
            "视图解绑保持独立 effect（既有决策），且必须仍然存在",
            source.contains("currentPlayer.detachViews()"),
        )
    }
}
