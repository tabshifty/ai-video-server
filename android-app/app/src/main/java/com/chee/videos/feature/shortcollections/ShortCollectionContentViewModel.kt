package com.chee.videos.feature.shortcollections

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.TvDeviceDto
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import com.chee.videos.core.ui.cast.buildCollectionShortTvRemoteSearchContext
import com.chee.videos.core.ui.cast.buildShortTvRemoteItems
import com.chee.videos.core.ui.cast.pickTvDeviceForLaunch
import com.chee.videos.core.ui.cast.resolveShortTvRemoteStartIndex
import com.chee.videos.core.ui.cast.resolveTvDeviceRequestTarget
import com.chee.videos.core.ui.cast.resolveTvDeviceSelection
import com.chee.videos.core.ui.cast.shouldPersistTvDevicePreference
import com.chee.videos.core.ui.cast.sortTvDevicesForSelection
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private const val ShortCollectionContentPageSize = 24

data class ShortCollectionContentUiState(
    val collectionId: String = "",
    val collectionName: String = "",
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
    val offlineMessage: String? = null,
)

@HiltViewModel
class ShortCollectionContentViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(ShortCollectionContentUiState())
    val uiState: StateFlow<ShortCollectionContentUiState> = _uiState.asStateFlow()
    private var loadJob: Job? = null

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

    fun bind(collectionId: String, collectionName: String) {
        if (collectionId.isBlank()) {
            _uiState.update { it.copy(offlineMessage = "合集参数缺失") }
            return
        }
        if (_uiState.value.collectionId == collectionId) return
        _uiState.update {
            it.copy(
                collectionId = collectionId,
                collectionName = collectionName,
                loading = true,
                loadingMore = false,
                loaded = false,
                page = 0,
                totalCount = 0,
                items = emptyList(),
                playingVideoId = null,
                errorMessage = null,
                offlineMessage = null,
            )
        }
        loadPage(page = 1, append = false)
    }

    fun retry() {
        if (_uiState.value.collectionId.isBlank()) return
        loadPage(page = 1, append = false)
    }

    fun loadMoreIfNeeded(currentIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.loadingMore || state.items.isEmpty()) return
        if (currentIndex < state.items.lastIndex - 5) return
        if (state.totalCount > 0 && state.items.size >= state.totalCount) return
        loadPage(page = state.page + 1, append = true)
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
        viewModelScope.launch { loadDetail(videoId, force = force) }
    }

    fun toggleLike(videoId: String) {
        toggleAction(videoId) { videoRepository.toggleLike(videoId) }
    }

    fun toggleFavorite(videoId: String) {
        toggleAction(videoId) { videoRepository.toggleFavorite(videoId) }
    }

    fun openCastSheet() {
        val startIndex = resolveShortTvRemoteStartIndex(_uiState.value.items, _uiState.value.playingVideoId)
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
                            selectedTvDeviceId = resolveTvDeviceSelection(
                                items = sorted,
                                currentSelectedDeviceId = it.selectedTvDeviceId,
                                preferredDeviceId = preferredDeviceId,
                            ),
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

    fun startCastFromCollection() {
        val state = _uiState.value
        val selectedDeviceId = state.selectedTvDeviceId?.trim().orEmpty()
        val currentIndex = resolveShortTvRemoteStartIndex(state.items, state.playingVideoId)
        if (selectedDeviceId.isBlank()) {
            _uiState.update { it.copy(castErrorMessage = "请选择要投放的电视") }
            return
        }
        if (currentIndex == null) {
            _uiState.update { it.copy(castErrorMessage = "当前视频未准备好，暂时无法投放") }
            return
        }
        val snapshotItems = buildShortTvRemoteItems(state.items)
        if (snapshotItems.isEmpty()) {
            _uiState.update { it.copy(castErrorMessage = "当前合集为空，无法发起投放") }
            return
        }
        _uiState.update { it.copy(castLaunching = true, castErrorMessage = null) }
        viewModelScope.launch {
            val launchTarget = resolveTvDeviceRequestTarget(state.tvDevices, selectedDeviceId)
            if (launchTarget.isNullOrBlank()) {
                _uiState.update { it.copy(castLaunching = false, castErrorMessage = "请选择要投放的电视") }
                return@launch
            }
            videoRepository.createTvRemoteSession(
                deviceId = launchTarget,
                items = snapshotItems,
                currentIndex = currentIndex,
                searchContext = buildCollectionShortTvRemoteSearchContext(
                    collectionId = state.collectionId,
                    page = state.page,
                    totalCount = state.totalCount,
                ),
            ).onSuccess { session ->
                if (shouldPersistTvDevicePreference(pickTvDeviceForLaunch(state.tvDevices, launchTarget))) {
                    videoRepository.saveLastTvRemoteDeviceId(launchTarget)
                }
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

    private fun loadPage(page: Int, append: Boolean) {
        val collectionId = _uiState.value.collectionId
        if (collectionId.isBlank()) return
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            if (append) {
                _uiState.update { it.copy(loadingMore = true, errorMessage = null) }
            } else {
                _uiState.update { it.copy(loading = true, errorMessage = null) }
            }
            videoRepository.fetchShortDiscover(mode = "collection", value = collectionId, page = page, pageSize = ShortCollectionContentPageSize)
                .onSuccess { payload ->
                    _uiState.update {
                        val merged = if (append) mergeItems(it.items, payload.items) else payload.items.distinctBy { row -> row.id }
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
                            offlineMessage = if (!append && merged.isEmpty()) {
                                "该合集已下线或暂无可播放内容"
                            } else {
                                null
                            },
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            loadingMore = false,
                            loaded = true,
                            errorMessage = if (it.items.isNotEmpty()) null else (err.message ?: "加载合集内容失败"),
                        )
                    }
                }
        }
    }

    private fun mergeItems(existing: List<VideoListItemDto>, incoming: List<VideoListItemDto>): List<VideoListItemDto> {
        return (existing + incoming).distinctBy { it.id }
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
