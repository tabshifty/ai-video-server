package com.chee.videos.feature.detail

import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import com.chee.videos.feature.tv.decodeTvRouteArg
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class DetailUiState(
    val loading: Boolean = true,
    val detail: VideoDetailDto? = null,
    val videoType: String = "",
    val baseUrl: String = "",
    val accessToken: String = "",
    val preferredPlaybackProfile: String = "",
    val errorMessage: String? = null,
)

@HiltViewModel
class DetailViewModel @Inject constructor(
    savedStateHandle: SavedStateHandle,
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
) : ViewModel() {

    private val videoId: String = decodeTvRouteArg(savedStateHandle["videoId"])
    private val videoType: String = savedStateHandle.get<String>("videoType").orEmpty()

    private val _uiState = MutableStateFlow(DetailUiState())
    val uiState: StateFlow<DetailUiState> = _uiState.asStateFlow()

    init {
        load()
    }

    fun load() {
        viewModelScope.launch {
            val baseUrl = videoRepository.readActiveBaseUrl().orEmpty()
            val accessToken = videoRepository.readAccessToken().orEmpty()

            _uiState.update {
                it.copy(
                    loading = true,
                    videoType = videoType,
                    baseUrl = baseUrl,
                    accessToken = accessToken,
                    preferredPlaybackProfile = videoRepository.preferredLongFormPlaybackProfile().wireValue,
                    errorMessage = null,
                )
            }
            if (videoId.isBlank()) {
                _uiState.update {
                    it.copy(
                        loading = false,
                        errorMessage = "视频不存在或播放目标无效",
                    )
                }
                return@launch
            }
            videoRepository.fetchDetail(videoId)
                .onSuccess { detail ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            detail = detail,
                            errorMessage = null,
                        )
                    }
                }
                .onFailure { err ->
                    if (err is AuthExpiredException) {
                        authRepository.logoutLocal()
                    }
                    _uiState.update {
                        it.copy(
                            loading = false,
                            errorMessage = err.message ?: "详情加载失败",
                        )
                    }
                }
        }
    }

    fun toggleLike() = toggleAction { videoRepository.toggleLike(videoId) }

    fun toggleFavorite() = toggleAction { videoRepository.toggleFavorite(videoId) }

    fun toggleDislike() = toggleAction { videoRepository.toggleDislike(videoId) }

    fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        if (videoId.isBlank() || watchSeconds <= 0) {
            return
        }
        viewModelScope.launch {
            videoRepository.reportHistory(videoId, watchSeconds, completed)
        }
    }

    private fun toggleAction(call: suspend () -> Result<ActionTogglePayload>) {
        viewModelScope.launch {
            call().onSuccess { result ->
                _uiState.update { state ->
                    val detail = state.detail ?: return@update state
                    val nextUserState = when (result.action) {
                        "like" -> detail.userState.copy(isLiked = result.enabled, isDisliked = false)
                        "favorite" -> detail.userState.copy(isFavorited = result.enabled)
                        "dislike" -> detail.userState.copy(isDisliked = result.enabled, isLiked = false, isFavorited = false)
                        else -> detail.userState
                    }
                    state.copy(detail = detail.copy(userState = nextUserState))
                }
            }.onFailure { err ->
                if (err is AuthExpiredException) {
                    authRepository.logoutLocal()
                }
                _uiState.update { it.copy(errorMessage = err.message ?: "操作失败") }
            }
        }
    }

}
