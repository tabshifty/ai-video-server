package com.chee.videos.feature.shorts

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ShortFeedUiState(
    val loading: Boolean = false,
    val loaded: Boolean = false,
    val items: List<FeedVideoDto> = emptyList(),
    val errorMessage: String? = null,
)

@HiltViewModel
class ShortFeedViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    store: AppPreferencesStore,
) : ViewModel() {

    private val _uiState = MutableStateFlow(ShortFeedUiState())
    val uiState: StateFlow<ShortFeedUiState> = _uiState.asStateFlow()

    val accessToken = store.accessTokenFlow.stateIn(
        scope = viewModelScope,
        started = kotlinx.coroutines.flow.SharingStarted.WhileSubscribed(5000),
        initialValue = null,
    )

    fun load(force: Boolean = false) {
        if (_uiState.value.loaded && !force) {
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, errorMessage = null) }
            videoRepository.fetchShortFeed()
                .onSuccess { items ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            loaded = true,
                            items = items,
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            loaded = true,
                            errorMessage = err.message ?: "短视频加载失败",
                        )
                    }
                }
        }
    }
}
