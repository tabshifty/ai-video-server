package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.runTest

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
}
