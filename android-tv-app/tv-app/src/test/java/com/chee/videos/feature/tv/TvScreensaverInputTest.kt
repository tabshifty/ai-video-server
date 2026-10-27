package com.chee.videos.feature.tv

import org.junit.Assert.*
import org.junit.Test
import java.util.Base64

class TvScreensaverInputTest {
    @Test
    fun tokenRefreshKeepsIdentityButAccountAndServerChangesDoNot() {
        fun token(uid: String, nonce: Int) = "header." + Base64.getUrlEncoder().withoutPadding()
            .encodeToString("{\"uid\":\"$uid\",\"iat\":$nonce}".toByteArray()) + ".signature"
        assertEquals(screensaverSessionKey("https://first.example", token("甲", 1)),
            screensaverSessionKey("https://first.example/", token("甲", 2)))
        assertNotEquals(screensaverSessionKey("https://first.example", token("甲", 1)),
            screensaverSessionKey("https://first.example", token("乙", 1)))
        assertNotEquals(screensaverSessionKey("https://first.example", token("甲", 1)),
            screensaverSessionKey("https://second.example", token("甲", 1)))
        assertNull(screensaverSessionKey("https://first.example", null))
    }

    @Test
    fun wakeKeyConsumesRepeatAndUpThenAllowsNextCompletePress() {
        val gate = TvScreensaverInputGate()
        assertTrue(gate.key(23, down = true, showing = true, systemKey = false))
        repeat(5) { assertTrue(gate.key(23, down = true, showing = false, systemKey = false)) }
        assertTrue(gate.key(23, down = false, showing = false, systemKey = false))
        assertFalse(gate.key(23, down = true, showing = false, systemKey = false))
        assertFalse(gate.key(23, down = false, showing = false, systemKey = false))
    }

    @Test
    fun systemKeysPassThroughAndWakeTouchConsumesEntireGesture() {
        val gate = TvScreensaverInputGate()
        assertFalse(gate.key(24, down = true, showing = true, systemKey = true))
        assertFalse(gate.key(24, down = false, showing = true, systemKey = true))
        assertTrue(gate.touch(start = true, end = false, showing = true))
        assertTrue(gate.touch(start = false, end = false, showing = false))
        assertTrue(gate.touch(start = false, end = true, showing = false))
        assertFalse(gate.touch(start = true, end = false, showing = false))
        // 原页面收到 DOWN 后长时间无移动，屏保出现；UP 仍只能唤醒。
        assertTrue(gate.touch(start = false, end = true, showing = true))
        assertFalse(gate.touch(start = true, end = false, showing = false))
    }

    @Test
    fun playbackOwnersCannotReleaseEachOthersBusyStateAndWakeIsSynchronous() {
        val controller = TvScreensaverController()
        val first = Any()
        val second = Any()
        controller.setPlaybackBusy(first, true)
        controller.setPlaybackBusy(second, true)
        controller.setPlaybackBusy(first, false)
        assertTrue(controller.runtime.value.playbackBusy)
        controller.setPlaybackBusy(second, false)
        assertFalse(controller.runtime.value.playbackBusy)
        controller.setShowing(true)
        val sequence = controller.runtime.value.interactionSequence
        controller.userActivity()
        assertFalse(controller.showing.value)
        assertEquals(sequence + 1, controller.runtime.value.interactionSequence)
    }
}
