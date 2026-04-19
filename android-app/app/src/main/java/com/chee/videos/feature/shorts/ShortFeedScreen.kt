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
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
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
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import androidx.compose.ui.layout.ContentScale
import coil.compose.AsyncImage
import coil.imageLoader
import coil.request.ImageRequest
import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.util.UrlBuilder
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

private val ShortsActionBg = Color(0x5A000000)
private val ShortsActionActiveTint = Color(0xFFFF5E84)

@Composable
@OptIn(ExperimentalFoundationApi::class)
fun ShortFeedScreen(
    baseUrl: String,
    accessToken: String,
    viewModel: ShortFeedViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.load()
    }

    when {
        uiState.loading && uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize().background(Color.Black), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = Color.White)
            }
        }

        !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize().background(Color.Black), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                    Button(onClick = { viewModel.load(force = true) }) { Text("重试") }
                }
            }
        }

        uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize().background(Color.Black), contentAlignment = Alignment.Center) {
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

                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(Color.Black),
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

                    if (sheetVideoId != null) {
                        val detail = uiState.detailByVideoId[sheetVideoId]
                        val loading = detail == null && sheetVideoId in uiState.detailLoadingVideoIds
                        ShortDetailSheet(
                            modifier = Modifier
                                .align(Alignment.BottomCenter)
                                .fillMaxWidth()
                                .fillMaxHeight(0.75f),
                            detail = detail,
                            loading = loading,
                            errorMessage = uiState.detailErrorMessage,
                            actionBusy = sheetVideoId in uiState.actionBusyVideoIds,
                            onClose = viewModel::closeDetailSheet,
                            onRetry = { viewModel.ensureDetailLoaded(sheetVideoId, force = true, reportError = true) },
                            onToggleDislike = { viewModel.toggleDislike(sheetVideoId) },
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

    Box(modifier = Modifier.fillMaxSize().background(Color.Black)) {
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
                            player = sharedPlayer
                        }
                    },
                    update = { view ->
                        view.player = sharedPlayer
                        view.resizeMode = when (fitMode) {
                            VideoFitMode.FILL -> AspectRatioFrameLayout.RESIZE_MODE_ZOOM
                            VideoFitMode.FIT -> AspectRatioFrameLayout.RESIZE_MODE_FIT
                        }
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
                } else {
                    if (showPoster) {
                        Box(modifier = Modifier.fillMaxSize().background(Color.Black))
                    }
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
        }

        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(
                            Color.Transparent,
                            Color.Transparent,
                            Color(0xAA000000),
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
                    .background(Color(0x99000000))
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

        Column(
            modifier = Modifier
                .align(Alignment.CenterEnd)
                .padding(end = 10.dp),
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

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomStart)
                .padding(horizontal = 16.dp, vertical = 20.dp),
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
                    color = Color.White.copy(alpha = 0.86f),
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
    detail: VideoDetailDto?,
    loading: Boolean,
    errorMessage: String?,
    actionBusy: Boolean,
    onClose: () -> Unit,
    onRetry: () -> Unit,
    onToggleDislike: () -> Unit,
) {
    Surface(
        modifier = modifier,
        color = Color(0xFF101114),
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
                    .padding(horizontal = 10.dp, vertical = 8.dp),
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
                    Column(
                        modifier = Modifier
                            .fillMaxSize()
                            .verticalScroll(rememberScrollState())
                            .padding(horizontal = 16.dp, vertical = 8.dp),
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                        Text(
                            text = detail.title,
                            color = Color.White,
                            style = MaterialTheme.typography.headlineSmall,
                            fontWeight = FontWeight.Bold,
                        )

                        if (!detail.description.isNullOrBlank()) {
                            Text(
                                text = detail.description,
                                color = Color(0xFFDDDDDD),
                                style = MaterialTheme.typography.bodyMedium,
                            )
                        }

                        Text(
                            text = "播放 ${detail.viewsCount}  点赞 ${detail.likesCount}  收藏 ${detail.favoritesCount}",
                            color = Color(0xFFB8B8B8),
                            style = MaterialTheme.typography.bodySmall,
                        )

                        if (actorNames.isNotEmpty()) {
                            Text(
                                text = "演员：${actorNames.joinToString(" / ")}",
                                color = Color.White,
                                style = MaterialTheme.typography.bodyMedium,
                            )
                        }

                        if (detail.collections.orEmpty().isNotEmpty()) {
                            Text(
                                text = "关联合集",
                                color = Color.White,
                                style = MaterialTheme.typography.titleMedium,
                                fontWeight = FontWeight.SemiBold,
                            )
                            FlowRow(
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                                verticalArrangement = Arrangement.spacedBy(8.dp),
                            ) {
                                detail.collections.orEmpty().forEach { collection ->
                                    AssistChip(
                                        onClick = {},
                                        enabled = false,
                                        label = { Text(collection.name) },
                                        colors = AssistChipDefaults.assistChipColors(
                                            disabledContainerColor = Color(0xFF22242A),
                                            disabledLabelColor = Color(0xFFE7E7E7),
                                        ),
                                    )
                                }
                            }
                        }

                        if (detail.tags.orEmpty().isNotEmpty()) {
                            FlowRow(
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                                verticalArrangement = Arrangement.spacedBy(8.dp),
                            ) {
                                detail.tags.orEmpty().forEach { tag ->
                                    AssistChip(
                                        onClick = {},
                                        enabled = false,
                                        label = { Text(tag) },
                                        colors = AssistChipDefaults.assistChipColors(
                                            disabledContainerColor = Color(0xFF1C1E23),
                                            disabledLabelColor = Color(0xFFBDBDBD),
                                        ),
                                    )
                                }
                            }
                        }

                        Button(
                            onClick = onToggleDislike,
                            enabled = !actionBusy,
                            modifier = Modifier.fillMaxWidth(),
                            colors = ButtonDefaults.buttonColors(
                                containerColor = if (dislikeActive) Color(0xFF5A0E1B) else Color(0xFF2B2E36),
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
private fun ShortsActionButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    active: Boolean,
    enabled: Boolean,
    onClick: () -> Unit,
    contentDescription: String,
) {
    Box(
        modifier = Modifier
            .size(44.dp)
            .clip(CircleShape)
            .background(ShortsActionBg)
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            tint = if (active) ShortsActionActiveTint else Color.White,
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

private fun readWatchSnapshot(player: ExoPlayer): Pair<Int, Boolean> {
    val watchSeconds = (player.currentPosition.coerceAtLeast(0L) / 1000L).toInt()
    val durationMs = player.duration
    val durationSeconds = if (durationMs > 0) (durationMs / 1000L).toInt() else 0
    val completedThreshold = (durationSeconds - 3).coerceAtLeast(1)
    val completed = durationSeconds > 0 && watchSeconds >= completedThreshold
    return watchSeconds to completed
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
