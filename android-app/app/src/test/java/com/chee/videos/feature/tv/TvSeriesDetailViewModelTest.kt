package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Test

class TvSeriesDetailViewModelTest {

    @Test
    fun selectSeason_resetsSelectedEpisodeToSeasonFirstEpisode() {
        val series = TvMockData.allSeries().first { it.seasons.size >= 2 }
        val viewModel = TvSeriesDetailViewModel(
            SavedStateHandle(
                mapOf(
                    TvSeriesIdArg to series.id,
                ),
            ),
        )
        val before = viewModel.uiState.value
        assertNotNull(before.series)

        val targetSeason = before.series!!.seasons[1]
        viewModel.selectSeason(targetSeason.number)

        val after = viewModel.uiState.value
        assertEquals(targetSeason.number, after.selectedSeasonNumber)
        assertEquals(targetSeason.episodes.first().number, after.selectedEpisodeNumber)
    }
}
