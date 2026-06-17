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
    val refreshing: Boolean = false,
    val series: TvSeriesUiModel? = null,
    val baseUrl: String = "",
    val selectedSeasonNumber: Int = 1,
    val selectedEpisodeNumber: Int = 1,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvSeriesDetailViewModel @Inject constructor(
    private val repository: TvRepository,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val seriesId = decodeTvRouteArg(savedStateHandle.get<String>(TvSeriesIdArg))
    private var requestVersion = 0

    private val _uiState = MutableStateFlow(TvSeriesDetailUiState())
    val uiState: StateFlow<TvSeriesDetailUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun retry() {
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
        val versionAtRequest = nextRequestVersion()
        viewModelScope.launch {
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            _uiState.update {
                val hasExistingData = it.series != null
                it.copy(
                    loading = !hasExistingData,
                    refreshing = hasExistingData,
                    baseUrl = baseUrl,
                    errorMessage = null,
                )
            }
            repository.fetchSeriesDetail(seriesId)
                .onSuccess { dto ->
                    if (versionAtRequest != requestVersion) {
                        return@onSuccess
                    }
                    val series = tvSeriesDetailToUiModel(dto)
                    _uiState.update { state ->
                        val resolvedSelection = resolveSeriesDetailSelection(
                            series = series,
                            currentSeasonNumber = state.selectedSeasonNumber,
                            currentEpisodeNumber = state.selectedEpisodeNumber,
                            preserveCurrentSelection = state.series != null,
                        )
                        state.copy(
                            loading = false,
                            refreshing = false,
                            series = series,
                            baseUrl = baseUrl,
                            selectedSeasonNumber = resolvedSelection.seasonNumber,
                            selectedEpisodeNumber = resolvedSelection.episodeNumber,
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { error ->
                    if (versionAtRequest != requestVersion) {
                        return@onFailure
                    }
                    _uiState.update { state ->
                        val hasExistingData = state.series != null
                        if (hasExistingData) {
                            state.copy(
                                loading = false,
                                refreshing = false,
                                baseUrl = baseUrl,
                                errorMessage = error.message ?: "电视剧详情加载失败",
                            )
                        } else {
                            TvSeriesDetailUiState(
                                loading = false,
                                refreshing = false,
                                baseUrl = baseUrl,
                                errorMessage = error.message ?: "电视剧详情加载失败",
                            )
                        }
                    }
                }
        }
    }

    private fun nextRequestVersion(): Int {
        requestVersion += 1
        return requestVersion
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

internal fun resolveSeriesDetailSelection(
    series: TvSeriesUiModel,
    currentSeasonNumber: Int,
    currentEpisodeNumber: Int,
    preserveCurrentSelection: Boolean,
): TvPreferredEpisodeSelection {
    if (!preserveCurrentSelection) {
        return defaultSeriesDetailSelection(series)
    }

    val currentSeason = series.seasons.firstOrNull { it.number == currentSeasonNumber }
    if (currentSeason != null) {
        val currentEpisode = currentSeason.episodes.firstOrNull { it.number == currentEpisodeNumber }
        if (currentEpisode != null) {
            return TvPreferredEpisodeSelection(currentSeason.number, currentEpisode.number)
        }
        return TvPreferredEpisodeSelection(
            seasonNumber = currentSeason.number,
            episodeNumber = findPreferredEpisodeNumber(currentSeason),
        )
    }

    return defaultSeriesDetailSelection(series)
}

private fun defaultSeriesDetailSelection(series: TvSeriesUiModel): TvPreferredEpisodeSelection {
    val defaultSelection = findPreferredSeriesSelection(series)
    val defaultSeason = series.seasons.firstOrNull { it.number == defaultSelection?.seasonNumber }
        ?: series.seasons.firstOrNull()
    return TvPreferredEpisodeSelection(
        seasonNumber = defaultSeason?.number ?: 1,
        episodeNumber = defaultSelection?.episodeNumber
            ?: defaultSeason?.let(::findPreferredEpisodeNumber)
            ?: 1,
    )
}
