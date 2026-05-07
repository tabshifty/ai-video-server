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
    val query: String = "",
    val baseUrl: String = "",
    val continueWatching: TvContinueWatchingUiModel? = null,
    val sections: List<TvCatalogSectionUiModel> = emptyList(),
    val tvSeries: List<TvHomeShelfItemUiModel> = emptyList(),
    val movies: List<TvHomeShelfItemUiModel> = emptyList(),
    val av: List<TvHomeShelfItemUiModel> = emptyList(),
    val searchResults: List<TvSearchResultUiModel> = emptyList(),
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

    fun updateQuery(query: String) {
        _uiState.update { it.copy(query = query) }
        if (query.isBlank()) {
            loadHome()
        } else {
            loadSearch(query)
        }
    }

    private fun loadHome() {
        viewModelScope.launch {
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            _uiState.update {
                it.copy(
                    loading = true,
                    baseUrl = baseUrl,
                    errorMessage = null,
                    searchResults = emptyList(),
                )
            }
            repository.fetchHome()
                .onSuccess { payload ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            baseUrl = baseUrl,
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
}
