package com.chee.videos.feature.home

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
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
)

@HiltViewModel
class HomeViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {

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
}
