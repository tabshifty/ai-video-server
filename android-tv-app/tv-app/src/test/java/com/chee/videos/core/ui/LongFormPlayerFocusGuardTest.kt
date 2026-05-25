package com.chee.videos.core.ui

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class LongFormPlayerFocusGuardTest {

    private val noOverlay = PlayerFocusGuardInput(
        controlsVisible = false,
        subtitleSheetVisible = false,
        audioTrackSheetVisible = false,
        resumePromptVisible = false,
        backConfirmPromptVisible = false,
        playerErrorVisible = false,
    )

    @Test
    fun anyOverlayVisible_falseWhenAllFalse() {
        assertFalse(noOverlay.anyOverlayVisible())
    }

    @Test
    fun anyOverlayVisible_trueWhenAnyOverlayFlagSet() {
        assertTrue(noOverlay.copy(subtitleSheetVisible = true).anyOverlayVisible())
        assertTrue(noOverlay.copy(audioTrackSheetVisible = true).anyOverlayVisible())
        assertTrue(noOverlay.copy(resumePromptVisible = true).anyOverlayVisible())
        assertTrue(noOverlay.copy(backConfirmPromptVisible = true).anyOverlayVisible())
        assertTrue(noOverlay.copy(playerErrorVisible = true).anyOverlayVisible())
    }

    @Test
    fun anyOverlayVisible_ignoresControlsVisible() {
        // controlsVisible 是播放器自身可见性，不是叠加层
        assertFalse(noOverlay.copy(controlsVisible = true).anyOverlayVisible())
    }

    @Test
    fun reclaim_resumePromptDismiss_withNoOtherOverlay() {
        val previous = noOverlay.copy(resumePromptVisible = true)
        val current = noOverlay
        assertTrue(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_resumePromptDismiss_withOtherOverlayStillVisible_doesNotReclaim() {
        val previous = noOverlay.copy(resumePromptVisible = true, subtitleSheetVisible = true)
        val current = noOverlay.copy(subtitleSheetVisible = true)
        assertFalse(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_subtitleSheetDismiss_withNoOtherOverlay() {
        val previous = noOverlay.copy(subtitleSheetVisible = true)
        val current = noOverlay
        assertTrue(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_audioSheetDismiss_withResumePromptStillVisible_doesNotReclaim() {
        val previous = noOverlay.copy(audioTrackSheetVisible = true, resumePromptVisible = true)
        val current = noOverlay.copy(resumePromptVisible = true)
        assertFalse(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_backConfirmPromptDismiss_withNoOtherOverlay() {
        val previous = noOverlay.copy(backConfirmPromptVisible = true)
        val current = noOverlay
        assertTrue(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_playerErrorDismiss_withNoOtherOverlay() {
        val previous = noOverlay.copy(playerErrorVisible = true)
        val current = noOverlay
        assertTrue(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_controlsAutoHide_withNoOverlay() {
        // 5 秒 auto-hide 同样应触发兜底（防 B2 / B5）
        val previous = noOverlay.copy(controlsVisible = true)
        val current = noOverlay
        assertTrue(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_noTransition_returnsFalse() {
        // input 完全没变 → 不重请求
        assertFalse(shouldReclaimRootFocus(noOverlay, noOverlay))

        val visible = noOverlay.copy(subtitleSheetVisible = true)
        assertFalse(shouldReclaimRootFocus(visible, visible))
    }

    @Test
    fun reclaim_overlayOpens_returnsFalse() {
        // 反向跃迁（false→true，新打开 overlay）→ 不抢它的焦点
        assertFalse(shouldReclaimRootFocus(noOverlay, noOverlay.copy(subtitleSheetVisible = true)))
        assertFalse(shouldReclaimRootFocus(noOverlay, noOverlay.copy(resumePromptVisible = true)))
    }

    @Test
    fun reclaim_multipleOverlaysAllClose_returnsTrueOnce() {
        // 字幕 + 音轨 picker 同时显示，同时关闭 → 触发一次 reclaim
        val previous = noOverlay.copy(subtitleSheetVisible = true, audioTrackSheetVisible = true)
        val current = noOverlay
        assertTrue(shouldReclaimRootFocus(previous, current))
    }

    @Test
    fun reclaim_resumePromptCloseAndSubtitleOpensSimultaneously_doesNotReclaim() {
        // 续播卡消失同时打开字幕 picker：focus 已经交给字幕 picker，不应抢回
        val previous = noOverlay.copy(resumePromptVisible = true)
        val current = noOverlay.copy(subtitleSheetVisible = true)
        assertFalse(shouldReclaimRootFocus(previous, current))
    }
}
