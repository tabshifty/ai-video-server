package com.chee.videos.feature.tv

import androidx.lifecycle.ViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow

data class TvCatalogUiState(
    val loading: Boolean = true,
    val continueWatching: TvContinueWatchingUiModel? = null,
    val sections: List<TvCatalogSectionUiModel> = emptyList(),
)

@HiltViewModel
class TvCatalogViewModel @Inject constructor() : ViewModel() {

    private val _uiState = MutableStateFlow(TvCatalogUiState())
    val uiState: StateFlow<TvCatalogUiState> = _uiState.asStateFlow()

    init {
        _uiState.value = TvCatalogUiState(
            loading = false,
            continueWatching = TvMockData.continueWatching(),
            sections = TvMockData.catalogSections(),
        )
    }
}
