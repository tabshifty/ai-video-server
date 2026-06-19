package com.chee.videos.feature.tv

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormPlayerSoftRetryLogicTest {
    @Test
    fun `soft retry feedback only shows after first frame and outside diagnostics`() {
        assertFalse(
            shouldUseTvLongFormSoftRetryFeedback(
                hasRenderedFirstFrame = false,
                softRetryUiState = TvLongFormSoftRetryUiState.Preparing(1),
                showDiagnostics = false,
            ),
        )
        assertFalse(
            shouldUseTvLongFormSoftRetryFeedback(
                hasRenderedFirstFrame = true,
                softRetryUiState = TvLongFormSoftRetryUiState.Preparing(1),
                showDiagnostics = true,
            ),
        )
        assertTrue(
            shouldUseTvLongFormSoftRetryFeedback(
                hasRenderedFirstFrame = true,
                softRetryUiState = TvLongFormSoftRetryUiState.Preparing(1),
                showDiagnostics = false,
            ),
        )
    }

    @Test
    fun `ignored retry error only matches latest canceled retry key`() {
        assertTrue(shouldIgnoreTvLongFormRetryError(ignoredRetryAttemptKey = 3, routeRetryNonce = 3))
        assertFalse(shouldIgnoreTvLongFormRetryError(ignoredRetryAttemptKey = 2, routeRetryNonce = 3))
        assertFalse(shouldIgnoreTvLongFormRetryError(ignoredRetryAttemptKey = null, routeRetryNonce = 3))
    }
}
