package com.chee.videos.feature.tvauth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.repository.TvAuthRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TvAuthApprovalUiState(
    val sessionId: String = "",
    val pairCode: String = "",
    val deviceName: String = "",
    val serverBaseUrl: String = "",
    val submitting: Boolean = false,
    val message: String? = null,
    val isError: Boolean = false,
)

@HiltViewModel
class TvAuthApprovalViewModel @Inject constructor(
    private val repository: TvAuthRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(TvAuthApprovalUiState())
    val uiState: StateFlow<TvAuthApprovalUiState> = _uiState.asStateFlow()

    fun bind(deepLink: TvAuthDeepLink) {
        _uiState.update {
            it.copy(
                sessionId = deepLink.sessionId,
                pairCode = deepLink.pairCode,
                deviceName = deepLink.deviceName,
                serverBaseUrl = deepLink.serverBaseUrl.orEmpty(),
                message = null,
                isError = false,
            )
        }
    }

    fun approve() {
        val sessionId = _uiState.value.sessionId
        if (sessionId.isBlank()) {
            _uiState.update { it.copy(message = "缺少配对会话", isError = true) }
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(submitting = true, message = null, isError = false) }
            repository.approve(sessionId)
                .onSuccess {
                    _uiState.update { it.copy(submitting = false, message = "已授权，电视端会自动继续登录") }
                }
                .onFailure { err ->
                    _uiState.update {
                        it.copy(
                            submitting = false,
                            message = err.message ?: "授权失败",
                            isError = true,
                        )
                    }
                }
        }
    }

    fun deny() {
        val sessionId = _uiState.value.sessionId
        if (sessionId.isBlank()) {
            return
        }
        viewModelScope.launch {
            repository.deny(sessionId)
        }
    }
}
