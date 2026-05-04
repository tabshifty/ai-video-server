package com.chee.videos.feature.shortsearch

import com.chee.videos.core.model.VideoListItemDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortSearchViewModelStateTest {
    @Test
    fun normalizeShortSearchQuery_trimsWhitespace() {
        assertEquals("标签", normalizeShortSearchQuery("  标签  "))
        assertEquals("", normalizeShortSearchQuery("   "))
    }

    @Test
    fun resetShortSearchForQuery_clearsPagingAndErrors() {
        val state = ShortSearchUiState(
            activeQuery = "旧词",
            loading = false,
            loaded = true,
            page = 3,
            totalCount = 99,
            items = listOf(VideoListItemDto(id = "a", title = "A", type = "short")),
            errorMessage = "old",
        )

        val next = resetShortSearchForQuery(state, "新词")

        assertEquals("新词", next.activeQuery)
        assertTrue(next.loading)
        assertFalse(next.loaded)
        assertEquals(0, next.page)
        assertEquals(0, next.totalCount)
        assertTrue(next.items.isEmpty())
        assertNull(next.errorMessage)
    }

    @Test
    fun mergeShortSearchItems_deduplicatesById() {
        val merged = mergeShortSearchItems(
            existing = listOf(VideoListItemDto(id = "a", title = "A", type = "short")),
            incoming = listOf(
                VideoListItemDto(id = "a", title = "A2", type = "short"),
                VideoListItemDto(id = "b", title = "B", type = "short"),
            ),
        )

        assertEquals(listOf("a", "b"), merged.map { it.id })
    }
}
