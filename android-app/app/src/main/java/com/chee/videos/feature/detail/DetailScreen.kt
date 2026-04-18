package com.chee.videos.feature.detail

import android.graphics.Color as AndroidColor
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
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
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
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
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
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.util.UrlBuilder

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun DetailScreen(
    onBack: () -> Unit,
    onOpenFullscreen: (String) -> Unit = {},
    viewModel: DetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val isAv = uiState.videoType.equals("av", ignoreCase = true)
    val pageBg = if (isAv) Color(0xFF0B0C0F) else Color(0xFFF7F8FA)
    val textColor = if (isAv) Color(0xFFEDEFF4) else Color.Unspecified
    val mutedTextColor = if (isAv) Color(0xFFB9C0CD) else MaterialTheme.colorScheme.onSurfaceVariant

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("视频详情") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "返回")
                    }
                },
            )
        },
        containerColor = pageBg,
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
                val context = LocalContext.current
                val lifecycleOwner = LocalLifecycleOwner.current
                val playUrl = resolvePlayUrl(uiState.baseUrl, detail)
                val posterUrl = resolvePosterUrl(uiState.baseUrl, detail)

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
                var userRequestedPlay by remember(detail.id) { mutableStateOf(false) }
                var preparedUrl by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }

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
                    onDispose {
                        exoPlayer.release()
                    }
                }

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
                        if (!posterUrl.isNullOrBlank()) {
                            AsyncImage(
                                model = posterUrl,
                                contentDescription = "海报",
                                contentScale = ContentScale.Crop,
                                modifier = Modifier.fillMaxSize(),
                            )
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

                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.spacedBy(10.dp),
                    ) {
                        Button(
                            onClick = { userRequestedPlay = !userRequestedPlay },
                            enabled = !playUrl.isNullOrBlank(),
                            modifier = Modifier.weight(1f),
                        ) {
                            Text(if (userRequestedPlay) "暂停播放" else "播放")
                        }
                        Button(
                            onClick = { onOpenFullscreen(detail.id) },
                            enabled = !playUrl.isNullOrBlank(),
                            modifier = Modifier.weight(1f),
                        ) {
                            Text("横屏全屏播放")
                        }
                    }

                    if (userRequestedPlay && !playUrl.isNullOrBlank()) {
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
                            modifier = Modifier
                                .fillMaxWidth()
                                .height(220.dp)
                                .clip(RoundedCornerShape(14.dp)),
                        )
                    }

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
