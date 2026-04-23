package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesPlayerViewModelTest {

    @Test
    fun init_prefersRouteSeasonAndEpisodeWhenExists() {
        val series = TvMockData.allSeries().first { item -> item.seasons.any { season -> season.number == 1 && season.episodes.size >= 4 } }
        val viewModel = TvSeriesPlayerViewModel(
            SavedStateHandle(
                mapOf(
                    TvSeriesIdArg to series.id,
                    TvSeasonArg to 1,
                    TvEpisodeArg to 4,
                ),
            ),
        )

        val state = viewModel.uiState.value
        assertEquals(1, state.selectedSeasonNumber)
        assertEquals(4, state.selectedEpisodeNumber)
    }

    @Test
    fun cycleSpeed_rotatesWithinSupportedValues() {
        val series = TvMockData.allSeries().first()
        val viewModel = TvSeriesPlayerViewModel(
            SavedStateHandle(
                mapOf(TvSeriesIdArg to series.id),
            ),
        )
        repeat(5) {
            viewModel.cycleSpeed()
        }

        val speed = viewModel.uiState.value.playbackSpeed
        assertTrue(speed == 1f || speed == 1.25f || speed == 1.5f || speed == 2f)
    }
}
