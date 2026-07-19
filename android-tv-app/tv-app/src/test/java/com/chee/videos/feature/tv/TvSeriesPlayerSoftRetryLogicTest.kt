package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvSeriesPlayerSoftRetryLogicTest {

    @Test
    fun `preparing 优先取消当前重试`() {
        assertEquals(
            SeriesSoftRetryBackAction.CancelPreparing,
            resolveSeriesSoftRetryBackAction(TvLongFormSoftRetryUiState.Preparing(1)),
        )
    }

    @Test
    fun `failed 优先关闭软失败态`() {
        assertEquals(
            SeriesSoftRetryBackAction.DismissFailure,
            resolveSeriesSoftRetryBackAction(TvLongFormSoftRetryUiState.Failed(1, "网络中断")),
        )
    }

    @Test
    fun `瞬态或空状态交回播放器原返回链路`() {
        assertEquals(
            SeriesSoftRetryBackAction.DelegateToPlayerBack,
            resolveSeriesSoftRetryBackAction(TvLongFormSoftRetryUiState.Succeeded(1, "已恢复播放")),
        )
        assertEquals(
            SeriesSoftRetryBackAction.DelegateToPlayerBack,
            resolveSeriesSoftRetryBackAction(TvLongFormSoftRetryUiState.Canceled(1, "已取消重试")),
        )
        assertEquals(
            SeriesSoftRetryBackAction.DelegateToPlayerBack,
            resolveSeriesSoftRetryBackAction(null),
        )
    }
}
