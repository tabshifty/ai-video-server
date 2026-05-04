package com.chee.videos.feature.shorts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortPlaybackMode
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
    val loadingMore: Boolean = false,
    val items: List<com.chee.videos.core.model.FeedVideoDto> = emptyList(),
    val errorMessage: String? = null,
    val loadMoreErrorMessage: String? = null,
    val detailErrorMessage: String? = null,
    val fitMode: VideoFitMode = VideoFitMode.FILL,
    val pausedByUserVideoIds: Set<String> = emptySet(),
    val detailSheetVideoId: String? = null,
    val detailByVideoId: Map<String, VideoDetailDto> = emptyMap(),
    val detailLoadingVideoIds: Set<String> = emptySet(),
    val actionBusyVideoIds: Set<String> = emptySet(),
    val pagerResetToken: Int = 0,
    val pagerInitialPage: Int = 0,
    val pagerAnchorVideoId: String? = null,
    val playbackMode: ShortPlaybackMode = ShortPlaybackMode.LOOP_ONE,
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
        viewModelScope.launch {
            store.shortPlaybackModeFlow.collect { mode ->
                _uiState.update { it.copy(playbackMode = mode) }
            }
        }
    }

    fun load(force: Boolean = false) {
        val currentState = _uiState.value
        if ((currentState.loaded || currentState.loading) && !force) {
            return
        }
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    loading = true,
                    loaded = false,
                    errorMessage = null,
                    loadMoreErrorMessage = null,
                )
            }
            videoRepository.fetchShortFeed(pageSize = ShortFeedWindowManager.LoadBatchSize)
                .onSuccess { items ->
                    val itemIDs = items.map { it.id }.toSet()
                    _uiState.update { state ->
                        val nextInitialPage = if (force) {
                            0
                        } else {
                            resolveShortFeedInitialPage(
                                incomingItems = items,
                                anchorVideoId = state.pagerAnchorVideoId,
                                fallbackPage = state.pagerInitialPage,
                            )
                        }
                        val nextAnchorVideoId = items.getOrNull(nextInitialPage)?.id
                        state.copy(
                            loading = false,
                            loaded = true,
                            loadingMore = false,
                            items = items,
                            errorMessage = null,
                            loadMoreErrorMessage = null,
                            pausedByUserVideoIds = state.pausedByUserVideoIds.filterTo(mutableSetOf()) { it in itemIDs },
                            detailByVideoId = state.detailByVideoId.filterKeys { it in itemIDs },
                            detailLoadingVideoIds = state.detailLoadingVideoIds.filterTo(mutableSetOf()) { it in itemIDs },
                            actionBusyVideoIds = state.actionBusyVideoIds.filterTo(mutableSetOf()) { it in itemIDs },
                            detailSheetVideoId = state.detailSheetVideoId?.takeIf { it in itemIDs },
                            pagerInitialPage = nextInitialPage,
                            pagerAnchorVideoId = nextAnchorVideoId,
                            pagerResetToken = if (force || nextInitialPage != state.pagerInitialPage) {
                                state.pagerResetToken + 1
                            } else {
                                state.pagerResetToken
                            },
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            loaded = true,
                            loadingMore = false,
                            errorMessage = err.message ?: "短视频加载失败",
                        )
                    }
                }
        }
    }

    fun loadMoreIfNeeded(currentIndex: Int, currentVideoId: String?) {
        val state = _uiState.value
        if (!ShortFeedWindowManager.shouldLoadMore(currentIndex, state.items.size, state.loadingMore)) {
            return
        }
        if (state.loading || !state.loaded || state.items.isEmpty()) {
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(loadingMore = true, loadMoreErrorMessage = null) }
            videoRepository.fetchShortFeed(
                pageSize = ShortFeedWindowManager.LoadBatchSize,
                excludeIds = state.items.map { it.id },
            ).onSuccess { incoming ->
                _uiState.update { latest ->
                    val merged = ShortFeedWindowManager.mergeAppend(
                        snapshot = latest.toWindowSnapshot(),
                        incoming = incoming,
                        anchorVideoId = currentVideoId,
                    )
                    latest.copy(
                        loadingMore = false,
                        loadMoreErrorMessage = null,
                        items = merged.items,
                        detailByVideoId = merged.detailByVideoId,
                        detailLoadingVideoIds = merged.detailLoadingVideoIds,
                        actionBusyVideoIds = merged.actionBusyVideoIds,
                        pausedByUserVideoIds = merged.pausedByUserVideoIds,
                        detailSheetVideoId = merged.detailSheetVideoId,
                        pagerInitialPage = merged.anchorPageAfterTrim ?: latest.pagerInitialPage,
                        pagerAnchorVideoId = merged.anchorPageAfterTrim
                            ?.let { merged.items.getOrNull(it)?.id }
                            ?: latest.pagerAnchorVideoId,
                        pagerResetToken = if (merged.anchorPageAfterTrim != null) latest.pagerResetToken + 1 else latest.pagerResetToken,
                    )
                }
            }.onFailure { err ->
                handleAuthError(err)
                _uiState.update {
                    it.copy(
                        loadingMore = false,
                        loadMoreErrorMessage = err.message ?: "短视频补货失败",
                    )
                }
            }
        }
    }

    fun rememberPagerAnchor(page: Int, videoId: String?) {
        _uiState.update { state ->
            val normalizedPage = page.coerceAtLeast(0)
            val normalizedVideoId = videoId?.takeIf { it.isNotBlank() }
            if (state.pagerInitialPage == normalizedPage && state.pagerAnchorVideoId == normalizedVideoId) {
                state
            } else {
                state.copy(
                    pagerInitialPage = normalizedPage,
                    pagerAnchorVideoId = normalizedVideoId,
                )
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

    fun togglePlaybackMode() {
        val next = if (_uiState.value.playbackMode == ShortPlaybackMode.LOOP_ONE) {
            ShortPlaybackMode.AUTO_NEXT
        } else {
            ShortPlaybackMode.LOOP_ONE
        }
        _uiState.update { it.copy(playbackMode = next) }
        viewModelScope.launch {
            store.saveShortPlaybackMode(next)
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

    fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        if (videoId.isBlank() || watchSeconds <= 0) {
            return
        }
        viewModelScope.launch {
            videoRepository.reportHistory(videoId, watchSeconds, completed)
        }
    }

    private fun toggleAction(videoId: String, actionCall: suspend () -> Result<ActionTogglePayload>) {
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

    private fun ShortFeedUiState.toWindowSnapshot(): ShortFeedWindowSnapshot {
        return ShortFeedWindowSnapshot(
            items = items,
            detailByVideoId = detailByVideoId,
            detailLoadingVideoIds = detailLoadingVideoIds,
            actionBusyVideoIds = actionBusyVideoIds,
            pausedByUserVideoIds = pausedByUserVideoIds,
            detailSheetVideoId = detailSheetVideoId,
        )
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch {
                authRepository.logoutLocal()
            }
        }
    }
}

internal fun resolveShortFeedInitialPage(
    incomingItems: List<com.chee.videos.core.model.FeedVideoDto>,
    anchorVideoId: String?,
    fallbackPage: Int,
): Int {
    if (incomingItems.isEmpty()) {
        return 0
    }
    val anchorIndex = anchorVideoId
        ?.takeIf { it.isNotBlank() }
        ?.let { id -> incomingItems.indexOfFirst { it.id == id } }
        ?: -1
    if (anchorIndex >= 0) {
        return anchorIndex
    }
    return fallbackPage.coerceIn(0, incomingItems.lastIndex)
}
