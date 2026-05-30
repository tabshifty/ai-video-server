package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Test

class TvEpisodeRailPolicyTest {
    private val items = (1..8).map { index ->
        TvEpisodeRailItem(
            id = "e$index",
            number = index,
            title = "第${index}集",
            playable = index != 5,
            current = index == 4,
        )
    }

    @Test
    fun currentIndexUsesCurrentEpisodeIdAndFallsBackToZero() {
        assertEquals(3, resolveEpisodeRailCurrentIndex(items, currentEpisodeId = "e4"))
        assertEquals(0, resolveEpisodeRailCurrentIndex(items, currentEpisodeId = "missing"))
        assertEquals(0, resolveEpisodeRailCurrentIndex(emptyList(), currentEpisodeId = "missing"))
    }

    @Test
    fun initialFirstVisibleItemIndexKeepsCurrentEpisodeNearMiddle() {
        assertEquals(0, resolveEpisodeRailInitialFirstVisibleItemIndex(items, currentEpisodeId = "e2"))
        assertEquals(1, resolveEpisodeRailInitialFirstVisibleItemIndex(items, currentEpisodeId = "e5"))
        assertEquals(0, resolveEpisodeRailInitialFirstVisibleItemIndex(items, currentEpisodeId = "missing"))
    }
}
