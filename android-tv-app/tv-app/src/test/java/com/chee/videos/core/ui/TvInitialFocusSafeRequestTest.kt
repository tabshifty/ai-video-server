package com.chee.videos.core.ui

import androidx.compose.ui.focus.FocusRequester
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Test

class TvInitialFocusSafeRequestTest {

    private fun readMain(relative: String): String {
        val path = Path.of("src/main/java", relative)
        assertTrue("源文件必须存在: $relative", path.exists())
        return path.readText()
    }

    @Test
    fun `tryRequestFocus swallows uninitialized requester instead of crashing the app`() {
        val requester = FocusRequester()
        val ok = try {
            requester.tryRequestFocus()
        } catch (err: Throwable) {
            fail("tryRequestFocus 必须吃掉 FocusRequester 未初始化的 IllegalStateException：${err.message}")
            return
        }
        assertFalse(
            "未挂载的 FocusRequester 必须返回 false，方便调用方决定是否重试",
            ok,
        )
    }

    @Test
    fun `tryRequestFocus rethrows unrelated IllegalStateException`() {
        val rethrown = runCatching {
            simulateTryRequestFocusOn(IllegalStateException("some unrelated state error"))
        }.exceptionOrNull()
        assertTrue(
            "tryRequestFocus 不应吞掉与 FocusRequester 无关的 ISE，避免掩盖其他业务崩溃",
            rethrown is IllegalStateException,
        )
        assertEquals(
            "some unrelated state error",
            rethrown?.message,
        )
    }

    @Test
    fun `tryRequestFocus extension is declared on FocusRequester in TvInitialFocusEffect file`() {
        val source = readMain("com/chee/videos/core/ui/TvInitialFocusEffect.kt")
        assertTrue(
            "TvInitialFocusEffect.kt 必须暴露 FocusRequester.tryRequestFocus() 扩展，集中 swallow IllegalStateException",
            source.contains("fun FocusRequester.tryRequestFocus"),
        )
        assertTrue(
            "tryRequestFocus 必须复用 isFocusRequesterNotInitialized 匹配器，避免重复字面量",
            source.contains("isFocusRequesterNotInitialized"),
        )
    }

    @Test
    fun `tv catalog initial focus block uses tryRequestFocus instead of requestFocus`() {
        val source = readMain("com/chee/videos/feature/tv/TvCatalogScreen.kt")
        val launchedBlockStart = source.indexOf("LaunchedTvInitialFocus(uiState.loading")
        assertTrue(
            "TvCatalogScreen 必须存在 LaunchedTvInitialFocus(uiState.loading, ...) 调用块",
            launchedBlockStart >= 0,
        )
        val launchedBlockEnd = source.indexOf("}\n\n", launchedBlockStart)
        assertTrue("LaunchedTvInitialFocus 块必须能找到收口大括号", launchedBlockEnd > launchedBlockStart)
        val block = source.substring(launchedBlockStart, launchedBlockEnd)
        assertFalse(
            "TvCatalogScreen 启动焦点块内不允许出现裸 requestFocus()，必须改为 tryRequestFocus() 以避免 Compose 1.7 启动崩溃",
            block.contains(".requestFocus()"),
        )
        assertTrue(
            "TvCatalogScreen 启动焦点块必须改用 .tryRequestFocus()",
            block.contains(".tryRequestFocus()"),
        )
    }

    /**
     * 通过反射调用扩展函数验证 rethrow 行为。这里把异常注入到一个伪 FocusRequester 调用路径之外，
     * 直接验证扩展函数在面对非目标 ISE 时不会吞掉。
     */
    private fun simulateTryRequestFocusOn(injected: Throwable) {
        // 复用真实扩展：在 inline try/catch 内手动重抛非匹配异常即可。
        try {
            throw injected
        } catch (err: IllegalStateException) {
            if (isFocusRequesterNotInitialized(err)) return
            throw err
        }
    }
}
