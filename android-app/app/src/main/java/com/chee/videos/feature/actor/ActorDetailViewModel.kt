package com.chee.videos.feature.actor

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.ActorDetailDto
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.PhoneContentSourcePage
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.model.loadPhoneContentBatch
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private const val ACTOR_WORKS_PAGE_SIZE = 24

data class ActorDetailUiState(
    val actorId: String = "",
    val actor: ActorDetailDto? = null,
    val items: List<VideoListItemDto> = emptyList(),
    val loading: Boolean = false,
    val refreshing: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val pageSize: Int = ACTOR_WORKS_PAGE_SIZE,
    val hasMore: Boolean = true,
    val errorMessage: String? = null,
    val loadMoreErrorMessage: String? = null,
)

@HiltViewModel
class ActorDetailViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(ActorDetailUiState())
    val uiState: StateFlow<ActorDetailUiState> = _uiState.asStateFlow()

    fun initialize(actorId: String) {
        val normalizedActorId = actorId.trim()
        if (normalizedActorId.isBlank()) {
            _uiState.value = ActorDetailUiState(
                loaded = true,
                errorMessage = "演员不存在",
            )
            return
        }
        val current = _uiState.value
        if (current.loaded && current.actorId == normalizedActorId) {
            return
        }
        _uiState.value = ActorDetailUiState(actorId = normalizedActorId, loading = true)
        loadPage(page = 1, append = false, refreshing = false)
    }

    fun retry() {
        val actorId = _uiState.value.actorId
        if (actorId.isBlank()) {
            return
        }
        _uiState.update { it.copy(loading = true, errorMessage = null, loadMoreErrorMessage = null) }
        loadPage(page = 1, append = false, refreshing = false)
    }

    fun refresh() {
        val actorId = _uiState.value.actorId
        if (actorId.isBlank()) {
            return
        }
        _uiState.update {
            if (it.items.isEmpty()) {
                it.copy(loading = true, errorMessage = null, loadMoreErrorMessage = null)
            } else {
                it.copy(refreshing = true, errorMessage = null, loadMoreErrorMessage = null)
            }
        }
        loadPage(page = 1, append = false, refreshing = true)
    }

    fun loadMoreIfNeeded(currentIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.refreshing || state.loadingMore || state.items.isEmpty() || !state.hasMore) {
            return
        }
        if (currentIndex < (state.items.lastIndex - 6).coerceAtLeast(0)) {
            return
        }
        _uiState.update { it.copy(loadingMore = true, loadMoreErrorMessage = null) }
        loadPage(page = state.page + 1, append = true, refreshing = false)
    }

    private fun loadPage(page: Int, append: Boolean, refreshing: Boolean) {
        val actorId = _uiState.value.actorId
        val pageSize = _uiState.value.pageSize.coerceAtLeast(ACTOR_WORKS_PAGE_SIZE)
        viewModelScope.launch {
            var loadedActor: ActorDetailDto? = null
            loadPhoneContentBatch(
                firstPage = page,
                minimumVisibleItems = 1,
                typeOf = VideoListItemDto::type,
                fetchPage = { sourcePage ->
                    videoRepository.fetchActorDetail(
                        actorId = actorId,
                        page = sourcePage,
                        pageSize = pageSize,
                    ).map { payload ->
                        loadedActor = payload.actor
                        PhoneContentSourcePage(
                            items = payload.items,
                            hasMore = hasMoreSourcePages(
                                page = payload.page,
                                pageSize = payload.pageSize,
                                totalCount = payload.totalCount,
                            ),
                        )
                    }
                },
            ).onSuccess { batch ->
                _uiState.update { state ->
                    val mergedItems = if (append) {
                        (state.items + batch.items).distinctBy { it.id }
                    } else {
                        batch.items.distinctBy { it.id }
                    }
                    val trailingErrorMessage = batch.trailingError?.message
                    state.copy(
                        actor = loadedActor ?: state.actor,
                        items = mergedItems,
                        loading = false,
                        refreshing = false,
                        loadingMore = false,
                        loaded = true,
                        page = batch.lastLoadedPage,
                        pageSize = pageSize,
                        hasMore = batch.hasMore,
                        errorMessage = trailingErrorMessage.takeIf { mergedItems.isEmpty() },
                        loadMoreErrorMessage = trailingErrorMessage.takeIf { mergedItems.isNotEmpty() },
                    )
                }
            }
                .onFailure { err ->
                    if (err is AuthExpiredException) {
                        authRepository.logoutLocal()
                    }
                    _uiState.update { state ->
                        if (append) {
                            state.copy(
                                loadingMore = false,
                                loadMoreErrorMessage = err.message ?: "加载更多失败",
                            )
                        } else {
                            state.copy(
                                loading = false,
                                refreshing = false,
                                loaded = true,
                                errorMessage = err.message ?: if (refreshing) "刷新失败" else "演员页加载失败",
                            )
                        }
                    }
                }
        }
    }

    private fun hasMoreSourcePages(page: Int, pageSize: Int, totalCount: Int): Boolean {
        val safePage = page.coerceAtLeast(1)
        val safePageSize = pageSize.coerceAtLeast(1)
        return safePage * safePageSize < totalCount.coerceAtLeast(0)
    }
}
