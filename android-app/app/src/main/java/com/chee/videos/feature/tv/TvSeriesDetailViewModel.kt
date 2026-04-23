package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update

data class TvSeriesDetailUiState(
    val loading: Boolean = true,
    val series: TvSeriesUiModel? = null,
    val selectedSeasonNumber: Int = 1,
    val selectedEpisodeNumber: Int = 1,
)

@HiltViewModel
class TvSeriesDetailViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    private val _uiState = MutableStateFlow(TvSeriesDetailUiState())
    val uiState: StateFlow<TvSeriesDetailUiState> = _uiState.asStateFlow()

    init {
        val seriesId = savedStateHandle.get<String>(TvSeriesIdArg).orEmpty()
        val series = TvMockData.findSeries(seriesId) ?: TvMockData.allSeries().firstOrNull()
        val defaultSeason = series?.seasons?.firstOrNull()
        _uiState.value = TvSeriesDetailUiState(
            loading = false,
            series = series,
            selectedSeasonNumber = defaultSeason?.number ?: 1,
            selectedEpisodeNumber = defaultSeason?.episodes?.firstOrNull()?.number ?: 1,
        )
    }

    fun selectSeason(seasonNumber: Int) {
        val series = _uiState.value.series ?: return
        val season = series.seasons.firstOrNull { it.number == seasonNumber } ?: return
        _uiState.update { state ->
            state.copy(
                selectedSeasonNumber = season.number,
                selectedEpisodeNumber = season.episodes.firstOrNull()?.number ?: 1,
            )
        }
    }

    fun selectEpisode(episodeNumber: Int) {
        val season = selectedDetailSeason(_uiState.value) ?: return
        if (season.episodes.none { it.number == episodeNumber }) {
            return
        }
        _uiState.update { state ->
            state.copy(selectedEpisodeNumber = episodeNumber)
        }
    }
}

internal fun selectedDetailSeason(state: TvSeriesDetailUiState): TvSeasonUiModel? {
    val series = state.series ?: return null
    return series.seasons.firstOrNull { it.number == state.selectedSeasonNumber }
}

internal fun selectedDetailEpisode(state: TvSeriesDetailUiState): TvEpisodeUiModel? {
    val season = selectedDetailSeason(state) ?: return null
    return season.episodes.firstOrNull { it.number == state.selectedEpisodeNumber }
}
