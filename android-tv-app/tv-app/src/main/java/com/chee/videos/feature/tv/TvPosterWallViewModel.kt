package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

@HiltViewModel
class TvPosterWallViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val repository: TvRepository,
) : ViewModel() {
    private val wallKind = decodeTvRouteArg(savedStateHandle[TvCatalogWallKindArg]).ifBlank { "tv" }
    private val wallTitle = decodeTvRouteArg(savedStateHandle[TvCatalogWallTitleArg])
    private val wallSpec = resolveTvCatalogWallSpec(wallKind, wallTitle)
    private var requestVersion = 0

    private val _uiState = MutableStateFlow(
        TvCatalogWallUiState(
            loading = true,
            kind = wallSpec.kind,
            title = wallSpec.title,
            subtitle = wallSpec.subtitle,
        ),
    )
    val uiState: StateFlow<TvCatalogWallUiState> = _uiState.asStateFlow()

    init {
        loadPage(page = 1, append = false)
    }

    internal constructor(
        repository: TvRepository,
        kind: String,
        title: String = "",
    ) : this(
        SavedStateHandle(
            mapOf(
                TvCatalogWallKindArg to kind,
                TvCatalogWallTitleArg to encodeTvRouteSegment(title),
            ),
        ),
        repository,
    )

    fun refresh() {
        val state = _uiState.value
        if (state.loading || state.refreshing) {
            return
        }
        requestVersion += 1
        _uiState.update {
            it.copy(
                loading = it.items.isEmpty(),
                loadingMore = false,
                refreshing = it.items.isNotEmpty(),
                errorMessage = null,
            )
        }
        loadPage(page = 1, append = false)
    }

    fun changeSort(sortBy: String, sortOrder: String) {
        val normalized = normalizeTvCatalogWallSort(sortBy, sortOrder)
        val state = _uiState.value
        if (state.sortBy == normalized.sortBy && state.sortOrder == normalized.sortOrder) {
            return
        }
        requestVersion += 1
        _uiState.update {
            it.copy(
                loading = it.items.isEmpty(),
                loadingMore = false,
                refreshing = it.items.isNotEmpty(),
                page = if (it.items.isEmpty()) 0 else it.page,
                totalCount = if (it.items.isEmpty()) 0 else it.totalCount,
                sortBy = normalized.sortBy,
                sortOrder = normalized.sortOrder,
                errorMessage = null,
            )
        }
        loadPage(page = 1, append = false)
    }

    fun loadMoreIfNeeded(currentIndex: Int) {
        val state = _uiState.value
        if (state.loading || state.loadingMore || state.refreshing || state.items.isEmpty() || !state.errorMessage.isNullOrBlank()) {
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

    private fun loadPage(page: Int, append: Boolean) {
        val versionAtRequest = requestVersion
        val sortBy = _uiState.value.sortBy
        val sortOrder = _uiState.value.sortOrder
        if (append) {
            _uiState.update { it.copy(loadingMore = true, errorMessage = null) }
        }
        viewModelScope.launch {
            repository.fetchCatalogWall(
                wallKind,
                page = page,
                pageSize = WALL_PAGE_SIZE,
                sortBy = sortBy,
                sortOrder = sortOrder,
            )
                .onSuccess { payload ->
                    if (versionAtRequest != requestVersion) {
                        return@onSuccess
                    }
                    val incoming = payload.items.map(::tvCatalogWallItemToUiModel)
                    _uiState.update { state ->
                        val merged = if (append) {
                            (state.items + incoming).distinctBy { it.id }
                        } else {
                            incoming.distinctBy { it.id }
                        }
                        state.copy(
                            loading = false,
                            loadingMore = false,
                            refreshing = false,
                            page = payload.page.coerceAtLeast(page),
                            totalCount = payload.totalCount.coerceAtLeast(merged.size),
                            items = merged,
                            errorMessage = null,
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
                            refreshing = false,
                            errorMessage = error.message ?: "海报墙加载失败",
                        )
                    }
                }
        }
    }

    private companion object {
        const val WALL_PAGE_SIZE = 24
    }
}

internal data class TvCatalogWallSortSelection(
    val sortBy: String,
    val sortOrder: String,
)

internal fun normalizeTvCatalogWallSort(sortBy: String, sortOrder: String): TvCatalogWallSortSelection {
    val normalizedBy = when (sortBy.trim().lowercase()) {
        "release" -> "release"
        else -> "added"
    }
    val normalizedOrder = when (sortOrder.trim().lowercase()) {
        "asc" -> "asc"
        else -> "desc"
    }
    return TvCatalogWallSortSelection(sortBy = normalizedBy, sortOrder = normalizedOrder)
}
