package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.ui.buildSubtitleTrackPreference
import com.chee.videos.core.ui.resolveSelectedSubtitleTrackByPreference
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.CancellationException
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
    val activeSeasonNumber: Int = 1,
    val activeEpisodeNumber: Int = 1,
    val selectedSubtitleTrackId: String? = null,
    val selectedAudioTrackId: String? = null,
    val selectedAudioPreference: TvTrackPreference? = null,
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
    val playbackPreparing: Boolean = false,
    val playbackBlockedMessage: String? = null,
    val episodeSwitchState: TvEpisodeSwitchUiState? = null,
    val errorMessage: String? = null,
)

sealed interface TvEpisodeSwitchUiState {
    data class Succeeded(
        val targetSeasonNumber: Int,
        val targetEpisodeNumber: Int,
        val message: String,
    ) : TvEpisodeSwitchUiState

    data class Preparing(
        val targetSeasonNumber: Int,
        val targetEpisodeNumber: Int,
    ) : TvEpisodeSwitchUiState

    data class Failed(
        val targetSeasonNumber: Int,
        val targetEpisodeNumber: Int,
        val message: String,
    ) : TvEpisodeSwitchUiState

    data class Canceled(
        val targetSeasonNumber: Int,
        val targetEpisodeNumber: Int,
        val message: String,
    ) : TvEpisodeSwitchUiState
}

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
        selectEpisode(season.number, episodeNumber)
    }

    fun selectEpisode(seasonNumber: Int, episodeNumber: Int) {
        val series = _uiState.value.series ?: return
        val season = series.seasons.firstOrNull { it.number == seasonNumber } ?: return
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
            currentSeasonNumber = state.activeSeasonNumber,
            currentEpisodeNumber = state.activeEpisodeNumber,
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

    fun clearEpisodeSwitchFeedback() {
        _uiState.update { state ->
            if (state.episodeSwitchState != null) {
                state.copy(episodeSwitchState = null)
            } else {
                state
            }
        }
    }

    fun cancelEpisodeSwitch() {
        playbackTargetRequestId += 1
        _uiState.update { state ->
            if (state.episodeSwitchState == null && state.selectedSeasonNumber == state.activeSeasonNumber && state.selectedEpisodeNumber == state.activeEpisodeNumber) {
                return@update state
            }
            state.copy(
                selectedSeasonNumber = state.activeSeasonNumber,
                selectedEpisodeNumber = state.activeEpisodeNumber,
                playbackPreparing = false,
                playbackBlockedMessage = null,
                episodeSwitchState = TvEpisodeSwitchUiState.Canceled(
                    targetSeasonNumber = state.selectedSeasonNumber,
                    targetEpisodeNumber = state.selectedEpisodeNumber,
                    message = "已取消切换到第 ${state.selectedEpisodeNumber} 集",
                ),
                startCurrentEpisodeFromBeginning = false,
            )
        }
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
            currentSeasonNumber = state.activeSeasonNumber,
            currentEpisodeNumber = state.activeEpisodeNumber,
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
        val selectedTrack = activeEpisode(_uiState.value)
            ?.subtitleTracks
            ?.firstOrNull { it.id == subtitleTrackId }
        val preference = buildSubtitleTrackPreference(selectedTrack)
        _uiState.update { it.copy(selectedSubtitleTrackId = subtitleTrackId ?: "") }
        viewModelScope.launch {
            repository.saveTvSubtitlePreference(currentVideoId, preference)
        }
    }

    fun selectAudioTrack(audioTrackId: String?, preference: TvTrackPreference?) {
        val currentVideoId = _uiState.value.currentVideoId
        if (currentVideoId.isBlank()) {
            return
        }
        _uiState.update {
            it.copy(
                selectedAudioTrackId = audioTrackId ?: "",
                selectedAudioPreference = preference,
            )
        }
        viewModelScope.launch {
            repository.saveTvAudioPreference(currentVideoId, preference)
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
                        ?: resolvedSeason?.episodes?.firstOrNull { isTvEpisodePlayableForPlayback(it) }
                        ?: resolvedSeason?.episodes?.firstOrNull()
                    _uiState.value = TvSeriesPlayerUiState(
                        loading = false,
                        series = series,
                        baseUrl = baseUrl,
                        selectedSeasonNumber = resolvedSeason?.number ?: 1,
                        selectedEpisodeNumber = resolvedEpisode?.number ?: 1,
                        activeSeasonNumber = resolvedSeason?.number ?: 1,
                        activeEpisodeNumber = resolvedEpisode?.number ?: 1,
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
        val state = _uiState.value
        val episode = selectedEpisode(state)
        if (episode == null || !episode.playable || episode.videoId.isBlank()) {
            val hasActivePlayback = state.currentSourceUrl.isNotBlank()
            _uiState.update { currentState ->
                if (hasActivePlayback) {
                    currentState.copy(
                        selectedSeasonNumber = currentState.activeSeasonNumber,
                        selectedEpisodeNumber = currentState.activeEpisodeNumber,
                        playbackPreparing = false,
                        playbackBlockedMessage = null,
                        episodeSwitchState = TvEpisodeSwitchUiState.Failed(
                            targetSeasonNumber = state.selectedSeasonNumber,
                            targetEpisodeNumber = state.selectedEpisodeNumber,
                            message = "当前分集暂无可播放视频",
                        ),
                        startCurrentEpisodeFromBeginning = false,
                    )
                } else {
                    currentState.copy(
                        activeSeasonNumber = episode?.let { state.selectedSeasonNumber } ?: currentState.activeSeasonNumber,
                        activeEpisodeNumber = episode?.let { state.selectedEpisodeNumber } ?: currentState.activeEpisodeNumber,
                        currentVideoId = episode?.videoId.orEmpty(),
                        currentSourceUrl = "",
                        selectedSubtitleTrackId = null,
                        selectedAudioTrackId = null,
                        selectedAudioPreference = null,
                        canPlayCurrentEpisode = false,
                        playbackPreparing = false,
                        playbackBlockedMessage = "当前分集暂无可播放视频",
                        episodeSwitchState = null,
                    )
                }
            }
            return
        }
        val candidateDecision = resolveTvPlaybackCandidateDecision(episode.metadata)
        if (!candidateDecision.allowed) {
            val hasActivePlayback = state.currentSourceUrl.isNotBlank()
            _uiState.update { currentState ->
                if (hasActivePlayback) {
                    currentState.copy(
                        selectedSeasonNumber = currentState.activeSeasonNumber,
                        selectedEpisodeNumber = currentState.activeEpisodeNumber,
                        playbackPreparing = false,
                        playbackBlockedMessage = null,
                        episodeSwitchState = TvEpisodeSwitchUiState.Failed(
                            targetSeasonNumber = state.selectedSeasonNumber,
                            targetEpisodeNumber = state.selectedEpisodeNumber,
                            message = candidateDecision.blockMessage ?: "当前分集暂无可播放视频",
                        ),
                        startCurrentEpisodeFromBeginning = false,
                    )
                } else {
                    currentState.copy(
                        activeSeasonNumber = state.selectedSeasonNumber,
                        activeEpisodeNumber = state.selectedEpisodeNumber,
                        currentVideoId = episode.videoId,
                        currentSourceUrl = "",
                        selectedSubtitleTrackId = null,
                        selectedAudioTrackId = null,
                        selectedAudioPreference = null,
                        canPlayCurrentEpisode = false,
                        playbackPreparing = false,
                        playbackBlockedMessage = candidateDecision.blockMessage,
                        episodeSwitchState = null,
                    )
                }
            }
            return
        }
        val hasActivePlayback = state.currentSourceUrl.isNotBlank()
        _uiState.update {
            if (hasActivePlayback) {
                it.copy(
                    playbackPreparing = true,
                    playbackBlockedMessage = null,
                    episodeSwitchState = TvEpisodeSwitchUiState.Preparing(
                        targetSeasonNumber = state.selectedSeasonNumber,
                        targetEpisodeNumber = state.selectedEpisodeNumber,
                    ),
                )
            } else {
                it.copy(
                    activeSeasonNumber = state.selectedSeasonNumber,
                    activeEpisodeNumber = state.selectedEpisodeNumber,
                    currentVideoId = "",
                    currentSourceUrl = "",
                    selectedSubtitleTrackId = null,
                    selectedAudioTrackId = null,
                    selectedAudioPreference = null,
                    canPlayCurrentEpisode = false,
                    playbackPreparing = true,
                    playbackBlockedMessage = null,
                    episodeSwitchState = TvEpisodeSwitchUiState.Preparing(
                        targetSeasonNumber = state.selectedSeasonNumber,
                        targetEpisodeNumber = state.selectedEpisodeNumber,
                    ),
                )
            }
        }
        viewModelScope.launch {
            val sourceResult = try {
                val sourceUrl = repository.buildSourceUrl(
                    episode.videoId,
                    resolveTvPlaybackSourceProfile(episode.metadata),
                )
                val preferredSubtitleTrackId = resolveSelectedSubtitleTrackByPreference(
                    tracks = episode.subtitleTracks,
                    preference = repository.readTvSubtitlePreference(episode.videoId),
                )?.id
                val preferredAudioPreference = repository.readTvAudioPreference(episode.videoId)
                TvSeriesPlaybackSourceResult.Ready(
                    sourceUrl = sourceUrl,
                    preferredSubtitleTrackId = preferredSubtitleTrackId,
                    preferredAudioPreference = preferredAudioPreference,
                )
            } catch (err: CancellationException) {
                throw err
            } catch (err: Exception) {
                TvSeriesPlaybackSourceResult.Failed(err.message ?: "播放源准备失败，请重试")
            }
            if (requestId != playbackTargetRequestId) {
                return@launch
            }
            _uiState.update {
                val selected = selectedEpisode(it)
                if (selected?.videoId != episode.videoId) {
                    return@update it
                }
                when (sourceResult) {
                    is TvSeriesPlaybackSourceResult.Ready -> it.copy(
                        activeSeasonNumber = it.selectedSeasonNumber,
                        activeEpisodeNumber = it.selectedEpisodeNumber,
                        currentVideoId = episode.videoId,
                        currentSourceUrl = sourceResult.sourceUrl,
                        selectedSubtitleTrackId = sourceResult.preferredSubtitleTrackId,
                        selectedAudioTrackId = null,
                        selectedAudioPreference = sourceResult.preferredAudioPreference,
                        canPlayCurrentEpisode = true,
                        playbackPreparing = false,
                        playbackBlockedMessage = null,
                        episodeSwitchState = TvEpisodeSwitchUiState.Succeeded(
                            targetSeasonNumber = it.selectedSeasonNumber,
                            targetEpisodeNumber = it.selectedEpisodeNumber,
                            message = "已切到第 ${it.selectedEpisodeNumber} 集",
                        ),
                    )

                    is TvSeriesPlaybackSourceResult.Failed -> {
                        val hadActivePlayback = it.currentSourceUrl.isNotBlank()
                        if (hadActivePlayback) {
                            it.copy(
                                selectedSeasonNumber = it.activeSeasonNumber,
                                selectedEpisodeNumber = it.activeEpisodeNumber,
                                playbackPreparing = false,
                                playbackBlockedMessage = null,
                                episodeSwitchState = TvEpisodeSwitchUiState.Failed(
                                    targetSeasonNumber = state.selectedSeasonNumber,
                                    targetEpisodeNumber = state.selectedEpisodeNumber,
                                    message = sourceResult.message,
                                ),
                                startCurrentEpisodeFromBeginning = false,
                            )
                        } else {
                            it.copy(
                                activeSeasonNumber = it.selectedSeasonNumber,
                                activeEpisodeNumber = it.selectedEpisodeNumber,
                                currentVideoId = episode.videoId,
                                currentSourceUrl = "",
                                selectedSubtitleTrackId = null,
                                selectedAudioTrackId = null,
                                selectedAudioPreference = null,
                                canPlayCurrentEpisode = false,
                                playbackPreparing = false,
                                playbackBlockedMessage = sourceResult.message,
                                episodeSwitchState = null,
                                startCurrentEpisodeFromBeginning = false,
                            )
                        }
                    }
                }
            }
        }
    }
}

private sealed interface TvSeriesPlaybackSourceResult {
    data class Ready(
        val sourceUrl: String,
        val preferredSubtitleTrackId: String?,
        val preferredAudioPreference: TvTrackPreference?,
    ) : TvSeriesPlaybackSourceResult

    data class Failed(val message: String) : TvSeriesPlaybackSourceResult
}

internal fun selectedSeason(state: TvSeriesPlayerUiState): TvSeasonUiModel? {
    val series = state.series ?: return null
    return series.seasons.firstOrNull { it.number == state.selectedSeasonNumber }
}

internal fun selectedEpisode(state: TvSeriesPlayerUiState): TvEpisodeUiModel? {
    val season = selectedSeason(state) ?: return null
    return season.episodes.firstOrNull { it.number == state.selectedEpisodeNumber }
}

internal fun activeSeason(state: TvSeriesPlayerUiState): TvSeasonUiModel? {
    val series = state.series ?: return null
    return series.seasons.firstOrNull { it.number == state.activeSeasonNumber }
}

internal fun activeEpisode(state: TvSeriesPlayerUiState): TvEpisodeUiModel? {
    val season = activeSeason(state) ?: return null
    return season.episodes.firstOrNull { it.number == state.activeEpisodeNumber }
}

internal fun TvSeriesPlayerUiState.hasNextPlayableEpisode(): Boolean {
    val currentSeries = series ?: return false
    return resolveNextPlayableEpisode(
        series = currentSeries,
        currentSeasonNumber = activeSeasonNumber,
        currentEpisodeNumber = activeEpisodeNumber,
    ) != null
}
