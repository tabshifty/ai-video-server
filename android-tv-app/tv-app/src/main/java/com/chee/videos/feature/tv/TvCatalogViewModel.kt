package com.chee.videos.feature.tv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.TvHomeVideoDto
import com.chee.videos.core.model.TvSectionDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import com.chee.videos.core.model.TvSearchResultDto
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TvCatalogUiState(
    val loading: Boolean = true,
    val selectedMenu: TvHomeMenuItem = TvHomeMenuItem.defaultSelected(),
    val query: String = "",
    val kind: String = "tv",
    val baseUrl: String = "",
    val featured: TvHomeShelfItemUiModel? = null,
    val recentWatching: List<TvHomeShelfItemUiModel> = emptyList(),
    val recentUpdates: List<TvHomeShelfItemUiModel> = emptyList(),
    val continueWatching: TvContinueWatchingUiModel? = null,
    val sections: List<TvCatalogSectionUiModel> = emptyList(),
    val tvSeries: List<TvHomeShelfItemUiModel> = emptyList(),
    val movies: List<TvHomeShelfItemUiModel> = emptyList(),
    val av: List<TvHomeShelfItemUiModel> = emptyList(),
    val searchResults: List<TvSearchResultUiModel> = emptyList(),
    val tvSeekStepSeconds: Int = TvPlaybackSeekStepSetting.defaultSeconds,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvCatalogViewModel @Inject constructor(
    private val repository: TvRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(TvCatalogUiState())
    val uiState: StateFlow<TvCatalogUiState> = _uiState.asStateFlow()

    init {
        loadHome()
    }

    fun selectMenu(menuItem: TvHomeMenuItem) {
        _uiState.update { it.copy(selectedMenu = menuItem) }
        when {
            menuItem.isContentKind -> loadHome(menuItem.homeKind)
            menuItem == TvHomeMenuItem.Search -> {
                _uiState.update {
                    it.copy(
                        loading = false,
                        query = "",
                        searchResults = emptyList(),
                        errorMessage = null,
                    )
                }
            }
            menuItem == TvHomeMenuItem.Settings -> {
                _uiState.update { it.copy(loading = false, query = "", searchResults = emptyList()) }
                loadTvPlaybackSettings()
            }
            menuItem == TvHomeMenuItem.Iptv -> {
                _uiState.update { it.copy(loading = false, query = "", searchResults = emptyList()) }
            }
        }
    }

    fun selectTvSeekStepSeconds(seconds: Int) {
        val normalized = TvPlaybackSeekStepSetting.normalize(seconds)
        _uiState.update { it.copy(tvSeekStepSeconds = normalized) }
        viewModelScope.launch {
            repository.saveTvSeekStepSeconds(normalized)
        }
    }

    fun updateQuery(query: String) {
        _uiState.update { it.copy(query = query, selectedMenu = TvHomeMenuItem.Search) }
        if (query.isBlank()) {
            _uiState.update { it.copy(loading = false, searchResults = emptyList(), errorMessage = null) }
        } else {
            loadSearch(query)
        }
    }

    private fun loadHome(kind: String = _uiState.value.selectedMenu.homeKind.ifBlank { "tv" }) {
        val normalizedKind = normalizeTvHomeKind(kind)
        viewModelScope.launch {
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            _uiState.update {
                it.copy(
                    loading = true,
                    kind = normalizedKind,
                    baseUrl = baseUrl,
                    errorMessage = null,
                    searchResults = emptyList(),
                )
            }
            repository.fetchHome(kind = normalizedKind)
                .onSuccess { payload ->
                    val recentUpdates = coerceListOrEmpty<TvHomeVideoDto>(payload.recentUpdates).map(::tvHomeVideoToUiModel)
                    val typedRecentWatching = coerceListOrEmpty<TvHomeVideoDto>(payload.recentWatching).map(::tvHomeVideoToUiModel)
                    val featured = payload.featured?.let(::tvHomeVideoToUiModel)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            baseUrl = baseUrl,
                            kind = normalizeTvHomeKind(payload.kind),
                            featured = featured,
                            recentWatching = typedRecentWatching,
                            recentUpdates = recentUpdates,
                            continueWatching = payload.continueWatching?.let(::tvContinueWatchingToUiModel),
                            sections = coerceListOrEmpty<TvSectionDto>(payload.sections).map(::tvSectionToUiModel),
                            tvSeries = coerceListOrEmpty<TvHomeVideoDto>(payload.tvSeries).map(::tvHomeVideoToUiModel),
                            movies = coerceListOrEmpty<TvHomeVideoDto>(payload.movies).map(::tvHomeVideoToUiModel),
                            av = coerceListOrEmpty<TvHomeVideoDto>(payload.av).map(::tvHomeVideoToUiModel),
                            searchResults = emptyList(),
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { error ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            baseUrl = baseUrl,
                            featured = null,
                            recentWatching = emptyList(),
                            recentUpdates = emptyList(),
                            continueWatching = null,
                            sections = emptyList(),
                            tvSeries = emptyList(),
                            movies = emptyList(),
                            av = emptyList(),
                            searchResults = emptyList(),
                            errorMessage = error.message ?: "TV 首页加载失败",
                        )
                    }
                }
        }
    }

    private fun loadSearch(query: String) {
        viewModelScope.launch {
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            _uiState.update {
                it.copy(
                    loading = true,
                    baseUrl = baseUrl,
                    errorMessage = null,
                )
            }
            repository.fetchSearch(query)
                .onSuccess { payload ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            baseUrl = baseUrl,
                            continueWatching = null,
                            searchResults = coerceListOrEmpty<TvSearchResultDto>(payload.items).map(::tvSearchResultToUiModel),
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { error ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            baseUrl = baseUrl,
                            searchResults = emptyList(),
                            errorMessage = error.message ?: "TV 搜索失败",
                        )
                    }
                }
        }
    }

    private fun loadTvPlaybackSettings() {
        viewModelScope.launch {
            val stepSeconds = TvPlaybackSeekStepSetting.normalize(repository.readTvSeekStepSeconds())
            _uiState.update { it.copy(tvSeekStepSeconds = stepSeconds) }
        }
    }
}
