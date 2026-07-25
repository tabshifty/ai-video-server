package com.chee.videos.core.ui

import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.ui.cast.buildCollectionShortTvRemoteSearchContext
import com.chee.videos.core.ui.cast.resolveShortTvRemoteStartIndex
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvCastTest {
    @Test
    fun collectionContext_keepsCollectionPagingBoundary() {
        val context = buildCollectionShortTvRemoteSearchContext(
            collectionId = " collection-1 ",
            page = 3,
            totalCount = 71,
        )

        assertEquals("", context?.query)
        assertEquals("short", context?.type)
        assertEquals("collection-1", context?.collectionId)
        assertEquals(3, context?.page)
        assertEquals(24, context?.pageSize)
        assertEquals(71, context?.totalCount)
    }

    @Test
    fun collectionContext_rejectsBlankCollectionId() {
        assertNull(buildCollectionShortTvRemoteSearchContext("  ", page = 1, totalCount = 0))
    }

    @Test
    fun startIndex_followsCurrentSnapshotOrder() {
        val items = listOf(
            VideoListItemDto(id = "v1", title = "第一条", type = "short"),
            VideoListItemDto(id = "v2", title = "第二条", type = "short"),
        )

        assertEquals(1, resolveShortTvRemoteStartIndex(items, "v2"))
        assertNull(resolveShortTvRemoteStartIndex(items, "missing"))
    }
}
