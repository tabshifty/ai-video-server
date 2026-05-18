package com.chee.videos.feature.tv

import android.app.Activity
import android.graphics.Color as AndroidColor
import android.view.KeyEvent as AndroidKeyEvent
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.tvFocusableGlow

@Composable
fun TvIptvScreen(
    onBack: () -> Unit,
    viewModel: TvIptvViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val activity = context as? Activity
    val rootFocusRequester = remember { FocusRequester() }
    val exoPlayer = remember { ExoPlayer.Builder(context).build() }
    val dataSourceFactory = remember {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true)
    }
    var channelListVisible by remember { mutableStateOf(false) }
    var focusedChannelIndex by remember { mutableIntStateOf(0) }
    var preparedChannelId by remember { mutableStateOf<String?>(null) }
    var playbackRetryNonce by remember { mutableIntStateOf(0) }
    var playerErrorMessage by remember { mutableStateOf<String?>(null) }

    fun closeOrExit() {
        if (channelListVisible) {
            channelListVisible = false
        } else {
            onBack()
        }
    }

    fun showChannelList() {
        val currentIndex = uiState.channels.indexOfFirst { it.id == uiState.currentChannel?.id }
        focusedChannelIndex = currentIndex.takeIf { it >= 0 } ?: 0
        channelListVisible = true
    }

    BackHandler(onBack = ::closeOrExit)

    DisposableEffect(activity) {
        if (activity == null) {
            onDispose { }
        } else {
            val window = activity.window
            val controller = WindowCompat.getInsetsController(window, window.decorView)
            controller.hide(WindowInsetsCompat.Type.systemBars())
            controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
            onDispose { controller.show(WindowInsetsCompat.Type.systemBars()) }
        }
    }

    DisposableEffect(exoPlayer) {
        val listener = object : Player.Listener {
            override fun onPlayerError(error: PlaybackException) {
                playerErrorMessage = "频道播放失败，请切换频道或重试"
            }
        }
        exoPlayer.addListener(listener)
        onDispose {
            exoPlayer.removeListener(listener)
            exoPlayer.release()
        }
    }

    LaunchedEffect(Unit) {
        rootFocusRequester.requestFocus()
    }

    LaunchedEffect(uiState.currentChannel?.id, playbackRetryNonce) {
        val channel = uiState.currentChannel
        if (channel == null) {
            exoPlayer.clearMediaItems()
            preparedChannelId = null
            return@LaunchedEffect
        }
        playerErrorMessage = null
        val mediaItem = MediaItem.Builder()
            .setUri(channel.url)
            .setMediaId(channel.id)
            .setMediaMetadata(
                androidx.media3.common.MediaMetadata.Builder()
                    .setTitle(channel.name)
                    .build(),
            )
            .build()
        val mediaSource = DefaultMediaSourceFactory(dataSourceFactory).createMediaSource(mediaItem)
        exoPlayer.setMediaSource(mediaSource, true)
        exoPlayer.prepare()
        exoPlayer.playWhenReady = true
        exoPlayer.play()
        preparedChannelId = channel.id
    }

    KeepScreenOnEffect(enabled = uiState.currentChannel != null && playerErrorMessage == null)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black)
            .focusRequester(rootFocusRequester)
            .focusable()
            .onPreviewKeyEvent { event ->
                if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                if (channelListVisible) {
                    when (nativeKeyToIptvRemoteKey(event.nativeKeyEvent.keyCode)) {
                        TvIptvRemoteKey.Up -> {
                            focusedChannelIndex = Math.floorMod(focusedChannelIndex - 1, uiState.channels.size.coerceAtLeast(1))
                            true
                        }

                        TvIptvRemoteKey.Down -> {
                            focusedChannelIndex = Math.floorMod(focusedChannelIndex + 1, uiState.channels.size.coerceAtLeast(1))
                            true
                        }

                        TvIptvRemoteKey.Ok -> {
                            uiState.channels.getOrNull(focusedChannelIndex)?.let { viewModel.selectChannel(it.id) }
                            channelListVisible = false
                            true
                        }

                        TvIptvRemoteKey.Back -> {
                            channelListVisible = false
                            true
                        }

                        else -> false
                    }
                } else {
                    when (nativeKeyToIptvRemoteKey(event.nativeKeyEvent.keyCode)) {
                        TvIptvRemoteKey.Up -> {
                            viewModel.stepChannel(-1)
                            true
                        }

                        TvIptvRemoteKey.Down -> {
                            viewModel.stepChannel(1)
                            true
                        }

                        TvIptvRemoteKey.Right -> {
                            showChannelList()
                            true
                        }

                        TvIptvRemoteKey.Back -> {
                            onBack()
                            true
                        }

                        else -> false
                    }
                }
            },
    ) {
        AndroidView(
            factory = { viewContext ->
                PlayerView(viewContext).apply {
                    useController = false
                    setShutterBackgroundColor(AndroidColor.BLACK)
                    setKeepContentOnPlayerReset(true)
                    resizeMode = AspectRatioFrameLayout.RESIZE_MODE_FIT
                    player = exoPlayer
                }
            },
            update = { view -> view.player = exoPlayer },
            modifier = Modifier.fillMaxSize(),
        )

        TvIptvTopOverlay(channel = uiState.currentChannel)

        if (uiState.loading) {
            TvIptvStatusOverlay(message = "正在加载 IPTV 频道", retryLabel = null, onRetry = {})
        } else if (uiState.currentChannel == null || uiState.statusMessage != null) {
            TvIptvStatusOverlay(
                message = uiState.statusMessage ?: "暂无可播放的 IPTV 频道",
                retryLabel = "重试",
                onRetry = viewModel::reload,
            )
        } else if (playerErrorMessage != null) {
            TvIptvStatusOverlay(
                message = playerErrorMessage.orEmpty(),
                retryLabel = "重试播放",
                onRetry = { playbackRetryNonce += 1 },
            )
        }

        if (channelListVisible) {
            TvIptvChannelListOverlay(
                groups = uiState.groups,
                focusedChannelId = uiState.channels.getOrNull(focusedChannelIndex)?.id,
                currentChannelId = uiState.currentChannel?.id,
                onSelect = { channel ->
                    viewModel.selectChannel(channel.id)
                    channelListVisible = false
                },
                modifier = Modifier.align(Alignment.CenterEnd),
            )
        }
    }
}

@Composable
private fun TvIptvTopOverlay(channel: TvIptvChannelUiModel?) {
    Surface(
        color = Color(0x8C0D1016),
        shape = RoundedCornerShape(18.dp),
        modifier = Modifier
            .padding(18.dp)
            .fillMaxWidth(),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Icon(Icons.Filled.Tv, contentDescription = null, tint = AppChrome.AccentWarm, modifier = Modifier.size(22.dp))
            Text(
                text = channel?.name ?: "IPTV",
                color = Color.White,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f),
            )
            channel?.group?.trim()?.takeIf { it.isNotBlank() }?.let { group ->
                Text(text = group, color = AppChrome.TextSecondary, style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}

@Composable
private fun TvIptvStatusOverlay(
    message: String,
    retryLabel: String?,
    onRetry: () -> Unit,
) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Surface(color = Color(0xD610131A), shape = RoundedCornerShape(18.dp)) {
            Column(
                modifier = Modifier.padding(horizontal = 26.dp, vertical = 22.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                if (retryLabel == null) {
                    CircularProgressIndicator(color = AppChrome.AccentStrong)
                }
                Text(
                    text = message,
                    color = Color.White,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                )
                if (retryLabel != null) {
                    Surface(
                        color = AppChrome.Accent,
                        shape = RoundedCornerShape(8.dp),
                        modifier = Modifier
                            .tvFocusableGlow(shape = RoundedCornerShape(8.dp), focusedScale = 1.04f)
                            .clickable(onClick = onRetry),
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 16.dp, vertical = 10.dp),
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                            verticalAlignment = Alignment.CenterVertically,
                        ) {
                            Icon(Icons.Filled.Refresh, contentDescription = null, tint = Color.White, modifier = Modifier.size(18.dp))
                            Text(text = retryLabel, color = Color.White, style = MaterialTheme.typography.bodyMedium)
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TvIptvChannelListOverlay(
    groups: List<TvIptvChannelGroupUiModel>,
    focusedChannelId: String?,
    currentChannelId: String?,
    onSelect: (TvIptvChannelUiModel) -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = Color(0xE80B0F17),
        shape = RoundedCornerShape(topStart = 18.dp, bottomStart = 18.dp),
        modifier = modifier
            .fillMaxHeight()
            .width(360.dp),
    ) {
        LazyColumn(
            contentPadding = PaddingValues(start = 14.dp, end = 14.dp, top = 18.dp, bottom = 18.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            item("title") {
                Text(
                    text = "频道列表",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.padding(bottom = 8.dp),
                )
            }
            groups.forEach { group ->
                item("group-${group.group}") {
                    Text(
                        text = group.group,
                        color = AppChrome.AccentWarm,
                        style = MaterialTheme.typography.labelLarge,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(top = 8.dp, bottom = 2.dp),
                    )
                }
                items(group.channels, key = { it.id }) { channel ->
                    TvIptvChannelRow(
                        channel = channel,
                        focused = channel.id == focusedChannelId,
                        current = channel.id == currentChannelId,
                        onClick = { onSelect(channel) },
                    )
                }
            }
        }
    }
}

@Composable
private fun TvIptvChannelRow(
    channel: TvIptvChannelUiModel,
    focused: Boolean,
    current: Boolean,
    onClick: () -> Unit,
) {
    val rowColor = when {
        current -> AppChrome.Accent.copy(alpha = 0.92f)
        focused -> AppChrome.SurfaceElevated.copy(alpha = 0.95f)
        else -> Color.Transparent
    }
    Surface(
        color = rowColor,
        shape = RoundedCornerShape(8.dp),
        modifier = Modifier
            .fillMaxWidth()
            .height(48.dp)
            .clickable(onClick = onClick),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Icon(Icons.Filled.Tv, contentDescription = null, tint = if (current) Color.White else AppChrome.TextMuted, modifier = Modifier.size(18.dp))
            Text(
                text = channel.name,
                color = if (current) Color.White else AppChrome.TextPrimary,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = if (current) FontWeight.SemiBold else FontWeight.Normal,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f),
            )
            if (current) {
                Text(text = "播放中", color = Color.White.copy(alpha = 0.78f), style = MaterialTheme.typography.labelSmall)
            }
            Spacer(modifier = Modifier.width(0.dp))
        }
    }
}

private fun nativeKeyToIptvRemoteKey(keyCode: Int): TvIptvRemoteKey? =
    when (keyCode) {
        AndroidKeyEvent.KEYCODE_DPAD_UP -> TvIptvRemoteKey.Up
        AndroidKeyEvent.KEYCODE_DPAD_DOWN -> TvIptvRemoteKey.Down
        AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> TvIptvRemoteKey.Right
        AndroidKeyEvent.KEYCODE_BACK -> TvIptvRemoteKey.Back
        AndroidKeyEvent.KEYCODE_DPAD_CENTER,
        AndroidKeyEvent.KEYCODE_ENTER,
        AndroidKeyEvent.KEYCODE_NUMPAD_ENTER,
        -> TvIptvRemoteKey.Ok

        else -> null
    }
