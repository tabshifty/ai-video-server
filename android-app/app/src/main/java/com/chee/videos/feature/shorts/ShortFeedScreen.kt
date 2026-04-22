package com.chee.videos.feature.shorts

import android.graphics.Color as AndroidColor
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AspectRatio
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Favorite
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.ThumbDown
import androidx.compose.material.icons.filled.ThumbUp
import androidx.compose.material3.AssistChip
import androidx.compose.material3.AssistChipDefaults
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
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
import androidx.compose.runtime.key
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
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
import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.VideoCollectionDto
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoImageCollectionDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.ShortVideoOverlayActionButton
import com.chee.videos.core.ui.ShortVideoBottomProgressBar
import com.chee.videos.core.ui.shortPosterContentScale as sharedShortPosterContentScale
import com.chee.videos.core.ui.shortScrubTargetFromDelta
import com.chee.videos.core.ui.shortVideoResizeMode
import com.chee.videos.core.util.UrlBuilder
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch

private val ShortsActionActiveTint = AppChrome.AccentStrong
private val ShortDetailSheetBg = AppChrome.CanvasRaised
private val ShortDetailCardBg = AppChrome.Surface
private val ShortDetailCardMutedBg = AppChrome.SurfaceElevated
private val ShortDetailPillBg = AppChrome.SurfaceStrong
private const val ProgressRevealDelayMs = 140L

internal fun shortPosterContentScale(fitMode: VideoFitMode): ContentScale {
    return sharedShortPosterContentScale(fitMode)
}

@Composable
@OptIn(ExperimentalFoundationApi::class)
fun ShortFeedScreen(
    baseUrl: String,
    accessToken: String,
    onOpenDiscover: (mode: String, value: String, title: String) -> Unit,
    onOpenImageCollectionViewer: (String) -> Unit,
    viewModel: ShortFeedViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.load()
    }

    when {
        uiState.loading && uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize().background(AppChrome.Canvas), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = Color.White)
            }
        }

        !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize().background(AppChrome.Canvas), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                    Button(onClick = { viewModel.load(force = true) }) { Text("重试") }
                }
            }
        }

        uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize().background(AppChrome.Canvas), contentAlignment = Alignment.Center) {
                Text("暂无短视频", color = Color.White)
            }
        }

        else -> {
            val context = LocalContext.current
            val lifecycleOwner = LocalLifecycleOwner.current
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
            var renderedVideoId by remember { mutableStateOf<String?>(null) }
            var lastHistoryVideoId by remember { mutableStateOf<String?>(null) }
            var isPlayerActuallyPlaying by remember { mutableStateOf(false) }
            val latestCurrentVideoId by rememberUpdatedState(sharedPlayer.currentMediaItem?.mediaId)

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
                    val finalVideoId = latestCurrentVideoId ?: sharedPlayer.currentMediaItem?.mediaId
                    if (!finalVideoId.isNullOrBlank()) {
                        val (watchSeconds, completed) = readWatchSnapshot(sharedPlayer)
                        viewModel.reportHistory(finalVideoId, watchSeconds, completed)
                    }
                    sharedPlayer.removeListener(listener)
                    isPlayerActuallyPlaying = false
                    sharedPlayer.release()
                }
            }

            key(uiState.pagerResetToken) {
                val pagerState = rememberPagerState(
                    initialPage = uiState.pagerInitialPage.coerceIn(0, (uiState.items.lastIndex).coerceAtLeast(0)),
                    pageCount = { uiState.items.size },
                )
                val sheetVideoId = uiState.detailSheetVideoId
                val sheetOpen = sheetVideoId != null
                val videoHeightFraction by animateFloatAsState(
                    targetValue = if (sheetOpen) 0.25f else 1f,
                    animationSpec = spring(stiffness = 450f),
                    label = "short_video_region_height",
                )
                val currentVideoId = uiState.items.getOrNull(pagerState.currentPage)?.id
                val currentVideoPausedByUser = currentVideoId?.let { it in uiState.pausedByUserVideoIds } ?: true
                var durationMs by remember(currentVideoId) { mutableStateOf(0L) }
                var positionMs by remember(currentVideoId) { mutableStateOf(0L) }
                var isScrubbingShort by remember(currentVideoId) { mutableStateOf(false) }
                var scrubAnchorMs by remember(currentVideoId) { mutableStateOf(0L) }
                var scrubTargetMs by remember(currentVideoId) { mutableStateOf(0L) }
                var progressBarSettled by remember(currentVideoId) { mutableStateOf(false) }

                LaunchedEffect(pagerState.currentPage, uiState.items) {
                    uiState.items.getOrNull(pagerState.currentPage)?.id?.let { currentVideoID ->
                        viewModel.ensureDetailLoaded(currentVideoID)
                        viewModel.loadMoreIfNeeded(
                            currentIndex = pagerState.currentPage,
                            currentVideoId = currentVideoID,
                        )
                    }
                }
                LaunchedEffect(pagerState.currentPage, uiState.items, baseUrl) {
                    val page = pagerState.currentPage
                    val start = (page - 2).coerceAtLeast(0)
                    val end = (page + 2).coerceAtMost(uiState.items.lastIndex)
                    for (idx in start..end) {
                        val item = uiState.items.getOrNull(idx) ?: continue
                        val thumbURL = resolveThumbnailUrl(baseUrl, item.thumbnailPath) ?: continue
                        imageLoader.enqueue(ImageRequest.Builder(context).data(thumbURL).build())
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
                    if (currentVideoId == null) {
                        sharedPlayer.stop()
                        sharedPlayer.clearMediaItems()
                        renderedVideoId = null
                        return@LaunchedEffect
                    }
                    renderedVideoId = null
                    sharedPlayer.stop()
                    sharedPlayer.clearMediaItems()
                    val sourceUrl = UrlBuilder.source(baseUrl, currentVideoId)
                    val mediaItem = MediaItem.Builder()
                        .setUri(sourceUrl)
                        .setMediaId(currentVideoId)
                        .build()
                    val mediaSource = ProgressiveMediaSource.Factory(dataSourceFactory)
                        .createMediaSource(mediaItem)
                    sharedPlayer.setMediaSource(mediaSource, true)
                    sharedPlayer.prepare()
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
                    if (currentVideoId == null) {
                        durationMs = 0L
                        positionMs = 0L
                        isScrubbingShort = false
                        scrubAnchorMs = 0L
                        scrubTargetMs = 0L
                        progressBarSettled = false
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
                LaunchedEffect(currentVideoId, pagerState.isScrollInProgress) {
                    if (currentVideoId.isNullOrBlank()) {
                        progressBarSettled = false
                        return@LaunchedEffect
                    }
                    if (pagerState.isScrollInProgress) {
                        progressBarSettled = false
                        return@LaunchedEffect
                    }
                    delay(ProgressRevealDelayMs)
                    progressBarSettled = true
                }
                LaunchedEffect(currentVideoId, currentVideoPausedByUser) {
                    val shouldPlay = currentVideoId != null && !currentVideoPausedByUser
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
                                if (currentVideoId != null && !currentVideoPausedByUser) {
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

                val showProgressBar = shouldShowShortFeedProgressBar(
                    currentVideoId = currentVideoId,
                    detailSheetOpen = sheetOpen,
                    pagerSettled = progressBarSettled,
                )
                val showActionRail = shouldShowShortFeedActionRail(
                    detailSheetOpen = sheetOpen,
                )

                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(AppChrome.Canvas),
                ) {
                    Box(
                        modifier = Modifier
                            .align(Alignment.TopCenter)
                            .fillMaxWidth()
                            .fillMaxHeight(videoHeightFraction),
                    ) {
                        VerticalPager(
                            state = pagerState,
                            modifier = Modifier.fillMaxSize(),
                            userScrollEnabled = !sheetOpen,
                        ) { page ->
                            val item = uiState.items[page]
                            val detail = uiState.detailByVideoId[item.id]
                            VerticalVideoPage(
                                item = item,
                                sharedPlayer = sharedPlayer,
                                active = pagerState.currentPage == page,
                                fitMode = uiState.fitMode,
                                pausedByUser = item.id in uiState.pausedByUserVideoIds,
                                posterUrl = resolveThumbnailUrl(baseUrl, item.thumbnailPath),
                                showPoster = pagerState.currentPage == page && renderedVideoId != item.id,
                                actorNames = extractActorNames(detail),
                                isLiked = detail?.userState?.isLiked == true,
                                isFavorited = detail?.userState?.isFavorited == true,
                                actionBusy = item.id in uiState.actionBusyVideoIds,
                                showActionRail = showActionRail,
                                onTogglePauseByUser = { viewModel.togglePauseByUser(item.id) },
                                onToggleLike = { viewModel.toggleLike(item.id) },
                                onToggleFavorite = { viewModel.toggleFavorite(item.id) },
                                onToggleMode = viewModel::toggleFitMode,
                                onOpenDetail = { viewModel.openDetailSheet(item.id) },
                            )
                        }
                    }

                    if (!uiState.loadMoreErrorMessage.isNullOrBlank()) {
                        Surface(
                            color = Color(0xCC11141A),
                            shape = RoundedCornerShape(18.dp),
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .padding(bottom = if (sheetOpen) 16.dp else 32.dp),
                        ) {
                            Text(
                                text = uiState.loadMoreErrorMessage.orEmpty(),
                                color = Color.White,
                                style = MaterialTheme.typography.bodySmall,
                                modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp),
                            )
                        }
                    }

                    if (showProgressBar) {
                        ShortVideoBottomProgressBar(
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth(),
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

                    if (sheetVideoId != null) {
                        val detail = uiState.detailByVideoId[sheetVideoId]
                        val loading = detail == null && sheetVideoId in uiState.detailLoadingVideoIds
                        ShortDetailSheet(
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth()
                                .fillMaxHeight(0.75f),
                            baseUrl = baseUrl,
                            detail = detail,
                            loading = loading,
                            errorMessage = uiState.detailErrorMessage,
                            actionBusy = sheetVideoId in uiState.actionBusyVideoIds,
                            onClose = viewModel::closeDetailSheet,
                            onRetry = { viewModel.ensureDetailLoaded(sheetVideoId, force = true, reportError = true) },
                            onToggleDislike = { viewModel.toggleDislike(sheetVideoId) },
                            onOpenTag = { tag ->
                                val normalized = tag.trim()
                                if (normalized.isBlank()) {
                                    return@ShortDetailSheet
                                }
                                viewModel.closeDetailSheet()
                                onOpenDiscover("tag", normalized, "#$normalized")
                            },
                            onOpenCollection = { collection ->
                                val id = collection.id.trim()
                                if (id.isBlank()) {
                                    return@ShortDetailSheet
                                }
                                viewModel.closeDetailSheet()
                                onOpenDiscover("collection", id, collection.name)
                            },
                            onOpenImageCollection = { imageCollection ->
                                val route = shortDetailImageCollectionViewerRoute(imageCollection)
                                    ?: return@ShortDetailSheet
                                onOpenImageCollectionViewer(route)
                            },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun VerticalVideoPage(
    item: FeedVideoDto,
    sharedPlayer: ExoPlayer,
    active: Boolean,
    fitMode: VideoFitMode,
    pausedByUser: Boolean,
    posterUrl: String?,
    showPoster: Boolean,
    actorNames: List<String>,
    isLiked: Boolean,
    isFavorited: Boolean,
    actionBusy: Boolean,
    showActionRail: Boolean,
    onTogglePauseByUser: () -> Boolean,
    onToggleLike: () -> Unit,
    onToggleFavorite: () -> Unit,
    onToggleMode: () -> Unit,
    onOpenDetail: () -> Unit,
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

    Box(modifier = Modifier.fillMaxSize().background(AppChrome.Canvas)) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .pointerInput(item.id, pausedByUser) {
                    detectTapGestures(
                        onTap = {
                            val pausedNow = onTogglePauseByUser()
                            centerIsPause = pausedNow
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
                } else {
                    if (showPoster) {
                        Box(modifier = Modifier.fillMaxSize().background(Color.Black))
                    }
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
        }

        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(
                            Color.Transparent,
                            Color.Transparent,
                            AppChrome.Canvas.copy(alpha = 0.88f),
                        ),
                    ),
                ),
        )

        AnimatedVisibility(
            visible = showCenterIndicator,
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.Center),
        ) {
            Box(
                modifier = Modifier
                    .clip(CircleShape)
                    .background(AppChrome.Surface.copy(alpha = 0.82f))
                    .padding(horizontal = 18.dp, vertical = 14.dp),
                contentAlignment = Alignment.Center,
            ) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
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

        AnimatedVisibility(
            visible = showActionRail,
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.CenterEnd),
        ) {
            Column(
                modifier = Modifier.padding(end = 10.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
            ) {
                ShortsActionButton(
                    icon = Icons.Filled.ThumbUp,
                    active = isLiked,
                    enabled = !actionBusy,
                    onClick = onToggleLike,
                    contentDescription = "喜欢",
                )
                ShortsActionButton(
                    icon = Icons.Filled.Favorite,
                    active = isFavorited,
                    enabled = !actionBusy,
                    onClick = onToggleFavorite,
                    contentDescription = "收藏",
                )
                ShortsActionButton(
                    icon = Icons.Filled.AspectRatio,
                    active = false,
                    enabled = true,
                    onClick = onToggleMode,
                    contentDescription = if (fitMode == VideoFitMode.FILL) "切换完整显示" else "切换铺满显示",
                )
                ShortsActionButton(
                    icon = Icons.Filled.Info,
                    active = false,
                    enabled = true,
                    onClick = onOpenDetail,
                    contentDescription = "详情",
                )
            }
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomStart)
                .padding(horizontal = 16.dp, vertical = 24.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text(
                text = item.title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = Color.White,
            )
            if (actorNames.isNotEmpty()) {
                Text(
                    text = "演员：${actorNames.joinToString(" / ")}",
                    color = AppChrome.TextSecondary,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@OptIn(ExperimentalLayoutApi::class)
@Composable
private fun ShortDetailSheet(
    modifier: Modifier = Modifier,
    baseUrl: String,
    detail: VideoDetailDto?,
    loading: Boolean,
    errorMessage: String?,
    actionBusy: Boolean,
    onClose: () -> Unit,
    onRetry: () -> Unit,
    onToggleDislike: () -> Unit,
    onOpenTag: (String) -> Unit,
    onOpenCollection: (VideoCollectionDto) -> Unit,
    onOpenImageCollection: (VideoImageCollectionDto) -> Unit,
) {
    Surface(
        modifier = modifier,
        color = ShortDetailSheetBg,
        shape = RoundedCornerShape(topStart = 22.dp, topEnd = 22.dp),
        tonalElevation = 0.dp,
        shadowElevation = 10.dp,
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            Box(
                modifier = Modifier
                    .padding(top = 8.dp)
                    .align(Alignment.CenterHorizontally)
                    .size(width = 36.dp, height = 4.dp)
                    .clip(CircleShape)
                    .background(Color(0x55FFFFFF)),
            )
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = 6.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Text(
                    text = "详情",
                    color = Color.White,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                )
                IconButton(onClick = onClose) {
                    Icon(imageVector = Icons.Filled.Close, contentDescription = "关闭", tint = Color.White)
                }
            }

            when {
                loading -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator(color = Color.White)
                    }
                }

                detail == null -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                            Text(errorMessage ?: "详情加载失败", color = MaterialTheme.colorScheme.error)
                            Button(onClick = onRetry) { Text("重试") }
                        }
                    }
                }

                else -> {
                    val actorNames = extractActorNames(detail)
                    val dislikeActive = detail.userState.isDisliked
                    val metaLine = buildList {
                        if (actorNames.isNotEmpty()) {
                            add("演员 ${actorNames.joinToString(" / ")}")
                        }
                        if (detail.duration > 0) {
                            add("时长 ${formatShortPlaybackTimeHms(detail.duration.toLong() * 1000L)}")
                        }
                    }.joinToString(" · ")
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .verticalScroll(rememberScrollState())
                            .padding(horizontal = 16.dp, vertical = 12.dp),
                        verticalArrangement = Arrangement.spacedBy(16.dp),
                    ) {
                        Surface(
                            color = ShortDetailCardBg,
                            shape = AppChrome.CardShape,
                        ) {
                            Column(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = 18.dp, vertical = 18.dp),
                                verticalArrangement = Arrangement.spacedBy(10.dp),
                            ) {
                                Text(
                                    text = detail.title,
                                    color = Color.White,
                                    style = MaterialTheme.typography.headlineSmall,
                                    fontWeight = FontWeight.Bold,
                                )
                                if (metaLine.isNotBlank()) {
                                    Text(
                                        text = metaLine,
                                        color = AppChrome.TextSecondary,
                                        style = MaterialTheme.typography.bodySmall,
                                    )
                                }
                            }
                        }

                        Surface(
                            color = ShortDetailCardBg,
                            shape = AppChrome.CardShape,
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = 12.dp, vertical = 12.dp),
                                horizontalArrangement = Arrangement.spacedBy(10.dp),
                            ) {
                                ShortDetailMetricCard(
                                    modifier = Modifier.weight(1f),
                                    label = "播放",
                                    value = detail.viewsCount.toString(),
                                )
                                ShortDetailMetricCard(
                                    modifier = Modifier.weight(1f),
                                    label = "点赞",
                                    value = detail.likesCount.toString(),
                                )
                                ShortDetailMetricCard(
                                    modifier = Modifier.weight(1f),
                                    label = "收藏",
                                    value = detail.favoritesCount.toString(),
                                )
                            }
                        }

                        ShortDetailInfoCard(
                            title = "简介",
                            body = detail.description.orEmpty().ifBlank { "暂无简介" },
                        )

                        val imageCollection = detail.imageCollection
                        if (shouldShowShortDetailImageCollection(imageCollection)) {
                            Surface(
                                color = ShortDetailCardBg,
                                shape = AppChrome.CardShape,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clip(AppChrome.CardShape)
                                    .clickable { onOpenImageCollection(checkNotNull(imageCollection)) },
                            ) {
                                Row(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(horizontal = 16.dp, vertical = 16.dp),
                                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                ) {
                                    val coverUrl = resolveThumbnailUrl(baseUrl, imageCollection?.coverUrl)
                                    if (!coverUrl.isNullOrBlank()) {
                                        AsyncImage(
                                            model = coverUrl,
                                            contentDescription = imageCollection?.name,
                                            modifier = Modifier
                                                .size(64.dp)
                                                .clip(RoundedCornerShape(16.dp)),
                                            contentScale = ContentScale.Crop,
                                        )
                                    } else {
                                        Box(
                                            modifier = Modifier
                                                .size(64.dp)
                                                .clip(RoundedCornerShape(16.dp))
                                                .background(ShortDetailCardMutedBg),
                                            contentAlignment = Alignment.Center,
                                        ) {
                                            Icon(
                                                imageVector = Icons.Filled.Info,
                                                contentDescription = null,
                                                tint = AppChrome.TextMuted,
                                            )
                                        }
                                    }
                                    Column(
                                        modifier = Modifier.weight(1f),
                                        verticalArrangement = Arrangement.spacedBy(4.dp),
                                    ) {
                                        ShortDetailSectionTitle("相关图片")
                                        Text(
                                            text = imageCollection?.name.orEmpty(),
                                            color = Color.White,
                                            style = MaterialTheme.typography.titleMedium,
                                            fontWeight = FontWeight.SemiBold,
                                        )
                                        Text(
                                            text = "点击进入图片预览",
                                            color = AppChrome.TextMuted,
                                            style = MaterialTheme.typography.bodySmall,
                                        )
                                    }
                                }
                            }
                        }

                        if (detail.collections.orEmpty().isNotEmpty()) {
                            Surface(
                                color = ShortDetailCardBg,
                                shape = AppChrome.CardShape,
                            ) {
                                Column(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(horizontal = 16.dp, vertical = 16.dp),
                                    verticalArrangement = Arrangement.spacedBy(10.dp),
                                ) {
                                    ShortDetailSectionTitle("关联合集")
                                    FlowRow(
                                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                                        verticalArrangement = Arrangement.spacedBy(8.dp),
                                    ) {
                                        detail.collections.orEmpty().forEach { collection ->
                                            ShortDetailPill(
                                                text = collection.name,
                                                emphasized = true,
                                                onClick = { onOpenCollection(collection) },
                                            )
                                        }
                                    }
                                }
                            }
                        }

                        if (detail.tags.orEmpty().isNotEmpty()) {
                            Surface(
                                color = ShortDetailCardBg,
                                shape = AppChrome.CardShape,
                            ) {
                                Column(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(horizontal = 16.dp, vertical = 16.dp),
                                    verticalArrangement = Arrangement.spacedBy(10.dp),
                                ) {
                                    ShortDetailSectionTitle("标签")
                                    FlowRow(
                                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                                        verticalArrangement = Arrangement.spacedBy(8.dp),
                                    ) {
                                        detail.tags.orEmpty().forEach { tag ->
                                            ShortDetailPill(
                                                text = tag,
                                                emphasized = false,
                                                onClick = { onOpenTag(tag) },
                                            )
                                        }
                                    }
                                }
                            }
                        }

                        Button(
                            onClick = onToggleDislike,
                            enabled = !actionBusy,
                            modifier = Modifier.fillMaxWidth(),
                            colors = ButtonDefaults.buttonColors(
                                containerColor = if (dislikeActive) Color(0xFF5A0E1B) else ShortDetailCardMutedBg,
                                contentColor = Color.White,
                            ),
                        ) {
                            Icon(imageVector = Icons.Filled.ThumbDown, contentDescription = null)
                            Text(
                                text = if (dislikeActive) "取消不喜欢" else "不喜欢",
                                modifier = Modifier.padding(start = 8.dp),
                            )
                        }

                        if (!errorMessage.isNullOrBlank()) {
                            Text(
                                text = errorMessage,
                                color = MaterialTheme.colorScheme.error,
                                style = MaterialTheme.typography.bodySmall,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ShortDetailMetricCard(
    modifier: Modifier = Modifier,
    label: String,
    value: String,
) {
    Surface(
        modifier = modifier,
        color = ShortDetailCardMutedBg,
        shape = RoundedCornerShape(18.dp),
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
                color = Color.White,
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
private fun ShortDetailInfoCard(
    title: String,
    body: String,
) {
    Surface(
        color = ShortDetailCardBg,
        shape = AppChrome.CardShape,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            ShortDetailSectionTitle(title)
            Text(
                text = body,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.bodyMedium,
            )
        }
    }
}

@Composable
private fun ShortDetailSectionTitle(title: String) {
    Text(
        text = title,
        color = Color.White,
        style = MaterialTheme.typography.titleSmall,
        fontWeight = FontWeight.SemiBold,
    )
}

@Composable
private fun ShortDetailPill(
    text: String,
    emphasized: Boolean,
    onClick: () -> Unit,
) {
    Box(
        modifier = Modifier
            .clip(AppChrome.PillShape)
            .clickable(onClick = onClick)
            .background(if (emphasized) AppChrome.AccentSoft else ShortDetailPillBg)
            .padding(horizontal = 12.dp, vertical = 10.dp),
    ) {
        Text(
            text = text,
            color = if (emphasized) Color.White else AppChrome.TextSecondary,
            style = MaterialTheme.typography.bodySmall,
            fontWeight = FontWeight.Medium,
        )
    }
}

@Composable
private fun ShortsActionButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    active: Boolean,
    enabled: Boolean,
    onClick: () -> Unit,
    contentDescription: String,
) {
    ShortVideoOverlayActionButton(
        icon = icon,
        active = active,
        enabled = enabled,
        onClick = onClick,
        contentDescription = contentDescription,
    )
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

private fun readWatchSnapshot(player: ExoPlayer): Pair<Int, Boolean> {
    val watchSeconds = (player.currentPosition.coerceAtLeast(0L) / 1000L).toInt()
    val durationMs = player.duration
    val durationSeconds = if (durationMs > 0) (durationMs / 1000L).toInt() else 0
    val completedThreshold = (durationSeconds - 3).coerceAtLeast(1)
    val completed = durationSeconds > 0 && watchSeconds >= completedThreshold
    return watchSeconds to completed
}

internal fun formatShortPlaybackTimeHms(ms: Long): String {
    val totalSeconds = (ms.coerceAtLeast(0L) / 1000L)
    val hours = totalSeconds / 3600L
    val minutes = (totalSeconds % 3600L) / 60L
    val seconds = totalSeconds % 60L
    return buildString {
        append(hours.toString().padStart(2, '0'))
        append(':')
        append(minutes.toString().padStart(2, '0'))
        append(':')
        append(seconds.toString().padStart(2, '0'))
    }
}

internal fun shouldShowShortDetailImageCollection(imageCollection: VideoImageCollectionDto?): Boolean {
    return !imageCollection?.id?.trim().isNullOrBlank()
}

internal fun shortDetailImageCollectionViewerRoute(imageCollection: VideoImageCollectionDto?): String? {
    val collectionId = imageCollection?.id?.trim().orEmpty()
    if (collectionId.isBlank()) {
        return null
    }
    return "image-collections/$collectionId"
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
