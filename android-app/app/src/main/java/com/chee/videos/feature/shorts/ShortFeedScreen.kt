package com.chee.videos.feature.shorts

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.pager.VerticalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
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
import com.chee.videos.core.util.UrlBuilder

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
    onOpenDetail: (String) -> Unit,
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current

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

    LaunchedEffect(active) {
        player.playWhenReady = active
        if (active) {
            player.play()
        } else {
            player.pause()
        }
    }

    DisposableEffect(player, lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> player.pause()
                Lifecycle.Event.ON_RESUME -> if (active) player.play()
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
            player.release()
        }
    }

    Box(modifier = Modifier.fillMaxSize().background(Color.Black)) {
        AndroidView(
            factory = {
                PlayerView(it).apply {
                    this.player = player
                    useController = false
                    resizeMode = AspectRatioFrameLayout.RESIZE_MODE_ZOOM
                    setShutterBackgroundColor(android.graphics.Color.BLACK)
                }
            },
            modifier = Modifier.fillMaxSize(),
        )

        Column(
            modifier = Modifier
                .fillMaxWidth()
                .align(Alignment.BottomStart)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Text(
                text = item.title,
                style = MaterialTheme.typography.titleMedium,
                color = Color.White,
            )
            Row(horizontalArrangement = Arrangement.spacedBy(16.dp)) {
                Text(
                    text = "查看详情",
                    color = Color.White,
                    modifier = Modifier.clickable { onOpenDetail(item.id) },
                )
                Text(
                    text = "类型：短视频",
                    color = Color.White.copy(alpha = 0.8f),
                )
            }
        }
    }
}
