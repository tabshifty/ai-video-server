package com.chee.videos.core.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.withFrameNanos
import androidx.compose.ui.focus.FocusRequester
import kotlin.coroutines.cancellation.CancellationException

private const val FOCUS_REQUESTER_NOT_INITIALIZED_MARKER =
    "FocusRequester is not initialized"

internal fun isFocusRequesterNotInitialized(err: Throwable): Boolean {
    if (err !is IllegalStateException) return false
    val message = err.message ?: return false
    return message.contains(FOCUS_REQUESTER_NOT_INITIALIZED_MARKER)
}

/**
 * 在 LaunchedTvInitialFocus 等首屏焦点协程里调用 requestFocus() 时使用，
 * 单调用点吃掉「FocusRequester is not initialized」未挂载异常，
 * 防止 Compose 1.7 的协程恢复路径绕开外层 runCatching 把 ISE 透到主 Looper。
 * 真正未关心的 ISE 仍会原样抛出。
 */
fun FocusRequester.tryRequestFocus(): Boolean {
    return try {
        requestFocus()
        true
    } catch (err: IllegalStateException) {
        if (isFocusRequesterNotInitialized(err)) {
            false
        } else {
            throw err
        }
    }
}

@Composable
fun LaunchedTvInitialFocus(
    vararg keys: Any?,
    block: suspend () -> Unit,
) {
    LaunchedEffect(*keys) {
        withFrameNanos { }
        runCatching { block() }
            .onFailure { err ->
                if (err is CancellationException) throw err
                if (isFocusRequesterNotInitialized(err)) {
                    return@onFailure
                }
                throw err
            }
    }
}
