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

data class CategoryState(
    val loading: Boolean = false,
    val loaded: Boolean = false,
    val items: List<VideoListItemDto> = emptyList(),
    val errorMessage: String? = null,
)

data class HomeUiState(
    val movie: CategoryState = CategoryState(),
    val episode: CategoryState = CategoryState(),
    val av: CategoryState = CategoryState(),
    val avSearch: AvSearchState = AvSearchState(),
)

data class AvSearchState(
    val query: String = "",
    val isSearchMode: Boolean = false,
    val loading: Boolean = false,
    val results: List<VideoListItemDto> = emptyList(),
    val totalCount: Int = 0,
    val lastCompletedQuery: String = "",
    val errorMessage: String? = null,
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
        val current = stateFor(type)
        if (current.loaded && !force) {
            return
        }

        viewModelScope.launch {
            updateState(type) { it.copy(loading = true, errorMessage = null) }
            videoRepository.fetchCategory(type)
                .onSuccess { items ->
                    updateState(type) {
                        it.copy(
                            loading = false,
                            loaded = true,
                            items = items,
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    if (err is AuthExpiredException) {
                        authRepository.logoutLocal()
                    }
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
                    errorMessage = null,
                ),
            )
        }

        val normalizedQuery = rawQuery.trim()
        if (normalizedQuery.isBlank()) {
            _uiState.update { state ->
                state.copy(avSearch = AvSearchState())
            }
            loadCategory("av")
            return
        }

        avSearchJob = viewModelScope.launch {
            delay(avSearchDebounceMs)
            performAvSearch(normalizedQuery)
        }
    }

    fun retryAvState() {
        val normalizedQuery = _uiState.value.avSearch.query.trim()
        if (normalizedQuery.isBlank()) {
            loadCategory("av", force = true)
            return
        }
        avSearchJob?.cancel()
        viewModelScope.launch {
            performAvSearch(normalizedQuery)
        }
    }

    private suspend fun performAvSearch(query: String) {
        _uiState.update { state ->
            state.copy(
                avSearch = state.avSearch.copy(
                    query = query,
                    isSearchMode = true,
                    loading = true,
                    errorMessage = null,
                ),
            )
        }
        videoRepository.searchAv(query)
            .onSuccess { payload ->
                if (_uiState.value.avSearch.query.trim() != query) {
                    return@onSuccess
                }
                applyAvSearchResult(query, payload)
            }
            .onFailure { err ->
                if (err is AuthExpiredException) {
                    authRepository.logoutLocal()
                }
                if (_uiState.value.avSearch.query.trim() != query) {
                    return@onFailure
                }
                _uiState.update { state ->
                    state.copy(
                        avSearch = state.avSearch.copy(
                            isSearchMode = true,
                            loading = false,
                            results = emptyList(),
                            totalCount = 0,
                            lastCompletedQuery = query,
                            errorMessage = err.message ?: "搜索失败",
                        ),
                    )
                }
            }
    }

    private fun applyAvSearchResult(query: String, payload: SearchPayload) {
        _uiState.update { state ->
            state.copy(
                avSearch = state.avSearch.copy(
                    isSearchMode = true,
                    loading = false,
                    results = payload.items,
                    totalCount = payload.totalCount,
                    lastCompletedQuery = query,
                    errorMessage = null,
                ),
            )
        }
    }

    private fun stateFor(type: String): CategoryState {
        return when (type) {
            "movie" -> _uiState.value.movie
            "episode" -> _uiState.value.episode
            "av" -> _uiState.value.av
            else -> CategoryState()
        }
    }

    private fun updateState(type: String, transform: (CategoryState) -> CategoryState) {
        _uiState.update { state ->
            when (type) {
                "movie" -> state.copy(movie = transform(state.movie))
                "episode" -> state.copy(episode = transform(state.episode))
                "av" -> state.copy(av = transform(state.av))
                else -> state
            }
        }
    }

    private companion object {
        const val AV_SEARCH_DEBOUNCE_MS = 350L
    }
}
