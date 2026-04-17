package com.chee.videos.feature.shorts

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AspectRatio
import androidx.compose.material.icons.filled.Info
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
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
import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.util.UrlBuilder
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@Composable
@OptIn(ExperimentalFoundationApi::class)
fun ShortFeedScreen(
    baseUrl: String,
    accessToken: String,
    onOpenDetail: (String) -> Unit,
    viewModel: ShortFeedViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.load()
    }

    when {
        uiState.loading && uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        }

        !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                    Button(onClick = { viewModel.load(force = true) }) { Text("重试") }
                }
            }
        }

        uiState.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("暂无短视频")
            }
        }

        else -> {
            val pagerState = rememberPagerState(pageCount = { uiState.items.size })
            VerticalPager(
                state = pagerState,
                modifier = Modifier.fillMaxSize(),
            ) { page ->
                val item = uiState.items[page]
                val sourceUrl = UrlBuilder.source(baseUrl, item.id)
                VerticalVideoPage(
                    item = item,
                    sourceUrl = sourceUrl,
                    accessToken = accessToken,
                    active = pagerState.currentPage == page,
                    fitMode = uiState.fitMode,
                    pausedByUser = item.id in uiState.pausedByUserVideoIds,
                    onTogglePauseByUser = { viewModel.togglePauseByUser(item.id) },
                    onToggleMode = viewModel::toggleFitMode,
                    onOpenDetail = onOpenDetail,
                )
            }
        }
    }
}

@Composable
private fun VerticalVideoPage(
    item: FeedVideoDto,
    sourceUrl: String,
    accessToken: String,
    active: Boolean,
    fitMode: VideoFitMode,
    pausedByUser: Boolean,
    onTogglePauseByUser: () -> Boolean,
    onToggleMode: () -> Unit,
    onOpenDetail: (String) -> Unit,
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    val scope = rememberCoroutineScope()

    var showCenterIndicator by remember { mutableStateOf(false) }
    var centerIsPause by remember { mutableStateOf(false) }
    var hideIndicatorJob by remember { mutableStateOf<Job?>(null) }

    val player = remember(sourceUrl, accessToken) {
        val dataSourceFactory = DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true)
        if (accessToken.isNotBlank()) {
            dataSourceFactory.setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
        }

        val mediaSource = ProgressiveMediaSource.Factory(dataSourceFactory)
            .createMediaSource(MediaItem.fromUri(sourceUrl))

        ExoPlayer.Builder(context).build().apply {
            setMediaSource(mediaSource)
            repeatMode = Player.REPEAT_MODE_ONE
            prepare()
        }
    }

    LaunchedEffect(active, pausedByUser) {
        val shouldPlay = active && !pausedByUser
        player.playWhenReady = shouldPlay
        if (shouldPlay) {
            player.play()
        } else {
            player.pause()
        }
    }

    DisposableEffect(player, lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> player.pause()
                Lifecycle.Event.ON_RESUME -> if (active && !pausedByUser) player.play()
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
            hideIndicatorJob?.cancel()
            player.release()
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
            AndroidView(
                factory = {
                    PlayerView(it).apply {
                        this.player = player
                        useController = false
                        setShutterBackgroundColor(android.graphics.Color.BLACK)
                    }
                },
                update = { view ->
                    view.resizeMode = when (fitMode) {
                        VideoFitMode.FILL -> AspectRatioFrameLayout.RESIZE_MODE_ZOOM
                        VideoFitMode.FIT -> AspectRatioFrameLayout.RESIZE_MODE_FIT
                    }
                },
                modifier = Modifier.fillMaxSize(),
            )
        }

        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    brush = Brush.verticalGradient(
                        colors = listOf(
                            Color.Transparent,
                            Color.Transparent,
                            Color(0x99000000),
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
            verticalArrangement = Arrangement.spacedBy(14.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            ShortsActionButton(
                icon = Icons.Filled.AspectRatio,
                label = if (fitMode == VideoFitMode.FILL) "铺满" else "完整",
                selected = true,
                onClick = onToggleMode,
            )
            ShortsActionButton(
                icon = Icons.Filled.Info,
                label = "详情",
                onClick = { onOpenDetail(item.id) },
            )
        }

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomStart)
                .padding(horizontal = 16.dp, vertical = 20.dp),
            verticalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            Text(
                text = item.title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
                color = Color.White,
            )
            Text(
                text = "轻触暂停/播放 · 上下滑切换",
                color = Color.White.copy(alpha = 0.82f),
                style = MaterialTheme.typography.bodySmall,
            )
        }
    }
}

@Composable
private fun ShortsActionButton(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    onClick: () -> Unit,
    selected: Boolean = false,
) {
    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(4.dp),
    ) {
        Box(
            modifier = Modifier
                .size(44.dp)
                .clip(CircleShape)
                .background(if (selected) Color(0xCCFE2C55) else Color(0x77000000))
                .clickable(
                    interactionSource = remember { MutableInteractionSource() },
                    indication = null,
                    onClick = onClick,
                ),
            contentAlignment = Alignment.Center,
        ) {
            Icon(icon, contentDescription = null, tint = Color.White)
        }
        Text(
            text = label,
            color = Color.White,
            style = MaterialTheme.typography.labelSmall,
        )
    }
}
