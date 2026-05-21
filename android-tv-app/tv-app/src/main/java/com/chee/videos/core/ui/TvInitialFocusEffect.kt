package com.chee.videos.core.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.withFrameNanos
import kotlin.coroutines.cancellation.CancellationException

private const val FOCUS_REQUESTER_NOT_INITIALIZED_MARKER =
    "FocusRequester is not initialized"

internal fun isFocusRequesterNotInitialized(err: Throwable): Boolean {
    if (err !is IllegalStateException) return false
    val message = err.message ?: return false
    return message.contains(FOCUS_REQUESTER_NOT_INITIALIZED_MARKER)
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
