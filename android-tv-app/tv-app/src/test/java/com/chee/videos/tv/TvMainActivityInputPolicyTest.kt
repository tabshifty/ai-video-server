package com.chee.videos.tv

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvMainActivityInputPolicyTest {
    @Test
    fun `swallows only compose hover exit uncleared crash`() {
        val throwable = IllegalStateException("The ACTION_HOVER_EXIT event was not cleared.").apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.platform.AndroidComposeView",
                    "sendHoverExitEvent",
                    "AndroidComposeView.android.kt",
                    565,
                ),
            )
        }

        assertTrue(shouldSwallowTvComposeHoverExitCrash(throwable))
    }

    @Test
    fun `does not swallow unrelated illegal state exception`() {
        val throwable = IllegalStateException("Other input failure").apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.platform.AndroidComposeView",
                    "dispatchHoverEvent",
                    "AndroidComposeView.android.kt",
                    1808,
                ),
            )
        }

        assertFalse(shouldSwallowTvComposeHoverExitCrash(throwable))
    }

    @Test
    fun `does not swallow matching message from non compose source`() {
        val throwable = IllegalStateException("The ACTION_HOVER_EXIT event was not cleared.").apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "com.chee.videos.tv.TvMainActivity",
                    "dispatchGenericMotionEvent",
                    "TvMainActivity.kt",
                    1,
                ),
            )
        }

        assertFalse(shouldSwallowTvComposeHoverExitCrash(throwable))
    }

    @Test
    fun `swallows compose hover exit from main looper lambda`() {
        val throwable = IllegalStateException("The ACTION_HOVER_EXIT event was not cleared.").apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.platform.AndroidComposeView",
                    "sendHoverExitEvent\$lambda\$5",
                    "AndroidComposeView.android.kt",
                    565,
                ),
                StackTraceElement(
                    "androidx.compose.ui.platform.AndroidComposeView\$\$ExternalSyntheticLambda3",
                    "run",
                    "D8\$\$SyntheticClass",
                    0,
                ),
                StackTraceElement(
                    "android.os.Handler",
                    "handleCallback",
                    "Handler.java",
                    883,
                ),
                StackTraceElement(
                    "android.os.Looper",
                    "loop",
                    "Looper.java",
                    214,
                ),
            )
        }

        assertTrue(shouldSwallowTvComposeHoverExitCrash(throwable))
    }

    @Test
    fun `swallows compose dispatch hover lambda variant`() {
        val throwable = IllegalStateException("The ACTION_HOVER_EXIT event was not cleared.").apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.platform.AndroidComposeView",
                    "dispatchHoverEvent\$lambda\$0",
                    "AndroidComposeView.android.kt",
                    1808,
                ),
            )
        }

        assertTrue(shouldSwallowTvComposeHoverExitCrash(throwable))
    }
}
