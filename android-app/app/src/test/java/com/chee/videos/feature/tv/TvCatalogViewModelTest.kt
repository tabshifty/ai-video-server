package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvSectionDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import com.google.gson.Gson
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.runTest

class TvCatalogViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun init_loadsBrowsePayload() = runTest {
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
    }

    @Test
    fun updateQuery_requestsSearchPayload() = runTest {
        val repository = FakeTvRepository(
            homePayload = TvHomePayload(),
            searchPayload = TvHomePayload(
                searchResults = listOf(tvSeriesSummary(id = "11", title = "静默轨道")),
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
    fun nullListsInPayload_doNotCrashAndFallbackToEmpty() = runTest {
        val payload = Gson().fromJson(
            """
            {
              "continue_watching": null,
              "sections": null,
              "search_results": null,
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
        assertTrue(state.searchResults.isEmpty())
        assertEquals(null, state.errorMessage)
    }
}
