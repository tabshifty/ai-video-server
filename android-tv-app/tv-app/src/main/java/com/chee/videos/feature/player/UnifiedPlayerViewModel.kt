package com.chee.videos.feature.player

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class PlayerVideoItem(
    val id: String,
    val title: String,
    val thumbnailPath: String?,
    val type: String,
    val duration: Int,
)

data class UnifiedPlayerUiState(
    val loading: Boolean = false,
    val loaded: Boolean = false,
    val source: String = "",
    val startVideoId: String = "",
    val preferredPlaybackProfile: String = "",
    val startIndex: Int = 0,
    val items: List<PlayerVideoItem> = emptyList(),
    val shortFitMode: VideoFitMode = VideoFitMode.FILL,
    val playbackMode: ShortPlaybackMode = ShortPlaybackMode.LOOP_ONE,
    val errorMessage: String? = null,
    val detailByVideoId: Map<String, VideoDetailDto> = emptyMap(),
    val detailLoadingVideoIds: Set<String> = emptySet(),
)

@HiltViewModel
class UnifiedPlayerViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(UnifiedPlayerUiState())
    val uiState: StateFlow<UnifiedPlayerUiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            store.unifiedShortFitModeFlow.collect { mode ->
                _uiState.update { it.copy(shortFitMode = mode) }
            }
        }
        viewModelScope.launch {
            store.shortPlaybackModeFlow.collect { mode ->
                _uiState.update { it.copy(playbackMode = mode) }
            }
        }
    }

    fun load(source: String, startVideoId: String, force: Boolean = false) {
        val current = _uiState.value
        if (
            !force &&
            current.loaded &&
            current.source == source &&
            current.startVideoId == startVideoId &&
            current.errorMessage.isNullOrBlank() &&
            current.items.isNotEmpty()
        ) {
            return
        }

        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    loading = true,
                    loaded = true,
                    source = source,
                    startVideoId = startVideoId,
                    preferredPlaybackProfile = videoRepository.preferredLongFormPlaybackProfile().wireValue,
                    errorMessage = null,
                )
            }

            val result = when (source) {
                "history" -> videoRepository.fetchContinueHistory(page = 1, limit = 120)
                    .map { payload ->
                        payload.items.map { row ->
                            PlayerVideoItem(
                                id = row.videoId,
                                title = row.title,
                                thumbnailPath = row.thumbnailPath,
                                type = row.type,
                                duration = row.duration,
                            )
                        }
                    }

                "favorite" -> videoRepository.fetchFavoritedVideos(page = 1, pageSize = 120)
                    .map { list ->
                        list.map { row ->
                            PlayerVideoItem(
                                id = row.id,
                                title = row.title,
                                thumbnailPath = row.thumbnailPath,
                                type = row.type,
                                duration = row.duration,
                            )
                        }
                    }

                "like" -> videoRepository.fetchLikedVideos(page = 1, pageSize = 120)
                    .map { list ->
                        list.map { row ->
                            PlayerVideoItem(
                                id = row.id,
                                title = row.title,
                                thumbnailPath = row.thumbnailPath,
                                type = row.type,
                                duration = row.duration,
                            )
                        }
                    }

                else -> Result.success(emptyList())
            }

            result
                .onSuccess { items ->
                    val index = items.indexOfFirst { it.id == startVideoId }.takeIf { it >= 0 } ?: 0
                    _uiState.update {
                        it.copy(
                            loading = false,
                            items = items,
                            startIndex = index,
                            errorMessage = null,
                            detailByVideoId = it.detailByVideoId.filterKeys { key -> items.any { row -> row.id == key } },
                            detailLoadingVideoIds = it.detailLoadingVideoIds.filter { id -> items.any { row -> row.id == id } }.toSet(),
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            errorMessage = err.message ?: "播放列表加载失败",
                            items = emptyList(),
                            startIndex = 0,
                        )
                    }
                }
        }
    }

    fun ensureDetailLoaded(videoId: String, force: Boolean = false) {
        val state = _uiState.value
        if (!force && state.detailByVideoId.containsKey(videoId)) {
            return
        }
        if (state.detailLoadingVideoIds.contains(videoId)) {
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(detailLoadingVideoIds = it.detailLoadingVideoIds + videoId) }
            videoRepository.fetchDetail(videoId)
                .onSuccess { detail ->
                    _uiState.update {
                        it.copy(
                            detailByVideoId = it.detailByVideoId + (videoId to detail),
                            detailLoadingVideoIds = it.detailLoadingVideoIds - videoId,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(detailLoadingVideoIds = it.detailLoadingVideoIds - videoId)
                    }
                }
        }
    }

    fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        if (videoId.isBlank() || watchSeconds <= 0) {
            return
        }
        viewModelScope.launch {
            videoRepository.reportHistory(videoId, watchSeconds, completed)
        }
    }

    fun toggleShortFitMode() {
        val next = if (_uiState.value.shortFitMode == VideoFitMode.FILL) VideoFitMode.FIT else VideoFitMode.FILL
        _uiState.update { it.copy(shortFitMode = next) }
        viewModelScope.launch {
            store.saveUnifiedShortFitMode(next)
        }
    }

    fun toggleShortPlaybackMode() {
        val next = if (_uiState.value.playbackMode == ShortPlaybackMode.LOOP_ONE) ShortPlaybackMode.AUTO_NEXT else ShortPlaybackMode.LOOP_ONE
        _uiState.update { it.copy(playbackMode = next) }
        viewModelScope.launch {
            store.saveShortPlaybackMode(next)
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch {
                authRepository.logoutLocal()
            }
        }
    }
}
