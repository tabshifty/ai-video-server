package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvCatalogFocusPolicyTest {

    @Test
    fun `prefers continue watching when present`() {
        assertEquals(
            TvCatalogInitialFocusTarget.CONTINUE_WATCHING,
            resolveTvCatalogInitialFocusTarget(
                hasContinueWatching = true,
                sectionItemCounts = listOf(3, 2),
                tvSeriesCount = 4,
                movieCount = 2,
                avCount = 1,
            ),
        )
    }

    @Test
    fun `falls back to first section item before shelves`() {
        assertEquals(
            TvCatalogInitialFocusTarget.FIRST_SECTION_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasContinueWatching = false,
                sectionItemCounts = listOf(0, 2),
                tvSeriesCount = 4,
                movieCount = 2,
                avCount = 1,
            ),
        )
    }

    @Test
    fun `falls back through shelves then search when content is empty`() {
        assertEquals(
            TvCatalogInitialFocusTarget.TV_SERIES_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 1,
                movieCount = 2,
                avCount = 1,
            ),
        )
        assertEquals(
            TvCatalogInitialFocusTarget.MOVIE_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 0,
                movieCount = 2,
                avCount = 1,
            ),
        )
        assertEquals(
            TvCatalogInitialFocusTarget.AV_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 0,
                movieCount = 0,
                avCount = 1,
            ),
        )
        assertEquals(
            TvCatalogInitialFocusTarget.SEARCH,
            resolveTvCatalogInitialFocusTarget(
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 0,
                movieCount = 0,
                avCount = 0,
            ),
        )
    }
}
