package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.runTest

class TvPosterWallViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun init_loadsFirstPage_andLoadMoreAppendsNextPage() = runTest {
        val viewModel = TvPosterWallViewModel(
            repository = FakeTvRepository(
                posterWallPages = listOf(
                    tvPosterWallPage(page = 1, totalCount = 10, items = listOf(
                        tvPosterWallItem(id = "item-1", type = "tv", title = "雾城档案"),
                        tvPosterWallItem(id = "item-2", type = "tv", title = "静默轨道"),
                    )),
                    tvPosterWallPage(page = 2, totalCount = 10, items = listOf(
                        tvPosterWallItem(id = "item-3", type = "tv", title = "旧城往事"),
                    )),
                ),
            ),
            kind = "recent",
        )

        viewModel.awaitIdle()

        val initial = viewModel.uiState.value
        assertFalse(initial.loading)
        assertEquals(1, initial.page)
        assertEquals(2, initial.items.size)

        viewModel.loadMoreIfNeeded(currentIndex = 1)
        viewModel.awaitIdle()

        val loaded = viewModel.uiState.value
        assertEquals(2, loaded.page)
        assertEquals(3, loaded.items.size)
        assertEquals("item-3", loaded.items.last().id)
    }
}
