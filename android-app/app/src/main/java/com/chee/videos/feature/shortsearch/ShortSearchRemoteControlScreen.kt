package com.chee.videos.feature.shortsearch

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import android.net.Uri
import android.widget.Toast
import com.chee.videos.core.model.TvRemoteSessionDto
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

const val ShortSearchRemoteSessionIdArg = "sessionId"
const val ShortSearchRemoteControlRoutePattern = "short-search-remote/{$ShortSearchRemoteSessionIdArg}"

fun buildShortSearchRemoteControlRoute(sessionId: String): String =
    "short-search-remote/${Uri.encode(sessionId)}"

data class ShortSearchRemoteControlUiState(
    val loading: Boolean = true,
    val session: TvRemoteSessionDto? = null,
    val actionLoading: Boolean = false,
    val errorMessage: String? = null,
    val endedMessage: String? = null,
)

@HiltViewModel
class ShortSearchRemoteControlViewModel @Inject constructor(
    private val videoRepository: VideoRepository,
    private val authRepository: AuthRepository,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val sessionId = savedStateHandle.get<String>(ShortSearchRemoteSessionIdArg).orEmpty()
    private val _uiState = MutableStateFlow(ShortSearchRemoteControlUiState())
    val uiState: StateFlow<ShortSearchRemoteControlUiState> = _uiState.asStateFlow()
    private var pollingJob: Job? = null

    init {
        startPolling()
    }

    fun previous() {
        runAction { videoRepository.tvRemotePrevious(sessionId) }
    }

    fun next() {
        runAction { videoRepository.tvRemoteNext(sessionId) }
    }

    fun consumeEndedMessage() {
        _uiState.update { it.copy(endedMessage = null) }
    }

    fun refreshNow() {
        viewModelScope.launch { refreshSession() }
    }

    private fun runAction(block: suspend () -> Result<TvRemoteSessionDto>) {
        if (_uiState.value.actionLoading) {
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(actionLoading = true, errorMessage = null) }
            block().onSuccess { session ->
                applySession(session)
                _uiState.update { it.copy(actionLoading = false) }
            }.onFailure { err ->
                handleAuthError(err)
                _uiState.update {
                    it.copy(
                        actionLoading = false,
                        errorMessage = err.message ?: "控制投放失败",
                    )
                }
            }
        }
    }

    private fun startPolling() {
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            while (true) {
                refreshSession()
                if (_uiState.value.endedMessage != null) {
                    return@launch
                }
                delay(5_000L)
            }
        }
    }

    private suspend fun refreshSession() {
        videoRepository.fetchTvRemoteSession(sessionId)
            .onSuccess { session -> applySession(session) }
            .onFailure { err ->
                handleAuthError(err)
                _uiState.update {
                    it.copy(
                        loading = false,
                        errorMessage = err.message ?: "加载投放控制失败",
                    )
                }
            }
    }

    private fun applySession(session: TvRemoteSessionDto) {
        _uiState.update { state ->
            state.copy(
                loading = false,
                session = session,
                errorMessage = null,
                endedMessage = if (session.status == "ended") "投放已结束" else state.endedMessage,
            )
        }
    }

    private fun handleAuthError(err: Throwable?) {
        if (err is com.chee.videos.core.model.AuthExpiredException) {
            viewModelScope.launch { authRepository.logoutLocal() }
        }
    }
}

@Composable
fun ShortSearchRemoteControlScreen(
    onBack: () -> Unit,
    viewModel: ShortSearchRemoteControlViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current

    LaunchedEffect(uiState.endedMessage) {
        val message = uiState.endedMessage ?: return@LaunchedEffect
        Toast.makeText(context, message as CharSequence, Toast.LENGTH_SHORT).show()
        viewModel.consumeEndedMessage()
        onBack()
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFF0B0E15)),
        contentAlignment = Alignment.Center,
    ) {
        when {
            uiState.loading -> CircularProgressIndicator(color = Color.White)
            uiState.session == null -> Text(
                text = uiState.errorMessage ?: "投放会话不存在",
                color = Color.White,
            )
            else -> {
                val session = uiState.session!!
                val currentItem = session.currentItem ?: session.items.getOrNull(session.currentIndex)
                Surface(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(20.dp),
                    shape = RoundedCornerShape(20.dp),
                    color = Color(0xFF141821),
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(20.dp)
                            .navigationBarsPadding(),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                    ) {
                        Text(
                            text = "电视投放控制",
                            style = MaterialTheme.typography.headlineSmall,
                            color = Color.White,
                            fontWeight = FontWeight.Bold,
                        )
                        Text(
                            text = "设备：${session.deviceName.ifBlank { session.deviceId }}",
                            color = Color(0xFFB9C0CC),
                        )
                        Text(
                            text = currentItem?.title?.ifBlank { "当前视频" } ?: "当前视频",
                            style = MaterialTheme.typography.titleLarge,
                            color = Color.White,
                            fontWeight = FontWeight.SemiBold,
                        )
                        Text(
                            text = "第 ${session.currentIndex + 1} / ${session.items.size.coerceAtLeast(1)} 条",
                            color = Color(0xFF8D95A3),
                        )
                        if (!uiState.errorMessage.isNullOrBlank()) {
                            Text(
                                text = uiState.errorMessage.orEmpty(),
                                color = MaterialTheme.colorScheme.error,
                            )
                        }
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            OutlinedButton(
                                modifier = Modifier.weight(1f),
                                enabled = session.hasPrevious && !uiState.actionLoading,
                                onClick = viewModel::previous,
                            ) {
                                Text("上一个")
                            }
                            Button(
                                modifier = Modifier.weight(1f),
                                enabled = session.hasNext && !uiState.actionLoading,
                                onClick = viewModel::next,
                            ) {
                                Text("下一个")
                            }
                        }
                        OutlinedButton(
                            modifier = Modifier.fillMaxWidth(),
                            enabled = !uiState.actionLoading,
                            onClick = onBack,
                        ) {
                            Text("返回")
                        }
                    }
                }
            }
        }
    }
}
