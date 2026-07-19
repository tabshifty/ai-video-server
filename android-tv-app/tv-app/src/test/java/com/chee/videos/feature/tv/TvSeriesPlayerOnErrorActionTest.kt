package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesPlayerOnErrorActionTest {

    @Test
    fun `首帧未现时返回硬错误，沿用原回退行为`() {
        val action = resolveSeriesOnErrorAction(
            hasRenderedFirstFrame = false,
            errorMessage = "解码失败",
        )
        assertEquals(SeriesOnErrorAction.HardError, action)
    }

    @Test
    fun `首帧已现且带错误信息时返回软重试，保留已渲染画面`() {
        val action = resolveSeriesOnErrorAction(
            hasRenderedFirstFrame = true,
            errorMessage = "网络中断",
        )
        assertTrue(action is SeriesOnErrorAction.SoftRetry)
        assertEquals("网络中断", (action as SeriesOnErrorAction.SoftRetry).message)
    }

    @Test
    fun `软重试非硬错误，触发器不应出现 hasStartedPlayback 被清的硬错误分支`() {
        val action = resolveSeriesOnErrorAction(
            hasRenderedFirstFrame = true,
            errorMessage = "网络中断",
        )
        assertFalse(action is SeriesOnErrorAction.HardError)
    }

    @Test
    fun `首帧已现但错误信息为空时仍按硬错误回退，避免空提示`() {
        val action = resolveSeriesOnErrorAction(
            hasRenderedFirstFrame = true,
            errorMessage = "",
        )
        assertEquals(SeriesOnErrorAction.HardError, action)
    }
}
