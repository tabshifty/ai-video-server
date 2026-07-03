package com.chee.videos.feature.shortsearch

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.TvDeviceDto
import com.chee.videos.core.model.TvRemoteSessionItemDto
import com.chee.videos.core.model.VideoDetailDto
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
    val detailByVideoId: Map<String, VideoDetailDto> = emptyMap(),
    val detailLoadingVideoIds: Set<String> = emptySet(),
    val actionBusyVideoIds: Set<String> = emptySet(),
    val tvDevices: List<TvDeviceDto> = emptyList(),
    val castSheetVisible: Boolean = false,
    val castDevicesLoading: Boolean = false,
    val castLaunching: Boolean = false,
    val selectedTvDeviceId: String? = null,
    val castErrorMessage: String? = null,
    val pendingRemoteSessionId: String? = null,
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

internal fun buildShortSearchRemoteItems(items: List<VideoListItemDto>): List<TvRemoteSessionItemDto> {
    return items
        .filter { it.id.isNotBlank() }
        .map { item ->
            TvRemoteSessionItemDto(
                videoId = item.id,
                title = item.title.trim(),
                thumbnailPath = item.thumbnailPath?.trim(),
                duration = item.duration,
                type = item.type.trim().ifBlank { "short" },
            )
        }
}

internal fun sortTvDevicesForSelection(items: List<TvDeviceDto>, preferredDeviceId: String?): List<TvDeviceDto> {
    val preferred = preferredDeviceId?.trim().orEmpty()
    return items.sortedWith(
        compareByDescending<TvDeviceDto> { it.deviceId == preferred }
            .thenByDescending { it.isOnline }
            .thenBy { it.deviceName },
    )
}

internal fun resolveShortSearchRemoteStartIndex(items: List<VideoListItemDto>, playingVideoId: String?): Int? {
    val target = playingVideoId?.trim().orEmpty()
    if (target.isBlank()) {
        return null
    }
    val index = items.indexOfFirst { it.id == target }
    return index.takeIf { it >= 0 }
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

    fun ensureDetailLoaded(videoId: String, force: Boolean = false) {
        viewModelScope.launch {
            loadDetail(videoId, force = force)
        }
    }

    fun toggleLike(videoId: String) {
        toggleAction(videoId) { videoRepository.toggleLike(videoId) }
    }

    fun toggleFavorite(videoId: String) {
        toggleAction(videoId) { videoRepository.toggleFavorite(videoId) }
    }

    fun openCastSheet() {
        val startIndex = resolveShortSearchRemoteStartIndex(_uiState.value.items, _uiState.value.playingVideoId)
        if (startIndex == null) {
            _uiState.update { it.copy(castErrorMessage = "当前视频未准备好，暂时无法投放") }
            return
        }
        _uiState.update { it.copy(castSheetVisible = true, castDevicesLoading = true, castErrorMessage = null) }
        viewModelScope.launch {
            val preferredDeviceId = videoRepository.readLastTvRemoteDeviceId()
            videoRepository.fetchTvDevices()
                .onSuccess { devices ->
                    val sorted = sortTvDevicesForSelection(devices, preferredDeviceId)
                    _uiState.update {
                        it.copy(
                            tvDevices = sorted,
                            castDevicesLoading = false,
                            selectedTvDeviceId = it.selectedTvDeviceId
                                ?.takeIf { selected -> sorted.any { device -> device.deviceId == selected } }
                                ?: preferredDeviceId?.takeIf { selected -> sorted.any { device -> device.deviceId == selected } }
                                ?: sorted.firstOrNull()?.deviceId,
                            castErrorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            castDevicesLoading = false,
                            castErrorMessage = err.message ?: "加载电视列表失败",
                        )
                    }
                }
        }
    }

    fun dismissCastSheet() {
        _uiState.update {
            it.copy(
                castSheetVisible = false,
                castDevicesLoading = false,
                castLaunching = false,
                castErrorMessage = null,
            )
        }
    }

    fun selectTvDevice(deviceId: String) {
        _uiState.update { it.copy(selectedTvDeviceId = deviceId.trim()) }
    }

    fun startCastFromCurrentSearch() {
        val state = _uiState.value
        val selectedDeviceId = state.selectedTvDeviceId?.trim().orEmpty()
        val currentIndex = resolveShortSearchRemoteStartIndex(state.items, state.playingVideoId)
        if (selectedDeviceId.isBlank()) {
            _uiState.update { it.copy(castErrorMessage = "请选择要投放的电视") }
            return
        }
        if (currentIndex == null) {
            _uiState.update { it.copy(castErrorMessage = "当前视频未准备好，暂时无法投放") }
            return
        }
        val snapshotItems = buildShortSearchRemoteItems(state.items)
        if (snapshotItems.isEmpty()) {
            _uiState.update { it.copy(castErrorMessage = "当前搜索结果为空，无法发起投放") }
            return
        }
        _uiState.update { it.copy(castLaunching = true, castErrorMessage = null) }
        viewModelScope.launch {
            videoRepository.createTvRemoteSession(
                deviceId = selectedDeviceId,
                items = snapshotItems,
                currentIndex = currentIndex,
            ).onSuccess { session ->
                videoRepository.saveLastTvRemoteDeviceId(selectedDeviceId)
                _uiState.update {
                    it.copy(
                        castSheetVisible = false,
                        castDevicesLoading = false,
                        castLaunching = false,
                        castErrorMessage = null,
                        pendingRemoteSessionId = session.sessionId,
                    )
                }
            }.onFailure { err ->
                handleAuthError(err)
                _uiState.update {
                    it.copy(
                        castLaunching = false,
                        castErrorMessage = err.message ?: "发起投放失败",
                    )
                }
            }
        }
    }

    fun consumePendingRemoteSession() {
        _uiState.update { it.copy(pendingRemoteSessionId = null) }
    }

    fun clearCastErrorMessage() {
        _uiState.update { it.copy(castErrorMessage = null) }
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
                            detailByVideoId = it.detailByVideoId.filterKeys { key -> merged.any { row -> row.id == key } },
                            detailLoadingVideoIds = it.detailLoadingVideoIds.filter { id -> merged.any { row -> row.id == id } }.toSet(),
                            actionBusyVideoIds = it.actionBusyVideoIds.filter { id -> merged.any { row -> row.id == id } }.toSet(),
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

    private fun toggleAction(videoId: String, actionCall: suspend () -> Result<ActionTogglePayload>) {
        if (_uiState.value.actionBusyVideoIds.contains(videoId)) {
            return
        }
        viewModelScope.launch {
            val hasDetail = _uiState.value.detailByVideoId.containsKey(videoId)
            if (!hasDetail) {
                val loaded = loadDetail(videoId, force = true)
                if (!loaded) {
                    return@launch
                }
            }

            _uiState.update { it.copy(actionBusyVideoIds = it.actionBusyVideoIds + videoId) }
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
                    _uiState.update { it.copy(actionBusyVideoIds = it.actionBusyVideoIds - videoId) }
                }
        }
    }

    private suspend fun loadDetail(videoId: String, force: Boolean): Boolean {
        if (videoId.isBlank()) {
            return false
        }
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
                )
            }
            true
        } else {
            val err = result.exceptionOrNull()
            handleAuthError(err)
            _uiState.update { it.copy(detailLoadingVideoIds = it.detailLoadingVideoIds - videoId) }
            false
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch { authRepository.logoutLocal() }
        }
    }
}
