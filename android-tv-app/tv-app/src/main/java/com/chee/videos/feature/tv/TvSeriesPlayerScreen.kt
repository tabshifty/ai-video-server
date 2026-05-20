package com.chee.videos.feature.tv

import android.app.Activity
import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.TvLayoutSpec
import com.chee.videos.core.ui.buildLongFormMediaItem
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolveSelectedSubtitleTrack
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.feature.detail.LongFormPlaybackSession
import kotlinx.coroutines.delay

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TvSeriesPlayerScreen(
    accessToken: String,
    onBack: () -> Unit,
    viewModel: TvSeriesPlayerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val activity = context as? Activity
    val lifecycleOwner = LocalLifecycleOwner.current
    var backPromptAtMillis by remember { mutableStateOf<Long?>(null) }
    var showBackConfirmPrompt by remember { mutableStateOf(false) }

    fun handlePlaybackBack() {
        val now = SystemClock.uptimeMillis()
        when (resolveTvPlayerBackAction(backPromptAtMillis, now)) {
            TvPlayerBackAction.ShowPrompt -> {
                backPromptAtMillis = now
                showBackConfirmPrompt = true
            }

            TvPlayerBackAction.Exit -> {
                backPromptAtMillis = null
                showBackConfirmPrompt = false
                onBack()
            }
        }
    }

    BackHandler(onBack = ::handlePlaybackBack)

    LaunchedEffect(showBackConfirmPrompt, backPromptAtMillis) {
        val promptAt = backPromptAtMillis
        if (showBackConfirmPrompt && promptAt != null) {
            delay(TvPlayerBackConfirmWindowMillis)
            if (backPromptAtMillis == promptAt) {
                showBackConfirmPrompt = false
            }
        }
    }

    if (uiState.loading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
            contentAlignment = Alignment.Center,
        ) {
            CircularProgressIndicator(color = AppChrome.AccentStrong)
            if (showBackConfirmPrompt) {
                TvPlayerBackConfirmPrompt(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 48.dp),
                )
            }
        }
        return
    }

    val series = uiState.series
    if (series == null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
            contentAlignment = Alignment.Center,
        ) {
            Text(uiState.errorMessage ?: "播放器数据不存在", color = AppChrome.TextSecondary)
            if (showBackConfirmPrompt) {
                TvPlayerBackConfirmPrompt(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 48.dp),
                )
            }
        }
        return
    }

    val currentEpisode = selectedEpisode(uiState)
    val canPlay = uiState.canPlayCurrentEpisode && uiState.currentSourceUrl.isNotBlank()
    val dataSourceFactory = remember(accessToken) {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
            if (accessToken.isNotBlank()) {
                setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
            }
        }
    }
    val exoPlayer = remember(accessToken) { ExoPlayer.Builder(context).build() }
    val latestCurrentVideoId by rememberUpdatedState(uiState.currentVideoId)
    var hasStartedPlayback by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var isPausedByUser by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var preparedUrl by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var preparedSubtitleTrackId by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var selectedSubtitleTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var selectedAudioTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var isPlayerActuallyPlaying by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var playerErrorMessage by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var lastHistoryVideoId by remember { mutableStateOf("") }
    var resumedFromHistoryVideoId by remember { mutableStateOf("") }

    val playbackSession = remember(hasStartedPlayback, isPausedByUser) {
        LongFormPlaybackSession(
            hasStartedPlayback = hasStartedPlayback,
            isPausedByUser = isPausedByUser,
        )
    }

    fun updatePlaybackSession(nextSession: LongFormPlaybackSession) {
        hasStartedPlayback = nextSession.hasStartedPlayback
        isPausedByUser = nextSession.isPausedByUser
    }

    DisposableEffect(activity) {
        if (activity == null) {
            onDispose { }
        } else {
            val window = activity.window
            val controller = WindowCompat.getInsetsController(window, window.decorView)
            controller.hide(WindowInsetsCompat.Type.systemBars())
            controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE

            onDispose {
                controller.show(WindowInsetsCompat.Type.systemBars())
            }
        }
    }

    LaunchedEffect(uiState.currentVideoId) {
        if (lastHistoryVideoId.isNotBlank() && lastHistoryVideoId != uiState.currentVideoId) {
            reportTvSeriesHistory(viewModel, lastHistoryVideoId, exoPlayer)
        }
        lastHistoryVideoId = uiState.currentVideoId
    }

    LaunchedEffect(uiState.currentSourceUrl, canPlay, selectedSubtitleTrackId, currentEpisode?.subtitleTracks) {
        val effectiveSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId)
        val updateDecision = resolveLongFormPlayerUpdate(
            preparedUrl = preparedUrl,
            nextUrl = uiState.currentSourceUrl,
            preparedSubtitleTrackId = preparedSubtitleTrackId,
            nextSubtitleTrackId = effectiveSubtitleTrackId,
        )
        if (!canPlay) {
            exoPlayer.pause()
            if (updateDecision.shouldClear) {
                exoPlayer.clearMediaItems()
            }
            preparedUrl = null
            preparedSubtitleTrackId = null
            return@LaunchedEffect
        }
        if (updateDecision.shouldReplaceSource) {
            playerErrorMessage = null
            val initialResumePositionMs = if (!updateDecision.preservePosition && resumedFromHistoryVideoId != uiState.currentVideoId) {
                currentEpisode?.watchSeconds?.coerceAtLeast(0)?.times(1000L) ?: 0L
            } else {
                0L
            }
            val restorePositionMs = if (updateDecision.preservePosition) {
                exoPlayer.currentPosition.coerceAtLeast(0L)
            } else {
                initialResumePositionMs
            }
            val mediaItem = buildLongFormMediaItem(
                sourceUrl = uiState.currentSourceUrl,
                mediaId = uiState.currentVideoId,
                title = currentEpisode?.title ?: series.title,
                baseUrl = uiState.baseUrl,
                selectedSubtitleTrack = resolveSelectedSubtitleTrack(currentEpisode?.subtitleTracks.orEmpty(), effectiveSubtitleTrackId),
            )
            val mediaSource = DefaultMediaSourceFactory(dataSourceFactory).createMediaSource(mediaItem)
            exoPlayer.setMediaSource(mediaSource, true)
            exoPlayer.prepare()
            if (restorePositionMs > 0L) {
                exoPlayer.seekTo(restorePositionMs)
            }
            if (!updateDecision.preservePosition && uiState.currentVideoId.isNotBlank()) {
                resumedFromHistoryVideoId = uiState.currentVideoId
            }
            preparedUrl = uiState.currentSourceUrl
            preparedSubtitleTrackId = effectiveSubtitleTrackId
            hasStartedPlayback = true
            isPausedByUser = false
        }
    }

    LaunchedEffect(uiState.currentVideoId, uiState.selectedSubtitleTrackId, currentEpisode?.subtitleTracks, hasStartedPlayback) {
        selectedSubtitleTrackId = resolveTvSubtitleSelectionOnTrackLoad(
            currentSelection = selectedSubtitleTrackId,
            storedSelection = uiState.selectedSubtitleTrackId,
            tracks = currentEpisode?.subtitleTracks.orEmpty(),
            hasStartedPlayback = hasStartedPlayback,
        )
    }

    LaunchedEffect(uiState.currentVideoId, uiState.selectedAudioTrackId) {
        selectedAudioTrackId = uiState.selectedAudioTrackId
    }

    LaunchedEffect(playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, canPlay) {
        if (!playbackSession.hasStartedPlayback || !canPlay) {
            exoPlayer.pause()
            return@LaunchedEffect
        }
        if (playbackSession.isPausedByUser) {
            exoPlayer.pause()
            return@LaunchedEffect
        }
        exoPlayer.playWhenReady = true
        exoPlayer.play()
    }

    LaunchedEffect(
        uiState.currentVideoId,
        canPlay,
        playbackSession.hasStartedPlayback,
        playbackSession.isPausedByUser,
    ) {
        if (!shouldStartPeriodicHistoryReport(
                videoId = uiState.currentVideoId,
                canPlay = canPlay,
                hasStartedPlayback = playbackSession.hasStartedPlayback,
                isPausedByUser = playbackSession.isPausedByUser,
            )
        ) {
            return@LaunchedEffect
        }
        while (true) {
            delay(TvPeriodicHistoryReportIntervalMillis)
            reportTvSeriesHistory(viewModel, uiState.currentVideoId, exoPlayer)
        }
    }

    DisposableEffect(exoPlayer) {
        val listener = object : androidx.media3.common.Player.Listener {
            override fun onIsPlayingChanged(isPlaying: Boolean) {
                isPlayerActuallyPlaying = isPlaying
            }

            override fun onRenderedFirstFrame() {
                playerErrorMessage = null
            }

            override fun onPlayerError(error: androidx.media3.common.PlaybackException) {
                playerErrorMessage = friendlyLongFormPlaybackErrorMessage(error)
                isPlayerActuallyPlaying = false
            }
        }
        exoPlayer.addListener(listener)
        onDispose {
            reportTvSeriesHistory(viewModel, latestCurrentVideoId, exoPlayer)
            exoPlayer.removeListener(listener)
            exoPlayer.release()
        }
    }

    DisposableEffect(lifecycleOwner, exoPlayer, playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, canPlay) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> {
                    reportTvSeriesHistory(viewModel, latestCurrentVideoId, exoPlayer)
                    exoPlayer.pause()
                }

                Lifecycle.Event.ON_RESUME -> {
                    if (playbackSession.shouldResumeOnLifecycle() && canPlay) {
                        exoPlayer.playWhenReady = true
                        exoPlayer.play()
                    }
                }

                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
        }
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (canPlay) {
            Box(modifier = Modifier.fillMaxSize()) {
                LongFormVideoPlayer(
                    title = currentEpisode?.title ?: series.title,
                    player = exoPlayer,
                    isFullscreen = true,
                    onBack = onBack,
                    onTogglePlayPause = {
                        updatePlaybackSession(playbackSession.togglePlayPause(canPlay = canPlay))
                    },
                    onToggleFullscreen = {},
                    modifier = Modifier.fillMaxSize(),
                    subtitleTracks = currentEpisode?.subtitleTracks.orEmpty(),
                    selectedSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId),
                    onSelectSubtitleTrack = {
                        selectedSubtitleTrackId = it ?: ""
                        viewModel.selectSubtitleTrack(it)
                    },
                    selectedAudioTrackId = selectedAudioTrackId,
                    onSelectAudioTrack = {
                        selectedAudioTrackId = it ?: ""
                        viewModel.selectAudioTrack(it)
                    },
                    showStatusBarPadding = false,
                    tvMode = true,
                    tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                    onOpenEpisodeSelector = { viewModel.setSelectorVisible(true) },
                    onNextEpisode = viewModel::nextEpisode,
                    onExitPlayback = onBack,
                )
                TvPlayerErrorBanner(
                    message = playerErrorMessage,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(12.dp),
                )
                if (showBackConfirmPrompt) {
                    TvPlayerBackConfirmPrompt(
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(bottom = 72.dp),
                    )
                }
            }
        } else {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                EmptyTvPlayerState()
                if (showBackConfirmPrompt) {
                    TvPlayerBackConfirmPrompt(
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(bottom = 48.dp),
                    )
                }
            }
        }
    }

    if (uiState.selectorVisible) {
        ModalBottomSheet(
            onDismissRequest = { viewModel.setSelectorVisible(false) },
            containerColor = AppChrome.Surface,
            contentColor = AppChrome.TextPrimary,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .navigationBarsPadding(),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Text(
                    text = "选集播放",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    modifier = Modifier.padding(horizontal = 16.dp),
                )
                LazyColumn(
                    contentPadding = PaddingValues(
                        start = 16.dp,
                        end = 16.dp,
                        bottom = TvLayoutSpec.scrollBottomSafePaddingDp.dp,
                    ),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    items(selectedSeason(uiState)?.episodes.orEmpty(), key = { episode -> episode.id }) { episode ->
                        val selected = episode.number == uiState.selectedEpisodeNumber
                        Surface(
                            color = if (selected) AppChrome.AccentSoft else AppChrome.SurfaceElevated,
                            shape = RoundedCornerShape(12.dp),
                            modifier = Modifier.fillMaxWidth(),
                        ) {
                            RowItem(
                                title = "E${episode.number} ${episode.title}",
                                subtitle = if (episode.playable) episode.durationLabel else "待绑定 / 未就绪",
                                selected = selected,
                                onClick = {
                                    viewModel.selectEpisode(episode.number)
                                    viewModel.setSelectorVisible(false)
                                },
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun RowItem(
    title: String,
    subtitle: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    Surface(
        color = Color.Transparent,
        modifier = Modifier.fillMaxWidth(),
        onClick = onClick,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text(
                text = title,
                color = if (selected) AppChrome.TextPrimary else AppChrome.TextSecondary,
                style = MaterialTheme.typography.bodyLarge,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = subtitle,
                color = AppChrome.TextMuted,
                style = MaterialTheme.typography.bodySmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun TvPlayerErrorBanner(
    message: String?,
    modifier: Modifier = Modifier,
) {
    if (message.isNullOrBlank()) {
        return
    }
    Surface(
        modifier = modifier,
        color = Color(0xCC2B0F12),
        shape = RoundedCornerShape(14.dp),
    ) {
        Text(
            text = message,
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
            color = Color.White,
            style = MaterialTheme.typography.bodySmall,
        )
    }
}

@Composable
private fun EmptyTvPlayerState() {
    Text(
        text = "当前分集暂无可播放视频",
        color = AppChrome.TextSecondary,
        style = MaterialTheme.typography.bodyLarge,
    )
}

private fun reportTvSeriesHistory(
    viewModel: TvSeriesPlayerViewModel,
    videoId: String,
    player: ExoPlayer,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = player.currentPosition,
        durationMs = player.duration,
    )
    viewModel.reportHistory(videoId, snapshot.watchSeconds, snapshot.completed)
}

private fun resolveTvSubtitleSelectionOnTrackLoad(
    currentSelection: String?,
    storedSelection: String?,
    tracks: List<com.chee.videos.core.model.SubtitleTrackDto>,
    hasStartedPlayback: Boolean,
): String? {
    val current = normalizeTvSubtitleSelection(currentSelection)
    if (hasStartedPlayback && !current.isNullOrBlank() && tracks.any { it.id == current }) {
        return current
    }
    val stored = normalizeTvSubtitleSelection(storedSelection)
    if (!stored.isNullOrBlank() && tracks.any { it.id == stored }) {
        return stored
    }
    return resolveSubtitleSelectionOnTrackLoad(
        currentSelection = current,
        tracks = tracks,
        hasStartedPlayback = hasStartedPlayback,
    )
}

private fun normalizeTvSubtitleSelection(selection: String?): String? =
    selection?.trim()?.takeIf { it.isNotBlank() }
