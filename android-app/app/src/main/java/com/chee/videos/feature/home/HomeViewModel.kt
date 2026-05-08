package com.chee.videos.feature.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.SearchPayload
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

private const val DEFAULT_PAGED_PAGE_SIZE = 30

data class PagedVideoListState(
    val loading: Boolean = false,
    val loaded: Boolean = false,
    val refreshing: Boolean = false,
    val loadingMore: Boolean = false,
    val items: List<VideoListItemDto> = emptyList(),
    val page: Int = 0,
    val pageSize: Int = DEFAULT_PAGED_PAGE_SIZE,
    val totalCount: Int = 0,
    val hasMore: Boolean = true,
    val errorMessage: String? = null,
    val loadMoreErrorMessage: String? = null,
)

data class HomeUiState(
    val movie: PagedVideoListState = PagedVideoListState(),
    val episode: PagedVideoListState = PagedVideoListState(),
    val av: PagedVideoListState = PagedVideoListState(),
    val avSearch: AvSearchState = AvSearchState(),
)

data class AvSearchState(
    val query: String = "",
    val isSearchMode: Boolean = false,
    val lastCompletedQuery: String = "",
    val resultState: PagedVideoListState = PagedVideoListState(),
)

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {
    private var avSearchJob: Job? = null
    internal var avSearchDebounceMs: Long = AV_SEARCH_DEBOUNCE_MS

    private val _uiState = MutableStateFlow(HomeUiState())
    val uiState: StateFlow<HomeUiState> = _uiState.asStateFlow()

    fun loadCategory(type: String, force: Boolean = false) {
        if (type == "av") {
            loadAvBrowse(forceRefresh = force)
            return
        }

        val current = stateFor(type)
        if (current.loaded && !force) {
            return
        }

        viewModelScope.launch {
            updateState(type) { it.copy(loading = true, errorMessage = null) }
            videoRepository.fetchCategory(type)
                .onSuccess { payload ->
                    updateState(type) {
                        it.copy(
                            loading = false,
                            loaded = true,
                            items = payload.items,
                            page = payload.page.coerceAtLeast(1),
                            pageSize = payload.pageSize.coerceAtLeast(1),
                            totalCount = payload.totalCount.coerceAtLeast(payload.items.size),
                            hasMore = payload.totalCount > payload.items.size,
                            errorMessage = null,
                            loadMoreErrorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    updateState(type) {
                        it.copy(
                            loading = false,
                            loaded = true,
                            errorMessage = err.message ?: "加载失败",
                        )
                    }
                }
        }
    }

    fun updateAvQuery(rawQuery: String) {
        avSearchJob?.cancel()
        _uiState.update { state ->
            state.copy(
                avSearch = state.avSearch.copy(
                    query = rawQuery,
                    resultState = state.avSearch.resultState.copy(
                        errorMessage = null,
                        loadMoreErrorMessage = null,
                    ),
                ),
            )
        }

        val normalizedQuery = rawQuery.trim()
        if (normalizedQuery.isBlank()) {
            _uiState.update { state ->
                state.copy(avSearch = AvSearchState())
            }
            if (!_uiState.value.av.loaded) {
                loadAvBrowse(forceRefresh = false)
            }
            return
        }

        avSearchJob = viewModelScope.launch {
            delay(avSearchDebounceMs)
            performAvSearch(normalizedQuery, forceRefresh = false)
        }
    }

    fun retryAvState() {
        val normalizedQuery = _uiState.value.avSearch.query.trim()
        if (normalizedQuery.isBlank()) {
            loadAvBrowse(forceRefresh = true)
            return
        }
        avSearchJob?.cancel()
        viewModelScope.launch {
            performAvSearch(normalizedQuery, forceRefresh = false)
        }
    }

    fun refreshAvState() {
        val normalizedQuery = _uiState.value.avSearch.query.trim()
        if (_uiState.value.avSearch.isSearchMode && normalizedQuery.isNotBlank()) {
            avSearchJob?.cancel()
            viewModelScope.launch {
                performAvSearch(normalizedQuery, forceRefresh = true)
            }
        } else {
            loadAvBrowse(forceRefresh = true)
        }
    }

    fun loadMoreAvIfNeeded(currentIndex: Int) {
        val searchState = _uiState.value.avSearch
        if (searchState.isSearchMode && searchState.query.trim().isNotBlank()) {
            val current = searchState.resultState
            if (!shouldLoadMore(current, currentIndex)) {
                return
            }
            loadAvSearchPage(
                query = searchState.query.trim(),
                page = current.page + 1,
                append = true,
                forceRefresh = false,
            )
            return
        }

        val browseState = _uiState.value.av
        if (!shouldLoadMore(browseState, currentIndex)) {
            return
        }
        loadAvBrowsePage(page = browseState.page + 1, append = true, forceRefresh = false)
    }

    private fun loadAvBrowse(forceRefresh: Boolean) {
        val current = _uiState.value.av
        if (current.loaded && !forceRefresh) {
            return
        }
        loadAvBrowsePage(page = 1, append = false, forceRefresh = forceRefresh)
    }

    private fun loadAvBrowsePage(
        page: Int,
        append: Boolean,
        forceRefresh: Boolean,
    ) {
        val current = _uiState.value.av
        val pageSize = current.pageSize.coerceAtLeast(DEFAULT_PAGED_PAGE_SIZE)

        _uiState.update { state ->
            state.copy(
                av = preparePagedState(
                    current = state.av,
                    append = append,
                    forceRefresh = forceRefresh,
                ),
            )
        }

        viewModelScope.launch {
            videoRepository.fetchCategory(type = "av", page = page, pageSize = pageSize)
                .onSuccess { payload ->
                    _uiState.update { state ->
                        state.copy(
                            av = consumePagedPayload(
                                current = state.av,
                                payload = payload,
                                append = append,
                            ),
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update { state ->
                        state.copy(
                            av = consumePagedError(
                                current = state.av,
                                message = err.message ?: "AV 海报墙加载失败",
                                append = append,
                            ),
                        )
                    }
                }
        }
    }

    private suspend fun performAvSearch(query: String, forceRefresh: Boolean) {
        loadAvSearchPage(query = query, page = 1, append = false, forceRefresh = forceRefresh)
    }

    private fun loadAvSearchPage(
        query: String,
        page: Int,
        append: Boolean,
        forceRefresh: Boolean,
    ) {
        val normalizedQuery = query.trim()
        if (normalizedQuery.isBlank()) {
            return
        }

        val currentSearch = _uiState.value.avSearch
        val pageSize = currentSearch.resultState.pageSize.coerceAtLeast(DEFAULT_PAGED_PAGE_SIZE)
        val keepExistingItems = append || (forceRefresh && currentSearch.resultState.items.isNotEmpty())

        _uiState.update { state ->
            val searchState = state.avSearch
            state.copy(
                avSearch = searchState.copy(
                    query = normalizedQuery,
                    isSearchMode = true,
                    resultState = preparePagedState(
                        current = searchState.resultState,
                        append = append,
                        forceRefresh = forceRefresh,
                        clearItems = !keepExistingItems && searchState.lastCompletedQuery != normalizedQuery,
                    ),
                ),
            )
        }

        viewModelScope.launch {
            videoRepository.searchAv(normalizedQuery, page = page, pageSize = pageSize)
                .onSuccess { payload ->
                    if (_uiState.value.avSearch.query.trim() != normalizedQuery) {
                        return@onSuccess
                    }
                    _uiState.update { state ->
                        state.copy(
                            avSearch = state.avSearch.copy(
                                isSearchMode = true,
                                lastCompletedQuery = normalizedQuery,
                                resultState = consumePagedPayload(
                                    current = state.avSearch.resultState,
                                    payload = payload,
                                    append = append,
                                ),
                            ),
                        )
                    }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    if (_uiState.value.avSearch.query.trim() != normalizedQuery) {
                        return@onFailure
                    }
                    _uiState.update { state ->
                        state.copy(
                            avSearch = state.avSearch.copy(
                                isSearchMode = true,
                                lastCompletedQuery = normalizedQuery,
                                resultState = consumePagedError(
                                    current = state.avSearch.resultState,
                                    message = err.message ?: "搜索失败",
                                    append = append,
                                ),
                            ),
                        )
                    }
                }
        }
    }

    private fun shouldLoadMore(state: PagedVideoListState, currentIndex: Int): Boolean {
        if (state.loading || state.refreshing || state.loadingMore || state.items.isEmpty()) {
            return false
        }
        if (!state.hasMore) {
            return false
        }
        return currentIndex >= (state.items.lastIndex - AV_LOAD_MORE_THRESHOLD).coerceAtLeast(0)
    }

    private fun preparePagedState(
        current: PagedVideoListState,
        append: Boolean,
        forceRefresh: Boolean,
        clearItems: Boolean = false,
    ): PagedVideoListState {
        return when {
            append -> current.copy(
                loadingMore = true,
                errorMessage = null,
                loadMoreErrorMessage = null,
            )
            forceRefresh && current.items.isNotEmpty() -> current.copy(
                loading = false,
                refreshing = true,
                errorMessage = null,
                loadMoreErrorMessage = null,
            )
            else -> current.copy(
                loading = true,
                refreshing = false,
                loadingMore = false,
                items = if (clearItems) emptyList() else current.items,
                page = if (clearItems) 0 else current.page,
                totalCount = if (clearItems) 0 else current.totalCount,
                hasMore = if (clearItems) true else current.hasMore,
                errorMessage = null,
                loadMoreErrorMessage = null,
            )
        }
    }

    private fun consumePagedPayload(
        current: PagedVideoListState,
        payload: SearchPayload,
        append: Boolean,
    ): PagedVideoListState {
        val mergedItems = if (append) {
            (current.items + payload.items).distinctBy { it.id }
        } else {
            payload.items.distinctBy { it.id }
        }
        return current.copy(
            loading = false,
            loaded = true,
            refreshing = false,
            loadingMore = false,
            items = mergedItems,
            page = payload.page.coerceAtLeast(1),
            pageSize = payload.pageSize.coerceAtLeast(1),
            totalCount = payload.totalCount.coerceAtLeast(mergedItems.size),
            hasMore = payload.totalCount > mergedItems.size,
            errorMessage = null,
            loadMoreErrorMessage = null,
        )
    }

    private fun consumePagedError(
        current: PagedVideoListState,
        message: String,
        append: Boolean,
    ): PagedVideoListState {
        return if (append) {
            current.copy(
                loading = false,
                refreshing = false,
                loadingMore = false,
                loadMoreErrorMessage = message,
            )
        } else {
            current.copy(
                loading = false,
                loaded = true,
                refreshing = false,
                loadingMore = false,
                errorMessage = message,
            )
        }
    }

    private fun stateFor(type: String): PagedVideoListState {
        return when (type) {
            "movie" -> _uiState.value.movie
            "episode" -> _uiState.value.episode
            "av" -> _uiState.value.av
            else -> PagedVideoListState()
        }
    }

    private fun updateState(type: String, transform: (PagedVideoListState) -> PagedVideoListState) {
        _uiState.update { state ->
            when (type) {
                "movie" -> state.copy(movie = transform(state.movie))
                "episode" -> state.copy(episode = transform(state.episode))
                "av" -> state.copy(av = transform(state.av))
                else -> state
            }
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is AuthExpiredException) {
            viewModelScope.launch {
                authRepository.logoutLocal()
            }
        }
    }

    private companion object {
        const val AV_SEARCH_DEBOUNCE_MS = 350L
        const val AV_LOAD_MORE_THRESHOLD = 5
    }
}
