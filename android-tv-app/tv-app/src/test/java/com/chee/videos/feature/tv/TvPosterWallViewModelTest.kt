package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest

@OptIn(ExperimentalCoroutinesApi::class)
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

    @Test
    fun init_keepsAvWallKindAndItems() = runTest {
        val viewModel = TvPosterWallViewModel(
            repository = FakeTvRepository(
                posterWallPages = listOf(
                    tvPosterWallPage(
                        page = 1,
                        totalCount = 1,
                        items = listOf(tvPosterWallItem(id = "av-1", type = "av", title = "SNIS-001")),
                    ),
                ),
            ),
            kind = "av",
            title = "全部18+",
        )

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals("av", state.kind)
        assertEquals("全部18+", state.title)
        assertEquals(1, state.items.size)
        assertEquals("av", state.items.first().type)
    }

    @Test
    fun init_andSortChange_passesSortToRepositoryAndReloadsFirstPage() = runTest {
        val repository = FakeTvRepository(
            posterWallPages = listOf(
                tvPosterWallPage(page = 1, totalCount = 10, items = listOf(
                    tvPosterWallItem(id = "item-1", type = "movie", title = "午夜列车"),
                )),
                tvPosterWallPage(page = 2, totalCount = 10, items = listOf(
                    tvPosterWallItem(id = "item-2", type = "movie", title = "旧城往事"),
                )),
            ),
        )
        val viewModel = TvPosterWallViewModel(repository = repository, kind = "movie")

        viewModel.awaitIdle()
        assertEquals("added", repository.posterWallRequests.last().sortBy)
        assertEquals("desc", repository.posterWallRequests.last().sortOrder)

        viewModel.loadMoreIfNeeded(currentIndex = 0)
        viewModel.awaitIdle()
        assertEquals(2, viewModel.uiState.value.page)

        viewModel.changeSort(sortBy = "release", sortOrder = "asc")
        viewModel.awaitIdle()

        val lastRequest = repository.posterWallRequests.last()
        assertEquals("release", lastRequest.sortBy)
        assertEquals("asc", lastRequest.sortOrder)
        assertEquals(1, lastRequest.page)
        assertEquals(1, viewModel.uiState.value.page)
        assertEquals("release", viewModel.uiState.value.sortBy)
        assertEquals("asc", viewModel.uiState.value.sortOrder)
    }

    @Test
    fun refresh_withExistingItemsKeepsGridVisibleWhileRequestIsPending() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvPosterWallViewModel(repository = repository, kind = "movie")

        repository.completePosterWall(
            kind = "movie",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(
                    tvPosterWallItem(id = "item-1", type = "movie", title = "午夜列车"),
                    tvPosterWallItem(id = "item-2", type = "movie", title = "旧城往事"),
                ),
            ),
        )
        advanceUntilIdle()

        viewModel.refresh()
        runCurrent()

        var state = viewModel.uiState.value
        assertFalse(state.loading)
        assertTrue(state.refreshing)
        assertEquals(listOf("item-1", "item-2"), state.items.map { it.id })

        repository.completePosterWall(
            kind = "movie",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(tvPosterWallItem(id = "item-3", type = "movie", title = "新片")),
            ),
        )
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertFalse(state.loading)
        assertFalse(state.refreshing)
        assertEquals(listOf("item-3"), state.items.map { it.id })
    }

    @Test
    fun changeSort_withExistingItemsKeepsGridVisibleUntilNewSortCompletes() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvPosterWallViewModel(repository = repository, kind = "movie")

        repository.completePosterWall(
            kind = "movie",
            sortBy = "added",
            sortOrder = "desc",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(
                    tvPosterWallItem(id = "item-1", type = "movie", title = "午夜列车"),
                    tvPosterWallItem(id = "item-2", type = "movie", title = "旧城往事"),
                ),
            ),
        )
        advanceUntilIdle()

        viewModel.changeSort(sortBy = "release", sortOrder = "asc")
        runCurrent()

        var state = viewModel.uiState.value
        assertFalse(state.loading)
        assertTrue(state.refreshing)
        assertEquals("release", state.sortBy)
        assertEquals("asc", state.sortOrder)
        assertEquals(listOf("item-1", "item-2"), state.items.map { it.id })

        repository.completePosterWall(
            kind = "movie",
            sortBy = "release",
            sortOrder = "asc",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(tvPosterWallItem(id = "item-3", type = "movie", title = "发售新片")),
            ),
        )
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertFalse(state.loading)
        assertFalse(state.refreshing)
        assertEquals(1, state.page)
        assertEquals(listOf("item-3"), state.items.map { it.id })
    }

    @Test
    fun staleSortResponse_doesNotOverrideLatestSortResults() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvPosterWallViewModel(repository = repository, kind = "movie")

        repository.completePosterWall(
            kind = "movie",
            sortBy = "added",
            sortOrder = "desc",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(tvPosterWallItem(id = "item-1", type = "movie", title = "午夜列车")),
            ),
        )
        advanceUntilIdle()

        viewModel.changeSort(sortBy = "release", sortOrder = "asc")
        runCurrent()
        viewModel.changeSort(sortBy = "added", sortOrder = "asc")
        runCurrent()

        repository.completePosterWall(
            kind = "movie",
            sortBy = "added",
            sortOrder = "asc",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(tvPosterWallItem(id = "current", type = "movie", title = "当前排序")),
            ),
        )
        advanceUntilIdle()

        var state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("added", state.sortBy)
        assertEquals("asc", state.sortOrder)
        assertEquals(listOf("current"), state.items.map { it.id })

        repository.completePosterWall(
            kind = "movie",
            sortBy = "release",
            sortOrder = "asc",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(tvPosterWallItem(id = "stale", type = "movie", title = "旧排序")),
            ),
        )
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("added", state.sortBy)
        assertEquals("asc", state.sortOrder)
        assertEquals(listOf("current"), state.items.map { it.id })
    }

    @Test
    fun refreshFailure_withExistingItemsKeepsGridVisibleAndExposesInlineError() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvPosterWallViewModel(repository = repository, kind = "movie")

        repository.completePosterWall(
            kind = "movie",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(tvPosterWallItem(id = "item-1", type = "movie", title = "午夜列车")),
            ),
        )
        advanceUntilIdle()

        viewModel.refresh()
        runCurrent()
        repository.completePosterWallFailure(kind = "movie", error = IllegalStateException("网络超时"))
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertFalse(state.refreshing)
        assertEquals(listOf("item-1"), state.items.map { it.id })
        assertEquals("网络超时", state.errorMessage)
    }

    @Test
    fun loadMore_isBlockedAfterSoftUpdateFailureUntilRefreshSucceeds() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvPosterWallViewModel(repository = repository, kind = "movie")

        repository.completePosterWall(
            kind = "movie",
            payload = tvPosterWallPage(
                page = 1,
                totalCount = 8,
                items = listOf(
                    tvPosterWallItem(id = "item-1", type = "movie", title = "午夜列车"),
                    tvPosterWallItem(id = "item-2", type = "movie", title = "旧城往事"),
                ),
            ),
        )
        advanceUntilIdle()

        viewModel.changeSort(sortBy = "release", sortOrder = "asc")
        runCurrent()
        repository.completePosterWallFailure(
            kind = "movie",
            sortBy = "release",
            sortOrder = "asc",
            error = IllegalStateException("网络超时"),
        )
        advanceUntilIdle()

        viewModel.loadMoreIfNeeded(currentIndex = 1)
        runCurrent()

        assertEquals(
            listOf(
                TvPosterWallRequest(kind = "movie", page = 1, pageSize = 24, sortBy = "added", sortOrder = "desc"),
                TvPosterWallRequest(kind = "movie", page = 1, pageSize = 24, sortBy = "release", sortOrder = "asc"),
            ),
            repository.posterWallRequests,
        )
    }
}
