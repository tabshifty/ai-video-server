package com.chee.videos.feature.tv

import android.app.Activity
import android.net.Uri
import android.os.Handler
import android.os.Looper
import android.util.Log
import android.view.LayoutInflater
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
import androidx.compose.foundation.lazy.rememberLazyListState
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
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
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
import coil.compose.AsyncImage
import com.chee.videos.tv.R
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.TvLayoutSpec
import com.chee.videos.core.ui.tvFocusableGlow
import java.util.ArrayList
import kotlinx.coroutines.delay
import org.videolan.libvlc.LibVLC
import org.videolan.libvlc.Media
import org.videolan.libvlc.MediaPlayer
import org.videolan.libvlc.util.VLCVideoLayout

private const val IPTV_LOG_TAG = "TvIptv"

@Composable
fun TvIptvScreen(
    onBack: () -> Unit,
    viewModel: TvIptvViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val currentChannelForLog by rememberUpdatedState(uiState.currentChannel)
    val context = LocalContext.current
    val activity = context as? Activity
    val rootFocusRequester = remember { FocusRequester() }
    val mainHandler = remember { Handler(Looper.getMainLooper()) }
    val libVlc = remember {
        LibVLC(
            context.applicationContext,
            ArrayList(
                listOf(
                    "--network-caching=1500",
                    "--clock-jitter=0",
                    "--clock-synchro=0",
                    "--avcodec-hw=none",
                ),
            ),
        )
    }
    val vlcPlayer = remember(libVlc) { MediaPlayer(libVlc) }
    var channelListVisible by remember { mutableStateOf(false) }
    var focusedChannelIndex by remember { mutableIntStateOf(0) }
    var channelListOpenNonce by remember { mutableIntStateOf(0) }
    var showChannelHint by remember { mutableStateOf(false) }
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
        channelListOpenNonce += 1
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

    DisposableEffect(vlcPlayer) {
        val listener = MediaPlayer.EventListener { event ->
            mainHandler.post {
                Log.i(
                    IPTV_LOG_TAG,
                    "event=${event.type} vout=${event.getVoutCount()} " +
                        "videoTracks=${vlcPlayer.videoTracksCount} audioTracks=${vlcPlayer.audioTracksCount} " +
                        "videoTrack=${vlcPlayer.videoTrack} url=${currentChannelForLog?.url.orEmpty()}",
                )
                when (event.type) {
                    MediaPlayer.Event.EncounteredError -> {
                        playerErrorMessage = "频道播放失败，请切换频道或重试"
                    }

                    MediaPlayer.Event.Playing -> {
                        playerErrorMessage = null
                        Log.i(IPTV_LOG_TAG, describeVlcTracks(vlcPlayer))
                    }

                    MediaPlayer.Event.Vout,
                    MediaPlayer.Event.ESAdded,
                    MediaPlayer.Event.ESSelected,
                    -> {
                        Log.i(IPTV_LOG_TAG, describeVlcTracks(vlcPlayer))
                    }
                }
            }
        }
        vlcPlayer.setEventListener(listener)
        onDispose {
            vlcPlayer.setEventListener(null)
            vlcPlayer.stop()
            vlcPlayer.detachViews()
            vlcPlayer.release()
            libVlc.release()
        }
    }

    LaunchedEffect(Unit) {
        rootFocusRequester.requestFocus()
    }

    LaunchedEffect(uiState.currentChannel?.id) {
        if (uiState.currentChannel == null) {
            showChannelHint = false
            return@LaunchedEffect
        }
        showChannelHint = true
        delay(3_000)
        showChannelHint = false
    }

    LaunchedEffect(uiState.currentChannel?.id, playbackRetryNonce) {
        val channel = uiState.currentChannel
        if (channel == null) {
            vlcPlayer.stop()
            return@LaunchedEffect
        }
        playerErrorMessage = null
        vlcPlayer.stop()
        val media = Media(libVlc, Uri.parse(channel.url)).apply {
            setHWDecoderEnabled(false, false)
            addOption(":network-caching=1500")
            addOption(":clock-jitter=0")
            addOption(":clock-synchro=0")
        }
        vlcPlayer.media = media
        media.release()
        vlcPlayer.play()
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
                (LayoutInflater.from(viewContext).inflate(R.layout.tv_iptv_player_view, null) as VLCVideoLayout).apply {
                    vlcPlayer.attachViews(this, null, false, true)
                }
            },
            modifier = Modifier.fillMaxSize(),
        )

        if (
            shouldShowIptvChannelHint(
                currentChannel = uiState.currentChannel,
                hintActive = showChannelHint,
                channelListVisible = channelListVisible,
                loading = uiState.loading,
                statusMessage = uiState.statusMessage,
                playerErrorMessage = playerErrorMessage,
            )
        ) {
            TvIptvTopOverlay(channel = uiState.currentChannel)
        }

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
                channelListOpenNonce = channelListOpenNonce,
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
        shape = RoundedCornerShape(12.dp),
        modifier = Modifier
            .padding(start = 18.dp, top = 18.dp),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            if (channel == null) {
                TvIptvLogoFallback(
                    modifier = Modifier.size(34.dp),
                    iconSize = 22.dp,
                    tint = AppChrome.AccentWarm,
                )
            } else {
                TvIptvChannelLogo(
                    channel = channel,
                    modifier = Modifier.size(34.dp),
                    iconSize = 22.dp,
                    tint = AppChrome.AccentWarm,
                )
            }
            Text(
                text = channel?.name ?: "IPTV",
                color = Color.White,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.width(220.dp),
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
    channelListOpenNonce: Int,
    onSelect: (TvIptvChannelUiModel) -> Unit,
    modifier: Modifier = Modifier,
) {
    val initialFirstVisibleItemIndex = remember(groups, currentChannelId, channelListOpenNonce) {
        resolveIptvChannelListInitialFirstVisibleItemIndex(groups, currentChannelId)
    }
    val listState = rememberLazyListState(initialFirstVisibleItemIndex = initialFirstVisibleItemIndex)
    var skipInitialFocusScroll by remember(channelListOpenNonce) { mutableStateOf(true) }

    LaunchedEffect(groups, focusedChannelId, channelListOpenNonce) {
        if (skipInitialFocusScroll) {
            skipInitialFocusScroll = false
            return@LaunchedEffect
        }
        resolveIptvChannelListItemIndex(groups, focusedChannelId)?.let { itemIndex ->
            listState.animateScrollToItem(itemIndex)
        }
    }

    Surface(
        color = Color(0xE80B0F17),
        shape = RoundedCornerShape(topStart = 18.dp, bottomStart = 18.dp),
        modifier = modifier
            .fillMaxHeight()
            .width(360.dp),
    ) {
        LazyColumn(
            state = listState,
            contentPadding = PaddingValues(
                start = 14.dp,
                end = 14.dp,
                top = 18.dp,
                bottom = TvLayoutSpec.scrollBottomSafePaddingDp.dp,
            ),
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
            TvIptvChannelLogo(
                channel = channel,
                modifier = Modifier.size(30.dp),
                iconSize = 18.dp,
                tint = if (current) Color.White else AppChrome.TextMuted,
            )
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

@Composable
private fun TvIptvChannelLogo(
    channel: TvIptvChannelUiModel,
    modifier: Modifier = Modifier,
    iconSize: androidx.compose.ui.unit.Dp,
    tint: Color,
) {
    var showFallback by remember(channel.logoUrl) { mutableStateOf(channel.logoUrl.isNullOrBlank()) }

    Box(
        modifier = modifier
            .clip(RoundedCornerShape(8.dp))
            .background(Color.White.copy(alpha = 0.08f)),
        contentAlignment = Alignment.Center,
    ) {
        if (showFallback) {
            Icon(Icons.Filled.Tv, contentDescription = null, tint = tint, modifier = Modifier.size(iconSize))
        } else {
            AsyncImage(
                model = channel.logoUrl,
                contentDescription = null,
                contentScale = ContentScale.Fit,
                onError = { showFallback = true },
                modifier = Modifier
                    .fillMaxSize()
                    .padding(4.dp),
            )
        }
    }
}

@Composable
private fun TvIptvLogoFallback(
    modifier: Modifier = Modifier,
    iconSize: androidx.compose.ui.unit.Dp,
    tint: Color,
) {
    Box(
        modifier = modifier
            .clip(RoundedCornerShape(8.dp))
            .background(Color.White.copy(alpha = 0.08f)),
        contentAlignment = Alignment.Center,
    ) {
        Icon(Icons.Filled.Tv, contentDescription = null, tint = tint, modifier = Modifier.size(iconSize))
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

private fun describeVlcTracks(player: MediaPlayer): String {
    val videoTracks = player.videoTracks.orEmpty().joinToString(prefix = "[", postfix = "]") { track ->
        "${track.id}:${track.name}"
    }
    val audioTracks = player.audioTracks.orEmpty().joinToString(prefix = "[", postfix = "]") { track ->
        "${track.id}:${track.name}"
    }
    val currentVideo = player.currentVideoTrack
    val currentVideoText = if (currentVideo == null) {
        "none"
    } else {
        "${currentVideo.codec} ${currentVideo.width}x${currentVideo.height} profile=${currentVideo.profile} level=${currentVideo.level}"
    }
    return "tracks videoCount=${player.videoTracksCount} audioCount=${player.audioTracksCount} " +
        "selectedVideo=${player.videoTrack} currentVideo=$currentVideoText " +
        "videoTracks=$videoTracks audioTracks=$audioTracks"
}
