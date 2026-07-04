package com.chee.videos.feature.tv

import android.graphics.Color as AndroidColor
import android.net.Uri
import android.view.KeyEvent as AndroidKeyEvent
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.widthIn
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
import androidx.compose.runtime.mutableLongStateOf
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
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.lifecycle.viewModelScope
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
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
import com.chee.videos.core.ui.TvMotionTokens
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
    val seekStepSeconds: Int = TvPlaybackSeekStepSetting.defaultSeconds,
    val actionLoading: Boolean = false,
    val errorMessage: String? = null,
)

@HiltViewModel
class TvRemotePlaybackViewModel @Inject constructor(
    private val repository: TvRepository,
    savedStateHandle: SavedStateHandle,
) : ViewModel() {
    private val sessionId = savedStateHandle.get<String>(TvRemotePlaybackSessionIdArg).orEmpty()
    val remoteSessionId: String get() = sessionId
    private val _uiState = MutableStateFlow(TvRemotePlaybackUiState())
    val uiState: StateFlow<TvRemotePlaybackUiState> = _uiState.asStateFlow()
    private var pollingJob: Job? = null

    init {
        viewModelScope.launch {
            _uiState.update {
                it.copy(seekStepSeconds = TvPlaybackSeekStepSetting.normalize(repository.readTvSeekStepSeconds()))
            }
        }
    }

    fun previous() {
        runAction { repository.tvRemotePrevious(sessionId) }
    }

    fun next() {
        runAction { repository.tvRemoteNext(sessionId) }
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

    private fun startPolling() {
        if (pollingJob?.isActive == true) {
            return
        }
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

    private fun stopPolling() {
        pollingJob?.cancel()
        pollingJob = null
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
    onSessionEnded: (String) -> Unit = {},
    onBack: () -> Unit,
    viewModel: TvRemotePlaybackViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
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
    var showChrome by rememberSaveable { mutableStateOf(true) }
    var chromeHideJob by remember { mutableStateOf<Job?>(null) }
    var showCenterIndicator by remember { mutableStateOf(false) }
    var centerIndicatorIsPause by remember { mutableStateOf(false) }
    var centerIndicatorHideJob by remember { mutableStateOf<Job?>(null) }
    var showSeekOverlay by remember { mutableStateOf(false) }
    var seekOverlayPositionMs by remember { mutableLongStateOf(0L) }
    var seekOverlayDurationMs by remember { mutableLongStateOf(0L) }
    var seekOverlayHideJob by remember { mutableStateOf<Job?>(null) }
    var playbackErrorMessage by remember { mutableStateOf<String?>(null) }
    val session = uiState.session
    val currentItem = session?.currentItem ?: session?.items?.getOrNull(session.currentIndex)
    val currentVideoId = session?.currentVideoId ?: currentItem?.videoId.orEmpty()
    val visibleErrorMessage = uiState.errorMessage ?: playbackErrorMessage
    val keepChromeVisible = !visibleErrorMessage.isNullOrBlank()

    fun scheduleChromeAutoHide() {
        chromeHideJob?.cancel()
        if (keepChromeVisible) {
            showChrome = true
            return
        }
        chromeHideJob = coroutineScope.launch {
            delay(TvRemoteChromeAutoHideDurationMillis)
            if (uiState.errorMessage.isNullOrBlank() && playbackErrorMessage.isNullOrBlank()) {
                showChrome = false
            }
        }
    }

    fun showChromeTemporarily() {
        showChrome = true
        scheduleChromeAutoHide()
    }

    fun togglePlaybackWithFeedback() {
        isPausedByUser = if (sharedPlayer.isPlaying) {
            sharedPlayer.pause()
            true
        } else {
            sharedPlayer.play()
            false
        }
        centerIndicatorIsPause = isPausedByUser
        showCenterIndicator = true
        centerIndicatorHideJob?.cancel()
        centerIndicatorHideJob = coroutineScope.launch {
            delay(TvRemoteCenterIndicatorDurationMillis)
            showCenterIndicator = false
        }
        showChromeTemporarily()
    }

    fun applySeek(forward: Boolean, repeatCount: Int) {
        val targetMs = calculateTvRemoteSeekTarget(
            currentPositionMs = sharedPlayer.currentPosition.coerceAtLeast(0L),
            durationMs = sharedPlayer.duration,
            stepSeconds = uiState.seekStepSeconds,
            repeatCount = repeatCount,
            forward = forward,
        ) ?: return
        val durationMs = sharedPlayer.duration.takeIf { it > 0L } ?: return
        sharedPlayer.seekTo(targetMs)
        seekOverlayPositionMs = targetMs
        seekOverlayDurationMs = durationMs
        showSeekOverlay = true
        seekOverlayHideJob?.cancel()
        seekOverlayHideJob = coroutineScope.launch {
            delay(TvRemoteSeekOverlayDurationMillis)
            showSeekOverlay = false
        }
        showChromeTemporarily()
    }

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

    BackHandler {
        onSessionEnded(viewModel.remoteSessionId)
        onBack()
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    DisposableEffect(sharedPlayer) {
        val listener = object : Player.Listener {
            override fun onRenderedFirstFrame() {
                renderedVideoId = sharedPlayer.currentMediaItem?.mediaId
                playbackErrorMessage = null
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                isPlayerActuallyPlaying = isPlaying
            }

            override fun onPlayerError(error: PlaybackException) {
                playbackErrorMessage = error.message?.trim().takeUnless { it.isNullOrBlank() } ?: "播放失败"
                showChrome = true
                chromeHideJob?.cancel()
            }
        }
        sharedPlayer.addListener(listener)
        onDispose {
            chromeHideJob?.cancel()
            centerIndicatorHideJob?.cancel()
            seekOverlayHideJob?.cancel()
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
        playbackErrorMessage = null
        showSeekOverlay = false
        showCenterIndicator = false
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

    LaunchedEffect(currentVideoId) {
        if (currentVideoId.isNotBlank()) {
            showChromeTemporarily()
        }
    }

    LaunchedEffect(keepChromeVisible) {
        if (keepChromeVisible) {
            showChrome = true
            chromeHideJob?.cancel()
        } else if (showChrome) {
            scheduleChromeAutoHide()
        }
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
                        showChromeTemporarily()
                        if (session?.hasPrevious == true && !uiState.actionLoading) {
                            viewModel.previous()
                        }
                        true
                    }

                    AndroidKeyEvent.KEYCODE_DPAD_DOWN -> {
                        showChromeTemporarily()
                        if (session?.hasNext == true && !uiState.actionLoading) {
                            viewModel.next()
                        }
                        true
                    }

                    AndroidKeyEvent.KEYCODE_DPAD_LEFT -> {
                        applySeek(forward = false, repeatCount = event.nativeKeyEvent.repeatCount)
                        true
                    }

                    AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> {
                        applySeek(forward = true, repeatCount = event.nativeKeyEvent.repeatCount)
                        true
                    }

                    AndroidKeyEvent.KEYCODE_DPAD_CENTER,
                    AndroidKeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
                    AndroidKeyEvent.KEYCODE_SPACE,
                    -> {
                        togglePlaybackWithFeedback()
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
                    onSessionEnded(session.sessionId)
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
                AnimatedVisibility(
                    visible = showChrome || keepChromeVisible,
                    enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                    exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                    modifier = Modifier.align(Alignment.TopStart),
                ) {
                    TvRemotePlaybackInfoOverlay(
                        title = currentItem?.title?.ifBlank { "短视频投放" } ?: "短视频投放",
                        subtitle = "${session.deviceName.ifBlank { session.deviceId }} · 第 ${session.currentIndex + 1} / ${session.items.size.coerceAtLeast(1)} 条",
                        errorMessage = visibleErrorMessage,
                    )
                }
                AnimatedVisibility(
                    visible = showCenterIndicator,
                    enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                    exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                    modifier = Modifier.align(Alignment.Center),
                ) {
                    Surface(
                        color = AppChrome.Surface.copy(alpha = 0.82f),
                        shape = CircleShape,
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 18.dp, vertical = 14.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            Icon(
                                imageVector = if (centerIndicatorIsPause) Icons.Filled.Pause else Icons.Filled.PlayArrow,
                                contentDescription = null,
                                tint = Color.White,
                            )
                            Text(
                                text = if (centerIndicatorIsPause) "已暂停" else "继续播放",
                                color = Color.White,
                                style = MaterialTheme.typography.bodyMedium,
                            )
                        }
                    }
                }
                AnimatedVisibility(
                    visible = showSeekOverlay && seekOverlayDurationMs > 0L,
                    enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                    exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .navigationBarsPadding()
                        .padding(bottom = 24.dp),
                ) {
                    TvRemotePlaybackBottomProgressBar(
                        positionMs = seekOverlayPositionMs,
                        durationMs = seekOverlayDurationMs,
                    )
                }
                AnimatedVisibility(
                    visible = showChrome || keepChromeVisible,
                    enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                    exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
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
                            onClick = ::togglePlaybackWithFeedback,
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
private fun TvRemotePlaybackInfoOverlay(
    title: String,
    subtitle: String,
    errorMessage: String?,
) {
    Column(
        modifier = Modifier
            .statusBarsPadding()
            .padding(start = 24.dp, top = 24.dp, end = 24.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Text(
            text = title,
            color = Color.White,
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.widthIn(max = 720.dp),
        )
        Text(
            text = subtitle,
            color = Color(0xFFD2D7DF),
            style = MaterialTheme.typography.bodyMedium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.widthIn(max = 720.dp),
        )
        if (!errorMessage.isNullOrBlank()) {
            Text(
                text = errorMessage,
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.widthIn(max = 720.dp),
            )
        }
    }
}

@Composable
private fun TvRemotePlaybackBottomProgressBar(
    positionMs: Long,
    durationMs: Long,
) {
    val progress = if (durationMs <= 0L) {
        0f
    } else {
        (positionMs.toFloat() / durationMs.toFloat()).coerceIn(0f, 1f)
    }

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Surface(
            color = AppChrome.Surface.copy(alpha = 0.92f),
            shape = AppChrome.SurfaceShape,
        ) {
            Text(
                text = "${formatTvRemotePlaybackTime(positionMs)} / ${formatTvRemotePlaybackTime(durationMs)}",
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.labelMedium,
                modifier = Modifier.padding(horizontal = 12.dp, vertical = 7.dp),
            )
        }
        Box(
            modifier = Modifier
                .widthIn(min = 320.dp, max = 560.dp)
                .fillMaxWidth()
                .height(6.dp)
                .clip(CircleShape)
                .background(AppChrome.TextPrimary.copy(alpha = 0.18f)),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(progress)
                    .fillMaxSize()
                    .clip(CircleShape)
                    .background(AppChrome.AccentStrong, CircleShape),
            )
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

internal fun calculateTvRemoteSeekTarget(
    currentPositionMs: Long,
    durationMs: Long,
    stepSeconds: Int,
    repeatCount: Int,
    forward: Boolean,
): Long? {
    if (durationMs <= 0L) {
        return null
    }
    val current = currentPositionMs.coerceAtLeast(0L)
    val stepMs = TvPlaybackSeekStepSetting.normalize(stepSeconds) * 1_000L
    val deltaMs = stepMs * if (repeatCount > 0) 3L else 1L
    return if (forward) {
        (current + deltaMs).coerceAtMost(durationMs)
    } else {
        (current - deltaMs).coerceAtLeast(0L)
    }
}

private fun formatTvRemotePlaybackTime(ms: Long): String {
    val totalSeconds = ms.coerceAtLeast(0L) / 1_000L
    val hours = totalSeconds / 3_600L
    val minutes = (totalSeconds % 3_600L) / 60L
    val seconds = totalSeconds % 60L
    return buildString {
        append(hours.toString().padStart(2, '0'))
        append(':')
        append(minutes.toString().padStart(2, '0'))
        append(':')
        append(seconds.toString().padStart(2, '0'))
    }
}

private const val TvRemoteChromeAutoHideDurationMillis = 3_000L
private const val TvRemoteCenterIndicatorDurationMillis = 700L
private const val TvRemoteSeekOverlayDurationMillis = 1_200L
