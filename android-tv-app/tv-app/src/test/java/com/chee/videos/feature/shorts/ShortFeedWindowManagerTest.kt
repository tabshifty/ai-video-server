package com.chee.videos.feature.shorts

import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.UserStateDto
import com.chee.videos.core.model.VideoDetailDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortFeedWindowManagerTest {

    @Test
    fun `should load more when current page reaches tail threshold`() {
        assertTrue(
            ShortFeedWindowManager.shouldLoadMore(
                currentIndex = 15,
                totalCount = 20,
                loadingMore = false,
            ),
        )
        assertFalse(
            ShortFeedWindowManager.shouldLoadMore(
                currentIndex = 10,
                totalCount = 20,
                loadingMore = false,
            ),
        )
    }

    @Test
    fun `merge append trims head and keeps anchor caches`() {
        val items = (0 until 60).map { feedItem(it) }
        val incoming = (60 until 80).map { feedItem(it) }
        val state = ShortFeedWindowSnapshot(
            items = items,
            detailByVideoId = mapOf(
                "id-10" to detail("id-10"),
                "id-55" to detail("id-55"),
            ),
            detailLoadingVideoIds = setOf("id-10", "id-55"),
            actionBusyVideoIds = setOf("id-10", "id-55"),
            pausedByUserVideoIds = setOf("id-10", "id-55"),
            detailSheetVideoId = "id-10",
        )

        val merged = ShortFeedWindowManager.mergeAppend(
            snapshot = state,
            incoming = incoming,
            anchorVideoId = "id-55",
        )

        assertEquals(20, merged.trimmedHeadCount)
        assertEquals(60, merged.items.size)
        assertEquals("id-20", merged.items.first().id)
        assertEquals("id-79", merged.items.last().id)
        assertEquals(35, merged.anchorPageAfterTrim)
        assertFalse(merged.detailByVideoId.containsKey("id-10"))
        assertTrue(merged.detailByVideoId.containsKey("id-55"))
        assertFalse("id-10" in merged.pausedByUserVideoIds)
        assertTrue("id-55" in merged.pausedByUserVideoIds)
        assertNull(merged.detailSheetVideoId)
    }

    @Test
    fun `merge append falls back to repeated batch when incoming only duplicates`() {
        val items = (0 until 20).map { feedItem(it) }
        val incoming = listOf(feedItem(1), feedItem(2), feedItem(2), feedItem(3))

        val merged = ShortFeedWindowManager.mergeAppend(
            snapshot = ShortFeedWindowSnapshot(items = items),
            incoming = incoming,
            anchorVideoId = "id-19",
        )

        assertEquals(23, merged.items.size)
        assertEquals(listOf("id-1", "id-2", "id-3"), merged.items.takeLast(3).map { it.id })
    }

    private fun feedItem(index: Int): FeedVideoDto {
        return FeedVideoDto(
            id = "id-$index",
            title = "title-$index",
            type = "short",
        )
    }

    private fun detail(id: String): VideoDetailDto {
        return VideoDetailDto(
            id = id,
            title = id,
            userState = UserStateDto(),
        )
    }
}
