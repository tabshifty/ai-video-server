package com.chee.videos.feature.tv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.FeedVideoDto
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TvShortFeedUiState(
    val loading: Boolean = true,
    val loadingMore: Boolean = false,
    val items: List<FeedVideoDto> = emptyList(),
    val currentIndex: Int = 0,
    val loadErrorMessage: String? = null,
    val seekStepSeconds: Int = TvPlaybackSeekStepSetting.defaultSeconds,
    val pausedByUserVideoIds: Set<String> = emptySet(),
)

@HiltViewModel
class TvShortFeedViewModel @Inject constructor(
    private val repository: TvRepository,
) : ViewModel() {
    private var requestVersion = 0
    private val seenVideoIds = linkedSetOf<String>()
    private val _uiState = MutableStateFlow(TvShortFeedUiState())
    val uiState: StateFlow<TvShortFeedUiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            _uiState.update {
                it.copy(seekStepSeconds = TvPlaybackSeekStepSetting.normalize(repository.readTvSeekStepSeconds()))
            }
        }
    }

    fun load(force: Boolean = false) {
        val state = _uiState.value
        if (state.items.isNotEmpty() && !force) {
            return
        }
        val versionAtRequest = nextRequestVersion()
        if (force) {
            seenVideoIds.clear()
        }
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    loading = true,
                    loadingMore = false,
                    loadErrorMessage = null,
                    items = if (force) emptyList() else it.items,
                    currentIndex = if (force) 0 else it.currentIndex,
                    pausedByUserVideoIds = if (force) emptySet() else it.pausedByUserVideoIds,
                )
            }
            repository.fetchShortFeed(pageSize = TvShortFeedLoadBatchSize, excludeIds = seenVideoIds.toList())
                .onSuccess { incoming ->
                    if (versionAtRequest != requestVersion) {
                        return@onSuccess
                    }
                    val nextItems = mergeIncomingShortFeedItems(
                        currentItems = if (force) emptyList() else _uiState.value.items,
                        incoming = incoming,
                    )
                    seenVideoIds.addAll(nextItems.map { it.id })
                    _uiState.update { stateAfterLoad ->
                        stateAfterLoad.copy(
                            loading = false,
                            loadingMore = false,
                            items = nextItems,
                            currentIndex = if (nextItems.isEmpty()) 0 else stateAfterLoad.currentIndex.coerceIn(0, nextItems.lastIndex),
                            loadErrorMessage = null,
                        )
                    }
                }
                .onFailure { error ->
                    if (versionAtRequest != requestVersion) {
                        return@onFailure
                    }
                    _uiState.update {
                        it.copy(
                            loading = false,
                            loadingMore = false,
                            loadErrorMessage = error.message ?: "短视频加载失败",
                        )
                    }
                }
        }
    }

    fun ensureMoreLoaded(currentIndex: Int) {
        val state = _uiState.value
        if (state.loadingMore || state.items.isEmpty()) {
            return
        }
        if (currentIndex < (state.items.size - TvShortFeedLoadMoreThreshold).coerceAtLeast(0)) {
            return
        }

        val versionAtRequest = requestVersion
        _uiState.update { it.copy(loadingMore = true) }
        viewModelScope.launch {
            repository.fetchShortFeed(
                pageSize = TvShortFeedLoadBatchSize,
                excludeIds = seenVideoIds.toList(),
            ).onSuccess { incoming ->
                if (versionAtRequest != requestVersion) {
                    return@onSuccess
                }
                val nextItems = mergeIncomingShortFeedItems(
                    currentItems = _uiState.value.items,
                    incoming = incoming,
                )
                seenVideoIds.addAll(nextItems.map { it.id })
                _uiState.update { latest ->
                    latest.copy(
                        loadingMore = false,
                        items = nextItems,
                        currentIndex = latest.currentIndex.coerceIn(0, nextItems.lastIndex),
                    )
                }
            }.onFailure {
                if (versionAtRequest != requestVersion) {
                    return@onFailure
                }
                _uiState.update { latest -> latest.copy(loadingMore = false) }
            }
        }
    }

    fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        if (videoId.isBlank()) {
            return
        }
        viewModelScope.launch {
            repository.reportHistory(videoId, watchSeconds, completed)
        }
    }

    fun movePrevious(): Boolean {
        var moved = false
        _uiState.update { state ->
            if (state.items.isEmpty() || state.currentIndex <= 0) {
                return@update state
            }
            moved = true
            state.copy(currentIndex = state.currentIndex - 1)
        }
        return moved
    }

    fun moveNext(): Boolean {
        var moved = false
        _uiState.update { state ->
            if (state.items.isEmpty() || state.currentIndex >= state.items.lastIndex) {
                return@update state
            }
            moved = true
            state.copy(currentIndex = state.currentIndex + 1)
        }
        return moved
    }

    fun togglePauseCurrent(videoId: String): Boolean {
        if (videoId.isBlank()) {
            return false
        }
        var pausedNow = false
        _uiState.update { state ->
            val next = state.pausedByUserVideoIds.toMutableSet()
            if (next.remove(videoId)) {
                pausedNow = false
            } else {
                next.add(videoId)
                pausedNow = true
            }
            state.copy(pausedByUserVideoIds = next)
        }
        return pausedNow
    }

    private fun nextRequestVersion(): Int {
        requestVersion += 1
        return requestVersion
    }

    private fun mergeIncomingShortFeedItems(
        currentItems: List<FeedVideoDto>,
        incoming: List<FeedVideoDto>,
    ): List<FeedVideoDto> {
        val currentIds = currentItems.map { it.id }.toSet()
        val uniqueBatch = incoming
            .filter { it.id.isNotBlank() }
            .distinctBy { it.id }
        val appendBatch = uniqueBatch.filterNot { it.id in currentIds }
            .ifEmpty { uniqueBatch }
        return currentItems + appendBatch
    }

    private companion object {
        const val TvShortFeedLoadBatchSize = 20
        const val TvShortFeedLoadMoreThreshold = 5
    }
}
