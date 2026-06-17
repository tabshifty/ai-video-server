package com.chee.videos.feature.tv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.TvHomeVideoDto
import com.chee.videos.core.model.TvSectionDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import com.chee.videos.core.model.TvSearchResultDto
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
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
    val searchLoading: Boolean = false,
    val tvSeekStepSeconds: Int = TvPlaybackSeekStepSetting.defaultSeconds,
    val seriesAutoplayEnabled: Boolean = TvSeriesAutoplaySetting.DEFAULT_ENABLED,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvCatalogViewModel @Inject constructor(
    private val repository: TvRepository,
) : ViewModel() {
    private var requestVersion = 0
    private var searchJob: Job? = null
    private val _uiState = MutableStateFlow(TvCatalogUiState())
    val uiState: StateFlow<TvCatalogUiState> = _uiState.asStateFlow()

    init {
        loadHome()
    }

    fun selectMenu(menuItem: TvHomeMenuItem) {
        when {
            menuItem.isContentKind -> {
                _uiState.update { it.copy(selectedMenu = menuItem) }
                loadHome(menuItem.homeKind)
            }
            menuItem == TvHomeMenuItem.Search -> {
                invalidateCatalogRequests()
                _uiState.update {
                    it.copy(
                        loading = false,
                        selectedMenu = TvHomeMenuItem.Search,
                        query = "",
                        searchResults = emptyList(),
                        searchLoading = false,
                        errorMessage = null,
                    )
                }
            }
            menuItem == TvHomeMenuItem.Settings -> {
                invalidateCatalogRequests()
                _uiState.update {
                    it.copy(
                        loading = false,
                        selectedMenu = TvHomeMenuItem.Settings,
                        query = "",
                        searchResults = emptyList(),
                        searchLoading = false,
                        errorMessage = null,
                    )
                }
                loadTvPlaybackSettings()
            }
            menuItem == TvHomeMenuItem.Iptv -> {
                openIptv()
            }
        }
    }

    fun openIptv() {
        invalidateCatalogRequests()
        _uiState.update {
            it.copy(
                loading = false,
                query = "",
                searchResults = emptyList(),
                searchLoading = false,
                errorMessage = null,
            )
        }
    }

    fun selectTvSeekStepSeconds(seconds: Int) {
        val normalized = TvPlaybackSeekStepSetting.normalize(seconds)
        _uiState.update { it.copy(tvSeekStepSeconds = normalized) }
        viewModelScope.launch {
            repository.saveTvSeekStepSeconds(normalized)
        }
    }

    fun setSeriesAutoplayEnabled(enabled: Boolean) {
        _uiState.update { it.copy(seriesAutoplayEnabled = enabled) }
        viewModelScope.launch {
            repository.saveTvSeriesAutoplayEnabled(enabled)
        }
    }

    fun updateQuery(query: String) {
        _uiState.update { it.copy(query = query, selectedMenu = TvHomeMenuItem.Search) }
        if (query.isBlank()) {
            searchJob?.cancel()
            searchJob = null
            invalidateCatalogRequests()
            _uiState.update {
                it.copy(
                    loading = false,
                    searchLoading = false,
                    searchResults = emptyList(),
                    errorMessage = null,
                )
            }
        } else {
            scheduleSearch(query)
        }
    }

    fun retry() {
        val state = _uiState.value
        if (state.selectedMenu == TvHomeMenuItem.Search && state.query.isNotBlank()) {
            loadSearch(state.query)
        } else {
            loadHome(state.kind)
        }
    }

    private fun loadHome(kind: String = _uiState.value.selectedMenu.homeKind.ifBlank { "tv" }) {
        val versionAtRequest = nextCatalogRequestVersion()
        val normalizedKind = normalizeTvHomeKind(kind)
        viewModelScope.launch {
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            if (versionAtRequest != requestVersion) {
                return@launch
            }
            _uiState.update {
                it.copy(
                    loading = true,
                    kind = normalizedKind,
                    baseUrl = baseUrl,
                    errorMessage = null,
                    searchResults = emptyList(),
                    searchLoading = false,
                )
            }
            repository.fetchHome(kind = normalizedKind)
                .onSuccess { payload ->
                    if (versionAtRequest != requestVersion) {
                        return@onSuccess
                    }
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
                            searchLoading = false,
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
                            searchLoading = false,
                            errorMessage = error.message ?: "TV 首页加载失败",
                        )
                    }
                }
        }
    }

    private fun scheduleSearch(query: String) {
        searchJob?.cancel()
        val versionAtSchedule = nextCatalogRequestVersion()
        _uiState.update {
            it.copy(
                loading = false,
                searchLoading = true,
                errorMessage = null,
            )
        }
        searchJob = viewModelScope.launch {
            delay(TvSearchInputDebounceMillis)
            if (versionAtSchedule != requestVersion) {
                return@launch
            }
            loadSearch(query, versionAtSchedule)
        }
    }

    private fun loadSearch(query: String) {
        searchJob?.cancel()
        searchJob = null
        val versionAtRequest = nextCatalogRequestVersion()
        loadSearch(query, versionAtRequest)
    }

    private fun loadSearch(query: String, versionAtRequest: Int) {
        viewModelScope.launch {
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            if (versionAtRequest != requestVersion) {
                return@launch
            }
            _uiState.update {
                it.copy(
                    loading = false,
                    searchLoading = true,
                    baseUrl = baseUrl,
                    errorMessage = null,
                )
            }
            repository.fetchSearch(query)
                .onSuccess { payload ->
                    if (versionAtRequest != requestVersion) {
                        return@onSuccess
                    }
                    _uiState.update {
                        it.copy(
                            loading = false,
                            searchLoading = false,
                            baseUrl = baseUrl,
                            continueWatching = null,
                            searchResults = coerceListOrEmpty<TvSearchResultDto>(payload.items).map(::tvSearchResultToUiModel),
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
                            searchLoading = false,
                            baseUrl = baseUrl,
                            searchResults = emptyList(),
                            errorMessage = error.message ?: "TV 搜索失败",
                        )
                    }
                }
        }
    }

    private fun nextCatalogRequestVersion(): Int {
        requestVersion += 1
        return requestVersion
    }

    private fun invalidateCatalogRequests() {
        searchJob?.cancel()
        searchJob = null
        requestVersion += 1
    }

    private fun loadTvPlaybackSettings() {
        viewModelScope.launch {
            val stepSeconds = TvPlaybackSeekStepSetting.normalize(repository.readTvSeekStepSeconds())
            val autoplayEnabled = TvSeriesAutoplaySetting.parse(repository.readTvSeriesAutoplayEnabled())
            _uiState.update {
                it.copy(
                    tvSeekStepSeconds = stepSeconds,
                    seriesAutoplayEnabled = autoplayEnabled,
                )
            }
        }
    }

    private companion object {
        const val TvSearchInputDebounceMillis = 300L
    }
}
