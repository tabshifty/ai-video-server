package com.chee.videos.feature.tv

import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class TvCatalogUiState(
    val loading: Boolean = true,
    val query: String = "",
    val continueWatching: TvContinueWatchingUiModel? = null,
    val sections: List<TvCatalogSectionUiModel> = emptyList(),
    val searchResults: List<TvSeriesUiModel> = emptyList(),
)

@HiltViewModel
class TvCatalogViewModel @Inject constructor() : ViewModel() {
    private val allSeries = TvMockData.allSeries()
    private val catalogSections = TvMockData.catalogSections()
    private val continueWatching = TvMockData.continueWatching()

    private val _uiState = MutableStateFlow(TvCatalogUiState())
    val uiState: StateFlow<TvCatalogUiState> = _uiState.asStateFlow()

    init {
        _uiState.value = TvCatalogUiState(
            loading = false,
            continueWatching = continueWatching,
            sections = catalogSections,
        )
    }

    fun updateQuery(query: String) {
        _uiState.value = _uiState.value.copy(
            query = query,
            searchResults = filterTvSeriesByQuery(allSeries, query),
        )
    }
}

internal fun filterTvSeriesByQuery(
    seriesList: List<TvSeriesUiModel>,
    query: String,
): List<TvSeriesUiModel> {
    val normalizedQuery = query.trim()
    if (normalizedQuery.isBlank()) {
        return emptyList()
    }
    return seriesList.filter { series ->
        series.title.contains(normalizedQuery, ignoreCase = true)
    }
}
