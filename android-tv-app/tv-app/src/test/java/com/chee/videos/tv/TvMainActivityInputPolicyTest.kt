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

    @Test
    fun `swallows compose focus requester not initialized from request focus path`() {
        val throwable = IllegalStateException(
            """
            FocusRequester is not initialized. Here are some possible fixes:
               1. Remember the FocusRequester
            """.trimIndent(),
        ).apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.focus.FocusRequester",
                    "focus\$ui_release",
                    "FocusRequester.kt",
                    259,
                ),
                StackTraceElement(
                    "androidx.compose.ui.focus.FocusRequester",
                    "requestFocus",
                    "FocusRequester.kt",
                    65,
                ),
                StackTraceElement(
                    "com.chee.videos.feature.tv.TvCatalogScreenKt\$TvCatalogScreen\$6\$1",
                    "invokeSuspend",
                    "TvCatalogScreen.kt",
                    124,
                ),
                StackTraceElement(
                    "kotlin.coroutines.jvm.internal.BaseContinuationImpl",
                    "resumeWith",
                    "ContinuationImpl.kt",
                    33,
                ),
            )
        }

        assertTrue(shouldSwallowTvComposeFocusRequesterCrash(throwable))
    }

    @Test
    fun `swallows compose focus requester crash from focus search dpad path`() {
        val throwable = IllegalStateException(
            "FocusRequester is not initialized.",
        ).apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.focus.FocusRequester",
                    "findFocusTargetNode\$ui_release",
                    "FocusRequester.kt",
                    254,
                ),
                StackTraceElement(
                    "androidx.compose.ui.focus.FocusOwnerImpl",
                    "focusSearch-ULY8qGw",
                    "FocusOwnerImpl.kt",
                    243,
                ),
            )
        }

        assertTrue(shouldSwallowTvComposeFocusRequesterCrash(throwable))
    }

    @Test
    fun `does not swallow focus requester crash with unrelated message`() {
        val throwable = IllegalStateException("Some other state failure").apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.focus.FocusRequester",
                    "requestFocus",
                    "FocusRequester.kt",
                    65,
                ),
            )
        }

        assertFalse(shouldSwallowTvComposeFocusRequesterCrash(throwable))
    }

    @Test
    fun `does not swallow focus requester message from non compose source`() {
        val throwable = IllegalStateException(
            "FocusRequester is not initialized. Caused by app bug.",
        ).apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "com.chee.videos.feature.tv.TvCatalogScreenKt",
                    "TvCatalogScreen",
                    "TvCatalogScreen.kt",
                    140,
                ),
            )
        }

        assertFalse(shouldSwallowTvComposeFocusRequesterCrash(throwable))
    }

    @Test
    fun `does not swallow hover exit message routed through focus requester matcher`() {
        val throwable = IllegalStateException("The ACTION_HOVER_EXIT event was not cleared.").apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.focus.FocusRequester",
                    "requestFocus",
                    "FocusRequester.kt",
                    65,
                ),
            )
        }

        assertFalse(shouldSwallowTvComposeFocusRequesterCrash(throwable))
    }

    @Test
    fun `swallows focus requester crash with only focus owner impl frame`() {
        val throwable = IllegalStateException(
            "FocusRequester is not initialized.",
        ).apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.focus.FocusOwnerImpl",
                    "focusSearch-ULY8qGw",
                    "FocusOwnerImpl.kt",
                    243,
                ),
                StackTraceElement(
                    "androidx.compose.ui.platform.AndroidComposeView\$keyInputModifier\$1",
                    "invoke-ZmokQxo",
                    "AndroidComposeView.android.kt",
                    342,
                ),
            )
        }

        assertTrue(shouldSwallowTvComposeFocusRequesterCrash(throwable))
    }

    @Test
    fun `does not swallow focus requester message when stack lacks compose focus package`() {
        val throwable = IllegalStateException(
            "FocusRequester is not initialized.",
        ).apply {
            stackTrace = arrayOf(
                StackTraceElement(
                    "androidx.compose.ui.platform.AndroidComposeView",
                    "dispatchKeyEvent",
                    "AndroidComposeView.android.kt",
                    945,
                ),
                StackTraceElement(
                    "com.chee.videos.feature.tv.TvCatalogScreenKt",
                    "TvCatalogScreen",
                    "TvCatalogScreen.kt",
                    140,
                ),
            )
        }

        assertFalse(shouldSwallowTvComposeFocusRequesterCrash(throwable))
    }
}
