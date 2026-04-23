package com.chee.videos.feature.tv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.TvSectionDto
import com.chee.videos.core.model.TvSeriesSummaryDto
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
    val continueWatching: TvContinueWatchingUiModel? = null,
    val sections: List<TvCatalogSectionUiModel> = emptyList(),
    val searchResults: List<TvSeriesUiModel> = emptyList(),
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
        loadHome(query)
    }

    private fun loadHome(query: String = _uiState.value.query) {
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    loading = true,
                    errorMessage = null,
                    searchResults = if (query.isBlank()) emptyList() else it.searchResults,
                )
            }
            repository.fetchHome(query)
                .onSuccess { payload ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            continueWatching = payload.continueWatching?.let(::tvContinueWatchingToUiModel),
                            sections = coerceListOrEmpty<TvSectionDto>(payload.sections).map(::tvSectionToUiModel),
                            searchResults = coerceListOrEmpty<TvSeriesSummaryDto>(payload.searchResults).map(::tvSeriesSummaryToUiModel),
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { error ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            continueWatching = null,
                            sections = emptyList(),
                            searchResults = emptyList(),
                            errorMessage = error.message ?: "电视剧专区加载失败",
                        )
                    }
                }
        }
    }
}
