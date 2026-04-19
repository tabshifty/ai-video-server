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
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Fullscreen
import androidx.compose.material.icons.filled.FullscreenExit
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Button
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
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
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.util.UrlBuilder

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun DetailScreen(
    onBack: () -> Unit,
    viewModel: DetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val isAv = uiState.videoType.equals("av", ignoreCase = true)
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
            if (!isFullscreen) {
                CenterAlignedTopAppBar(
                    title = { Text("视频详情") },
                    modifier = Modifier.statusBarsPadding(),
                    windowInsets = WindowInsets(0, 0, 0, 0),
                    navigationIcon = {
                        IconButton(onClick = onBack) {
                            Icon(Icons.Default.ArrowBack, contentDescription = "返回")
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
                val playUrl = resolvePlayUrl(uiState.baseUrl, detail)
                val posterUrl = resolvePosterUrl(uiState.baseUrl, detail)
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
                var userRequestedPlay by rememberSaveable(detail.id) { mutableStateOf(false) }
                var preparedUrl by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
                var isPlayerActuallyPlaying by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }

                LaunchedEffect(detail.id) {
                    isFullscreen = false
                }

                LaunchedEffect(userRequestedPlay, playUrl, dataSourceFactory) {
                    if (!userRequestedPlay) {
                        exoPlayer.pause()
                        return@LaunchedEffect
                    }
                    if (playUrl.isNullOrBlank()) {
                        return@LaunchedEffect
                    }

                    if (preparedUrl != playUrl) {
                        exoPlayer.stop()
                        exoPlayer.clearMediaItems()
                        val mediaItem = MediaItem.fromUri(playUrl)
                        val mediaSource = ProgressiveMediaSource.Factory(dataSourceFactory).createMediaSource(mediaItem)
                        exoPlayer.setMediaSource(mediaSource, true)
                        exoPlayer.prepare()
                        preparedUrl = playUrl
                    }

                    exoPlayer.playWhenReady = true
                    exoPlayer.play()
                }

                DisposableEffect(lifecycleOwner, exoPlayer, userRequestedPlay) {
                    val observer = LifecycleEventObserver { _, event ->
                        when (event) {
                            Lifecycle.Event.ON_PAUSE -> exoPlayer.pause()
                            Lifecycle.Event.ON_RESUME -> {
                                if (userRequestedPlay) {
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
                val showPlayer = userRequestedPlay && canPlay

                if (isFullscreen) {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(innerPadding)
                            .background(Color.Black),
                    ) {
                        if (showPlayer) {
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
                        } else {
                            Text(
                                text = "暂无可播放视频",
                                color = Color(0xFFEDEFF4),
                                modifier = Modifier.align(Alignment.Center),
                            )
                        }
                    }
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
                                .clip(RoundedCornerShape(14.dp))
                                .background(if (isAv) Color(0xFF181B21) else Color(0xFF1A1C20)),
                        ) {
                            if (showPlayer) {
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
                                    onClick = { userRequestedPlay = true },
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
                        Text("时长：${if (isAv) formatDurationHms(detail.duration) else "${detail.duration} 秒"}", color = mutedTextColor)
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
    icon: ImageVector,
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

private fun resolvePlayUrl(baseUrl: String, detail: VideoDetailDto): String? {
    val raw = detail.playUrl?.trim().orEmpty()
    if (raw.isNotBlank()) {
        return resolveResourceUrl(baseUrl, raw)
    }
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) {
        return null
    }
    return UrlBuilder.source(normalizedBase, detail.id)
}

private fun resolvePosterUrl(baseUrl: String, detail: VideoDetailDto): String? {
    val posterUrl = anyString(detail.metadata?.get("poster_url"))
    val posterPath = anyString(detail.metadata?.get("poster_path"))
    val candidates = listOf(posterUrl, posterPath, detail.thumbnailPath)
    for (candidate in candidates) {
        val resolved = resolveResourceUrl(baseUrl, candidate)
        if (!resolved.isNullOrBlank()) {
            return resolved
        }
    }
    return null
}

private fun resolveResourceUrl(baseUrl: String, raw: String?): String? {
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

private fun anyString(value: Any?): String? {
    return (value as? String)?.trim()?.takeIf { it.isNotBlank() }
}

private fun formatDurationHms(totalSeconds: Int): String {
    val safeSeconds = totalSeconds.coerceAtLeast(0)
    val hours = safeSeconds / 3600
    val minutes = (safeSeconds % 3600) / 60
    val seconds = safeSeconds % 60
    return String.format("%02d:%02d:%02d", hours, minutes, seconds)
}
