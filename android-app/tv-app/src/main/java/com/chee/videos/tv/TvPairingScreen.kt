package com.chee.videos.tv

import android.os.Build
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.ViewModel
import com.chee.videos.core.model.TvAuthSessionCreatePayload
import com.chee.videos.core.repository.TvAuthRepository
import com.chee.videos.core.ui.AppChrome
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TvPairingUiState(
    val loading: Boolean = true,
    val pairCode: String = "",
    val qrContent: String = "",
    val sessionId: String = "",
    val statusMessage: String? = null,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvPairingViewModel @Inject constructor(
    private val repository: TvAuthRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(TvPairingUiState())
    val uiState: StateFlow<TvPairingUiState> = _uiState.asStateFlow()

    fun startPairing() {
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, errorMessage = null, statusMessage = null) }
            repository.createSession(
                deviceId = "android-tv-${Build.DEVICE}-${Build.MODEL}".lowercase(),
                deviceName = listOf(Build.MANUFACTURER, Build.MODEL).joinToString(" ").trim(),
            ).onSuccess { payload ->
                _uiState.update {
                    it.copy(
                        loading = false,
                        pairCode = payload.pairCode,
                        qrContent = payload.qrContent,
                        sessionId = payload.sessionId,
                        statusMessage = "请用手机 App 扫码或输入配对码确认登录",
                    )
                }
                pollUntilApproved(payload)
            }.onFailure { err ->
                _uiState.update {
                    it.copy(
                        loading = false,
                        errorMessage = err.message ?: "创建配对会话失败",
                    )
                }
            }
        }
    }

    private fun pollUntilApproved(payload: TvAuthSessionCreatePayload) {
        viewModelScope.launch {
            while (true) {
                delay((payload.pollIntervalSeconds.coerceAtLeast(3) * 1000).toLong())
                val result = repository.fetchSession(payload.sessionId)
                result.onSuccess { session ->
                    when (session.status) {
                        "approved" -> {
                            repository.saveApprovedSession(session)
                            _uiState.update { it.copy(statusMessage = "已授权，正在进入 TV 首页") }
                            return@launch
                        }
                        "expired" -> {
                            _uiState.update { it.copy(errorMessage = "配对已过期，请重新生成", loading = false) }
                            return@launch
                        }
                        "denied" -> {
                            _uiState.update { it.copy(errorMessage = "手机端已拒绝本次登录", loading = false) }
                            return@launch
                        }
                    }
                }.onFailure { err ->
                    _uiState.update { it.copy(errorMessage = err.message ?: "轮询配对状态失败", loading = false) }
                    return@launch
                }
            }
        }
    }
}

@Composable
fun TvPairingScreen(
    viewModel: TvPairingViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        if (uiState.sessionId.isBlank()) {
            viewModel.startPairing()
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .statusBarsPadding()
            .padding(32.dp),
        contentAlignment = Alignment.Center,
    ) {
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = AppChrome.Surface,
            shape = AppChrome.CardShape,
            shadowElevation = 12.dp,
        ) {
            Column(
                modifier = Modifier.padding(24.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp),
            ) {
                Text(
                    text = "TV 配对登录",
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold,
                    color = AppChrome.TextPrimary,
                )
                Text(
                    text = "先在这台电视上选择服务器，再用已登录的手机 App 确认授权。",
                    color = AppChrome.TextSecondary,
                )
                if (uiState.pairCode.isNotBlank()) {
                    Text(
                        text = "配对码：${uiState.pairCode}",
                        style = MaterialTheme.typography.headlineSmall,
                        color = AppChrome.AccentStrong,
                    )
                    Text(
                        text = "扫码内容：${uiState.qrContent}",
                        color = AppChrome.TextMuted,
                    )
                }
                if (!uiState.statusMessage.isNullOrBlank()) {
                    Text(uiState.statusMessage.orEmpty(), color = AppChrome.TextSecondary)
                }
                if (!uiState.errorMessage.isNullOrBlank()) {
                    Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                }
                Button(
                    onClick = viewModel::startPairing,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = AppChrome.Accent,
                        contentColor = AppChrome.TextPrimary,
                    ),
                ) {
                    Text(if (uiState.loading) "生成中..." else "重新生成配对码")
                }
            }
        }
    }
}
