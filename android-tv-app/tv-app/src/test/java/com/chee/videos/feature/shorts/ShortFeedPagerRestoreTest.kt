package com.chee.videos.feature.shorts

import com.chee.videos.core.model.FeedVideoDto
import org.junit.Assert.assertEquals
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

    private fun feedItem(id: String): FeedVideoDto {
        return FeedVideoDto(
            id = id,
            title = id,
            type = "short",
        )
    }
}
