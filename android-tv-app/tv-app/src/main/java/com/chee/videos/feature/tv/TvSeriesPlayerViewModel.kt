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

private val supportedPlaybackSpeeds = listOf(1f, 1.25f, 1.5f, 2f)

data class TvSeriesPlayerUiState(
    val loading: Boolean = true,
    val series: TvSeriesUiModel? = null,
    val baseUrl: String = "",
    val selectedSeasonNumber: Int = 1,
    val selectedEpisodeNumber: Int = 1,
    val selectedSubtitleTrackId: String? = null,
    val selectedAudioTrackId: String? = null,
    val playbackSpeed: Float = 1f,
    val tvSeekStepSeconds: Int = TvPlaybackSeekStepSetting.defaultSeconds,
    val autoplayEnabled: Boolean = TvSeriesAutoplaySetting.DEFAULT_ENABLED,
    val autoplayCanceledForCurrentEpisode: Boolean = false,
    val pendingEndOverlayKind: TvEndOverlayKind? = null,
    val startCurrentEpisodeFromBeginning: Boolean = false,
    val selectorVisible: Boolean = false,
    val currentVideoId: String = "",
    val currentSourceUrl: String = "",
    val canPlayCurrentEpisode: Boolean = false,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvSeriesPlayerViewModel @Inject constructor(
    private val repository: TvRepository,
    private val savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val seriesId = decodeTvRouteArg(savedStateHandle.get<String>(TvSeriesIdArg))
    private val requestedSeason = savedStateHandle.get<Int>(TvSeasonArg)
    private val requestedEpisode = savedStateHandle.get<Int>(TvEpisodeArg)

    private val _uiState = MutableStateFlow(TvSeriesPlayerUiState())
    val uiState: StateFlow<TvSeriesPlayerUiState> = _uiState.asStateFlow()
    private var playbackTargetRequestId: Long = 0

    init {
        load()
    }

    fun retry() {
        load()
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
        updateSelectedEpisode(season.number, findPreferredEpisodeNumber(season))
    }

    fun selectEpisode(episodeNumber: Int) {
        val season = selectedSeason(_uiState.value) ?: return
        if (season.episodes.none { it.number == episodeNumber }) {
            return
        }
        updateSelectedEpisode(season.number, episodeNumber)
    }

    fun nextEpisode() {
        val state = _uiState.value
        val series = state.series ?: return
        val next = resolveNextPlayableEpisode(
            series = series,
            currentSeasonNumber = state.selectedSeasonNumber,
            currentEpisodeNumber = state.selectedEpisodeNumber,
        ) ?: return
        updateSelectedEpisode(next.seasonNumber, next.episodeNumber, startFromBeginning = true)
    }

    fun advanceToNextEpisodeFromAutoplay() {
        nextEpisode()
    }

    fun cancelAutoplayForCurrentEpisode() {
        _uiState.update { state ->
            state.copy(autoplayCanceledForCurrentEpisode = true)
        }
    }

    fun showEndOverlay(kind: TvEndOverlayKind) {
        _uiState.update { state -> state.copy(pendingEndOverlayKind = kind) }
    }

    fun dismissEndOverlay() {
        _uiState.update { state -> state.copy(pendingEndOverlayKind = null) }
    }

    fun resolveEndOverlayKind(): TvEndOverlayKind =
        if (_uiState.value.hasNextPlayableEpisode()) {
            TvEndOverlayKind.CURRENT_FINISHED
        } else {
            TvEndOverlayKind.SERIES_FINISHED
        }

    fun nextEpisodeRef(): TvNextEpisodeRef? {
        val state = _uiState.value
        val series = state.series ?: return null
        return resolveNextPlayableEpisode(
            series = series,
            currentSeasonNumber = state.selectedSeasonNumber,
            currentEpisodeNumber = state.selectedEpisodeNumber,
        )
    }

    fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        if (videoId.isBlank() || watchSeconds <= 0) {
            return
        }
        viewModelScope.launch {
            repository.reportHistory(videoId, watchSeconds, completed)
        }
    }

    fun selectSubtitleTrack(subtitleTrackId: String?) {
        val currentVideoId = _uiState.value.currentVideoId
        if (currentVideoId.isBlank()) {
            return
        }
        _uiState.update { it.copy(selectedSubtitleTrackId = subtitleTrackId ?: "") }
        viewModelScope.launch {
            repository.saveTvSubtitlePreference(currentVideoId, subtitleTrackId ?: "")
        }
    }

    fun selectAudioTrack(audioTrackId: String?) {
        val currentVideoId = _uiState.value.currentVideoId
        if (currentVideoId.isBlank()) {
            return
        }
        _uiState.update { it.copy(selectedAudioTrackId = audioTrackId ?: "") }
        viewModelScope.launch {
            repository.saveTvAudioPreference(currentVideoId, audioTrackId ?: "")
        }
    }

    private fun load() {
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, errorMessage = null) }
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            val tvSeekStepSeconds = TvPlaybackSeekStepSetting.normalize(repository.readTvSeekStepSeconds())
            val autoplayEnabled = TvSeriesAutoplaySetting.parse(repository.readTvSeriesAutoplayEnabled())
            repository.fetchSeriesDetail(seriesId)
                .onSuccess { dto ->
                    val series = tvSeriesDetailToUiModel(dto)
                    val preferredSelection = if (requestedSeason != null && requestedEpisode != null) {
                        TvPreferredEpisodeSelection(
                            seasonNumber = requestedSeason.coerceAtLeast(1),
                            episodeNumber = requestedEpisode.coerceAtLeast(1),
                        )
                    } else {
                        findPreferredSeriesSelection(series)
                    }
                    val resolvedSeason = series.seasons.firstOrNull { it.number == preferredSelection?.seasonNumber }
                        ?: series.seasons.firstOrNull()
                    val resolvedEpisode = resolvedSeason?.episodes?.firstOrNull { it.number == preferredSelection?.episodeNumber }
                        ?: resolvedSeason?.episodes?.firstOrNull { it.playable }
                        ?: resolvedSeason?.episodes?.firstOrNull()
                    _uiState.value = TvSeriesPlayerUiState(
                        loading = false,
                        series = series,
                        baseUrl = baseUrl,
                        selectedSeasonNumber = resolvedSeason?.number ?: 1,
                        selectedEpisodeNumber = resolvedEpisode?.number ?: 1,
                        playbackSpeed = 1f,
                        tvSeekStepSeconds = tvSeekStepSeconds,
                        autoplayEnabled = autoplayEnabled,
                        autoplayCanceledForCurrentEpisode = false,
                        pendingEndOverlayKind = null,
                        startCurrentEpisodeFromBeginning = false,
                        selectorVisible = false,
                        errorMessage = null,
                    )
                    updatePlaybackTarget()
                }
                .onFailure { error ->
                    _uiState.value = TvSeriesPlayerUiState(
                        loading = false,
                        baseUrl = baseUrl,
                        tvSeekStepSeconds = tvSeekStepSeconds,
                        autoplayEnabled = autoplayEnabled,
                        startCurrentEpisodeFromBeginning = false,
                        errorMessage = error.message ?: "电视剧播放页加载失败",
                    )
                }
        }
    }

    private fun updateSelectedEpisode(
        seasonNumber: Int,
        episodeNumber: Int,
        startFromBeginning: Boolean = false,
    ) {
        val currentState = _uiState.value
        if (currentState.selectedSeasonNumber == seasonNumber && currentState.selectedEpisodeNumber == episodeNumber) {
            return
        }
        _uiState.update {
            it.copy(
                selectedSeasonNumber = seasonNumber,
                selectedEpisodeNumber = episodeNumber,
                autoplayCanceledForCurrentEpisode = false,
                pendingEndOverlayKind = null,
                startCurrentEpisodeFromBeginning = startFromBeginning,
            )
        }
        updatePlaybackTarget()
    }

    private fun updatePlaybackTarget() {
        val requestId = ++playbackTargetRequestId
        val episode = selectedEpisode(_uiState.value)
        if (episode == null || !episode.playable || episode.videoId.isBlank()) {
            _uiState.update {
                it.copy(
                    currentVideoId = "",
                    currentSourceUrl = "",
                    selectedSubtitleTrackId = null,
                    selectedAudioTrackId = null,
                    canPlayCurrentEpisode = false,
                )
            }
            return
        }
        _uiState.update {
            it.copy(
                currentVideoId = "",
                currentSourceUrl = "",
                selectedSubtitleTrackId = null,
                selectedAudioTrackId = null,
                canPlayCurrentEpisode = false,
            )
        }
        viewModelScope.launch {
            val sourceUrl = repository.buildSourceUrl(episode.videoId)
            val preferredSubtitleTrackId = repository.readTvSubtitlePreference(episode.videoId)
            val preferredAudioTrackId = repository.readTvAudioPreference(episode.videoId)
            if (requestId != playbackTargetRequestId) {
                return@launch
            }
            _uiState.update {
                val selected = selectedEpisode(it)
                if (selected?.videoId != episode.videoId) {
                    return@update it
                }
                it.copy(
                    currentVideoId = episode.videoId,
                    currentSourceUrl = sourceUrl,
                    selectedSubtitleTrackId = preferredSubtitleTrackId,
                    selectedAudioTrackId = preferredAudioTrackId,
                    canPlayCurrentEpisode = true,
                )
            }
        }
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

internal fun TvSeriesPlayerUiState.hasNextPlayableEpisode(): Boolean {
    val currentSeries = series ?: return false
    return resolveNextPlayableEpisode(
        series = currentSeries,
        currentSeasonNumber = selectedSeasonNumber,
        currentEpisodeNumber = selectedEpisodeNumber,
    ) != null
}
