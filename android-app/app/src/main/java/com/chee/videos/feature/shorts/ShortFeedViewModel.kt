package com.chee.videos.feature.shorts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ShortFeedUiState(
    val loading: Boolean = false,
    val loaded: Boolean = false,
    val items: List<FeedVideoDto> = emptyList(),
    val errorMessage: String? = null,
    val detailErrorMessage: String? = null,
    val fitMode: VideoFitMode = VideoFitMode.FILL,
    val pausedByUserVideoIds: Set<String> = emptySet(),
    val detailSheetVideoId: String? = null,
    val detailByVideoId: Map<String, VideoDetailDto> = emptyMap(),
    val detailLoadingVideoIds: Set<String> = emptySet(),
    val actionBusyVideoIds: Set<String> = emptySet(),
)

@HiltViewModel
class ShortFeedViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(ShortFeedUiState())
    val uiState: StateFlow<ShortFeedUiState> = _uiState.asStateFlow()

    val accessToken = store.accessTokenFlow.stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = null,
    )

    init {
        viewModelScope.launch {
            store.shortFitModeFlow.collect { mode ->
                _uiState.update { it.copy(fitMode = mode) }
            }
        }
    }

    fun load(force: Boolean = false) {
        if (_uiState.value.loaded && !force) {
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, errorMessage = null) }
            videoRepository.fetchShortFeed()
                .onSuccess { items ->
                    val itemIDs = items.map { it.id }.toSet()
                    _uiState.update { state ->
                        state.copy(
                            loading = false,
                            loaded = true,
                            items = items,
                            errorMessage = null,
                            pausedByUserVideoIds = state.pausedByUserVideoIds.filter { id -> itemIDs.contains(id) }.toSet(),
                            detailByVideoId = state.detailByVideoId.filterKeys { id -> itemIDs.contains(id) },
                            detailLoadingVideoIds = state.detailLoadingVideoIds.filter { id -> itemIDs.contains(id) }.toSet(),
                            actionBusyVideoIds = state.actionBusyVideoIds.filter { id -> itemIDs.contains(id) }.toSet(),
                            detailSheetVideoId = state.detailSheetVideoId?.takeIf { id -> itemIDs.contains(id) },
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            loaded = true,
                            errorMessage = err.message ?: "短视频加载失败",
                        )
                    }
                }
        }
    }

    fun toggleFitMode() {
        val next = if (_uiState.value.fitMode == VideoFitMode.FILL) VideoFitMode.FIT else VideoFitMode.FILL
        _uiState.update { it.copy(fitMode = next) }
        viewModelScope.launch {
            store.saveShortFitMode(next)
        }
    }

    fun togglePauseByUser(videoId: String): Boolean {
        var pausedNow = false
        _uiState.update { state ->
            val nextSet = state.pausedByUserVideoIds.toMutableSet()
            if (nextSet.contains(videoId)) {
                nextSet.remove(videoId)
                pausedNow = false
            } else {
                nextSet.add(videoId)
                pausedNow = true
            }
            state.copy(pausedByUserVideoIds = nextSet)
        }
        return pausedNow
    }

    fun ensureDetailLoaded(videoId: String, force: Boolean = false, reportError: Boolean = false) {
        viewModelScope.launch {
            loadDetail(videoId, force = force, reportError = reportError)
        }
    }

    fun openDetailSheet(videoId: String) {
        _uiState.update {
            it.copy(
                detailSheetVideoId = videoId,
                detailErrorMessage = null,
            )
        }
        ensureDetailLoaded(videoId, reportError = true)
    }

    fun closeDetailSheet() {
        _uiState.update { it.copy(detailSheetVideoId = null, detailErrorMessage = null) }
    }

    fun toggleLike(videoId: String) {
        toggleAction(videoId) { videoRepository.toggleLike(videoId) }
    }

    fun toggleFavorite(videoId: String) {
        toggleAction(videoId) { videoRepository.toggleFavorite(videoId) }
    }

    fun toggleDislike(videoId: String) {
        toggleAction(videoId) { videoRepository.toggleDislike(videoId) }
    }

    private fun toggleAction(videoId: String, actionCall: suspend () -> Result<com.chee.videos.core.model.ActionTogglePayload>) {
        if (_uiState.value.actionBusyVideoIds.contains(videoId)) {
            return
        }
        viewModelScope.launch {
            val hasDetail = _uiState.value.detailByVideoId.containsKey(videoId)
            if (!hasDetail) {
                val loaded = loadDetail(videoId, force = true, reportError = true)
                if (!loaded) {
                    return@launch
                }
            }

            _uiState.update {
                it.copy(
                    actionBusyVideoIds = it.actionBusyVideoIds + videoId,
                    detailErrorMessage = null,
                )
            }

            actionCall()
                .onSuccess { result ->
                    _uiState.update { state ->
                        val detail = state.detailByVideoId[videoId]
                        if (detail == null) {
                            return@update state.copy(actionBusyVideoIds = state.actionBusyVideoIds - videoId)
                        }
                        val nextUserState = when (result.action) {
                            "like" -> detail.userState.copy(isLiked = result.enabled, isDisliked = false)
                            "favorite" -> detail.userState.copy(isFavorited = result.enabled)
                            "dislike" -> detail.userState.copy(isDisliked = result.enabled, isLiked = false, isFavorited = false)
                            else -> detail.userState
                        }
                        state.copy(
                            detailByVideoId = state.detailByVideoId + (videoId to detail.copy(userState = nextUserState)),
                            actionBusyVideoIds = state.actionBusyVideoIds - videoId,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            actionBusyVideoIds = it.actionBusyVideoIds - videoId,
                            detailErrorMessage = err.message ?: "操作失败",
                        )
                    }
                }
        }
    }

    private suspend fun loadDetail(videoId: String, force: Boolean, reportError: Boolean): Boolean {
        val state = _uiState.value
        if (!force && state.detailByVideoId.containsKey(videoId)) {
            return true
        }
        if (state.detailLoadingVideoIds.contains(videoId)) {
            return false
        }

        _uiState.update { it.copy(detailLoadingVideoIds = it.detailLoadingVideoIds + videoId) }
        val result = videoRepository.fetchDetail(videoId)

        return if (result.isSuccess) {
            val detail = result.getOrNull()!!
            _uiState.update {
                it.copy(
                    detailByVideoId = it.detailByVideoId + (videoId to detail),
                    detailLoadingVideoIds = it.detailLoadingVideoIds - videoId,
                    detailErrorMessage = if (it.detailSheetVideoId == videoId) null else it.detailErrorMessage,
                )
            }
            true
        } else {
            val err = result.exceptionOrNull()
            handleAuthError(err)
            _uiState.update {
                it.copy(
                    detailLoadingVideoIds = it.detailLoadingVideoIds - videoId,
                    detailErrorMessage = if (reportError) err?.message ?: "详情加载失败" else it.detailErrorMessage,
                )
            }
            false
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
