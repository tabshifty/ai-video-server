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

    @Test
    fun followScrollOnlyRepositionsWhenFocusedEpisodeNearViewportEdge() {
        assertEquals(
            1,
            resolveEpisodeRailFollowScrollFirstVisibleItemIndex(
                items = items,
                focusedEpisodeId = "e5",
                firstVisibleItemIndex = 0,
                lastVisibleItemIndex = 4,
            ),
        )
        assertEquals(
            0,
            resolveEpisodeRailFollowScrollFirstVisibleItemIndex(
                items = items,
                focusedEpisodeId = "e1",
                firstVisibleItemIndex = 1,
                lastVisibleItemIndex = 5,
            ),
        )
        assertEquals(
            null,
            resolveEpisodeRailFollowScrollFirstVisibleItemIndex(
                items = items,
                focusedEpisodeId = "e3",
                firstVisibleItemIndex = 0,
                lastVisibleItemIndex = 4,
            ),
        )
    }

    @Test
    fun episodeRailLabelUsesChineseOrdinals() {
        assertEquals("第一集", formatTvEpisodeRailLabel(1))
        assertEquals("第十集", formatTvEpisodeRailLabel(10))
        assertEquals("第十一集", formatTvEpisodeRailLabel(11))
        assertEquals("第二十集", formatTvEpisodeRailLabel(20))
        assertEquals("第一百零二集", formatTvEpisodeRailLabel(102))
    }
}
