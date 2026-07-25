package com.chee.videos.feature.shortcollections

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortCollectionListItemDto
import com.chee.videos.core.model.ShortCollectionsPayload
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

private const val ShortCollectionsPageSize = 20

data class ShortCollectionsUiState(
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val totalCount: Int = 0,
    val items: List<ShortCollectionListItemDto> = emptyList(),
    val refreshing: Boolean = false,
    val errorMessage: String? = null,
)

@HiltViewModel
class ShortCollectionsViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(ShortCollectionsUiState())
    val uiState: StateFlow<ShortCollectionsUiState> = _uiState.asStateFlow()

    init {
        loadFirst()
    }

    fun loadFirst() {
        loadPage(page = 1, append = false)
    }

    fun retry() {
        if (_uiState.value.items.isNotEmpty()) {
            refresh()
        } else {
            loadPage(page = 1, append = false)
        }
    }

    fun refresh() {
        if (_uiState.value.loading) return
        _uiState.update { it.copy(refreshing = true, errorMessage = null) }
        loadPage(page = 1, append = false, isRefresh = true)
    }

    fun loadMoreIfNeeded(visibleIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.loadingMore || state.items.isEmpty()) return
        if (visibleIndex < state.items.lastIndex - 5) return
        if (state.totalCount > 0 && state.items.size >= state.totalCount) return
        loadPage(page = state.page + 1, append = true)
    }

    private fun loadPage(page: Int, append: Boolean, isRefresh: Boolean = false) {
        if (page < 1) return
        viewModelScope.launch {
            if (!append) {
                _uiState.update {
                    it.copy(
                        loading = !isRefresh && it.items.isEmpty(),
                        loadingMore = false,
                        errorMessage = null,
                    )
                }
            } else {
                _uiState.update { it.copy(loadingMore = true, errorMessage = null) }
            }
            videoRepository.fetchShortCollections(page = page, pageSize = ShortCollectionsPageSize)
                .onSuccess { payload ->
                    _uiState.update {
                        val merged = if (append) mergeItems(it.items, payload.items) else payload.items
                        it.copy(
                            loading = false,
                            loadingMore = false,
                            loaded = true,
                            refreshing = false,
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
                        it.copy(
                            loading = false,
                            loadingMore = false,
                            loaded = true,
                            refreshing = false,
                            errorMessage = err.message ?: "加载合集目录失败",
                        )
                    }
                }
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch { authRepository.logoutLocal() }
        }
    }

    private fun mergeItems(
        existing: List<ShortCollectionListItemDto>,
        incoming: List<ShortCollectionListItemDto>,
    ): List<ShortCollectionListItemDto> {
        return (existing + incoming).distinctBy { it.id }
    }
}
