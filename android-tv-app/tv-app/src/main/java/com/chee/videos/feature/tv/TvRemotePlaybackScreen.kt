package com.chee.videos.feature.tv

import android.graphics.Color as AndroidColor
import android.net.Uri
import android.view.KeyEvent as AndroidKeyEvent
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.SkipPrevious
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import com.chee.videos.core.model.TvRemoteSessionDto
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.util.UrlBuilder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

const val TvRemotePlaybackSessionIdArg = "sessionId"
const val TvRemotePlaybackRoutePattern = "tv/remote-shorts/{$TvRemotePlaybackSessionIdArg}"

fun buildTvRemotePlaybackRoute(sessionId: String): String =
    "tv/remote-shorts/${Uri.encode(sessionId)}"

data class TvRemotePlaybackUiState(
    val loading: Boolean = true,
    val session: TvRemoteSessionDto? = null,
    val currentSourceUrl: String = "",
    val currentPosterUrl: String = "",
    val actionLoading: Boolean = false,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvRemotePlaybackViewModel @Inject constructor(
    private val repository: TvRepository,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val sessionId = savedStateHandle.get<String>(TvRemotePlaybackSessionIdArg).orEmpty()
    private val _uiState = MutableStateFlow(TvRemotePlaybackUiState())
    val uiState: StateFlow<TvRemotePlaybackUiState> = _uiState.asStateFlow()
    private var pollingJob: Job? = null

    init {
        startPolling()
    }

    fun previous() {
        runAction { repository.tvRemotePrevious(sessionId) }
    }

    fun next() {
        runAction { repository.tvRemoteNext(sessionId) }
    }

    suspend fun endSession() {
        repository.endTvRemoteSession(sessionId)
    }

    private fun startPolling() {
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            while (true) {
                refreshSession()
                if (_uiState.value.session?.status == "ended") {
                    return@launch
                }
                delay(5_000L)
            }
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
                _uiState.update {
                    it.copy(
                        actionLoading = false,
                        errorMessage = err.message ?: "切换投放视频失败",
                    )
                }
            }
        }
    }

    private suspend fun refreshSession() {
        repository.fetchTvRemoteSession(sessionId)
            .onSuccess { session -> applySession(session) }
            .onFailure { err ->
                _uiState.update {
                    it.copy(
                        loading = false,
                        errorMessage = err.message ?: "加载投放会话失败",
                    )
                }
            }
    }

    private fun applySession(session: TvRemoteSessionDto) {
        viewModelScope.launch {
            val currentVideoId = session.currentVideoId ?: session.currentItem?.videoId
            val sourceUrl = if (session.status == "active" && !currentVideoId.isNullOrBlank()) {
                repository.buildSourceUrl(currentVideoId)
            } else {
                ""
            }
            val baseUrl = repository.readActiveBaseUrl().orEmpty()
            val posterUrl = session.currentItem?.thumbnailPath
                ?.takeIf { it.isNotBlank() }
                ?.let { raw ->
                    if (raw.startsWith("http://") || raw.startsWith("https://")) raw
                    else UrlBuilder.normalizeBaseUrl(baseUrl) + raw
                }
                .orEmpty()
            _uiState.update {
                it.copy(
                    loading = false,
                    session = session,
                    currentSourceUrl = sourceUrl,
                    currentPosterUrl = posterUrl,
                    errorMessage = if (session.status == "ended") null else it.errorMessage,
                )
            }
        }
    }
}

@Composable
fun TvRemotePlaybackScreen(
    accessToken: String,
    onBack: () -> Unit,
    viewModel: TvRemotePlaybackViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val coroutineScope = rememberCoroutineScope()
    val dataSourceFactory = remember(accessToken) {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
            if (accessToken.isNotBlank()) {
                setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
            }
        }
    }
    val sharedPlayer = remember(accessToken) {
        ExoPlayer.Builder(context).build().apply {
            repeatMode = Player.REPEAT_MODE_OFF
        }
    }

    var renderedVideoId by remember { mutableStateOf<String?>(null) }
    var isPlayerActuallyPlaying by remember { mutableStateOf(false) }
    var isPausedByUser by rememberSaveable { mutableStateOf(false) }
    val session = uiState.session
    val currentItem = session?.currentItem ?: session?.items?.getOrNull(session.currentIndex)
    val currentVideoId = session?.currentVideoId ?: currentItem?.videoId.orEmpty()

    BackHandler {
        coroutineScope.launch {
            viewModel.endSession()
            onBack()
        }
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    DisposableEffect(sharedPlayer) {
        val listener = object : Player.Listener {
            override fun onRenderedFirstFrame() {
                renderedVideoId = sharedPlayer.currentMediaItem?.mediaId
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                isPlayerActuallyPlaying = isPlaying
            }
        }
        sharedPlayer.addListener(listener)
        onDispose {
            sharedPlayer.removeListener(listener)
            sharedPlayer.release()
        }
    }

    LaunchedEffect(currentVideoId, uiState.currentSourceUrl) {
        if (currentVideoId.isBlank() || uiState.currentSourceUrl.isBlank()) {
            sharedPlayer.stop()
            sharedPlayer.clearMediaItems()
            renderedVideoId = null
            return@LaunchedEffect
        }
        renderedVideoId = null
        sharedPlayer.stop()
        sharedPlayer.clearMediaItems()
        val mediaItem = MediaItem.Builder()
            .setUri(uiState.currentSourceUrl)
            .setMediaId(currentVideoId)
            .build()
        val mediaSource = ProgressiveMediaSource.Factory(dataSourceFactory).createMediaSource(mediaItem)
        sharedPlayer.setMediaSource(mediaSource, true)
        sharedPlayer.prepare()
        sharedPlayer.playWhenReady = !isPausedByUser
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .onPreviewKeyEvent { event ->
                if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                when (event.nativeKeyEvent.keyCode) {
                    AndroidKeyEvent.KEYCODE_DPAD_UP -> {
                        if (session?.hasPrevious == true && !uiState.actionLoading) {
                            viewModel.previous()
                        }
                        true
                    }

                    AndroidKeyEvent.KEYCODE_DPAD_DOWN -> {
                        if (session?.hasNext == true && !uiState.actionLoading) {
                            viewModel.next()
                        }
                        true
                    }

                    AndroidKeyEvent.KEYCODE_DPAD_CENTER,
                    AndroidKeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
                    AndroidKeyEvent.KEYCODE_SPACE,
                    -> {
                        isPausedByUser = if (sharedPlayer.isPlaying) {
                            sharedPlayer.pause()
                            true
                        } else {
                            sharedPlayer.play()
                            false
                        }
                        true
                    }

                    else -> false
                }
            },
    ) {
        when {
            uiState.loading -> {
                Text("正在准备投放播放", modifier = Modifier.align(Alignment.Center), color = Color.White)
            }

            session == null -> {
                Text(uiState.errorMessage ?: "投放会话不存在", modifier = Modifier.align(Alignment.Center), color = Color.White)
            }

            session.status == "ended" -> {
                LaunchedEffect(session.sessionId, session.status) {
                    onBack()
                }
            }

            else -> {
                if (uiState.currentSourceUrl.isNotBlank()) {
                    AndroidView(
                        factory = {
                            PlayerView(it).apply {
                                player = sharedPlayer
                                useController = false
                                setShutterBackgroundColor(AndroidColor.BLACK)
                                resizeMode = AspectRatioFrameLayout.RESIZE_MODE_FIT
                            }
                        },
                        update = { view -> view.player = sharedPlayer },
                        modifier = Modifier.fillMaxSize(),
                    )
                }
                if (renderedVideoId != currentVideoId && uiState.currentPosterUrl.isNotBlank()) {
                    AsyncImage(
                        model = uiState.currentPosterUrl,
                        contentDescription = "${currentItem?.title.orEmpty()} 封面",
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                }
                Box(modifier = Modifier.fillMaxSize().background(Color(0x30000000)))
                Column(
                    modifier = Modifier
                        .align(Alignment.BottomStart)
                        .navigationBarsPadding()
                        .padding(28.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Text(
                        text = currentItem?.title?.ifBlank { "短视频投放" } ?: "短视频投放",
                        color = Color.White,
                        style = MaterialTheme.typography.headlineMedium,
                        fontWeight = FontWeight.Bold,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Text(
                        text = "${session.deviceName.ifBlank { session.deviceId }} · 第 ${session.currentIndex + 1} / ${session.items.size.coerceAtLeast(1)} 条",
                        color = Color(0xFFD2D7DF),
                        style = MaterialTheme.typography.bodyMedium,
                    )
                    if (!uiState.errorMessage.isNullOrBlank()) {
                        Text(
                            text = uiState.errorMessage.orEmpty(),
                            color = MaterialTheme.colorScheme.error,
                            style = MaterialTheme.typography.bodySmall,
                        )
                    }
                }
                AnimatedVisibility(
                    visible = true,
                    enter = fadeIn(),
                    exit = fadeOut(),
                    modifier = Modifier.align(Alignment.CenterEnd),
                ) {
                    Column(
                        modifier = Modifier.padding(end = 24.dp),
                        verticalArrangement = Arrangement.spacedBy(14.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                    ) {
                        TvRemoteActionBubble(
                            icon = Icons.Filled.SkipPrevious,
                            label = "上一个",
                            enabled = session.hasPrevious && !uiState.actionLoading,
                            onClick = viewModel::previous,
                        )
                        TvRemoteActionBubble(
                            icon = if (sharedPlayer.isPlaying) Icons.Filled.Pause else Icons.Filled.PlayArrow,
                            label = if (sharedPlayer.isPlaying) "暂停" else "播放",
                            enabled = true,
                            onClick = {
                                if (sharedPlayer.isPlaying) {
                                    sharedPlayer.pause()
                                    isPausedByUser = true
                                } else {
                                    sharedPlayer.play()
                                    isPausedByUser = false
                                }
                            },
                        )
                        TvRemoteActionBubble(
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

@Composable
private fun TvRemoteActionBubble(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    enabled: Boolean,
    onClick: () -> Unit,
) {
    val alpha = if (enabled) 1f else 0.45f
    Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Box(
            modifier = Modifier
                .clip(CircleShape)
                .background(Color(0xD9151A22))
                .then(if (enabled) Modifier.clickable(onClick = onClick) else Modifier)
                .padding(18.dp),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = label, tint = Color.White.copy(alpha = alpha))
        }
        Text(label, color = Color.White.copy(alpha = alpha), style = MaterialTheme.typography.bodySmall)
    }
}
