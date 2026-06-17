package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvHomeVideoDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSectionDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import com.google.gson.Gson
import kotlinx.coroutines.ExperimentalCoroutinesApi
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest

@OptIn(ExperimentalCoroutinesApi::class)
class TvCatalogViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun init_loadsBrowsePayloadWithAvContent() = runTest {
        val viewModel = TvCatalogViewModel(
            repository = FakeTvRepository(
                homePayload = TvHomePayload(
                    continueWatching = TvContinueWatchingDto(
                        seriesId = "7",
                        seriesTitle = "雾城档案",
                        seasonNumber = 2,
                        episodeNumber = 4,
                        episodeTitle = "暗线浮现",
                        posterUrl = "/poster.jpg",
                        backdropUrl = "/backdrop.jpg",
                        watchSeconds = 128,
                        progressPercent = 64,
                    ),
                    sections = listOf(
                        TvSectionDto(
                            title = "最近更新",
                            subtitle = "按最近播出时间排序",
                            items = listOf(tvSeriesSummary(id = "7", title = "雾城档案")),
                        ),
                    ),
                    movies = listOf(
                        TvHomeVideoDto(id = "movie-1", type = "movie", title = "午夜列车"),
                    ),
                    av = listOf(
                        TvHomeVideoDto(id = "av-1", type = "av", title = "SNIS-001"),
                    ),
                ),
            ),
        )

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("https://example.com", state.baseUrl)
        assertEquals("雾城档案", state.continueWatching?.seriesTitle)
        assertEquals("/poster.jpg", state.continueWatching?.posterUrl)
        assertEquals("/backdrop.jpg", state.continueWatching?.backdropUrl)
        assertEquals(128, state.continueWatching?.watchSeconds)
        assertEquals(1, state.sections.size)
        assertEquals(1, state.movies.size)
        assertEquals(1, state.av.size)
        assertEquals("av", state.av.first().type)
        assertEquals("SNIS-001", state.av.first().title)
    }

    @Test
    fun updateQuery_requestsSearchPayload() = runTest {
        val repository = FakeTvRepository(
            homePayload = TvHomePayload(),
            searchPayload = TvSearchPayload(
                items = listOf(tvSearchResult(id = "11", type = "tv", title = "静默轨道")),
            ),
        )
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.awaitIdle()
        viewModel.updateQuery("静默")
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals("静默", state.query)
        assertEquals(1, state.searchResults.size)
        assertEquals("静默轨道", state.searchResults.first().title)
        assertEquals("tv", state.searchResults.first().type)
    }

    @Test
    fun staleHomeResponse_afterMenuSwitchDoesNotOverrideCurrentMenu() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.selectMenu(TvHomeMenuItem.Movie)
        repository.completeHome(
            kind = "movie",
            payload = TvHomePayload(
                kind = "movie",
                movies = listOf(TvHomeVideoDto(id = "movie-current", type = "movie", title = "当前电影")),
            ),
        )
        advanceUntilIdle()

        var state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals(TvHomeMenuItem.Movie, state.selectedMenu)
        assertEquals("movie", state.kind)
        assertEquals(listOf("当前电影"), state.movies.map { it.title })

        repository.completeHome(
            kind = "tv",
            payload = TvHomePayload(
                kind = "tv",
                tvSeries = listOf(TvHomeVideoDto(id = "tv-stale", type = "tv", title = "旧电视剧")),
            ),
        )
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals(TvHomeMenuItem.Movie, state.selectedMenu)
        assertEquals("movie", state.kind)
        assertTrue(state.tvSeries.isEmpty())
        assertEquals(listOf("当前电影"), state.movies.map { it.title })
    }

    @Test
    fun staleSearchResponse_afterQueryChangeDoesNotOverrideCurrentResults() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvCatalogViewModel(repository = repository)

        repository.completeHome()
        advanceUntilIdle()
        viewModel.updateQuery("雾")
        viewModel.updateQuery("雾城")
        repository.completeSearch(
            query = "雾城",
            payload = TvSearchPayload(
                items = listOf(tvSearchResult(id = "new", type = "tv", title = "雾城档案")),
            ),
        )
        advanceUntilIdle()

        var state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("雾城", state.query)
        assertEquals(listOf("雾城档案"), state.searchResults.map { it.title })

        repository.completeSearch(
            query = "雾",
            payload = TvSearchPayload(
                items = listOf(tvSearchResult(id = "old", type = "tv", title = "旧搜索")),
            ),
        )
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("雾城", state.query)
        assertEquals(listOf("雾城档案"), state.searchResults.map { it.title })
    }

    @Test
    fun staleFailures_afterLeavingCatalogDoNotOverrideCurrentState() = runTest {
        val repository = DelayedCatalogTvRepository()
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.selectMenu(TvHomeMenuItem.Movie)
        viewModel.selectMenu(TvHomeMenuItem.Settings)
        repository.completeHomeFailure(kind = "movie", error = IllegalStateException("旧首页失败"))
        advanceUntilIdle()

        var state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals(TvHomeMenuItem.Settings, state.selectedMenu)
        assertEquals(null, state.errorMessage)

        viewModel.updateQuery("雾")
        viewModel.openIptv()
        repository.completeSearchFailure(query = "雾", error = IllegalStateException("旧搜索失败"))
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals(TvHomeMenuItem.Search, state.selectedMenu)
        assertEquals("", state.query)
        assertTrue(state.searchResults.isEmpty())
        assertEquals(null, state.errorMessage)
    }

    @Test
    fun settingsStateLoadsAndSavesTvSeekStepSeconds() = runTest {
        val repository = FakeTvRepository(tvSeekStepSeconds = 15, tvSeriesAutoplayEnabled = false)
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.awaitIdle()
        viewModel.selectMenu(TvHomeMenuItem.Settings)
        viewModel.awaitIdle()

        assertEquals(15, viewModel.uiState.value.tvSeekStepSeconds)
        assertFalse(viewModel.uiState.value.seriesAutoplayEnabled)

        viewModel.selectTvSeekStepSeconds(30)
        viewModel.setSeriesAutoplayEnabled(true)
        viewModel.awaitIdle()

        assertEquals(30, viewModel.uiState.value.tvSeekStepSeconds)
        assertTrue(viewModel.uiState.value.seriesAutoplayEnabled)
        assertEquals(30, repository.readTvSeekStepSeconds())
        assertEquals(true, repository.readTvSeriesAutoplayEnabled())
    }

    @Test
    fun updateQuery_keepsAvResultsInTvCatalog() = runTest {
        val repository = FakeTvRepository(
            homePayload = TvHomePayload(),
            searchPayload = TvSearchPayload(
                items = listOf(
                    tvSearchResult(id = "11", type = "tv", title = "静默轨道"),
                    tvSearchResult(id = "av-11", type = "av", title = "SNIS-001"),
                ),
            ),
        )
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.awaitIdle()
        viewModel.updateQuery("静默")
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.searchResults.size)
        assertEquals(listOf("tv", "av"), state.searchResults.map { it.type })
    }

    @Test
    fun init_keepsAvContinueWatchingInTvCatalog() = runTest {
        val viewModel = TvCatalogViewModel(
            repository = FakeTvRepository(
                homePayload = TvHomePayload(
                    continueWatching = TvContinueWatchingDto(
                        type = "av",
                        seriesId = "av-1",
                        seriesTitle = "SNIS-001",
                    ),
                ),
            ),
        )

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals("av", state.continueWatching?.type)
        assertEquals("SNIS-001", state.continueWatching?.seriesTitle)
    }

    @Test
    fun repositoryFailure_exposesErrorState() = runTest {
        val viewModel = TvCatalogViewModel(
            repository = FakeTvRepository(homeError = IllegalStateException("加载失败")),
        )

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("加载失败", state.errorMessage)
        assertTrue(state.sections.isEmpty())
    }

    @Test
    fun retry_afterHomeErrorRequestsCurrentKindAgain() = runTest {
        val repository = FakeTvRepository(homeError = IllegalStateException("加载失败"))
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.awaitIdle()
        viewModel.retry()
        viewModel.awaitIdle()

        assertEquals(2, repository.homeRequests.size)
        assertEquals("tv", repository.homeRequests.last().kind)
    }

    @Test
    fun retry_afterSearchErrorRequestsCurrentQueryAgain() = runTest {
        val repository = FakeTvRepository(homeError = IllegalStateException("搜索失败"))
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.awaitIdle()
        viewModel.selectMenu(TvHomeMenuItem.Search)
        viewModel.updateQuery("静默")
        viewModel.awaitIdle()
        viewModel.retry()
        viewModel.awaitIdle()

        assertEquals(2, repository.searchRequests.size)
        assertEquals("静默", repository.searchRequests.last().query)
    }

    @Test
    fun nullListsInPayload_doNotCrashAndFallbackToEmpty() = runTest {
        val payload = Gson().fromJson(
            """
            {
              "continue_watching": null,
              "sections": null,
              "tv_series": null,
              "movies": null,
              "av": null,
              "page": 1,
              "page_size": 20
            }
            """.trimIndent(),
            TvHomePayload::class.java,
        )
        val viewModel = TvCatalogViewModel(
            repository = FakeTvRepository(homePayload = payload),
        )

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertTrue(state.sections.isEmpty())
        assertTrue(state.movies.isEmpty())
        assertTrue(state.av.isEmpty())
        assertTrue(state.searchResults.isEmpty())
        assertEquals(null, state.errorMessage)
    }
}
