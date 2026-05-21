package com.chee.videos.core.ui

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvInitialFocusRequesterMatcherTest {

    @Test
    fun `matches compose 1_6 raw multiline message with leading newline and indent`() {
        val realComposeMessage = """
           FocusRequester is not initialized. Here are some possible fixes:

           1. Remember the FocusRequester: val focusRequester = remember { FocusRequester() }
           2. Did you forget to add a Modifier.focusRequester() ?
           3. Are you attempting to request focus during composition? Focus requests should be made in
           response to some event. Eg Modifier.clickable { focusRequester.requestFocus() }
        """
        val err = IllegalStateException(realComposeMessage)
        assertTrue(
            "matcher 必须能识别 Compose 1.6 抛出的原始多行 message（含前导换行和缩进）",
            isFocusRequesterNotInitialized(err),
        )
    }

    @Test
    fun `matches trimmed message variant just in case compose trimIndents it`() {
        val err = IllegalStateException("FocusRequester is not initialized.")
        assertTrue(
            "matcher 也要能识别紧凑形态的同语义异常",
            isFocusRequesterNotInitialized(err),
        )
    }

    @Test
    fun `does not match other IllegalStateException messages`() {
        assertFalse(
            "matcher 不应吞 ACTION_HOVER_EXIT 这种同类型但不同语义的异常",
            isFocusRequesterNotInitialized(
                IllegalStateException("The ACTION_HOVER_EXIT event was not cleared."),
            ),
        )
        assertFalse(
            isFocusRequesterNotInitialized(IllegalStateException("some other crash")),
        )
        assertFalse(
            isFocusRequesterNotInitialized(IllegalStateException(null as String?)),
        )
    }

    @Test
    fun `does not match non IllegalStateException`() {
        assertFalse(
            "matcher 只接受 IllegalStateException，避免误吞同消息的 RuntimeException",
            isFocusRequesterNotInitialized(
                RuntimeException("FocusRequester is not initialized."),
            ),
        )
    }
}
