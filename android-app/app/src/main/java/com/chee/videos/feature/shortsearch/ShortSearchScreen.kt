package com.chee.videos.feature.shortsearch

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
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridItemSpan
import androidx.compose.foundation.lazy.staggeredgrid.itemsIndexed
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AspectRatio
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.ThumbUp
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
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
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.model.toPlayerRepeatMode
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.ShortPlaybackModeToggleButton
import com.chee.videos.core.ui.ShortOverlayFullscreenButton
import com.chee.videos.core.ui.ShortOverlayFullscreenHost
import com.chee.videos.core.ui.ShortVideoBottomProgressBar
import com.chee.videos.core.ui.ShortVideoOverlayActionButton
import com.chee.videos.core.ui.shouldShowShortOverlayProgressBar
import com.chee.videos.core.ui.shortNonHomeProgressBarPadding
import com.chee.videos.core.ui.shortPosterContentScale
import com.chee.videos.core.ui.shortPlaybackModeLabel
import com.chee.videos.core.ui.shortScrubTargetFromDelta
import com.chee.videos.core.ui.shortVideoResizeMode
import com.chee.videos.core.util.UrlBuilder
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun ShortSearchScreen(
    baseUrl: String,
    accessToken: String,
    onFullscreenChange: (Boolean) -> Unit = {},
    viewModel: ShortSearchViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val playerStartIndex = remember(uiState.playingVideoId, uiState.items) {
        val id = uiState.playingVideoId ?: return@remember -1
        uiState.items.indexOfFirst { it.id == id }
    }

    BackHandler(enabled = uiState.playingVideoId != null) { viewModel.closePlayer() }

    Column(modifier = Modifier.fillMaxSize().background(Color(0xFF0B0E15)).statusBarsPadding()) {
        OutlinedTextField(
            value = uiState.queryInput,
            onValueChange = viewModel::onQueryInputChange,
            label = { Text("搜索短视频（标题/标签）") },
            singleLine = true,
            modifier = Modifier.fillMaxWidth().padding(12.dp),
        )

        when {
            uiState.loading -> Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator(color = Color.White) }
            uiState.errorMessage != null && uiState.items.isEmpty() -> Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                    Text("点击重试", color = Color.White, modifier = Modifier.clip(RoundedCornerShape(8.dp)).clickable(onClick = viewModel::retry).padding(horizontal = 14.dp, vertical = 8.dp))
                }
            }
            uiState.activeQuery.isBlank() -> Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { Text("输入关键词开始搜索", color = Color(0xFFB9C0CC)) }
            uiState.items.isEmpty() -> Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { Text("暂无匹配结果", color = Color.White) }
            else -> LazyVerticalStaggeredGrid(
                columns = StaggeredGridCells.Fixed(2),
                modifier = Modifier.fillMaxSize(),
                verticalItemSpacing = 10.dp,
                horizontalArrangement = Arrangement.spacedBy(10.dp),
                contentPadding = PaddingValues(start = 12.dp, end = 12.dp, bottom = 24.dp, top = 8.dp),
            ) {
                itemsIndexed(uiState.items, key = { _, item -> item.id }) { index, item ->
                    if (index >= uiState.items.lastIndex - 5) {
                        LaunchedEffect(index, uiState.items.size, uiState.loadingMore) {
                            viewModel.loadMoreIfNeeded(index)
                        }
                    }
                    SearchCoverCard(baseUrl = baseUrl, item = item, onClick = { viewModel.enterPlayer(item.id) })
                }
                if (uiState.loadingMore) {
                    item(span = StaggeredGridItemSpan.FullLine) {
                        Row(modifier = Modifier.fillMaxWidth().padding(vertical = 8.dp), horizontalArrangement = Arrangement.Center) {
                            CircularProgressIndicator(color = Color.White, strokeWidth = 2.dp)
                        }
                    }
                }
            }
        }
    }

    if (uiState.playingVideoId != null && playerStartIndex >= 0) {
        ShortSearchPlayerOverlay(
            baseUrl = baseUrl,
            accessToken = accessToken,
            items = uiState.items,
            initialIndex = playerStartIndex,
            fitMode = uiState.fitMode,
            playbackMode = uiState.playbackMode,
            detailByVideoId = uiState.detailByVideoId,
            detailLoadingVideoIds = uiState.detailLoadingVideoIds,
            actionBusyVideoIds = uiState.actionBusyVideoIds,
            onNeedMore = viewModel::loadMoreIfNeeded,
            onEnsureDetailLoaded = viewModel::ensureDetailLoaded,
            onToggleLike = viewModel::toggleLike,
            onToggleFavorite = viewModel::toggleFavorite,
            onClose = viewModel::closePlayer,
            onToggleFitMode = viewModel::toggleFitMode,
            onTogglePlaybackMode = viewModel::togglePlaybackMode,
            onFullscreenChange = onFullscreenChange,
        )
    }
}

@Composable
private fun SearchCoverCard(baseUrl: String, item: VideoListItemDto, onClick: () -> Unit) {
    val thumbUrl = remember(baseUrl, item.thumbnailPath) { resolveThumbnailUrl(baseUrl, item.thumbnailPath) }
    Surface(color = Color(0xFF141821), shape = RoundedCornerShape(8.dp), modifier = Modifier.fillMaxWidth().clip(RoundedCornerShape(8.dp)).clickable(onClick = onClick)) {
        Box {
            AsyncImage(model = thumbUrl, contentDescription = item.title, contentScale = ContentScale.Crop, modifier = Modifier.fillMaxWidth().aspectRatio(0.72f))
            Box(modifier = Modifier.matchParentSize().background(Brush.verticalGradient(listOf(Color.Transparent, Color.Transparent, Color(0xBF090B11)))))
            Text(text = item.title, color = Color.White, style = MaterialTheme.typography.bodySmall, maxLines = 2, overflow = TextOverflow.Ellipsis, fontWeight = FontWeight.SemiBold, modifier = Modifier.align(Alignment.BottomStart).padding(10.dp))
        }
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun ShortSearchPlayerOverlay(
    baseUrl: String,
    accessToken: String,
    items: List<VideoListItemDto>,
    initialIndex: Int,
    fitMode: VideoFitMode,
    playbackMode: ShortPlaybackMode,
    detailByVideoId: Map<String, VideoDetailDto>,
    detailLoadingVideoIds: Set<String>,
    actionBusyVideoIds: Set<String>,
    onNeedMore: (Int) -> Unit,
    onEnsureDetailLoaded: (String) -> Unit,
    onToggleLike: (String) -> Unit,
    onToggleFavorite: (String) -> Unit,
    onClose: () -> Unit,
    onToggleFitMode: () -> Unit,
    onTogglePlaybackMode: () -> Unit,
    onFullscreenChange: (Boolean) -> Unit,
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    val pagerState = rememberPagerState(initialPage = initialIndex.coerceIn(0, items.lastIndex), pageCount = { items.size })
    val coroutineScope = rememberCoroutineScope()
    var pausedByUserVideoIds by remember { mutableStateOf(setOf<String>()) }
    var renderedVideoId by remember { mutableStateOf<String?>(null) }
    var isPlayerActuallyPlaying by remember { mutableStateOf(false) }
    var showController by rememberSaveable { mutableStateOf(true) }
    var durationMs by remember { mutableStateOf(0L) }
    var positionMs by remember { mutableStateOf(0L) }
    var isScrubbingShort by remember { mutableStateOf(false) }
    var scrubAnchorMs by remember { mutableStateOf(0L) }
    var scrubTargetMs by remember { mutableStateOf(0L) }
    var isFullscreen by rememberSaveable { mutableStateOf(false) }

    val dataSourceFactory = remember(accessToken) {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
            if (accessToken.isNotBlank()) setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
        }
    }
    val sharedPlayer = remember(accessToken) { ExoPlayer.Builder(context).build() }
    val latestMode by rememberUpdatedState(playbackMode)
    val latestPage by rememberUpdatedState(pagerState.currentPage)
    val latestLastIndex by rememberUpdatedState(items.lastIndex)
    val latestIsFullscreen by rememberUpdatedState(isFullscreen)
    val latestOnFullscreenChange by rememberUpdatedState(onFullscreenChange)

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    LaunchedEffect(isFullscreen) {
        latestOnFullscreenChange(isFullscreen)
    }

    DisposableEffect(Unit) {
        onDispose { latestOnFullscreenChange(false) }
    }

    LaunchedEffect(playbackMode, isFullscreen) {
        if (!isFullscreen) {
            sharedPlayer.repeatMode = playbackMode.toPlayerRepeatMode()
        }
    }

    DisposableEffect(sharedPlayer) {
        val listener = object : Player.Listener {
            override fun onRenderedFirstFrame() {
                renderedVideoId = sharedPlayer.currentMediaItem?.mediaId
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                isPlayerActuallyPlaying = isPlaying
            }

            override fun onEvents(player: Player, events: Player.Events) {
                durationMs = player.duration.takeIf { it > 0L } ?: 0L
                if (!isScrubbingShort) {
                    positionMs = player.currentPosition.coerceAtLeast(0L)
                }
            }

            override fun onPlaybackStateChanged(playbackState: Int) {
                if (
                    playbackState == Player.STATE_ENDED &&
                    !latestIsFullscreen &&
                    latestMode == ShortPlaybackMode.AUTO_NEXT &&
                    latestPage < latestLastIndex
                ) {
                    coroutineScope.launch { pagerState.animateScrollToPage(latestPage + 1) }
                }
            }
        }
        sharedPlayer.addListener(listener)
        onDispose {
            sharedPlayer.removeListener(listener)
            sharedPlayer.release()
        }
    }

    DisposableEffect(lifecycleOwner, sharedPlayer) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> sharedPlayer.pause()
                Lifecycle.Event.ON_RESUME -> sharedPlayer.playWhenReady = true
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    LaunchedEffect(pagerState.currentPage, items.size) {
        if (pagerState.currentPage >= items.lastIndex - 5) onNeedMore(pagerState.currentPage)
    }

    val currentItem = items.getOrNull(pagerState.currentPage)
    val currentPausedByUser = currentItem?.id?.let { it in pausedByUserVideoIds } ?: false

    LaunchedEffect(currentItem?.id) {
        val videoId = currentItem?.id
        if (!videoId.isNullOrBlank()) {
            onEnsureDetailLoaded(videoId)
        }
    }

    LaunchedEffect(currentItem?.id) {
        if (currentItem == null) {
            durationMs = 0L
            positionMs = 0L
            isScrubbingShort = false
            scrubAnchorMs = 0L
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

    LaunchedEffect(currentItem?.id, baseUrl, dataSourceFactory) {
        val item = currentItem ?: return@LaunchedEffect
        val sourceUrl = UrlBuilder.source(baseUrl, item.id)
        val mediaSource = ProgressiveMediaSource.Factory(dataSourceFactory).createMediaSource(MediaItem.Builder().setUri(sourceUrl).setMediaId(item.id).build())
        sharedPlayer.setMediaSource(mediaSource, true)
        sharedPlayer.prepare()
        sharedPlayer.playWhenReady = !currentPausedByUser
        renderedVideoId = null
    }

    val overlayModifier = if (isFullscreen) {
        Modifier.fillMaxSize().background(Color.Black)
    } else {
        Modifier.fillMaxSize().background(Color.Black).statusBarsPadding()
    }

    Box(modifier = overlayModifier) {
        if (isFullscreen) {
            ShortOverlayFullscreenHost(
                isFullscreen = isFullscreen,
                onFullscreenChange = { isFullscreen = it },
                player = sharedPlayer,
                title = currentItem?.title.orEmpty(),
                subtitleTracks = currentItem?.id?.let { detailByVideoId[it]?.subtitleTracks }.orEmpty(),
                fallbackPlaybackMode = playbackMode,
            )
        } else {
            VerticalPager(
                state = pagerState,
                modifier = Modifier.fillMaxSize(),
            ) { page ->
                val item = items[page]
                val isActive = page == pagerState.currentPage
                val detail = detailByVideoId[item.id]
                val actionBusy = item.id in actionBusyVideoIds || item.id in detailLoadingVideoIds
                val posterUrl = remember(baseUrl, item.thumbnailPath) { resolveThumbnailUrl(baseUrl, item.thumbnailPath) }
                Box(modifier = Modifier.fillMaxSize().pointerInput(item.id) { detectTapGestures(onTap = {
                    showController = !showController
                    if (isActive) {
                        val nextPaused = pausedByUserVideoIds.toMutableSet().apply { if (contains(item.id)) remove(item.id) else add(item.id) }.toSet()
                        pausedByUserVideoIds = nextPaused
                        sharedPlayer.playWhenReady = !nextPaused.contains(item.id)
                    }
                }) }) {
                    if (isActive) {
                        AndroidView(factory = {
                            PlayerView(it).apply {
                                player = sharedPlayer
                                useController = false
                                setShutterBackgroundColor(AndroidColor.BLACK)
                                resizeMode = shortVideoResizeMode(fitMode)
                            }
                        }, update = { view ->
                            view.player = sharedPlayer
                            view.resizeMode = shortVideoResizeMode(fitMode)
                        }, modifier = Modifier.fillMaxSize())
                        if (renderedVideoId != item.id && !posterUrl.isNullOrBlank()) {
                            AsyncImage(model = posterUrl, contentDescription = "${item.title} 封面", contentScale = shortPosterContentScale(fitMode), modifier = Modifier.fillMaxSize())
                        }
                    } else {
                        AsyncImage(model = posterUrl, contentDescription = "${item.title} 封面", contentScale = shortPosterContentScale(fitMode), modifier = Modifier.fillMaxSize())
                    }
                    Box(modifier = Modifier.fillMaxSize().background(Brush.verticalGradient(listOf(Color.Transparent, Color.Transparent, Color(0xB0000000)))))
                    AnimatedVisibility(visible = showController, enter = fadeIn(), exit = fadeOut(), modifier = Modifier.align(Alignment.CenterEnd)) {
                        Column(
                            modifier = Modifier.padding(end = 10.dp),
                            verticalArrangement = Arrangement.spacedBy(12.dp),
                            horizontalAlignment = Alignment.CenterHorizontally,
                        ) {
                            ShortVideoOverlayActionButton(
                                icon = Icons.Filled.ThumbUp,
                                active = detail?.userState?.isLiked == true,
                                enabled = !actionBusy,
                                onClick = { onToggleLike(item.id) },
                                contentDescription = "喜欢",
                            )
                            ShortVideoOverlayActionButton(
                                icon = Icons.Filled.Favorite,
                                active = detail?.userState?.isFavorited == true,
                                enabled = !actionBusy,
                                onClick = { onToggleFavorite(item.id) },
                                contentDescription = "收藏",
                            )
                            ShortVideoOverlayActionButton(icon = Icons.Filled.AspectRatio, active = false, enabled = true, onClick = onToggleFitMode, contentDescription = if (fitMode == VideoFitMode.FILL) "切换完整显示" else "切换铺满显示")
                            ShortPlaybackModeToggleButton(playbackMode = playbackMode, onClick = onTogglePlaybackMode)
                            ShortOverlayFullscreenButton(onClick = { isFullscreen = true })
                        }
                    }
                    Text(text = item.title, color = Color.White, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.Bold, maxLines = 2, overflow = TextOverflow.Ellipsis, modifier = Modifier.align(Alignment.BottomStart).navigationBarsPadding().padding(horizontal = 16.dp).padding(bottom = 24.dp))
                }
            }

            Text(text = shortPlaybackModeLabel(playbackMode), color = Color.White, style = MaterialTheme.typography.bodySmall, modifier = Modifier.align(Alignment.TopEnd).padding(12.dp))
            Text(text = "关闭", color = Color.White, modifier = Modifier.align(Alignment.TopStart).padding(12.dp).clip(RoundedCornerShape(8.dp)).clickable(onClick = onClose).padding(horizontal = 10.dp, vertical = 6.dp))

            if (shouldShowShortOverlayProgressBar(currentItem?.id)) {
                ShortVideoBottomProgressBar(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .fillMaxWidth()
                        .shortNonHomeProgressBarPadding(),
                    positionMs = positionMs,
                    durationMs = durationMs,
                    isScrubbing = isScrubbingShort,
                    scrubTargetMs = scrubTargetMs,
                    onScrubStart = {
                        if (durationMs <= 0L) {
                            return@ShortVideoBottomProgressBar
                        }
                        isScrubbingShort = true
                        scrubAnchorMs = positionMs.coerceIn(0L, durationMs)
                        scrubTargetMs = scrubAnchorMs
                    },
                    onScrubByDeltaFraction = { deltaFraction ->
                        if (durationMs <= 0L || !isScrubbingShort) {
                            return@ShortVideoBottomProgressBar
                        }
                        scrubTargetMs = shortScrubTargetFromDelta(
                            anchorMs = scrubAnchorMs,
                            durationMs = durationMs,
                            deltaFraction = deltaFraction,
                        )
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

private fun resolveThumbnailUrl(baseUrl: String, rawPath: String?): String? {
    val path = rawPath?.trim().orEmpty()
    if (path.isBlank()) return null
    if (path.startsWith("http://") || path.startsWith("https://")) return path
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) return null
    return if (path.startsWith('/')) "$normalizedBase$path" else "$normalizedBase/$path"
}
