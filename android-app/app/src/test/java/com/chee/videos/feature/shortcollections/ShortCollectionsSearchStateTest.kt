package com.chee.videos.feature.shortcollections

import com.chee.videos.core.model.ShortCollectionListItemDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortCollectionsSearchStateTest {
    @Test
    fun normalizeShortCollectionsQuery_trimsWhitespaceAndDropsBlank() {
        assertEquals("海边合集", normalizeShortCollectionsQuery("  海边合集  "))
        assertNull(normalizeShortCollectionsQuery("   "))
    }

    @Test
    fun resetShortCollectionsForSearch_restartsPagingAndKeepsExistingItemsForRetry() {
        val existing = listOf(ShortCollectionListItemDto(id = "c1", name = "旧合集"))
        val state = ShortCollectionsUiState(
            loading = false,
            loadingMore = true,
            loaded = true,
            page = 3,
            totalCount = 66,
            query = "",
            submittedQuery = "",
            items = existing,
            errorMessage = "old",
        )

        val next = resetShortCollectionsForSearch(state, " 海边 ", "海边")

        assertEquals(" 海边 ", next.query)
        assertEquals("海边", next.submittedQuery)
        assertEquals(0, next.page)
        assertEquals(0, next.totalCount)
        assertEquals(existing, next.items)
        assertTrue(next.loading)
        assertFalse(next.loadingMore)
        assertFalse(next.loaded)
        assertNull(next.errorMessage)
    }

    @Test
    fun shortCollectionsEmptyMessage_distinguishesBrowseAndSearch() {
        assertEquals("暂无合集", shortCollectionsEmptyMessage(""))
        assertEquals("没有找到相关合集", shortCollectionsEmptyMessage("海边"))
    }
}
