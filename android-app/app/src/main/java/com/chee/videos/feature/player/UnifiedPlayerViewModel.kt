package com.chee.videos.feature.player

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.PhoneContentSourcePage
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.model.isPhoneSupportedVideoType
import com.chee.videos.core.model.loadPhoneContentBatch
import com.chee.videos.core.model.resolvePhoneContentStartIndex
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
    val contentUnavailable: Boolean = false,
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
                    contentUnavailable = false,
                )
            }

            val result = loadPhoneContentBatch(
                firstPage = 1,
                minimumVisibleItems = 1,
                typeOf = PlayerVideoItem::type,
                fetchPage = { page -> fetchPlayerSourcePage(source, page) },
            )

            result
                .onSuccess { batch ->
                    val items = batch.items.distinctBy { it.id }
                    val index = resolvePhoneContentStartIndex(
                        items = items,
                        startItemId = startVideoId,
                        idOf = PlayerVideoItem::id,
                    )
                    if (index == null) {
                        val trailingError = batch.trailingError
                        if (items.isEmpty() && trailingError != null) {
                            handleAuthError(trailingError)
                            _uiState.update {
                                it.copy(
                                    loading = false,
                                    items = emptyList(),
                                    startIndex = 0,
                                    errorMessage = trailingError.message ?: "播放列表加载失败",
                                )
                            }
                        } else {
                            _uiState.update {
                                it.copy(
                                    loading = false,
                                    items = emptyList(),
                                    startIndex = 0,
                                    errorMessage = null,
                                    contentUnavailable = true,
                                )
                            }
                        }
                        return@onSuccess
                    }
                    _uiState.update {
                        it.copy(
                            loading = false,
                            items = items,
                            startIndex = index,
                            errorMessage = null,
                            contentUnavailable = false,
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
                            contentUnavailable = false,
                        )
                    }
                }
        }
    }

    private suspend fun fetchPlayerSourcePage(
        source: String,
        page: Int,
    ): Result<PhoneContentSourcePage<PlayerVideoItem>> {
        return when (source) {
            "history" -> videoRepository.fetchContinueHistory(page = page, limit = PLAYER_QUEUE_PAGE_SIZE)
                .map { payload ->
                    PhoneContentSourcePage(
                        items = payload.items.map { row ->
                            PlayerVideoItem(
                                id = row.videoId,
                                title = row.title,
                                thumbnailPath = row.thumbnailPath,
                                type = row.type,
                                duration = row.duration,
                            )
                        },
                        hasMore = hasMoreSourcePages(payload.page, payload.pageSize, payload.totalCount),
                    )
                }

            "favorite" -> videoRepository.fetchFavoritedVideos(page = page, pageSize = PLAYER_QUEUE_PAGE_SIZE)
                .map { payload ->
                    PhoneContentSourcePage(
                        items = payload.items.map { row ->
                            PlayerVideoItem(
                                id = row.id,
                                title = row.title,
                                thumbnailPath = row.thumbnailPath,
                                type = row.type,
                                duration = row.duration,
                            )
                        },
                        hasMore = hasMoreSourcePages(payload.page, payload.pageSize, payload.totalCount),
                    )
                }

            "like" -> videoRepository.fetchLikedVideos(page = page, pageSize = PLAYER_QUEUE_PAGE_SIZE)
                .map { payload ->
                    PhoneContentSourcePage(
                        items = payload.items.map { row ->
                            PlayerVideoItem(
                                id = row.id,
                                title = row.title,
                                thumbnailPath = row.thumbnailPath,
                                type = row.type,
                                duration = row.duration,
                            )
                        },
                        hasMore = hasMoreSourcePages(payload.page, payload.pageSize, payload.totalCount),
                    )
                }

            else -> Result.success(PhoneContentSourcePage(items = emptyList(), hasMore = false))
        }
    }

    fun ensureDetailLoaded(videoId: String, force: Boolean = false) {
        val state = _uiState.value
        val item = state.items.firstOrNull { it.id == videoId }
        if (item == null || !isPhoneSupportedVideoType(item.type)) {
            return
        }
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

    private fun hasMoreSourcePages(page: Int, pageSize: Int, totalCount: Int): Boolean {
        val safePage = page.coerceAtLeast(1)
        val safePageSize = pageSize.coerceAtLeast(1)
        return safePage * safePageSize < totalCount.coerceAtLeast(0)
    }

    private companion object {
        const val PLAYER_QUEUE_PAGE_SIZE = 120
    }
}
