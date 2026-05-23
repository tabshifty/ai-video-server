package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvResumePromptTest {
    @Test
    fun shouldShowResumePromptCard_requiresAllowedStateAndPositiveRemainingTime() {
        val allowed = resumePromptGuardInput()

        assertTrue(shouldShowResumePromptCard(allowed))
        assertFalse(shouldShowResumePromptCard(allowed.copy(hasResumeSeekTriggered = false)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(promptPermanentlyDismissed = true)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(isPlayerError = true)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(isBackConfirmVisible = true)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(isEpisodeSelectorVisible = true)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(isEndOverlayVisible = true)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(isAutoplayPromptVisible = true)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(isPausedByUser = true)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(remainingMs = 0L)))
        assertFalse(shouldShowResumePromptCard(allowed.copy(remainingMs = -1L)))
    }

    @Test
    fun shouldTickResumePromptCountdown_ignoresRemainingTimeButHonorsGuards() {
        val allowed = resumePromptGuardInput()

        assertTrue(shouldTickResumePromptCountdown(allowed))
        assertTrue(shouldTickResumePromptCountdown(allowed.copy(remainingMs = 0L)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(hasResumeSeekTriggered = false)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(promptPermanentlyDismissed = true)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(isPlayerError = true)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(isBackConfirmVisible = true)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(isEpisodeSelectorVisible = true)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(isEndOverlayVisible = true)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(isAutoplayPromptVisible = true)))
        assertFalse(shouldTickResumePromptCountdown(allowed.copy(isPausedByUser = true)))
    }

    @Test
    fun resumePromptCountdownTickRemaining_usesCeilingSecondsWithinFiveSecondWindow() {
        assertEquals(5, resumePromptCountdownTickRemaining(5_000L))
        assertEquals(5, resumePromptCountdownTickRemaining(4_999L))
        assertEquals(5, resumePromptCountdownTickRemaining(4_001L))
        assertEquals(4, resumePromptCountdownTickRemaining(4_000L))
        assertEquals(2, resumePromptCountdownTickRemaining(1_001L))
        assertEquals(1, resumePromptCountdownTickRemaining(1_000L))
        assertEquals(1, resumePromptCountdownTickRemaining(1L))
        assertEquals(0, resumePromptCountdownTickRemaining(0L))
        assertEquals(0, resumePromptCountdownTickRemaining(-100L))
    }

    @Test
    fun shouldTriggerResumePrompt_requiresAtLeastTenSecondsOfHistoryPosition() {
        assertFalse(shouldTriggerResumePrompt(9_999L))
        assertTrue(shouldTriggerResumePrompt(10_000L))
        assertTrue(shouldTriggerResumePrompt(10_001L))
        assertTrue(shouldTriggerResumePrompt(Long.MAX_VALUE))
        assertFalse(shouldTriggerResumePrompt(0L))
        assertFalse(shouldTriggerResumePrompt(-1L))
    }

    @Test
    fun formatResumePromptTimestamp_usesMinuteSecondOrHourMinuteSecond() {
        assertEquals("0:00", formatResumePromptTimestamp(0L))
        assertEquals("12:34", formatResumePromptTimestamp(754_000L))
        assertEquals("1:23:45", formatResumePromptTimestamp(5_025_000L))
        assertEquals("0:00", formatResumePromptTimestamp(-100L))
        assertEquals("59:59", formatResumePromptTimestamp(3_599_000L))
        assertEquals("1:00:00", formatResumePromptTimestamp(3_600_000L))
    }
}

private fun resumePromptGuardInput(): ResumePromptGuardInput = ResumePromptGuardInput(
    hasResumeSeekTriggered = true,
    promptPermanentlyDismissed = false,
    isPlayerError = false,
    isBackConfirmVisible = false,
    isEpisodeSelectorVisible = false,
    isEndOverlayVisible = false,
    isAutoplayPromptVisible = false,
    isPausedByUser = false,
    remainingMs = 5_000L,
)
