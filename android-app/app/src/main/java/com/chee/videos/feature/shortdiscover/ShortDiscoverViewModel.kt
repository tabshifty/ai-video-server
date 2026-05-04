package com.chee.videos.feature.shortdiscover

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.VideoListItemDto
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

data class ShortDiscoverUiState(
    val mode: String = "",
    val value: String = "",
    val title: String = "",
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val totalCount: Int = 0,
    val items: List<VideoListItemDto> = emptyList(),
    val fitMode: VideoFitMode = VideoFitMode.FILL,
    val playbackMode: ShortPlaybackMode = ShortPlaybackMode.LOOP_ONE,
    val errorMessage: String? = null,
)

@HiltViewModel
class ShortDiscoverViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(ShortDiscoverUiState())
    val uiState: StateFlow<ShortDiscoverUiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            store.shortDiscoverFitModeFlow.collect { mode ->
                _uiState.update { it.copy(fitMode = mode) }
            }
        }
        viewModelScope.launch {
            store.shortPlaybackModeFlow.collect { mode ->
                _uiState.update { it.copy(playbackMode = mode) }
            }
        }
    }

    fun initialize(mode: String, value: String, title: String) {
        val normalizedMode = mode.trim().lowercase()
        val normalizedValue = value.trim()
        val normalizedTitle = title.trim().ifBlank {
            if (normalizedMode == "tag") "#$normalizedValue" else normalizedValue
        }
        val current = _uiState.value
        if (
            current.loaded &&
            current.mode == normalizedMode &&
            current.value == normalizedValue &&
            current.title == normalizedTitle
        ) {
            return
        }

        _uiState.value = ShortDiscoverUiState(
            mode = normalizedMode,
            value = normalizedValue,
            title = normalizedTitle,
            loading = true,
        )
        loadPage(page = 1, append = false)
    }

    fun retry() {
        val state = _uiState.value
        if (state.mode.isBlank() || state.value.isBlank()) {
            return
        }
        _uiState.update { it.copy(loading = true, errorMessage = null) }
        loadPage(page = 1, append = false)
    }

    fun loadMoreIfNeeded(currentIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.loadingMore || state.items.isEmpty()) {
            return
        }
        if (currentIndex < state.items.size - 6) {
            return
        }
        if (state.totalCount > 0 && state.items.size >= state.totalCount) {
            return
        }
        loadPage(page = state.page + 1, append = true)
    }

    fun toggleFitMode() {
        val next = if (_uiState.value.fitMode == VideoFitMode.FILL) VideoFitMode.FIT else VideoFitMode.FILL
        _uiState.update { it.copy(fitMode = next) }
        viewModelScope.launch {
            store.saveShortDiscoverFitMode(next)
        }
    }

    fun togglePlaybackMode() {
        val next = if (_uiState.value.playbackMode == ShortPlaybackMode.LOOP_ONE) ShortPlaybackMode.AUTO_NEXT else ShortPlaybackMode.LOOP_ONE
        _uiState.update { it.copy(playbackMode = next) }
        viewModelScope.launch {
            store.saveShortPlaybackMode(next)
        }
    }

    private fun loadPage(page: Int, append: Boolean) {
        val state = _uiState.value
        if (state.mode.isBlank() || state.value.isBlank()) {
            return
        }
        viewModelScope.launch {
            if (append) {
                _uiState.update { it.copy(loadingMore = true, errorMessage = null) }
            }
            videoRepository.fetchShortDiscover(
                mode = state.mode,
                value = state.value,
                page = page,
                pageSize = 24,
            ).onSuccess { payload ->
                _uiState.update {
                    val mergedItems = if (append) {
                        (it.items + payload.items).distinctBy { row -> row.id }
                    } else {
                        payload.items.distinctBy { row -> row.id }
                    }
                    it.copy(
                        loading = false,
                        loadingMore = false,
                        loaded = true,
                        page = payload.page.coerceAtLeast(page),
                        totalCount = payload.totalCount.coerceAtLeast(mergedItems.size),
                        items = mergedItems,
                        errorMessage = null,
                    )
                }
            }.onFailure { err ->
                handleAuthError(err)
                _uiState.update {
                    it.copy(
                        loading = false,
                        loadingMore = false,
                        loaded = true,
                        errorMessage = err.message ?: "短视频发现加载失败",
                    )
                }
            }
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
