package com.chee.videos.feature.shortcollections

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ShortCollectionListItemDto
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
private const val ShortCollectionsSearchQueryMaxRunes = 200

data class ShortCollectionsUiState(
    val loading: Boolean = false,
    val loadingMore: Boolean = false,
    val loaded: Boolean = false,
    val page: Int = 0,
    val totalCount: Int = 0,
    val query: String = "",
    val submittedQuery: String = "",
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
    private var requestVersion = 0L

    init {
        loadFirst()
    }

    fun loadFirst() {
        requestPage(query = _uiState.value.submittedQuery, page = 1, append = false, isRefresh = false)
    }

    fun updateQuery(query: String) {
        if (query == _uiState.value.query) return
        _uiState.update { it.copy(query = query, errorMessage = null) }
        if (normalizeShortCollectionsQuery(query) == null && _uiState.value.submittedQuery.isNotBlank()) {
            submitQuery()
        }
    }

    fun submitQuery() {
        val state = _uiState.value
        val normalized = normalizeShortCollectionsQuery(state.query)
        if (normalized != null && normalized.codePointCount(0, normalized.length) > ShortCollectionsSearchQueryMaxRunes) {
            _uiState.update { it.copy(errorMessage = "搜索内容过长，最多 200 个字符") }
            return
        }
        val submitted = normalized.orEmpty()
        _uiState.value = resetShortCollectionsForSearch(state, state.query, submitted)
        requestPage(query = submitted, page = 1, append = false, isRefresh = false)
    }

    fun retry() {
        requestPage(query = _uiState.value.submittedQuery, page = 1, append = false, isRefresh = false)
    }

    fun refresh() {
        if (_uiState.value.loading || _uiState.value.loadingMore) return
        _uiState.update { it.copy(refreshing = true, errorMessage = null) }
        requestPage(query = _uiState.value.submittedQuery, page = 1, append = false, isRefresh = true)
    }

    fun loadMoreIfNeeded(visibleIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.loadingMore || state.items.isEmpty()) return
        if (visibleIndex < state.items.lastIndex - 5) return
        if (state.totalCount > 0 && state.items.size >= state.totalCount) return
        requestPage(
            query = state.submittedQuery,
            page = state.page + 1,
            append = true,
            isRefresh = false,
        )
    }

    private fun requestPage(query: String, page: Int, append: Boolean, isRefresh: Boolean) {
        if (page < 1) return
        requestVersion += 1
        val version = requestVersion
        _uiState.update {
            if (append) {
                it.copy(loadingMore = true, errorMessage = null)
            } else {
                it.copy(
                    loading = it.items.isEmpty(),
                    loadingMore = false,
                    refreshing = isRefresh,
                    errorMessage = null,
                )
            }
        }
        viewModelScope.launch {
            videoRepository.fetchShortCollections(
                query = query,
                page = page,
                pageSize = ShortCollectionsPageSize,
            ).onSuccess { payload ->
                if (version != requestVersion) return@onSuccess
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
            }.onFailure { err ->
                if (version != requestVersion) return@onFailure
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
    ): List<ShortCollectionListItemDto> = (existing + incoming).distinctBy { it.id }
}

internal fun normalizeShortCollectionsQuery(raw: String): String? = raw.trim().takeIf { it.isNotEmpty() }

internal fun resetShortCollectionsForSearch(
    state: ShortCollectionsUiState,
    query: String,
    submittedQuery: String,
): ShortCollectionsUiState = state.copy(
    loading = true,
    loadingMore = false,
    loaded = false,
    page = 0,
    totalCount = 0,
    query = query,
    submittedQuery = submittedQuery,
    refreshing = false,
    errorMessage = null,
)

internal fun shortCollectionsEmptyMessage(query: String): String {
    return if (normalizeShortCollectionsQuery(query) == null) "暂无合集" else "没有找到相关合集"
}
