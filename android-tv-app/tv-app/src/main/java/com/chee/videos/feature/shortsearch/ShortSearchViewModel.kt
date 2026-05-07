package com.chee.videos.feature.shortsearch

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ShortSearchUiState(
    val queryInput: String = "",
    val activeQuery: String = "",
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val totalCount: Int = 0,
    val items: List<VideoListItemDto> = emptyList(),
    val fitMode: VideoFitMode = VideoFitMode.FILL,
    val playbackMode: ShortPlaybackMode = ShortPlaybackMode.LOOP_ONE,
    val playingVideoId: String? = null,
    val errorMessage: String? = null,
)

internal fun normalizeShortSearchQuery(query: String): String = query.trim()

internal fun resetShortSearchForQuery(state: ShortSearchUiState, query: String): ShortSearchUiState {
    return state.copy(
        activeQuery = query,
        loading = true,
        loadingMore = false,
        loaded = false,
        page = 0,
        totalCount = 0,
        items = emptyList(),
        errorMessage = null,
    )
}

internal fun mergeShortSearchItems(existing: List<VideoListItemDto>, incoming: List<VideoListItemDto>): List<VideoListItemDto> {
    return (existing + incoming).distinctBy { it.id }
}

@HiltViewModel
class ShortSearchViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(ShortSearchUiState())
    val uiState: StateFlow<ShortSearchUiState> = _uiState.asStateFlow()
    private var searchJob: Job? = null

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

    fun onQueryInputChange(value: String) {
        _uiState.update { it.copy(queryInput = value) }
        searchJob?.cancel()
        searchJob = viewModelScope.launch {
            delay(350)
            val query = normalizeShortSearchQuery(value)
            if (query == _uiState.value.activeQuery) return@launch
            if (query.isBlank()) {
                _uiState.update {
                    it.copy(activeQuery = "", loaded = false, loading = false, items = emptyList(), page = 0, totalCount = 0, errorMessage = null)
                }
            } else {
                search(query)
            }
        }
    }

    fun retry() {
        val query = _uiState.value.activeQuery.ifBlank { _uiState.value.queryInput.trim() }
        if (query.isNotBlank()) search(query, force = true)
    }

    fun loadMoreIfNeeded(currentIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.loadingMore || state.items.isEmpty()) return
        if (currentIndex < state.items.lastIndex - 5) return
        if (state.totalCount > 0 && state.items.size >= state.totalCount) return
        loadPage(state.activeQuery, page = state.page + 1, append = true)
    }

    fun enterPlayer(videoId: String) {
        _uiState.update { it.copy(playingVideoId = videoId) }
    }

    fun closePlayer() {
        _uiState.update { it.copy(playingVideoId = null) }
    }

    fun toggleFitMode() {
        val next = if (_uiState.value.fitMode == VideoFitMode.FILL) VideoFitMode.FIT else VideoFitMode.FILL
        _uiState.update { it.copy(fitMode = next) }
        viewModelScope.launch { store.saveShortDiscoverFitMode(next) }
    }

    fun togglePlaybackMode() {
        val next = if (_uiState.value.playbackMode == ShortPlaybackMode.LOOP_ONE) ShortPlaybackMode.AUTO_NEXT else ShortPlaybackMode.LOOP_ONE
        _uiState.update { it.copy(playbackMode = next) }
        viewModelScope.launch { store.saveShortPlaybackMode(next) }
    }

    private fun search(query: String, force: Boolean = false) {
        val state = _uiState.value
        if (!force && state.loading) return
        _uiState.update { resetShortSearchForQuery(it, query) }
        loadPage(query, page = 1, append = false)
    }

    private fun loadPage(query: String, page: Int, append: Boolean) {
        if (query.isBlank()) return
        viewModelScope.launch {
            if (append) _uiState.update { it.copy(loadingMore = true, errorMessage = null) }
            videoRepository.searchShort(query = query, page = page, pageSize = 24)
                .onSuccess { payload ->
                    _uiState.update {
                        val merged = if (append) mergeShortSearchItems(it.items, payload.items) else payload.items.distinctBy { row -> row.id }
                        it.copy(
                            loading = false,
                            loadingMore = false,
                            loaded = true,
                            page = payload.page.coerceAtLeast(page),
                            totalCount = payload.totalCount.coerceAtLeast(merged.size),
                            items = merged,
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(loading = false, loadingMore = false, loaded = true, errorMessage = err.message ?: "短视频搜索失败")
                    }
                }
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch { authRepository.logoutLocal() }
        }
    }
}
