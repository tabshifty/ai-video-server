package com.chee.videos.feature.imagecollections

import com.chee.videos.core.model.ImageCollectionListItemDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class ImageCollectionsSearchStateTest {
    @Test
    fun normalizeImageCollectionsQuery_trimsWhitespaceAndDropsBlank() {
        assertEquals("旅行写真", normalizeImageCollectionsQuery("  旅行写真  "))
        assertNull(normalizeImageCollectionsQuery("   "))
    }

    @Test
    fun resetImageCollectionsForQuery_restartsPagingForNewSearch() {
        val state = ImageCollectionsUiState(
            loading = false,
            loadingMore = true,
            loaded = true,
            page = 3,
            totalCount = 66,
            query = "",
            items = listOf(ImageCollectionListItemDto(id = "c1", name = "旧合集")),
            errorMessage = "old",
        )

        val next = resetImageCollectionsForQuery(state, " 海边 ")

        assertEquals(" 海边 ", next.query)
        assertEquals(0, next.page)
        assertEquals(0, next.totalCount)
        assertEquals(emptyList<ImageCollectionListItemDto>(), next.items)
        assertEquals(true, next.loading)
        assertEquals(false, next.loadingMore)
        assertEquals(false, next.loaded)
        assertNull(next.errorMessage)
    }
}
