package com.chee.videos.feature.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.SlowMotionVideo
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
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
import androidx.compose.runtime.setValue
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.buildLongFormMediaItem
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolveSelectedSubtitleTrack
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.feature.detail.LongFormPlaybackSession

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TvSeriesPlayerScreen(
    accessToken: String,
    onBack: () -> Unit,
    viewModel: TvSeriesPlayerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current

    if (uiState.loading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
            contentAlignment = Alignment.Center,
        ) {
            CircularProgressIndicator(color = AppChrome.AccentStrong)
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
        }
        return
    }

    val currentEpisode = selectedEpisode(uiState)
    val currentSeason = selectedSeason(uiState)
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

    LaunchedEffect(uiState.currentVideoId) {
        if (lastHistoryVideoId.isNotBlank() && lastHistoryVideoId != uiState.currentVideoId) {
            val (watchSeconds, completed) = readWatchSnapshot(exoPlayer)
            viewModel.reportHistory(lastHistoryVideoId, watchSeconds, completed)
        }
        lastHistoryVideoId = uiState.currentVideoId
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
            if (latestCurrentVideoId.isNotBlank()) {
                val (watchSeconds, completed) = readWatchSnapshot(exoPlayer)
                viewModel.reportHistory(latestCurrentVideoId, watchSeconds, completed)
            }
            exoPlayer.removeListener(listener)
            exoPlayer.release()
        }
    }

    DisposableEffect(lifecycleOwner, exoPlayer, playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, canPlay) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> exoPlayer.pause()
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
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(Color(0xFF111827), Color(0xFF070B13), Color.Black),
                    ),
                ),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .statusBarsPadding()
                    .padding(horizontal = 8.dp, vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Surface(color = Color(0x4DFFFFFF), shape = CircleShape) {
                    IconButton(onClick = onBack) {
                        Icon(
                            imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "返回",
                            tint = Color.White,
                        )
                    }
                }
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = series.title,
                        color = Color.White,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Text(
                        text = "S${uiState.selectedSeasonNumber} · E${uiState.selectedEpisodeNumber}  ${currentEpisode?.title.orEmpty()}",
                        color = AppChrome.TextSecondary,
                        style = MaterialTheme.typography.bodySmall,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(260.dp)
                    .padding(horizontal = 12.dp)
                    .clip(RoundedCornerShape(18.dp))
                    .background(Color(0xFF0D111A)),
            ) {
                if (canPlay) {
                    Box(modifier = Modifier.fillMaxSize()) {
                        LongFormVideoPlayer(
                            title = currentEpisode?.title ?: series.title,
                            player = exoPlayer,
                            isFullscreen = false,
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
                            showStatusBarPadding = false,
                        )
                        if (!playerErrorMessage.isNullOrBlank()) {
                            Surface(
                                modifier = Modifier
                                    .align(Alignment.BottomCenter)
                                    .padding(12.dp),
                                color = Color(0xCC2B0F12),
                                shape = RoundedCornerShape(14.dp),
                            ) {
                                Text(
                                    text = playerErrorMessage.orEmpty(),
                                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
                                    color = Color.White,
                                    style = MaterialTheme.typography.bodySmall,
                                )
                            }
                        }
                    }
                } else {
                    EmptyTvPlayerState()
                }
            }

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 14.dp, vertical = 14.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Surface(color = AppChrome.SurfaceElevated, shape = RoundedCornerShape(14.dp)) {
                    Text(
                        text = currentEpisode?.summary ?: "暂无剧情简介",
                        color = AppChrome.TextSecondary,
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 12.dp),
                    )
                }
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                    TvPlayerActionButton(
                        label = if (playbackSession.isPausedByUser) "继续" else "暂停",
                        icon = Icons.Filled.PlayArrow,
                        modifier = Modifier.weight(1f),
                        enabled = canPlay,
                        onClick = {
                            updatePlaybackSession(playbackSession.togglePlayPause(canPlay = canPlay))
                        },
                    )
                    TvPlayerActionButton(
                        label = "${uiState.playbackSpeed}x",
                        icon = Icons.Filled.SlowMotionVideo,
                        modifier = Modifier.weight(1f),
                        onClick = viewModel::cycleSpeed,
                    )
                    TvPlayerActionButton(
                        label = "下一集",
                        icon = Icons.Filled.SkipNext,
                        modifier = Modifier.weight(1f),
                        onClick = viewModel::nextEpisode,
                    )
                }
                Surface(
                    color = AppChrome.AccentSoft,
                    shape = AppChrome.PillShape,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { viewModel.setSelectorVisible(true) },
                ) {
                    Text(
                        text = "打开选集抽屉（${currentSeason?.title.orEmpty()}）",
                        color = AppChrome.TextPrimary,
                        style = MaterialTheme.typography.labelLarge,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
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
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.padding(horizontal = 16.dp),
                ) {
                    series.seasons.forEach { season ->
                        val selected = season.number == uiState.selectedSeasonNumber
                        Surface(
                            color = if (selected) AppChrome.AccentSoft else AppChrome.SurfaceElevated,
                            shape = RoundedCornerShape(10.dp),
                            modifier = Modifier.clickable { viewModel.selectSeason(season.number) },
                        ) {
                            Text(
                                text = season.title,
                                color = if (selected) AppChrome.TextPrimary else AppChrome.TextMuted,
                                style = MaterialTheme.typography.labelLarge,
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                            )
                        }
                    }
                }
                LazyColumn(
                    contentPadding = PaddingValues(start = 16.dp, end = 16.dp, bottom = 18.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    items(selectedSeason(uiState)?.episodes.orEmpty(), key = { episode -> episode.id }) { episode ->
                        val selected = episode.number == uiState.selectedEpisodeNumber
                        Surface(
                            color = if (selected) AppChrome.AccentSoft else AppChrome.SurfaceElevated,
                            shape = RoundedCornerShape(12.dp),
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    viewModel.selectEpisode(episode.number)
                                    viewModel.setSelectorVisible(false)
                                },
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = 12.dp, vertical = 10.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                            ) {
                                Text(
                                    text = "E${episode.number}",
                                    color = if (selected) AppChrome.TextPrimary else AppChrome.TextSecondary,
                                    style = MaterialTheme.typography.labelLarge,
                                    fontWeight = FontWeight.SemiBold,
                                    modifier = Modifier.width(42.dp),
                                )
                                Column(
                                    modifier = Modifier.weight(1f),
                                    verticalArrangement = Arrangement.spacedBy(2.dp),
                                ) {
                                    Text(
                                        text = episode.title,
                                        color = AppChrome.TextPrimary,
                                        style = MaterialTheme.typography.bodyMedium,
                                        maxLines = 1,
                                        overflow = TextOverflow.Ellipsis,
                                    )
                                    Text(
                                        text = if (episode.playable) episode.durationLabel else "待绑定 / 未就绪",
                                        color = AppChrome.TextMuted,
                                        style = MaterialTheme.typography.bodySmall,
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun EmptyTvPlayerState() {
    Column(
        modifier = Modifier.fillMaxSize(),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Icon(
            imageVector = Icons.Filled.Tv,
            contentDescription = null,
            tint = Color.White.copy(alpha = 0.82f),
            modifier = Modifier.size(40.dp),
        )
        Text("当前分集暂无可播放视频", color = Color.White, style = MaterialTheme.typography.bodyMedium)
    }
}

@Composable
private fun TvPlayerActionButton(
    label: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    onClick: () -> Unit,
) {
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(12.dp),
        modifier = modifier.clickable(enabled = enabled, onClick = onClick),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 10.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center,
        ) {
            Icon(icon, contentDescription = null, tint = AppChrome.TextSecondary, modifier = Modifier.size(16.dp))
            Text(
                text = label,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.labelLarge,
                modifier = Modifier.padding(start = 6.dp),
            )
        }
    }
}

private fun readWatchSnapshot(player: ExoPlayer): Pair<Int, Boolean> {
    val watchSeconds = (player.currentPosition.coerceAtLeast(0L) / 1000L).toInt()
    val durationMs = player.duration
    val durationSeconds = if (durationMs > 0) (durationMs / 1000L).toInt() else 0
    val completedThreshold = (durationSeconds - 3).coerceAtLeast(1)
    val completed = durationSeconds > 0 && watchSeconds >= completedThreshold
    return watchSeconds to completed
}

private fun resolveTvSubtitleSelectionOnTrackLoad(
    currentSelection: String?,
    storedSelection: String?,
    tracks: List<SubtitleTrackDto>,
    hasStartedPlayback: Boolean,
): String? {
    val preferredSelection = currentSelection ?: storedSelection
    if (preferredSelection != null && preferredSelection.isBlank()) {
        return ""
    }
    return resolveSubtitleSelectionOnTrackLoad(
        currentSelection = preferredSelection,
        tracks = tracks,
        hasStartedPlayback = hasStartedPlayback,
    )
}

private fun normalizeTvSubtitleSelection(selection: String?): String? =
    selection?.takeIf { it.isNotBlank() }
