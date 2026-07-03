package com.chee.videos.feature.shorts

import com.chee.videos.core.model.FeedVideoDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortFeedPagerRestoreTest {

    @Test
    fun returnsAnchorIndexWhenAnchorExists() {
        val items = listOf(feedItem("a"), feedItem("b"), feedItem("c"))

        val page = resolveShortFeedInitialPage(
            incomingItems = items,
            anchorVideoId = "b",
            fallbackPage = 0,
        )

        assertEquals(1, page)
    }

    @Test
    fun fallsBackToPreviousPageWhenAnchorMissing() {
        val items = listOf(feedItem("a"), feedItem("b"), feedItem("c"))

        val page = resolveShortFeedInitialPage(
            incomingItems = items,
            anchorVideoId = "x",
            fallbackPage = 2,
        )

        assertEquals(2, page)
    }

    @Test
    fun clampsFallbackPageToLastIndex() {
        val items = listOf(feedItem("a"), feedItem("b"), feedItem("c"))

        val page = resolveShortFeedInitialPage(
            incomingItems = items,
            anchorVideoId = null,
            fallbackPage = 99,
        )

        assertEquals(2, page)
    }

    @Test
    fun pendingDeleteRemovalSelectsNextAvailableVideo() {
        val items = listOf(feedItem("a"), feedItem("b"), feedItem("c"))

        val window = resolveShortFeedPendingDeleteWindow(
            items = items,
            removedVideoId = "b",
        )

        assertTrue(window.removed)
        assertEquals(listOf("a", "c"), window.items.map { it.id })
        assertEquals(1, window.pagerInitialPage)
        assertEquals("c", window.pagerAnchorVideoId)
    }

    @Test
    fun pendingDeleteRemovalFallsBackToPreviousWhenLastVideoRemoved() {
        val items = listOf(feedItem("a"), feedItem("b"))

        val window = resolveShortFeedPendingDeleteWindow(
            items = items,
            removedVideoId = "b",
        )

        assertEquals(listOf("a"), window.items.map { it.id })
        assertEquals(0, window.pagerInitialPage)
        assertEquals("a", window.pagerAnchorVideoId)
    }

    private fun feedItem(id: String): FeedVideoDto {
        return FeedVideoDto(
            id = id,
            title = id,
            type = "short",
        )
    }
}
