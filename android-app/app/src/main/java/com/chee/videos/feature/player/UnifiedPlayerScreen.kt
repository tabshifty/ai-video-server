package com.chee.videos.feature.player

import android.app.Activity
import android.content.pm.ActivityInfo
import android.graphics.Color as AndroidColor
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
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
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import coil.imageLoader
import coil.request.ImageRequest
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.ShortVideoBottomProgressBar
import com.chee.videos.core.util.UrlBuilder
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun UnifiedPlayerScreen(
    baseUrl: String,
    accessToken: String,
    source: String,
    startVideoId: String,
    onBack: () -> Unit,
    viewModel: UnifiedPlayerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(source, startVideoId) {
        viewModel.load(source = source, startVideoId = startVideoId)
    }

    when {
        uiState.loading && uiState.items.isEmpty() -> {
            Box(
                modifier = Modifier.fillMaxSize().background(Color.Black).statusBarsPadding(),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = Color.White)
            }
        }

        !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
            Box(
                modifier = Modifier.fillMaxSize().background(Color.Black).statusBarsPadding(),
                contentAlignment = Alignment.Center,
            ) {
                Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
            }
        }

        uiState.items.isEmpty() -> {
            Box(
                modifier = Modifier.fillMaxSize().background(Color.Black).statusBarsPadding(),
                contentAlignment = Alignment.Center,
            ) {
                Text("暂无可播放内容", color = Color.White)
            }
        }

        else -> {
            val context = LocalContext.current
            val lifecycleOwner = LocalLifecycleOwner.current
            val pagerState = rememberPagerState(pageCount = { uiState.items.size })
            var pausedByUserVideoIds by remember { mutableStateOf(setOf<String>()) }
            var renderedVideoId by remember { mutableStateOf<String?>(null) }
            var lastHistoryVideoId by remember { mutableStateOf<String?>(null) }
            var isPlayerActuallyPlaying by remember { mutableStateOf(false) }

            val dataSourceFactory = remember(accessToken) {
                DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
                    if (accessToken.isNotBlank()) {
                        setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
                    }
                }
            }
            val sharedPlayer = remember(accessToken) {
                ExoPlayer.Builder(context).build().apply {
                    repeatMode = Player.REPEAT_MODE_ONE
                }
            }
            val imageLoader = context.imageLoader
            val currentVideoId = uiState.items.getOrNull(pagerState.currentPage)?.id
            val currentVideoType = uiState.items.getOrNull(pagerState.currentPage)?.type.orEmpty()
            val currentIsLongForm = isLongFormVideoType(currentVideoType)
            val currentVideoPausedByUser = currentVideoId?.let { id -> pausedByUserVideoIds.contains(id) } ?: false
            var durationMs by remember(currentVideoId) { mutableStateOf(0L) }
            var positionMs by remember(currentVideoId) { mutableStateOf(0L) }
            var isScrubbingShort by remember(currentVideoId) { mutableStateOf(false) }
            var scrubTargetMs by remember(currentVideoId) { mutableStateOf(0L) }
            val latestCurrentVideoId by rememberUpdatedState(currentVideoId)
            val activity = context as? Activity
            var isFullscreen by rememberSaveable { mutableStateOf(false) }

            BackHandler(enabled = isFullscreen && currentIsLongForm) {
                isFullscreen = false
            }

            DisposableEffect(activity, isFullscreen, currentIsLongForm) {
                if (activity == null || !isFullscreen || !currentIsLongForm) {
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
                isPlayerActuallyPlaying = sharedPlayer.isPlaying
                sharedPlayer.addListener(listener)
                onDispose {
                    val finalVideoID = latestCurrentVideoId ?: sharedPlayer.currentMediaItem?.mediaId
                    if (!finalVideoID.isNullOrBlank()) {
                        val (watchSeconds, completed) = readWatchSnapshot(sharedPlayer)
                        viewModel.reportHistory(finalVideoID, watchSeconds, completed)
                    }
                    sharedPlayer.removeListener(listener)
                    isPlayerActuallyPlaying = false
                    sharedPlayer.release()
                }
            }
            DisposableEffect(sharedPlayer, currentVideoId) {
                val listener = object : Player.Listener {
                    override fun onEvents(player: Player, events: Player.Events) {
                        val playerDuration = player.duration
                        durationMs = if (playerDuration > 0L) playerDuration else 0L
                        if (!isScrubbingShort) {
                            positionMs = player.currentPosition.coerceAtLeast(0L)
                        }
                    }
                }
                durationMs = sharedPlayer.duration.takeIf { it > 0L } ?: 0L
                positionMs = sharedPlayer.currentPosition.coerceAtLeast(0L)
                sharedPlayer.addListener(listener)
                onDispose {
                    sharedPlayer.removeListener(listener)
                }
            }
            LaunchedEffect(currentVideoId) {
                if (currentVideoId.isNullOrBlank()) {
                    durationMs = 0L
                    positionMs = 0L
                    isScrubbingShort = false
                    scrubTargetMs = 0L
                    return@LaunchedEffect
                }
                while (isActive) {
                    if (!isScrubbingShort) {
                        durationMs = sharedPlayer.duration.takeIf { it > 0L } ?: 0L
                        positionMs = sharedPlayer.currentPosition.coerceAtLeast(0L)
                    }
                    delay(220)
                }
            }

            LaunchedEffect(uiState.startIndex, uiState.items.size) {
                if (uiState.items.isEmpty()) {
                    return@LaunchedEffect
                }
                val targetIndex = uiState.startIndex.coerceIn(0, uiState.items.lastIndex)
                if (pagerState.currentPage != targetIndex) {
                    pagerState.scrollToPage(targetIndex)
                }
            }

            LaunchedEffect(pagerState.currentPage, uiState.items, baseUrl) {
                if (uiState.items.isEmpty()) {
                    return@LaunchedEffect
                }
                val page = pagerState.currentPage
                val start = (page - 2).coerceAtLeast(0)
                val end = (page + 2).coerceAtMost(uiState.items.lastIndex)
                for (idx in start..end) {
                    val item = uiState.items.getOrNull(idx) ?: continue
                    val thumbUrl = resolveThumbnailUrl(baseUrl, item.thumbnailPath) ?: continue
                    imageLoader.enqueue(ImageRequest.Builder(context).data(thumbUrl).build())
                }
            }

            LaunchedEffect(currentVideoId) {
                if (!currentVideoId.isNullOrBlank()) {
                    viewModel.ensureDetailLoaded(currentVideoId)
                }
            }
            LaunchedEffect(currentVideoType) {
                if (!isLongFormVideoType(currentVideoType)) {
                    isFullscreen = false
                }
            }
            LaunchedEffect(currentVideoId) {
                val previousVideoId = lastHistoryVideoId
                if (!previousVideoId.isNullOrBlank() && previousVideoId != currentVideoId) {
                    val (watchSeconds, completed) = readWatchSnapshot(sharedPlayer)
                    viewModel.reportHistory(previousVideoId, watchSeconds, completed)
                }
                lastHistoryVideoId = currentVideoId
            }

            LaunchedEffect(currentVideoId, baseUrl, dataSourceFactory) {
                if (currentVideoId.isNullOrBlank()) {
                    sharedPlayer.stop()
                    sharedPlayer.clearMediaItems()
                    renderedVideoId = null
                    return@LaunchedEffect
                }
                renderedVideoId = null
                sharedPlayer.stop()
                sharedPlayer.clearMediaItems()

                val mediaItem = MediaItem.Builder()
                    .setUri(UrlBuilder.source(baseUrl, currentVideoId))
                    .setMediaId(currentVideoId)
                    .build()
                val mediaSource = ProgressiveMediaSource.Factory(dataSourceFactory).createMediaSource(mediaItem)
                sharedPlayer.setMediaSource(mediaSource, true)
                sharedPlayer.prepare()
            }

            LaunchedEffect(currentVideoId, currentVideoPausedByUser) {
                val shouldPlay = !currentVideoId.isNullOrBlank() && !currentVideoPausedByUser
                sharedPlayer.playWhenReady = shouldPlay
                if (shouldPlay) {
                    sharedPlayer.play()
                } else {
                    sharedPlayer.pause()
                }
            }

            DisposableEffect(lifecycleOwner, sharedPlayer, currentVideoId, currentVideoPausedByUser) {
                val observer = LifecycleEventObserver { _, event ->
                    when (event) {
                        Lifecycle.Event.ON_PAUSE -> sharedPlayer.pause()
                        Lifecycle.Event.ON_RESUME -> {
                            if (!currentVideoId.isNullOrBlank() && !currentVideoPausedByUser) {
                                sharedPlayer.play()
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

            val fullscreenLongForm = isFullscreen && currentIsLongForm
            val containerModifier = if (fullscreenLongForm) {
                Modifier.fillMaxSize().background(Color.Black)
            } else {
                Modifier.fillMaxSize().background(Color.Black).statusBarsPadding()
            }

            Box(modifier = containerModifier) {
                VerticalPager(
                    state = pagerState,
                    modifier = Modifier.fillMaxSize(),
                    userScrollEnabled = !fullscreenLongForm,
                ) { page ->
                    val item = uiState.items[page]
                    val isActive = pagerState.currentPage == page
                    val isLongForm = isLongFormVideoType(item.type)
                    val togglePauseState: () -> Unit = {
                        val next = pausedByUserVideoIds.toMutableSet()
                        if (next.contains(item.id)) {
                            next.remove(item.id)
                        } else {
                            next.add(item.id)
                        }
                        pausedByUserVideoIds = next
                    }
                    if (isLongForm) {
                        UnifiedLongFormPage(
                            item = item,
                            sharedPlayer = sharedPlayer,
                            active = isActive,
                            posterUrl = resolveThumbnailUrl(baseUrl, item.thumbnailPath),
                            showPoster = isActive && renderedVideoId != item.id,
                            isFullscreen = fullscreenLongForm && isActive,
                            onBack = {
                                if (fullscreenLongForm && isActive) {
                                    isFullscreen = false
                                } else {
                                    onBack()
                                }
                            },
                            onTogglePauseByUser = togglePauseState,
                            onToggleFullscreen = { isFullscreen = !(fullscreenLongForm && isActive) },
                        )
                    } else {
                        UnifiedShortVideoPage(
                            item = item,
                            detail = uiState.detailByVideoId[item.id],
                            sharedPlayer = sharedPlayer,
                            active = isActive,
                            pausedByUser = item.id in pausedByUserVideoIds,
                            posterUrl = resolveThumbnailUrl(baseUrl, item.thumbnailPath),
                            showPoster = isActive && renderedVideoId != item.id,
                            onTogglePauseByUser = togglePauseState,
                        )
                    }
                }

                if (!currentIsLongForm && !fullscreenLongForm) {
                    Row(
                        modifier = Modifier
                            .align(Alignment.TopStart)
                            .fillMaxWidth()
                            .padding(top = 16.dp, start = 6.dp, end = 12.dp),
                        horizontalArrangement = Arrangement.spacedBy(10.dp),
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        IconButton(onClick = onBack) {
                            Icon(Icons.Filled.ArrowBack, contentDescription = "返回", tint = Color.White)
                        }
                        Text(
                            text = sourceLabel(source),
                            color = Color.White,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                        )
                    }

                    if (!currentVideoId.isNullOrBlank()) {
                        ShortVideoBottomProgressBar(
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth()
                                .navigationBarsPadding(),
                            positionMs = positionMs,
                            durationMs = durationMs,
                            isScrubbing = isScrubbingShort,
                            scrubTargetMs = scrubTargetMs,
                            onScrubStart = {
                                if (durationMs <= 0L) {
                                    return@ShortVideoBottomProgressBar
                                }
                                isScrubbingShort = true
                                scrubTargetMs = positionMs.coerceIn(0L, durationMs)
                            },
                            onScrubToFraction = { fraction ->
                                if (durationMs <= 0L || !isScrubbingShort) {
                                    return@ShortVideoBottomProgressBar
                                }
                                scrubTargetMs = (durationMs * fraction).toLong().coerceIn(0L, durationMs)
                            },
                            onScrubEnd = {
                                if (!isScrubbingShort) {
                                    return@ShortVideoBottomProgressBar
                                }
                                if (durationMs > 0L) {
                                    val target = scrubTargetMs.coerceIn(0L, durationMs)
                                    sharedPlayer.seekTo(target)
                                    positionMs = target
                                }
                                isScrubbingShort = false
                            },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun UnifiedShortVideoPage(
    item: PlayerVideoItem,
    detail: VideoDetailDto?,
    sharedPlayer: ExoPlayer,
    active: Boolean,
    pausedByUser: Boolean,
    posterUrl: String?,
    showPoster: Boolean,
    onTogglePauseByUser: () -> Unit,
) {
    val scope = rememberCoroutineScope()
    var showCenterIndicator by remember { mutableStateOf(false) }
    var centerIsPause by remember { mutableStateOf(false) }
    var hideIndicatorJob by remember { mutableStateOf<Job?>(null) }

    DisposableEffect(Unit) {
        onDispose {
            hideIndicatorJob?.cancel()
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
            .pointerInput(item.id, pausedByUser) {
                detectTapGestures(
                    onTap = {
                        val nextPaused = !pausedByUser
                        onTogglePauseByUser()
                        centerIsPause = nextPaused
                        showCenterIndicator = true
                        hideIndicatorJob?.cancel()
                        hideIndicatorJob = scope.launch {
                            delay(700)
                            showCenterIndicator = false
                        }
                    },
                )
            },
    ) {
        if (active) {
            AndroidView(
                factory = {
                    PlayerView(it).apply {
                        useController = false
                        setShutterBackgroundColor(AndroidColor.BLACK)
                        setKeepContentOnPlayerReset(false)
                        resizeMode = AspectRatioFrameLayout.RESIZE_MODE_ZOOM
                        player = sharedPlayer
                    }
                },
                update = { view ->
                    view.player = sharedPlayer
                },
                modifier = Modifier.fillMaxSize(),
            )
            if (!posterUrl.isNullOrBlank()) {
                if (showPoster) {
                    AsyncImage(
                        model = posterUrl,
                        contentDescription = "${item.title} 封面",
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                }
            } else if (showPoster) {
                Box(modifier = Modifier.fillMaxSize().background(Color.Black))
            }
        } else if (!posterUrl.isNullOrBlank()) {
            AsyncImage(
                model = posterUrl,
                contentDescription = "${item.title} 封面",
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize(),
            )
        } else {
            Box(modifier = Modifier.fillMaxSize().background(Color.Black))
        }

        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(Color.Transparent, Color.Transparent, Color(0xB0000000)),
                    ),
                ),
        )

        AnimatedVisibility(
            visible = showCenterIndicator,
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.Center),
        ) {
            Surface(
                shape = CircleShape,
                color = Color(0x88000000),
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    Icon(
                        imageVector = if (centerIsPause) Icons.Filled.Pause else Icons.Filled.PlayArrow,
                        contentDescription = null,
                        tint = Color.White,
                    )
                    Text(
                        text = if (centerIsPause) "已暂停" else "继续播放",
                        color = Color.White,
                        style = MaterialTheme.typography.bodyMedium,
                    )
                }
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomStart)
                .padding(horizontal = 16.dp, vertical = 22.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text(
                text = item.title,
                color = Color.White,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
            )
            val actorNames = extractActorNames(detail)
            if (actorNames.isNotEmpty()) {
                Text(
                    text = "演员：${actorNames.joinToString(" / ")}",
                    color = Color.White.copy(alpha = 0.84f),
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun UnifiedLongFormPage(
    item: PlayerVideoItem,
    sharedPlayer: ExoPlayer,
    active: Boolean,
    posterUrl: String?,
    showPoster: Boolean,
    isFullscreen: Boolean,
    onBack: () -> Unit,
    onTogglePauseByUser: () -> Unit,
    onToggleFullscreen: () -> Unit,
) {
    if (active) {
        LongFormVideoPlayer(
            title = item.title,
            player = sharedPlayer,
            isFullscreen = isFullscreen,
            onBack = onBack,
            onTogglePlayPause = onTogglePauseByUser,
            onToggleFullscreen = onToggleFullscreen,
            modifier = Modifier.fillMaxSize(),
            showStatusBarPadding = false,
            posterUrl = posterUrl,
            showPoster = showPoster,
        )
        return
    }

    if (!posterUrl.isNullOrBlank()) {
        AsyncImage(
            model = posterUrl,
            contentDescription = "${item.title} 封面",
            contentScale = ContentScale.Crop,
            modifier = Modifier.fillMaxSize(),
        )
    } else {
        Box(modifier = Modifier.fillMaxSize().background(Color.Black))
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

private fun resolveThumbnailUrl(baseUrl: String, rawPath: String?): String? {
    val path = rawPath?.trim().orEmpty()
    if (path.isBlank()) {
        return null
    }
    if (path.startsWith("http://") || path.startsWith("https://")) {
        return path
    }
    val normalized = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalized.isBlank()) {
        return null
    }
    return if (path.startsWith('/')) "$normalized$path" else "$normalized/$path"
}

private fun extractActorNames(detail: VideoDetailDto?): List<String> {
    if (detail == null) {
        return emptyList()
    }

    val relationNames = detail.actors.orEmpty()
        .mapNotNull { actor -> actor.name.trim().takeIf { it.isNotBlank() } }
        .distinct()
    if (relationNames.isNotEmpty()) {
        return relationNames
    }

    val raw = detail.metadata?.get("actors")
    return when (raw) {
        is List<*> -> raw.mapNotNull { row ->
            when (row) {
                is String -> row.trim().takeIf { it.isNotBlank() }
                is Map<*, *> -> (row["name"] as? String)?.trim()?.takeIf { it.isNotBlank() }
                else -> null
            }
        }.distinct()

        is String -> raw
            .split(",", "/", "|", "，", "、")
            .map { it.trim() }
            .filter { it.isNotBlank() }
            .distinct()

        else -> emptyList()
    }
}

private fun sourceLabel(source: String): String {
    return when (source) {
        "history" -> "播放历史"
        "favorite" -> "我的收藏"
        "like" -> "我的喜欢"
        else -> "视频播放"
    }
}

private fun isLongFormVideoType(type: String): Boolean {
    val normalized = type.trim().lowercase()
    return normalized == "movie" || normalized == "episode" || normalized == "av"
}
