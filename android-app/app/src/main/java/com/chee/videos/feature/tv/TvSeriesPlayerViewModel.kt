package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update

private val supportedPlaybackSpeeds = listOf(1f, 1.25f, 1.5f, 2f)

data class TvSeriesPlayerUiState(
    val loading: Boolean = true,
    val series: TvSeriesUiModel? = null,
    val selectedSeasonNumber: Int = 1,
    val selectedEpisodeNumber: Int = 1,
    val isPlaying: Boolean = true,
    val playbackSpeed: Float = 1f,
    val selectorVisible: Boolean = false,
)

@HiltViewModel
class TvSeriesPlayerViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
) : ViewModel() {

    private val _uiState = MutableStateFlow(TvSeriesPlayerUiState())
    val uiState: StateFlow<TvSeriesPlayerUiState> = _uiState.asStateFlow()

    init {
        val seriesId = savedStateHandle.get<String>(TvSeriesIdArg).orEmpty()
        val requestedSeason = savedStateHandle.get<Int>(TvSeasonArg) ?: 1
        val requestedEpisode = savedStateHandle.get<Int>(TvEpisodeArg) ?: 1
        val series = TvMockData.findSeries(seriesId) ?: TvMockData.allSeries().firstOrNull()
        val firstSeason = series?.seasons?.firstOrNull()
        val resolvedSeason = series?.seasons?.firstOrNull { it.number == requestedSeason.coerceAtLeast(1) } ?: firstSeason
        val resolvedEpisodeNumber = resolvedSeason?.episodes
            ?.firstOrNull { it.number == requestedEpisode.coerceAtLeast(1) }
            ?.number ?: resolvedSeason?.episodes?.firstOrNull()?.number ?: 1
        _uiState.value = TvSeriesPlayerUiState(
            loading = false,
            series = series,
            selectedSeasonNumber = resolvedSeason?.number ?: 1,
            selectedEpisodeNumber = resolvedEpisodeNumber,
            isPlaying = true,
            playbackSpeed = 1f,
            selectorVisible = false,
        )
    }

    fun togglePlay() {
        _uiState.update { state -> state.copy(isPlaying = !state.isPlaying) }
    }

    fun cycleSpeed() {
        _uiState.update { state ->
            val currentIndex = supportedPlaybackSpeeds.indexOf(state.playbackSpeed).coerceAtLeast(0)
            val nextSpeed = supportedPlaybackSpeeds[(currentIndex + 1) % supportedPlaybackSpeeds.size]
            state.copy(playbackSpeed = nextSpeed)
        }
    }

    fun setSelectorVisible(visible: Boolean) {
        _uiState.update { state -> state.copy(selectorVisible = visible) }
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
        val season = selectedSeason(_uiState.value) ?: return
        if (season.episodes.none { it.number == episodeNumber }) {
            return
        }
        _uiState.update { state ->
            state.copy(selectedEpisodeNumber = episodeNumber)
        }
    }

    fun nextEpisode() {
        val state = _uiState.value
        val season = selectedSeason(state) ?: return
        val currentIndex = season.episodes.indexOfFirst { it.number == state.selectedEpisodeNumber }
        val next = season.episodes.getOrNull(currentIndex + 1) ?: return
        _uiState.update { it.copy(selectedEpisodeNumber = next.number, isPlaying = true) }
    }
}

internal fun selectedSeason(state: TvSeriesPlayerUiState): TvSeasonUiModel? {
    val series = state.series ?: return null
    return series.seasons.firstOrNull { it.number == state.selectedSeasonNumber }
}

internal fun selectedEpisode(state: TvSeriesPlayerUiState): TvEpisodeUiModel? {
    val season = selectedSeason(state) ?: return null
    return season.episodes.firstOrNull { it.number == state.selectedEpisodeNumber }
}
