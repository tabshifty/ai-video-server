package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvInitialFocusEffectShapeTest {

    private val sourcePath: Path =
        Path.of("src/main/java/com/chee/videos/core/ui/TvInitialFocusEffect.kt")

    private val source: String by lazy {
        assertTrue("TV 初始焦点 helper 源文件必须存在", sourcePath.exists())
        sourcePath.readText()
    }

    @Test
    fun `helper waits for a compose frame before requesting focus`() {
        assertTrue(
            "TV 初始焦点 helper 必须先等过一帧再调用 requestFocus，避免 LazyColumn 延迟组合下 FocusRequester 未挂载",
            source.contains("withFrameNanos"),
        )
    }

    @Test
    fun `helper guards focus request with runCatching`() {
        assertTrue(
            "TV 初始焦点 helper 必须用 runCatching 包裹 block，集中过滤 FocusRequester 未初始化异常",
            source.contains("runCatching { block() }"),
        )
    }

    @Test
    fun `helper rethrows cancellation exception to preserve coroutine semantics`() {
        assertTrue(
            "TV 初始焦点 helper 必须重抛 CancellationException，避免吞掉协程取消信号",
            source.contains("if (err is CancellationException) throw err"),
        )
    }

    @Test
    fun `helper filters only focus requester not initialized message`() {
        assertTrue(
            "TV 初始焦点 helper 必须按消息前缀精确过滤，而不是吞掉所有 IllegalStateException",
            source.contains("\"FocusRequester is not initialized\"") ||
                source.contains("FOCUS_REQUESTER_NOT_INITIALIZED_PREFIX"),
        )
        assertTrue(
            "TV 初始焦点 helper 必须使用 startsWith 前缀匹配以兼容消息追加细节",
            source.contains("startsWith(FOCUS_REQUESTER_NOT_INITIALIZED_PREFIX)") ||
                source.contains("startsWith(\"FocusRequester is not initialized\")"),
        )
    }

    @Test
    fun `helper rethrows all other exceptions`() {
        assertTrue(
            "TV 初始焦点 helper 不允许吞掉非 FocusRequester 未初始化异常",
            source.contains("throw err"),
        )
    }

    @Test
    fun `helper is exposed as composable launchable from caller sites`() {
        assertTrue(
            "helper 必须以 @Composable 形式暴露，便于业务屏幕直接替换 LaunchedEffect",
            source.contains("@Composable") && source.contains("fun LaunchedTvInitialFocus"),
        )
        assertTrue(
            "helper 必须接受可变 keys 数组以匹配 LaunchedEffect 的失效语义",
            source.contains("vararg keys: Any?"),
        )
    }
}
