package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import kotlinx.coroutines.ExperimentalCoroutinesApi
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest

@OptIn(ExperimentalCoroutinesApi::class)
class TvSeriesPlayerViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun init_prefersRouteSeasonAndEpisodeWhenPlayable() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                                tvEpisode(id = "e4", number = 4, title = "第4集", videoId = "video-4", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(
                mapOf(
                    TvSeriesIdArg to "series-1",
                    TvSeasonArg to 1,
                    TvEpisodeArg to 4,
                ),
            ),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals(1, state.selectedSeasonNumber)
        assertEquals(4, state.selectedEpisodeNumber)
        assertEquals("video-4", state.currentVideoId)
    }

    @Test
    fun init_exposesActiveBaseUrlForSubtitleResolution() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                baseUrl = "https://media.example.com",
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
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals("https://media.example.com", viewModel.uiState.value.baseUrl)
    }

    @Test
    fun init_exposesEpisodeWatchSecondsForResume() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
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
                                    watchSeconds = 187,
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals(187, selectedEpisode(viewModel.uiState.value)?.watchSeconds)
    }

    @Test
    fun init_restoresSavedSubtitleSelectionForCurrentVideo() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                subtitlePreferences = mapOf("video-1" to "subtitle-zh"),
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
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals("subtitle-zh", viewModel.uiState.value.selectedSubtitleTrackId)
    }

    @Test
    fun nextEpisode_skipsToFollowingEpisodeAndUpdatesVideoId() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                                tvEpisode(id = "e2", number = 2, title = "第2集", videoId = "video-2", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(
                mapOf(TvSeriesIdArg to "series-1"),
            ),
        )
        viewModel.awaitIdle()
        viewModel.nextEpisode()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedEpisodeNumber)
        assertEquals("video-2", state.currentVideoId)
    }

    @Test
    fun unboundEpisode_cannotPlay() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集"),
                            ),
                        ),
                    ),
                ),
            ),
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertFalse(state.canPlayCurrentEpisode)
        assertEquals("", state.currentVideoId)
    }

    @Test
    fun selectEpisode_clearsPreviousPlaybackTargetUntilNewSourceReady() = runTest {
        val repository = DelayedSourceTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                            tvEpisode(id = "e2", number = 2, title = "第2集", videoId = "video-2", videoStatus = "ready"),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeSourceUrl("video-1")
        advanceUntilIdle()
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)

        viewModel.selectEpisode(2)
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedEpisodeNumber)
        assertEquals("", state.currentVideoId)
        assertEquals("", state.currentSourceUrl)
        assertFalse(state.canPlayCurrentEpisode)
    }

    @Test
    fun selectEpisode_ignoresLateSourceUrlFromPreviousEpisode() = runTest {
        val repository = DelayedSourceTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                            tvEpisode(id = "e2", number = 2, title = "第2集", videoId = "video-2", videoStatus = "ready"),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeSourceUrl("video-1")
        advanceUntilIdle()
        assertEquals(1, viewModel.uiState.value.selectedEpisodeNumber)
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)

        viewModel.selectEpisode(2)
        advanceUntilIdle()
        viewModel.selectEpisode(1)
        advanceUntilIdle()
        assertEquals(1, viewModel.uiState.value.selectedEpisodeNumber)
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)

        repository.completeSourceUrl("video-2")
        advanceUntilIdle()
        assertEquals(1, viewModel.uiState.value.selectedEpisodeNumber)
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)
        assertEquals("https://example.com/video-1.m3u8", viewModel.uiState.value.currentSourceUrl)
        assertTrue(viewModel.uiState.value.canPlayCurrentEpisode)
    }
}
