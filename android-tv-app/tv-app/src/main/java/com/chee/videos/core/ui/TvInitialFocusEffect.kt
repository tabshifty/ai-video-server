package com.chee.videos.core.ui

import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.withFrameNanos
import kotlin.coroutines.cancellation.CancellationException

private const val FOCUS_REQUESTER_NOT_INITIALIZED_PREFIX =
    "FocusRequester is not initialized"

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
                if (err is IllegalStateException &&
                    err.message?.startsWith(FOCUS_REQUESTER_NOT_INITIALIZED_PREFIX) == true
                ) {
                    return@onFailure
                }
                throw err
            }
    }
}
