package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TvSeriesDetailUiState(
    val loading: Boolean = true,
    val series: TvSeriesUiModel? = null,
    val selectedSeasonNumber: Int = 1,
    val selectedEpisodeNumber: Int = 1,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvSeriesDetailViewModel @Inject constructor(
    private val repository: TvRepository,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val seriesId = savedStateHandle.get<String>(TvSeriesIdArg).orEmpty()

    private val _uiState = MutableStateFlow(TvSeriesDetailUiState())
    val uiState: StateFlow<TvSeriesDetailUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun selectSeason(seasonNumber: Int) {
        val series = _uiState.value.series ?: return
        val season = series.seasons.firstOrNull { it.number == seasonNumber } ?: return
        _uiState.update { state ->
            state.copy(
                selectedSeasonNumber = season.number,
                selectedEpisodeNumber = findPreferredEpisodeNumber(season),
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

    private fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, errorMessage = null) }
            repository.fetchSeriesDetail(seriesId)
                .onSuccess { dto ->
                    val series = tvSeriesDetailToUiModel(dto)
                    val defaultSelection = findPreferredSeriesSelection(series)
                    val defaultSeason = series.seasons.firstOrNull { it.number == defaultSelection?.seasonNumber }
                        ?: series.seasons.firstOrNull()
                    _uiState.value = TvSeriesDetailUiState(
                        loading = false,
                        series = series,
                        selectedSeasonNumber = defaultSeason?.number ?: 1,
                        selectedEpisodeNumber = defaultSelection?.episodeNumber
                            ?: defaultSeason?.let(::findPreferredEpisodeNumber)
                            ?: 1,
                        errorMessage = null,
                    )
                }
                .onFailure { error ->
                    _uiState.value = TvSeriesDetailUiState(
                        loading = false,
                        errorMessage = error.message ?: "电视剧详情加载失败",
                    )
                }
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
