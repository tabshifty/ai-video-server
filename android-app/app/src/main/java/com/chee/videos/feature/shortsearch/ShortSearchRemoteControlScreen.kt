package com.chee.videos.feature.shortsearch

import android.net.Uri
import android.widget.Toast
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.SkipPrevious
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
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
    val autoplayNextLoading: Boolean = false,
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

    fun previous() {
        runAction { videoRepository.tvRemotePrevious(sessionId) }
    }

    fun next() {
        runAction { videoRepository.tvRemoteNext(sessionId) }
    }

    fun setAutoplayNextEnabled(enabled: Boolean) {
        val session = _uiState.value.session ?: return
        if (_uiState.value.autoplayNextLoading || session.isAutoplayNextEnabled == enabled) {
            return
        }
        viewModelScope.launch {
            _uiState.update { it.copy(autoplayNextLoading = true, errorMessage = null) }
            videoRepository.tvRemoteAutoplayNext(sessionId, enabled)
                .onSuccess { updated ->
                    applySession(updated)
                    _uiState.update { it.copy(autoplayNextLoading = false) }
                }
                .onFailure { err ->
                    handleAuthError(err)
                    _uiState.update {
                        it.copy(
                            autoplayNextLoading = false,
                            errorMessage = err.message ?: "自动播放下一条设置失败",
                        )
                    }
                }
        }
    }

    fun consumeEndedMessage() {
        _uiState.update { it.copy(endedMessage = null) }
    }

    fun refreshNow() {
        viewModelScope.launch { refreshSession() }
    }

    fun setPollingEnabled(enabled: Boolean) {
        if (enabled) {
            startPolling()
        } else {
            stopPolling()
        }
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
        if (pollingJob?.isActive == true) {
            return
        }
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

    private fun stopPolling() {
        pollingJob?.cancel()
        pollingJob = null
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
    val lifecycleOwner = LocalLifecycleOwner.current

    DisposableEffect(lifecycleOwner, viewModel) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_START -> {
                    viewModel.setPollingEnabled(true)
                    viewModel.refreshNow()
                }

                Lifecycle.Event.ON_STOP -> viewModel.setPollingEnabled(false)
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        if (lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.STARTED)) {
            viewModel.setPollingEnabled(true)
            viewModel.refreshNow()
        }
        onDispose {
            viewModel.setPollingEnabled(false)
            lifecycleOwner.lifecycle.removeObserver(observer)
        }
    }

    LaunchedEffect(uiState.endedMessage) {
        val message = uiState.endedMessage ?: return@LaunchedEffect
        Toast.makeText(context, message as CharSequence, Toast.LENGTH_SHORT).show()
        viewModel.consumeEndedMessage()
        onBack()
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    listOf(
                        Color(0xFF05070C),
                        Color(0xFF111824),
                        Color(0xFF070A10),
                    ),
                ),
            ),
        contentAlignment = Alignment.Center,
    ) {
        when {
            uiState.loading -> Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                CircularProgressIndicator(color = Color.White)
                Text("正在同步投放状态", color = Color(0xFFC8D2E2))
            }

            uiState.session == null -> Text(
                text = uiState.errorMessage ?: "投放会话不存在",
                color = Color.White,
                modifier = Modifier.padding(horizontal = 24.dp),
                textAlign = TextAlign.Center,
            )

            else -> {
                val session = uiState.session!!
                val currentItem = session.currentItem ?: session.items.getOrNull(session.currentIndex)
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .statusBarsPadding()
                        .navigationBarsPadding()
                        .padding(horizontal = 20.dp, vertical = 16.dp),
                    verticalArrangement = Arrangement.SpaceBetween,
                ) {
                    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.SpaceBetween,
                        ) {
                            IconButton(onClick = onBack) {
                                Icon(
                                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                                    contentDescription = "返回",
                                    tint = Color.White,
                                )
                            }
                            TvRemoteControlStatusPill(text = if (session.status == "active") "投放中" else "已结束")
                        }
                        Row(verticalAlignment = Alignment.CenterVertically) {
                            Icon(
                                imageVector = Icons.Filled.Tv,
                                contentDescription = null,
                                tint = Color(0xFF7DD3FC),
                                modifier = Modifier.size(20.dp),
                            )
                            Spacer(Modifier.width(8.dp))
                            Text(
                                text = session.deviceName.ifBlank { session.deviceId },
                                color = Color(0xFFC8D2E2),
                                style = MaterialTheme.typography.bodyMedium,
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                            )
                        }
                    }

                    Column(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(10.dp),
                    ) {
                        Text(
                            text = currentItem?.title?.ifBlank { "当前视频" } ?: "当前视频",
                            style = MaterialTheme.typography.headlineSmall,
                            color = Color.White,
                            fontWeight = FontWeight.Bold,
                            textAlign = TextAlign.Center,
                            maxLines = 3,
                            overflow = TextOverflow.Ellipsis,
                        )
                        Text(
                            text = "第 ${session.currentIndex + 1} / ${session.items.size.coerceAtLeast(1)} 条",
                            color = Color(0xFF93A4B8),
                            style = MaterialTheme.typography.titleMedium,
                        )
                    }

                    Column(verticalArrangement = Arrangement.spacedBy(14.dp)) {
                        Surface(
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(8.dp),
                            color = Color(0x1AFFFFFF),
                            border = BorderStroke(1.dp, Color(0x24FFFFFF)),
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = 16.dp, vertical = 14.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                Column(
                                    modifier = Modifier.weight(1f),
                                    verticalArrangement = Arrangement.spacedBy(4.dp),
                                ) {
                                    Text(
                                        text = "自动播放下一条",
                                        color = Color.White,
                                        style = MaterialTheme.typography.titleMedium,
                                        fontWeight = FontWeight.SemiBold,
                                    )
                                    Text(
                                        text = "当前条结束后自动切到下一条",
                                        color = Color(0xFFAAB6C5),
                                        style = MaterialTheme.typography.bodySmall,
                                    )
                                }
                                if (uiState.autoplayNextLoading) {
                                    CircularProgressIndicator(
                                        color = Color(0xFF7DD3FC),
                                        strokeWidth = 2.dp,
                                        modifier = Modifier.size(22.dp),
                                    )
                                }
                                Switch(
                                    checked = session.isAutoplayNextEnabled,
                                    onCheckedChange = viewModel::setAutoplayNextEnabled,
                                    enabled = !uiState.autoplayNextLoading,
                                    colors = SwitchDefaults.colors(
                                        checkedThumbColor = Color.White,
                                        checkedTrackColor = Color(0xFF0EA5E9),
                                        uncheckedThumbColor = Color(0xFFE2E8F0),
                                        uncheckedTrackColor = Color(0xFF475569),
                                    ),
                                )
                            }
                        }
                        if (!uiState.errorMessage.isNullOrBlank()) {
                            Text(
                                text = uiState.errorMessage.orEmpty(),
                                color = MaterialTheme.colorScheme.error,
                                style = MaterialTheme.typography.bodyMedium,
                            )
                        }
                        Row(
                            modifier = Modifier.fillMaxWidth(),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            TvRemoteControlActionButton(
                                modifier = Modifier.weight(1f),
                                icon = Icons.Filled.SkipPrevious,
                                label = "上一个",
                                enabled = session.hasPrevious && !uiState.actionLoading,
                                onClick = viewModel::previous,
                            )
                            TvRemoteControlActionButton(
                                modifier = Modifier.weight(1f),
                                icon = Icons.Filled.SkipNext,
                                label = "下一个",
                                enabled = session.hasNext && !uiState.actionLoading,
                                onClick = viewModel::next,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TvRemoteControlStatusPill(text: String) {
    Surface(
        shape = RoundedCornerShape(8.dp),
        color = Color(0x1F7DD3FC),
        border = BorderStroke(1.dp, Color(0x557DD3FC)),
    ) {
        Text(
            text = text,
            color = Color(0xFFBAE6FD),
            style = MaterialTheme.typography.labelMedium,
            modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
        )
    }
}

@Composable
private fun TvRemoteControlActionButton(
    icon: ImageVector,
    label: String,
    enabled: Boolean,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val foreground = if (enabled) Color.White else Color(0xFF7B8798)
    Surface(
        modifier = modifier
            .height(76.dp)
            .then(if (enabled) Modifier.clickable(onClick = onClick) else Modifier),
        shape = RoundedCornerShape(8.dp),
        color = if (enabled) Color(0xFF182232) else Color(0xFF101722),
        border = BorderStroke(1.dp, if (enabled) Color(0x337DD3FC) else Color(0x1AFFFFFF)),
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 14.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center,
        ) {
            Icon(
                imageVector = icon,
                contentDescription = label,
                tint = foreground,
                modifier = Modifier.size(26.dp),
            )
            Spacer(Modifier.width(8.dp))
            Text(
                text = label,
                color = foreground,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
            )
        }
    }
}
