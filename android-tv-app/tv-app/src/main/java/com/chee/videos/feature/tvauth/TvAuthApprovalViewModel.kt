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
    val sessionStatus: String = "pending",
    val loading: Boolean = false,
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
                sessionStatus = "pending",
                loading = true,
                message = null,
                isError = false,
            )
        }
        viewModelScope.launch {
            repository.fetchSession(
                sessionId = deepLink.sessionId,
                explicitBaseUrl = deepLink.serverBaseUrl,
            ).onSuccess { payload ->
                _uiState.update {
                    it.copy(
                        pairCode = payload.pairCode?.takeIf(String::isNotBlank) ?: it.pairCode,
                        deviceName = payload.deviceName?.takeIf(String::isNotBlank) ?: it.deviceName,
                        serverBaseUrl = payload.serverBaseUrl?.takeIf(String::isNotBlank) ?: it.serverBaseUrl,
                        sessionStatus = payload.status,
                        loading = false,
                        message = sessionStatusMessage(payload.status),
                        isError = payload.status == "expired" || payload.status == "denied",
                    )
                }
            }.onFailure { err ->
                _uiState.update {
                    it.copy(
                        loading = false,
                        message = err.message ?: "读取配对会话失败",
                        isError = true,
                    )
                }
            }
        }
    }

    fun approve() {
        val state = _uiState.value
        val sessionId = state.sessionId
        if (sessionId.isBlank()) {
            _uiState.update { it.copy(message = "缺少配对会话", isError = true) }
            return
        }
        if (state.sessionStatus != "pending") {
            _uiState.update { it.copy(message = sessionStatusMessage(state.sessionStatus), isError = true) }
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(submitting = true, message = null, isError = false) }
            repository.approve(sessionId)
                .onSuccess {
                    _uiState.update {
                        it.copy(
                            submitting = false,
                            sessionStatus = "approved",
                            message = "已授权，电视端会自动继续登录",
                        )
                    }
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
        val state = _uiState.value
        val sessionId = state.sessionId
        if (sessionId.isBlank()) {
            return
        }
        if (state.sessionStatus != "pending") {
            return
        }
        viewModelScope.launch {
            repository.deny(sessionId)
            _uiState.update {
                it.copy(
                    sessionStatus = "denied",
                    message = "已拒绝本次 TV 登录请求",
                    isError = true,
                )
            }
        }
    }

    private fun sessionStatusMessage(status: String): String? = when (status) {
        "approved" -> "这台 TV 已完成授权"
        "expired" -> "本次配对已过期，请让电视端重新生成"
        "denied" -> "本次 TV 登录请求已被拒绝"
        else -> null
    }
}
