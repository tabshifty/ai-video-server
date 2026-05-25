package com.chee.videos.feature.detail

import android.app.Activity
import android.content.pm.ActivityInfo
import android.graphics.Color as AndroidColor
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Fullscreen
import androidx.compose.material.icons.filled.FullscreenExit
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
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
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.resolveAvPosterUrl
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.buildLongFormMediaItem
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolveSelectedSubtitleTrack
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.core.util.UrlBuilder

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun DetailScreen(
    onBack: () -> Unit,
    viewModel: DetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val isAv = uiState.videoType.equals("av", ignoreCase = true)
    val useLongFormPlayerControls = isLongFormVideoType(uiState.videoType)
    val pageBg = if (isAv) Color(0xFF0B0C0F) else Color(0xFFF7F8FA)
    val textColor = if (isAv) Color(0xFFEDEFF4) else Color.Unspecified
    val mutedTextColor = if (isAv) Color(0xFFB9C0CD) else MaterialTheme.colorScheme.onSurfaceVariant
    val context = LocalContext.current
    val activity = context as? Activity
    var isFullscreen by rememberSaveable { mutableStateOf(false) }

    BackHandler(enabled = isFullscreen) {
        isFullscreen = false
    }

    DisposableEffect(activity, isFullscreen) {
        if (activity == null || !isFullscreen) {
            onDispose { }
        } else {
            val previousOrientation = activity.requestedOrientation
            val window = activity.window
            val controller = WindowCompat.getInsetsController(window, window.decorView)
            activity.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_USER_LANDSCAPE
            controller.hide(WindowInsetsCompat.Type.systemBars())
            controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE

            onDispose {
                controller.show(WindowInsetsCompat.Type.systemBars())
                activity.requestedOrientation = previousOrientation
            }
        }
    }

    Scaffold(
        topBar = {
            if (!isFullscreen && !useLongFormPlayerControls) {
                CenterAlignedTopAppBar(
                    title = { Text("视频详情") },
                    modifier = Modifier.statusBarsPadding(),
                    windowInsets = WindowInsets(0, 0, 0, 0),
                    navigationIcon = {
                        IconButton(onClick = onBack) {
                            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回")
                        }
                    }
                )
            }
        },
        containerColor = if (isFullscreen) Color.Black else pageBg,
        contentWindowInsets = WindowInsets(0, 0, 0, 0),
    ) { innerPadding ->
        when {
            uiState.loading -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(innerPadding),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator()
                }
            }

            !uiState.errorMessage.isNullOrBlank() && uiState.detail == null -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(innerPadding),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                        Button(onClick = viewModel::load) { Text("重试") }
                    }
                }
            }

            uiState.detail != null -> {
                val detail = uiState.detail!!
                val lifecycleOwner = LocalLifecycleOwner.current
                val playUrl = resolvePlayUrl(uiState.baseUrl, detail, uiState.preferredPlaybackProfile)
                val posterUrl = resolvePosterUrl(uiState.baseUrl, detail)
                val actorModels = remember(uiState.baseUrl, detail) {
                    buildAvDetailActorModels(uiState.baseUrl, detail)
                }
                val canPlay = !playUrl.isNullOrBlank()

                val dataSourceFactory = remember(uiState.accessToken) {
                    DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
                        if (uiState.accessToken.isNotBlank()) {
                            setDefaultRequestProperties(mapOf("Authorization" to "Bearer ${uiState.accessToken}"))
                        }
                    }
                }
                val exoPlayer = remember(uiState.accessToken) {
                    ExoPlayer.Builder(context).build()
                }
                var hasStartedPlayback by rememberSaveable(detail.id) { mutableStateOf(false) }
                var isPausedByUser by rememberSaveable(detail.id) { mutableStateOf(false) }
                var preparedUrl by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
                var preparedSubtitleTrackId by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
                var selectedSubtitleTrackId by rememberSaveable(detail.id) { mutableStateOf<String?>(null) }
                var isPlayerActuallyPlaying by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
                var playerErrorMessage by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
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

                LaunchedEffect(detail.id) {
                    isFullscreen = false
                }

                LaunchedEffect(detail.id, detail.subtitleTracks, hasStartedPlayback) {
                    selectedSubtitleTrackId = resolveSubtitleSelectionOnTrackLoad(
                        currentSelection = selectedSubtitleTrackId,
                        tracks = detail.subtitleTracks,
                        hasStartedPlayback = hasStartedPlayback,
                    )
                }

                LaunchedEffect(playbackSession.hasStartedPlayback, playUrl, dataSourceFactory, selectedSubtitleTrackId) {
                    val updateDecision = resolveLongFormPlayerUpdate(
                        preparedUrl = preparedUrl,
                        nextUrl = playUrl,
                        preparedSubtitleTrackId = preparedSubtitleTrackId,
                        nextSubtitleTrackId = selectedSubtitleTrackId,
                    )
                    if (!playbackSession.hasStartedPlayback || playUrl.isNullOrBlank()) {
                        return@LaunchedEffect
                    }

                    if (updateDecision.shouldReplaceSource) {
                        playerErrorMessage = null
                        val restorePositionMs = if (updateDecision.preservePosition) exoPlayer.currentPosition.coerceAtLeast(0L) else 0L
                        val mediaItem = buildLongFormMediaItem(
                            sourceUrl = playUrl,
                            mediaId = detail.id,
                            title = detail.title,
                            baseUrl = uiState.baseUrl,
                            selectedSubtitleTrack = resolveSelectedSubtitleTrack(detail.subtitleTracks, selectedSubtitleTrackId),
                        )
                        val mediaSource = DefaultMediaSourceFactory(dataSourceFactory).createMediaSource(mediaItem)
                        exoPlayer.setMediaSource(mediaSource, true)
                        exoPlayer.prepare()
                        if (restorePositionMs > 0L) {
                            exoPlayer.seekTo(restorePositionMs)
                        }
                        preparedUrl = playUrl
                        preparedSubtitleTrackId = selectedSubtitleTrackId
                    }

                    if (!playbackSession.isPausedByUser) {
                        exoPlayer.playWhenReady = true
                        exoPlayer.play()
                    }
                }

                LaunchedEffect(playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, playUrl) {
                    if (!playbackSession.hasStartedPlayback || playUrl.isNullOrBlank()) {
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

                DisposableEffect(
                    lifecycleOwner,
                    exoPlayer,
                    playbackSession.hasStartedPlayback,
                    playbackSession.isPausedByUser,
                ) {
                    val observer = LifecycleEventObserver { _, event ->
                        when (event) {
                            Lifecycle.Event.ON_PAUSE -> exoPlayer.pause()
                            Lifecycle.Event.ON_RESUME -> {
                                if (playbackSession.shouldResumeOnLifecycle()) {
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
                    isPlayerActuallyPlaying = exoPlayer.isPlaying
                    exoPlayer.addListener(listener)
                    onDispose {
                        exoPlayer.removeListener(listener)
                        isPlayerActuallyPlaying = false
                        exoPlayer.release()
                    }
                }

                KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)
                val showPlayer = playbackSession.shouldShowPlayer(canPlay)

                if (isFullscreen) {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(innerPadding)
                            .background(Color.Black),
                    ) {
                        if (showPlayer) {
                            if (useLongFormPlayerControls) {
                                Box(modifier = Modifier.fillMaxSize()) {
                                    LongFormVideoPlayer(
                                        title = detail.title,
                                        player = exoPlayer,
                                        isFullscreen = true,
                                        onBack = { isFullscreen = false },
                                        onTogglePlayPause = {
                                            val nextSession = playbackSession.togglePlayPause(canPlay = canPlay)
                                            updatePlaybackSession(nextSession)
                                        },
                                        onToggleFullscreen = { isFullscreen = false },
                                        modifier = Modifier.fillMaxSize(),
                                        showStatusBarPadding = false,
                                        subtitleTracks = detail.subtitleTracks,
                                        selectedSubtitleTrackId = selectedSubtitleTrackId,
                                        onSelectSubtitleTrack = { selectedSubtitleTrackId = it },
                                    )
                                    PlaybackErrorBanner(
                                        message = playerErrorMessage,
                                        modifier = Modifier
                                            .align(Alignment.BottomCenter)
                                            .padding(12.dp),
                                    )
                                }
                            } else {
                                VideoPlayerSurface(
                                    exoPlayer = exoPlayer,
                                    modifier = Modifier.fillMaxSize(),
                                )
                                PlayerOverlayIconButton(
                                    icon = Icons.Filled.FullscreenExit,
                                    contentDescription = "退出全屏",
                                    onClick = { isFullscreen = false },
                                    modifier = Modifier
                                        .align(Alignment.TopEnd)
                                        .statusBarsPadding()
                                        .padding(12.dp),
                                )
                            }
                        } else {
                            Text(
                                text = "暂无可播放视频",
                                color = Color(0xFFEDEFF4),
                                modifier = Modifier.align(Alignment.Center),
                            )
                        }
                    }
                } else {
                    if (isAv) {
                        AvDetailPage(
                            detail = detail,
                            posterUrl = posterUrl,
                            actorModels = actorModels,
                            showPlayer = showPlayer,
                            canPlay = canPlay,
                            useLongFormPlayerControls = useLongFormPlayerControls,
                            exoPlayer = exoPlayer,
                            playerErrorMessage = playerErrorMessage,
                            selectedSubtitleTrackId = selectedSubtitleTrackId,
                            onSelectSubtitleTrack = { selectedSubtitleTrackId = it },
                            onBack = onBack,
                            onToggleFullscreen = { isFullscreen = true },
                            onTogglePlayPause = {
                                val nextSession = playbackSession.togglePlayPause(canPlay = canPlay)
                                updatePlaybackSession(nextSession)
                            },
                            onRequestPlay = {
                                updatePlaybackSession(
                                    playbackSession.requestPlay(canPlay = canPlay),
                                )
                            },
                            errorMessage = uiState.errorMessage,
                            onToggleLike = viewModel::toggleLike,
                            onToggleFavorite = viewModel::toggleFavorite,
                            onToggleDislike = viewModel::toggleDislike,
                            modifier = Modifier
                                .fillMaxSize()
                                .padding(innerPadding),
                        )
                    } else {
                        Column(
                            modifier = Modifier
                                .fillMaxSize()
                                .padding(innerPadding)
                                .background(pageBg)
                                .verticalScroll(rememberScrollState())
                                .padding(16.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(220.dp)
                                    .clip(AppChrome.SurfaceShape)
                                    .background(Color(0xFF1A1C20)),
                            ) {
                                if (showPlayer) {
                                    if (useLongFormPlayerControls) {
                                        Box(modifier = Modifier.fillMaxSize()) {
                                            LongFormVideoPlayer(
                                                title = detail.title,
                                                player = exoPlayer,
                                                isFullscreen = false,
                                                onBack = onBack,
                                                onTogglePlayPause = {
                                                    val nextSession = playbackSession.togglePlayPause(canPlay = canPlay)
                                                    updatePlaybackSession(nextSession)
                                                },
                                                onToggleFullscreen = { isFullscreen = true },
                                                modifier = Modifier.fillMaxSize(),
                                                showStatusBarPadding = false,
                                                subtitleTracks = detail.subtitleTracks,
                                                selectedSubtitleTrackId = selectedSubtitleTrackId,
                                                onSelectSubtitleTrack = { selectedSubtitleTrackId = it },
                                            )
                                            PlaybackErrorBanner(
                                                message = playerErrorMessage,
                                                modifier = Modifier
                                                    .align(Alignment.BottomCenter)
                                                    .padding(12.dp),
                                            )
                                        }
                                    } else {
                                        VideoPlayerSurface(
                                            exoPlayer = exoPlayer,
                                            modifier = Modifier.fillMaxSize(),
                                        )
                                        PlayerOverlayIconButton(
                                            icon = Icons.Filled.Fullscreen,
                                            contentDescription = "全屏播放",
                                            onClick = { isFullscreen = true },
                                            modifier = Modifier
                                                .align(Alignment.BottomEnd)
                                                .padding(10.dp),
                                        )
                                    }
                                } else {
                                    if (!posterUrl.isNullOrBlank()) {
                                        AsyncImage(
                                            model = posterUrl,
                                            contentDescription = "海报",
                                            contentScale = ContentScale.Crop,
                                            modifier = Modifier.fillMaxSize(),
                                        )
                                    }
                                    PlayerOverlayIconButton(
                                        icon = Icons.Filled.PlayArrow,
                                        contentDescription = "播放",
                                        onClick = {
                                            updatePlaybackSession(
                                                playbackSession.requestPlay(canPlay = canPlay),
                                            )
                                        },
                                        enabled = canPlay,
                                        modifier = Modifier.align(Alignment.Center),
                                    )
                                    if (!canPlay) {
                                        Text(
                                            text = "暂无可播放视频",
                                            color = Color(0xFFB9C0CD),
                                            modifier = Modifier
                                                .align(Alignment.BottomCenter)
                                                .padding(bottom = 12.dp),
                                        )
                                    }
                                }
                            }

                            Text(
                                detail.title,
                                style = MaterialTheme.typography.headlineSmall,
                                fontWeight = FontWeight.Bold,
                                color = textColor,
                            )
                            Text(
                                text = detail.description.orEmpty().ifBlank { "暂无简介" },
                                style = MaterialTheme.typography.bodyMedium,
                                color = textColor,
                            )

                            Text("播放数据", color = textColor)
                            Text("时长：${detail.duration} 秒", color = mutedTextColor)
                            Text(
                                "播放：${detail.viewsCount}  点赞：${detail.likesCount}  收藏：${detail.favoritesCount}",
                                color = mutedTextColor,
                            )

                            if (detail.tags.orEmpty().isNotEmpty()) {
                                FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                                    detail.tags.orEmpty().forEach { tag ->
                                        FilterChip(
                                            selected = false,
                                            onClick = {},
                                            label = { Text(tag) },
                                        )
                                    }
                                }
                            }

                            Column(verticalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                                Button(onClick = viewModel::toggleLike, modifier = Modifier.fillMaxWidth()) {
                                    Text(if (detail.userState.isLiked) "取消点赞" else "点赞")
                                }
                                Button(onClick = viewModel::toggleFavorite, modifier = Modifier.fillMaxWidth()) {
                                    Text(if (detail.userState.isFavorited) "取消收藏" else "收藏")
                                }
                                Button(onClick = viewModel::toggleDislike, modifier = Modifier.fillMaxWidth()) {
                                    Text(if (detail.userState.isDisliked) "取消不喜欢" else "不喜欢")
                                }
                            }

                            if (!uiState.errorMessage.isNullOrBlank()) {
                                Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun AvDetailPage(
    detail: VideoDetailDto,
    posterUrl: String?,
    actorModels: List<AvDetailActorModel>,
    showPlayer: Boolean,
    canPlay: Boolean,
    useLongFormPlayerControls: Boolean,
    exoPlayer: ExoPlayer,
    playerErrorMessage: String?,
    selectedSubtitleTrackId: String?,
    onSelectSubtitleTrack: (String?) -> Unit,
    onBack: () -> Unit,
    onToggleFullscreen: () -> Unit,
    onTogglePlayPause: () -> Unit,
    onRequestPlay: () -> Unit,
    errorMessage: String?,
    onToggleLike: () -> Unit,
    onToggleFavorite: () -> Unit,
    onToggleDislike: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val heroModel = buildAvDetailHeroModel(detail)
    val mediaLayoutSpec = buildAvDetailMediaLayoutSpec()

    Column(
        modifier = modifier
            .background(Color(0xFF0B0C0F))
            .then(
                if (mediaLayoutSpec.applyStatusBarPadding) {
                    Modifier.statusBarsPadding()
                } else {
                    Modifier
                },
            )
            .verticalScroll(rememberScrollState())
            .padding(start = 16.dp, end = 16.dp, top = 12.dp, bottom = 20.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(mediaLayoutSpec.aspectRatio)
                .clip(AppChrome.SurfaceShape)
                .background(AppChrome.CanvasRaised),
        ) {
            if (showPlayer) {
                if (useLongFormPlayerControls) {
                    Box(modifier = Modifier.fillMaxSize()) {
                        LongFormVideoPlayer(
                            title = detail.title,
                            player = exoPlayer,
                            isFullscreen = false,
                            onBack = onBack,
                            onTogglePlayPause = onTogglePlayPause,
                            onToggleFullscreen = onToggleFullscreen,
                            modifier = Modifier.fillMaxSize(),
                            showStatusBarPadding = false,
                            subtitleTracks = detail.subtitleTracks,
                            selectedSubtitleTrackId = selectedSubtitleTrackId,
                            onSelectSubtitleTrack = onSelectSubtitleTrack,
                        )
                        PlaybackErrorBanner(
                            message = playerErrorMessage,
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .padding(12.dp),
                        )
                    }
                } else {
                    VideoPlayerSurface(
                        exoPlayer = exoPlayer,
                        modifier = Modifier.fillMaxSize(),
                    )
                    PlayerOverlayIconButton(
                        icon = Icons.Filled.Fullscreen,
                        contentDescription = "全屏播放",
                        onClick = onToggleFullscreen,
                        modifier = Modifier
                            .align(Alignment.BottomEnd)
                            .padding(10.dp),
                    )
                }
            } else {
                if (!posterUrl.isNullOrBlank()) {
                    AsyncImage(
                        model = posterUrl,
                        contentDescription = "海报",
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                }
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.verticalGradient(
                                colors = listOf(
                                    Color(0x22090A0D),
                                    Color(0x55090A0D),
                                    Color(0xE5090A0D),
                                ),
                            ),
                        ),
                )
                PlayerOverlayIconButton(
                    icon = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "返回",
                    onClick = onBack,
                    modifier = Modifier
                        .align(Alignment.TopStart)
                        .padding(14.dp),
                )
                PlayerOverlayIconButton(
                    icon = Icons.Filled.PlayArrow,
                    contentDescription = "播放",
                    onClick = onRequestPlay,
                    enabled = canPlay,
                    modifier = Modifier.align(Alignment.Center),
                )
                Column(
                    modifier = Modifier
                        .align(Alignment.BottomStart)
                        .padding(18.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    Text(
                        text = heroModel.primaryText,
                        style = MaterialTheme.typography.headlineMedium,
                        fontWeight = FontWeight.Bold,
                        color = AppChrome.TextPrimary,
                    )
                    Text(
                        text = heroModel.secondaryTitle,
                        style = MaterialTheme.typography.titleMedium,
                        color = AppChrome.TextSecondary,
                    )
                    if (heroModel.metaItems.isNotEmpty()) {
                        Text(
                            text = heroModel.metaItems.joinToString(" · "),
                            style = MaterialTheme.typography.bodySmall,
                            color = AppChrome.TextMuted,
                        )
                    }
                }
                if (!canPlay) {
                    Text(
                        text = "暂无可播放视频",
                        color = Color(0xFFB9C0CD),
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(bottom = 16.dp),
                    )
                }
            }
        }

        AvDetailOverview(
            detail = detail,
            heroModel = heroModel,
            actorModels = actorModels,
            errorMessage = errorMessage,
            onToggleLike = onToggleLike,
            onToggleFavorite = onToggleFavorite,
            onToggleDislike = onToggleDislike,
        )
    }
}

@Composable
private fun VideoPlayerSurface(
    exoPlayer: ExoPlayer,
    modifier: Modifier = Modifier,
) {
    AndroidView(
        factory = {
            PlayerView(it).apply {
                useController = true
                setShutterBackgroundColor(AndroidColor.BLACK)
                player = exoPlayer
            }
        },
        update = { view ->
            view.player = exoPlayer
        },
        modifier = modifier.background(Color.Black),
    )
}

@Composable
private fun PlayerOverlayIconButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
) {
    IconButton(
        onClick = onClick,
        enabled = enabled,
        modifier = modifier
            .size(46.dp)
            .background(Color(0x5A0A0B0E), CircleShape),
    ) {
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            tint = if (enabled) Color(0xFFEDEFF4) else Color(0xFF7D8593),
        )
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun AvDetailOverview(
    detail: VideoDetailDto,
    heroModel: AvDetailHeroModel,
    actorModels: List<AvDetailActorModel>,
    errorMessage: String?,
    onToggleLike: () -> Unit,
    onToggleFavorite: () -> Unit,
    onToggleDislike: () -> Unit,
) {
    val actionSpecs = buildDetailActionSpecs(detail.userState)

    Column(
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Surface(
            color = AppChrome.Surface,
            shape = AppChrome.SurfaceShape,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 18.dp, vertical = 18.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Text(
                    text = heroModel.primaryText,
                    style = MaterialTheme.typography.headlineMedium,
                    fontWeight = FontWeight.Bold,
                    color = AppChrome.TextPrimary,
                )
                Text(
                    text = heroModel.secondaryTitle,
                    style = MaterialTheme.typography.titleMedium,
                    color = AppChrome.TextSecondary,
                )
                if (heroModel.metaItems.isNotEmpty()) {
                    Text(
                        text = heroModel.metaItems.joinToString(" · "),
                        style = MaterialTheme.typography.bodySmall,
                        color = AppChrome.TextSecondary,
                    )
                }
            }
        }

        Surface(
            color = AppChrome.Surface,
            shape = AppChrome.SurfaceShape,
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 12.dp),
                horizontalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                AvMetricCard(
                    modifier = Modifier.weight(1f),
                    label = "播放",
                    value = detail.viewsCount.toString(),
                )
                AvMetricCard(
                    modifier = Modifier.weight(1f),
                    label = "点赞",
                    value = detail.likesCount.toString(),
                )
                AvMetricCard(
                    modifier = Modifier.weight(1f),
                    label = "收藏",
                    value = detail.favoritesCount.toString(),
                )
            }
        }

        Surface(
            color = AppChrome.Surface,
            shape = AppChrome.SurfaceShape,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                AvSectionTitle("作品信息")
                AvInfoRow(label = "番号", value = heroModel.primaryText)
                AvInfoRow(label = "标题", value = heroModel.secondaryTitle)
                AvInfoRow(label = "时长", value = formatDurationHms(detail.duration))
                heroModel.releaseDate?.let {
                    AvInfoRow(label = "发布时间", value = it)
                }
            }
        }

        Surface(
            color = AppChrome.Surface,
            shape = AppChrome.SurfaceShape,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                AvSectionTitle("演员")
                if (actorModels.isEmpty()) {
                    Text(
                        text = "暂无演员信息",
                        style = MaterialTheme.typography.bodyMedium,
                        color = AppChrome.TextMuted,
                    )
                } else {
                    FlowRow(
                        horizontalArrangement = Arrangement.spacedBy(10.dp),
                        verticalArrangement = Arrangement.spacedBy(10.dp),
                    ) {
                        actorModels.forEach { actor ->
                            AvActorCard(actor = actor)
                        }
                    }
                }
            }
        }

        Surface(
            color = AppChrome.SurfaceElevated,
            shape = AppChrome.SurfaceShape,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 16.dp, vertical = 16.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                AvSectionTitle("简介")
                Text(
                    text = heroModel.overviewText,
                    style = MaterialTheme.typography.bodyMedium,
                    color = AppChrome.TextSecondary,
                )
            }
        }

        if (detail.tags.orEmpty().isNotEmpty()) {
            Surface(
                color = AppChrome.Surface,
                shape = AppChrome.SurfaceShape,
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 16.dp, vertical = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    AvSectionTitle("标签")
                    FlowRow(
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        detail.tags.orEmpty().forEach { tag ->
                            Surface(
                                color = AppChrome.SurfaceStrong,
                                shape = AppChrome.PillShape,
                            ) {
                                Text(
                                    text = tag,
                                    color = AppChrome.TextSecondary,
                                    style = MaterialTheme.typography.bodySmall,
                                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
                                )
                            }
                        }
                    }
                }
            }
        }

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            actionSpecs.take(2).forEach { spec ->
                DetailActionButton(
                    modifier = Modifier.weight(1f),
                    spec = spec,
                    onClick = when (spec.action) {
                        DetailAction.Like -> onToggleLike
                        DetailAction.Favorite -> onToggleFavorite
                        DetailAction.Dislike -> onToggleDislike
                    },
                )
            }
        }
        DetailActionButton(
            modifier = Modifier.fillMaxWidth(),
            spec = actionSpecs.last(),
            onClick = onToggleDislike,
        )

        if (!errorMessage.isNullOrBlank()) {
            Text(
                text = errorMessage,
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
        }
    }
}

@Composable
private fun AvActorCard(actor: AvDetailActorModel) {
    Surface(
        color = AppChrome.SurfaceStrong,
        shape = AppChrome.SurfaceShape,
    ) {
        Column(
            modifier = Modifier
                .width(92.dp)
                .padding(horizontal = 10.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            if (actor.hasAvatar && !actor.avatarUrl.isNullOrBlank()) {
                AsyncImage(
                    model = actor.avatarUrl,
                    contentDescription = "${actor.name}头像",
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .size(56.dp)
                        .clip(AppChrome.SurfaceShape),
                )
            } else {
                Box(
                    modifier = Modifier
                        .size(56.dp)
                        .clip(AppChrome.SurfaceShape)
                        .background(AppChrome.CanvasRaised),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = actor.name.take(1),
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.Bold,
                        color = AppChrome.TextSecondary,
                    )
                }
            }
            Text(
                text = actor.name,
                style = MaterialTheme.typography.bodySmall,
                color = AppChrome.TextPrimary,
                textAlign = TextAlign.Center,
            )
        }
    }
}

@Composable
private fun AvInfoRow(
    label: String,
    value: String,
) {
    Column(
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelMedium,
            color = AppChrome.TextMuted,
        )
        Text(
            text = value,
            style = MaterialTheme.typography.bodyLarge,
            color = AppChrome.TextPrimary,
        )
    }
}

@Composable
private fun AvMetricCard(
    modifier: Modifier = Modifier,
    label: String,
    value: String,
) {
    Surface(
        modifier = modifier,
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.SurfaceShape,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 10.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = value,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
            )
            Text(
                text = label,
                color = AppChrome.TextMuted,
                style = MaterialTheme.typography.labelSmall,
            )
        }
    }
}

@Composable
private fun AvSectionTitle(title: String) {
    Text(
        text = title,
        color = AppChrome.TextPrimary,
        style = MaterialTheme.typography.titleSmall,
        fontWeight = FontWeight.SemiBold,
    )
}

@Composable
private fun DetailActionButton(
    modifier: Modifier = Modifier,
    spec: DetailActionSpec,
    onClick: () -> Unit,
) {
    val containerColor = when (spec.tone) {
        DetailActionTone.Primary -> if (spec.active) AppChrome.Accent else AppChrome.SurfaceStrong
        DetailActionTone.SecondaryDanger -> if (spec.active) Color(0xFF5A0E1B) else AppChrome.SurfaceElevated
    }

    Button(
        onClick = onClick,
        modifier = modifier,
        shape = AppChrome.SurfaceShape,
        colors = ButtonDefaults.buttonColors(
            containerColor = containerColor,
            contentColor = AppChrome.TextPrimary,
        ),
    ) {
        Text(text = spec.label)
    }
}

private fun isLongFormVideoType(type: String): Boolean {
    val normalized = type.trim().lowercase()
    return normalized == "movie" || normalized == "episode" || normalized == "av"
}

@Composable
private fun PlaybackErrorBanner(
    message: String?,
    modifier: Modifier = Modifier,
) {
    val text = message?.trim().orEmpty()
    if (text.isBlank()) {
        return
    }
    Surface(
        modifier = modifier,
        color = Color(0xCC2B0F12),
        shape = AppChrome.SurfaceShape,
    ) {
        Text(
            text = text,
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
            color = Color.White,
            style = MaterialTheme.typography.bodySmall,
        )
    }
}

private fun resolvePlayUrl(baseUrl: String, detail: VideoDetailDto, preferredPlaybackProfile: String): String? {
    val raw = detail.playUrl?.trim().orEmpty()
    if (raw.isNotBlank()) {
        val resolved = resolveResourceUrl(baseUrl, raw) ?: return null
        return appendPlaybackProfileQuery(resolved, preferredPlaybackProfile)
    }
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) {
        return null
    }
    return UrlBuilder.source(normalizedBase, detail.id, preferredPlaybackProfile)
}

private fun resolvePosterUrl(baseUrl: String, detail: VideoDetailDto): String? {
    return resolveAvPosterUrl(baseUrl, detail)
}

internal fun resolveResourceUrl(baseUrl: String, raw: String?): String? {
    val path = raw?.trim().orEmpty()
    if (path.isBlank()) {
        return null
    }
    if (path.startsWith("http://") || path.startsWith("https://")) {
        return path
    }
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) {
        return null
    }
    return if (path.startsWith("/")) "$normalizedBase$path" else "$normalizedBase/$path"
}

private fun appendPlaybackProfileQuery(rawUrl: String, preferredPlaybackProfile: String): String {
    val normalizedProfile = preferredPlaybackProfile.trim()
    if (normalizedProfile.isBlank() || !rawUrl.contains("/source")) {
        return rawUrl
    }
    val separator = if (rawUrl.contains("?")) "&" else "?"
    return "$rawUrl${separator}profile=$normalizedProfile"
}

private fun formatDurationHms(totalSeconds: Int): String {
    val safeSeconds = totalSeconds.coerceAtLeast(0)
    val hours = safeSeconds / 3600
    val minutes = (safeSeconds % 3600) / 60
    val seconds = safeSeconds % 60
    return String.format("%02d:%02d:%02d", hours, minutes, seconds)
}
