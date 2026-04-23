package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Rule
import org.junit.Test
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
}
