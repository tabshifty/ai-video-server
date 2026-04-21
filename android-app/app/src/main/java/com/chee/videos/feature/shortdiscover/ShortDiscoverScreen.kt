package com.chee.videos.feature.shortdiscover

import android.graphics.Color as AndroidColor
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridItemSpan
import androidx.compose.foundation.lazy.staggeredgrid.itemsIndexed
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.AspectRatio
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
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
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
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import coil.imageLoader
import coil.request.ImageRequest
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.ShortVideoOverlayActionButton
import com.chee.videos.core.ui.ShortVideoBottomProgressBar
import com.chee.videos.core.ui.shortPosterContentScale
import com.chee.videos.core.ui.shortVideoResizeMode
import com.chee.videos.core.util.UrlBuilder
import kotlin.math.absoluteValue
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun ShortDiscoverScreen(
    baseUrl: String,
    accessToken: String,
    mode: String,
    value: String,
    title: String,
    onBack: () -> Unit,
    viewModel: ShortDiscoverViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    var playingVideoId by rememberSaveable { mutableStateOf<String?>(null) }

    LaunchedEffect(mode, value, title) {
        viewModel.initialize(mode, value, title)
    }

    val playerStartIndex = remember(playingVideoId, uiState.items) {
        val targetID = playingVideoId ?: return@remember -1
        uiState.items.indexOfFirst { it.id == targetID }
    }

    BackHandler(enabled = playingVideoId != null) {
        playingVideoId = null
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
            .statusBarsPadding(),
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 8.dp, vertical = 6.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                IconButton(onClick = onBack) {
                    Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回", tint = Color.White)
                }
                Column(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(2.dp),
                ) {
                    Text(
                        text = uiState.title.ifBlank { "短视频发现" },
                        color = Color.White,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    val subtitle = if (uiState.totalCount > 0) "共 ${uiState.totalCount} 条内容" else "下滑探索更多内容"
                    Text(
                        text = subtitle,
                        color = Color(0xFFB9C0CC),
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
            }

            when {
                uiState.loading && uiState.items.isEmpty() -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = Color.White)
                    }
                }

                !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Column(
                            horizontalAlignment = Alignment.CenterHorizontally,
                            verticalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                            Surface(
                                shape = RoundedCornerShape(12.dp),
                                color = Color(0xFF242833),
                                modifier = Modifier.clickable(onClick = viewModel::retry),
                            ) {
                                Text(
                                    text = "重试",
                                    color = Color.White,
                                    style = MaterialTheme.typography.bodyMedium,
                                    modifier = Modifier.padding(horizontal = 18.dp, vertical = 10.dp),
                                )
                            }
                        }
                    }
                }

                uiState.items.isEmpty() -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Text("暂无匹配短视频", color = Color.White)
                    }
                }

                else -> {
                    LazyVerticalStaggeredGrid(
                        columns = StaggeredGridCells.Fixed(2),
                        modifier = Modifier.fillMaxSize(),
                        verticalItemSpacing = 10.dp,
                        horizontalArrangement = Arrangement.spacedBy(10.dp),
                        contentPadding = androidx.compose.foundation.layout.PaddingValues(
                            start = 12.dp,
                            end = 12.dp,
                            bottom = 24.dp,
                            top = 8.dp,
                        ),
                    ) {
                        itemsIndexed(uiState.items, key = { _, item -> item.id }) { index, item ->
                            if (index >= uiState.items.lastIndex - 5) {
                                LaunchedEffect(index, uiState.items.size, uiState.loadingMore) {
                                    viewModel.loadMoreIfNeeded(index)
                                }
                            }
                            ShortDiscoverCoverCard(
                                baseUrl = baseUrl,
                                item = item,
                                onClick = { playingVideoId = item.id },
                            )
                        }
                        if (uiState.loadingMore) {
                            item(span = StaggeredGridItemSpan.FullLine) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(vertical = 8.dp),
                                    horizontalArrangement = Arrangement.Center,
                                    verticalAlignment = Alignment.CenterVertically,
                                ) {
                                    CircularProgressIndicator(
                                        modifier = Modifier.width(18.dp),
                                        color = Color.White,
                                        strokeWidth = 2.dp,
                                    )
                                    Text(
                                        text = "正在加载更多",
                                        color = Color(0xFFB9C0CC),
                                        style = MaterialTheme.typography.bodySmall,
                                        modifier = Modifier.padding(start = 8.dp),
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }

        if (playingVideoId != null && playerStartIndex >= 0) {
            ShortDiscoverPlayerOverlay(
                baseUrl = baseUrl,
                accessToken = accessToken,
                items = uiState.items,
                initialIndex = playerStartIndex,
                fitMode = uiState.fitMode,
                onClose = { playingVideoId = null },
                onNeedMore = viewModel::loadMoreIfNeeded,
                onToggleFitMode = viewModel::toggleFitMode,
            )
        }
    }
}

@Composable
private fun ShortDiscoverCoverCard(
    baseUrl: String,
    item: VideoListItemDto,
    onClick: () -> Unit,
) {
    val ratio = remember(item.id) { discoverCardRatio(item.id) }
    val thumbUrl = remember(baseUrl, item.thumbnailPath) {
        resolveThumbnailUrl(baseUrl, item.thumbnailPath)
    }
    Surface(
        color = Color(0xFF141821),
        shape = RoundedCornerShape(14.dp),
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(14.dp))
            .clickable(onClick = onClick),
    ) {
        Box {
            if (thumbUrl.isNullOrBlank()) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(ratio)
                        .background(Color(0xFF0E1118)),
                )
            } else {
                AsyncImage(
                    model = thumbUrl,
                    contentDescription = item.title,
                    contentScale = ContentScale.Crop,
                    modifier = Modifier
                        .fillMaxWidth()
                        .aspectRatio(ratio),
                )
            }
            Box(
                modifier = Modifier
                    .matchParentSize()
                    .background(
                        Brush.verticalGradient(
                            colors = listOf(Color.Transparent, Color.Transparent, Color(0xBF090B11)),
                        ),
                    ),
            )
            Column(
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .padding(horizontal = 10.dp, vertical = 10.dp),
                verticalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                Text(
                    text = item.title,
                    color = Color.White,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    fontWeight = FontWeight.SemiBold,
                )
                if (item.duration > 0) {
                    Text(
                        text = formatDurationHms(item.duration),
                        color = Color(0xFFD2D7E2),
                        style = MaterialTheme.typography.labelSmall,
                    )
                }
            }
        }
    }
}

private fun discoverCardRatio(videoID: String): Float {
    return when (videoID.hashCode().absoluteValue % 3) {
        0 -> 0.72f
        1 -> 0.82f
        else -> 0.92f
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun ShortDiscoverPlayerOverlay(
    baseUrl: String,
    accessToken: String,
    items: List<VideoListItemDto>,
    initialIndex: Int,
    fitMode: VideoFitMode,
    onClose: () -> Unit,
    onNeedMore: (Int) -> Unit,
    onToggleFitMode: () -> Unit,
) {
    if (items.isEmpty() || initialIndex !in items.indices) {
        return
    }
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    val imageLoader = context.imageLoader
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
    val pagerState = rememberPagerState(
        initialPage = initialIndex.coerceIn(0, items.lastIndex),
        pageCount = { items.size },
    )
    var renderedVideoId by remember { mutableStateOf<String?>(null) }
    var pausedByUserVideoIds by remember { mutableStateOf(setOf<String>()) }
    var isPlayerActuallyPlaying by remember { mutableStateOf(false) }

    val currentVideoId = items.getOrNull(pagerState.currentPage)?.id
    val currentVideoPausedByUser = currentVideoId?.let { it in pausedByUserVideoIds } ?: true
    var durationMs by remember(currentVideoId) { mutableStateOf(0L) }
    var positionMs by remember(currentVideoId) { mutableStateOf(0L) }
    var isScrubbingShort by remember(currentVideoId) { mutableStateOf(false) }
    var scrubTargetMs by remember(currentVideoId) { mutableStateOf(0L) }
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

    LaunchedEffect(pagerState.currentPage, items.size, baseUrl) {
        onNeedMore(pagerState.currentPage)
        val page = pagerState.currentPage
        val start = (page - 2).coerceAtLeast(0)
        val end = (page + 2).coerceAtMost(items.lastIndex)
        for (idx in start..end) {
            val item = items.getOrNull(idx) ?: continue
            val thumbUrl = resolveThumbnailUrl(baseUrl, item.thumbnailPath) ?: continue
            imageLoader.enqueue(ImageRequest.Builder(context).data(thumbUrl).build())
        }
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

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        VerticalPager(
            state = pagerState,
            modifier = Modifier.fillMaxSize(),
        ) { page ->
            val item = items[page]
            ShortDiscoverPlayerPage(
                item = item,
                sharedPlayer = sharedPlayer,
                active = pagerState.currentPage == page,
                fitMode = fitMode,
                pausedByUser = item.id in pausedByUserVideoIds,
                posterUrl = resolveThumbnailUrl(baseUrl, item.thumbnailPath),
                showPoster = pagerState.currentPage == page && renderedVideoId != item.id,
                titleBottomPadding = 34.dp,
                onTogglePauseByUser = {
                    val next = pausedByUserVideoIds.toMutableSet()
                    if (next.contains(item.id)) {
                        next.remove(item.id)
                    } else {
                        next.add(item.id)
                    }
                    pausedByUserVideoIds = next
                },
                onToggleFitMode = onToggleFitMode,
            )
        }

        Row(
            modifier = Modifier
                .align(Alignment.TopStart)
                .fillMaxWidth()
                .statusBarsPadding()
                .padding(top = 6.dp, start = 6.dp, end = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            IconButton(onClick = onClose) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回", tint = Color.White)
            }
            Text(
                text = "瀑布流浏览",
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

@Composable
private fun ShortDiscoverPlayerPage(
    item: VideoListItemDto,
    sharedPlayer: ExoPlayer,
    active: Boolean,
    fitMode: VideoFitMode,
    pausedByUser: Boolean,
    posterUrl: String?,
    showPoster: Boolean,
    titleBottomPadding: androidx.compose.ui.unit.Dp,
    onTogglePauseByUser: () -> Unit,
    onToggleFitMode: () -> Unit,
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
                        resizeMode = shortVideoResizeMode(fitMode)
                        player = sharedPlayer
                    }
                },
                update = { view ->
                    view.player = sharedPlayer
                    view.resizeMode = shortVideoResizeMode(fitMode)
                },
                modifier = Modifier.fillMaxSize(),
            )
            if (!posterUrl.isNullOrBlank()) {
                if (showPoster) {
                    AsyncImage(
                        model = posterUrl,
                        contentDescription = "${item.title} 封面",
                        contentScale = shortPosterContentScale(fitMode),
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
                contentScale = shortPosterContentScale(fitMode),
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

        Box(
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .padding(end = 10.dp),
        ) {
            ShortVideoOverlayActionButton(
                icon = Icons.Filled.AspectRatio,
                active = false,
                enabled = true,
                onClick = onToggleFitMode,
                contentDescription = if (fitMode == VideoFitMode.FILL) "切换完整显示" else "切换铺满显示",
            )
        }

        Text(
            text = item.title,
            color = Color.White,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.Bold,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier
                .align(Alignment.BottomStart)
                .navigationBarsPadding()
                .padding(horizontal = 16.dp)
                .padding(bottom = titleBottomPadding),
        )
    }
}

private fun resolveThumbnailUrl(baseUrl: String, rawPath: String?): String? {
    val path = rawPath?.trim().orEmpty()
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

private fun formatDurationHms(totalSeconds: Int): String {
    val safe = totalSeconds.coerceAtLeast(0)
    val hours = safe / 3600
    val minutes = (safe % 3600) / 60
    val seconds = safe % 60
    return if (hours > 0) {
        "%02d:%02d:%02d".format(hours, minutes, seconds)
    } else {
        "%02d:%02d".format(minutes, seconds)
    }
}
