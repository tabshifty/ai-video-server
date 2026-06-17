package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest

class TvSeriesDetailViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun selectSeason_resetsSelectedEpisodeToSeasonFirstPlayableEpisode() = runTest {
        val viewModel = TvSeriesDetailViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                            ),
                        ),
                        TvSeasonDto(
                            id = "s2",
                            seasonNumber = 2,
                            title = "第二季",
                            episodes = listOf(
                                tvEpisode(id = "e2", number = 1, title = "第1集"),
                                tvEpisode(id = "e3", number = 2, title = "第2集", videoId = "video-2", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(
                mapOf(
                    TvSeriesIdArg to "series-1",
                ),
            ),
        )
        viewModel.awaitIdle()

        val before = viewModel.uiState.value
        assertNotNull(before.series)

        val targetSeason = before.series!!.seasons[1]
        viewModel.selectSeason(targetSeason.number)

        val after = viewModel.uiState.value
        assertEquals(targetSeason.number, after.selectedSeasonNumber)
        assertEquals(2, after.selectedEpisodeNumber)
    }

    @Test
    fun init_prefersMostRecentlyWatchedEpisodeAcrossSeasons() = runTest {
        val viewModel = TvSeriesDetailViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e1",
                                    number = 1,
                                    title = "第1集",
                                    videoId = "video-1",
                                    videoStatus = "ready",
                                    watchSeconds = 90,
                                    lastWatchedAt = "2026-04-25T08:00:00Z",
                                ),
                            ),
                        ),
                        TvSeasonDto(
                            id = "s2",
                            seasonNumber = 2,
                            title = "第二季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e2",
                                    number = 3,
                                    title = "第3集",
                                    videoId = "video-2",
                                    videoStatus = "ready",
                                    watchSeconds = 45,
                                    lastWatchedAt = "2026-04-26T09:30:00Z",
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedSeasonNumber)
        assertEquals(3, state.selectedEpisodeNumber)
    }

    @Test
    fun retry_afterErrorRequestsDetailAgain() = runTest {
        val repository = FakeTvRepository(detailError = IllegalStateException("详情失败"))
        val viewModel = TvSeriesDetailViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        viewModel.awaitIdle()
        viewModel.retry()
        viewModel.awaitIdle()

        assertEquals(listOf("series-1", "series-1"), repository.detailRequests)
    }

    @Test
    fun reload_withExistingSeriesKeepsContentAndUsesRefreshingState() = runTest {
        val repository = DelayedSeriesDetailTvRepository()
        val viewModel = TvSeriesDetailViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 1,
                        episodes = listOf(
                            playableEpisode(id = "e1", number = 1, title = "第1集"),
                        ),
                    ),
                ),
            ),
        )
        advanceUntilIdle()

        viewModel.retry()
        runCurrent()

        val refreshingState = viewModel.uiState.value
        assertFalse(refreshingState.loading)
        assertTrue(refreshingState.refreshing)
        assertNotNull(refreshingState.series)
        assertEquals("series-1", refreshingState.series?.id)

        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 1,
                        episodes = listOf(
                            playableEpisode(id = "e1b", number = 1, title = "第1集-新版"),
                        ),
                    ),
                ),
            ),
        )
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertFalse(state.refreshing)
        assertEquals("第1集-新版", selectedDetailEpisode(state)?.title)
    }

    @Test
    fun reload_successKeepsCurrentSeasonAndEpisodeWhenStillAvailable() = runTest {
        val repository = DelayedSeriesDetailTvRepository()
        val viewModel = TvSeriesDetailViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 1,
                        episodes = listOf(playableEpisode(id = "s1e1", number = 1, title = "第一季第1集")),
                    ),
                    season(
                        seasonNumber = 2,
                        episodes = listOf(
                            playableEpisode(id = "s2e1", number = 1, title = "第二季第1集"),
                            playableEpisode(id = "s2e2", number = 2, title = "第二季第2集"),
                        ),
                    ),
                ),
            ),
        )
        advanceUntilIdle()
        viewModel.selectSeason(2)
        viewModel.selectEpisode(2)

        viewModel.retry()
        runCurrent()
        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 1,
                        episodes = listOf(playableEpisode(id = "s1e1-new", number = 1, title = "第一季第1集")),
                    ),
                    season(
                        seasonNumber = 2,
                        episodes = listOf(
                            playableEpisode(id = "s2e1-new", number = 1, title = "第二季第1集"),
                            playableEpisode(id = "s2e2-new", number = 2, title = "第二季第2集-新版"),
                        ),
                    ),
                ),
            ),
        )
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedSeasonNumber)
        assertEquals(2, state.selectedEpisodeNumber)
        assertEquals("第二季第2集-新版", selectedDetailEpisode(state)?.title)
    }

    @Test
    fun reload_invalidCurrentEpisodeFallsBackInsideCurrentSeasonBeforeGlobalPreferred() = runTest {
        val repository = DelayedSeriesDetailTvRepository()
        val viewModel = TvSeriesDetailViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 1,
                        episodes = listOf(playableEpisode(id = "s1e1", number = 1, title = "第一季第1集")),
                    ),
                    season(
                        seasonNumber = 2,
                        episodes = listOf(
                            playableEpisode(id = "s2e1", number = 1, title = "第二季第1集"),
                            playableEpisode(id = "s2e2", number = 2, title = "第二季第2集"),
                        ),
                    ),
                ),
            ),
        )
        advanceUntilIdle()
        viewModel.selectSeason(2)
        viewModel.selectEpisode(2)

        viewModel.retry()
        runCurrent()
        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 1,
                        episodes = listOf(
                            playableEpisode(
                                id = "s1e9",
                                number = 9,
                                title = "第一季第9集",
                                watchSeconds = 90,
                                lastWatchedAt = "2026-06-16T08:00:00Z",
                            ),
                        ),
                    ),
                    season(
                        seasonNumber = 2,
                        episodes = listOf(
                            pendingEpisode(id = "s2e1", number = 1, title = "第二季第1集"),
                            playableEpisode(id = "s2e3", number = 3, title = "第二季第3集"),
                        ),
                    ),
                ),
            ),
        )
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedSeasonNumber)
        assertEquals(3, state.selectedEpisodeNumber)
    }

    @Test
    fun reload_invalidCurrentSeasonFallsBackToGlobalPreferredSelection() = runTest {
        val repository = DelayedSeriesDetailTvRepository()
        val viewModel = TvSeriesDetailViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 1,
                        episodes = listOf(playableEpisode(id = "s1e1", number = 1, title = "第一季第1集")),
                    ),
                    season(
                        seasonNumber = 2,
                        episodes = listOf(playableEpisode(id = "s2e1", number = 1, title = "第二季第1集")),
                    ),
                ),
            ),
        )
        advanceUntilIdle()
        viewModel.selectSeason(2)

        viewModel.retry()
        runCurrent()
        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(
                    season(
                        seasonNumber = 3,
                        episodes = listOf(
                            playableEpisode(
                                id = "s3e5",
                                number = 5,
                                title = "第三季第5集",
                                watchSeconds = 200,
                                lastWatchedAt = "2026-06-16T10:00:00Z",
                            ),
                        ),
                    ),
                    season(
                        seasonNumber = 4,
                        episodes = listOf(playableEpisode(id = "s4e1", number = 1, title = "第四季第1集")),
                    ),
                ),
            ),
        )
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(3, state.selectedSeasonNumber)
        assertEquals(5, state.selectedEpisodeNumber)
    }

    @Test
    fun reload_failureWithExistingSeriesKeepsOldSelectionAndSeries() = runTest {
        val repository = DelayedSeriesDetailTvRepository()
        val viewModel = TvSeriesDetailViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeDetail(
            tvSeriesDetail(
                title = "旧详情",
                seasons = listOf(
                    season(
                        seasonNumber = 2,
                        episodes = listOf(
                            playableEpisode(id = "s2e1", number = 1, title = "旧第1集"),
                            playableEpisode(id = "s2e2", number = 2, title = "旧第2集"),
                        ),
                    ),
                ),
            ),
        )
        advanceUntilIdle()
        viewModel.selectSeason(2)
        viewModel.selectEpisode(2)

        viewModel.retry()
        runCurrent()
        repository.completeDetailFailure(IllegalStateException("刷新失败"))
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertFalse(state.refreshing)
        assertEquals("旧详情", state.series?.title)
        assertEquals(2, state.selectedSeasonNumber)
        assertEquals(2, state.selectedEpisodeNumber)
        assertEquals("刷新失败", state.errorMessage)
    }

    @Test
    fun staleReloadResponse_doesNotOverrideNewerSeriesState() = runTest {
        val repository = DelayedSeriesDetailTvRepository()
        val viewModel = TvSeriesDetailViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeDetail(
            tvSeriesDetail(
                seasons = listOf(season(1, listOf(playableEpisode(id = "base-e1", number = 1, title = "基础集")))),
            ),
        )
        advanceUntilIdle()

        viewModel.retry()
        runCurrent()
        viewModel.retry()
        runCurrent()

        repository.completeLatestDetail(
            tvSeriesDetail(
                title = "新详情",
                seasons = listOf(season(1, listOf(playableEpisode(id = "fresh-e1", number = 1, title = "新集")))),
            ),
        )
        advanceUntilIdle()

        repository.completeDetail(
            tvSeriesDetail(
                title = "旧详情",
                seasons = listOf(season(1, listOf(playableEpisode(id = "stale-e1", number = 1, title = "旧集")))),
            ),
        )
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals("新详情", state.series?.title)
        assertEquals("fresh-e1", selectedDetailEpisode(state)?.id)
    }

    private fun season(
        seasonNumber: Int,
        episodes: List<TvEpisodeDto>,
    ) = TvSeasonDto(
        id = "season-$seasonNumber",
        seasonNumber = seasonNumber,
        title = "第${seasonNumber}季",
        episodes = episodes,
    )

    private fun playableEpisode(
        id: String,
        number: Int,
        title: String,
        watchSeconds: Int = 0,
        lastWatchedAt: String? = null,
    ): TvEpisodeDto = tvEpisode(
        id = id,
        number = number,
        title = title,
        videoId = "video-$id",
        videoStatus = "ready",
        watchSeconds = watchSeconds,
        lastWatchedAt = lastWatchedAt,
    )

    private fun pendingEpisode(
        id: String,
        number: Int,
        title: String,
    ): TvEpisodeDto = tvEpisode(
        id = id,
        number = number,
        title = title,
    )
}
