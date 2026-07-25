package com.chee.videos.feature.shortcollections

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridItemSpan
import androidx.compose.foundation.lazy.staggeredgrid.itemsIndexed
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.cast.TvCastDeviceDialog
import com.chee.videos.core.ui.shorts.ShortVerticalFeedPlayer
import com.chee.videos.core.ui.shorts.resolveShortFeedThumbnailUrl

@Composable
fun ShortCollectionContentScreen(
    baseUrl: String,
    accessToken: String,
    collectionId: String,
    collectionName: String,
    onBack: () -> Unit,
    onCollectionUnavailable: () -> Unit,
    onFullscreenChange: (Boolean) -> Unit = {},
    onOpenRemoteControl: (String) -> Unit = {},
    viewModel: ShortCollectionContentViewModel = hiltViewModel(),
) {
    LaunchedEffect(collectionId, collectionName) {
        viewModel.bind(collectionId, collectionName)
    }

    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val offlineMessage = uiState.offlineMessage
    val errorMessage = uiState.errorMessage
    val playerStartIndex = remember(uiState.playingVideoId, uiState.items) {
        val id = uiState.playingVideoId ?: return@remember -1
        uiState.items.indexOfFirst { it.id == id }
    }

    BackHandler(enabled = uiState.playingVideoId != null) { viewModel.closePlayer() }
    BackHandler(enabled = uiState.playingVideoId == null) {
        if (offlineMessage != null) onCollectionUnavailable() else onBack()
    }

    LaunchedEffect(uiState.pendingRemoteSessionId) {
        val sessionId = uiState.pendingRemoteSessionId ?: return@LaunchedEffect
        onOpenRemoteControl(sessionId)
        viewModel.consumePendingRemoteSession()
    }

    Column(modifier = Modifier.fillMaxSize().background(AppChrome.PageGradient).statusBarsPadding()) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(horizontal = 4.dp, vertical = 4.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            IconButton(onClick = { if (offlineMessage != null) onCollectionUnavailable() else onBack() }) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回", tint = Color.White)
            }
            Text(
                text = uiState.collectionName.ifBlank { collectionName },
                color = Color.White,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f).padding(end = 14.dp),
            )
        }

        when {
            offlineMessage != null -> CenterMessage(
                message = offlineMessage,
                actionLabel = "返回合集",
                onAction = onCollectionUnavailable,
            )
            uiState.loading && uiState.items.isEmpty() -> Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) { CircularProgressIndicator(color = Color.White) }
            errorMessage != null && uiState.items.isEmpty() -> CenterMessage(errorMessage, actionLabel = "重试", onAction = viewModel::retry)
            uiState.items.isEmpty() -> CenterMessage("该合集已下线或暂无可播放内容", actionLabel = "返回合集", onAction = onCollectionUnavailable)
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
                    CollectionCoverCard(baseUrl = baseUrl, item = item, onClick = { viewModel.enterPlayer(item.id) })
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
        ShortVerticalFeedPlayer(
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
            onOpenCastSheet = viewModel::openCastSheet,
            onClose = viewModel::closePlayer,
            onToggleFitMode = viewModel::toggleFitMode,
            onTogglePlaybackMode = viewModel::togglePlaybackMode,
            onFullscreenChange = onFullscreenChange,
        )
    }

    if (uiState.castSheetVisible) {
        TvCastDeviceDialog(
            devices = uiState.tvDevices,
            loading = uiState.castDevicesLoading,
            launching = uiState.castLaunching,
            selectedDeviceId = uiState.selectedTvDeviceId,
            errorMessage = uiState.castErrorMessage,
            onDismiss = viewModel::dismissCastSheet,
            onSelectDevice = viewModel::selectTvDevice,
            onConfirm = viewModel::startCastFromCollection,
        )
    }
}

@Composable
private fun CenterMessage(
    message: String,
    actionLabel: String? = null,
    onAction: (() -> Unit)? = null,
) {
    Column(
        modifier = Modifier.fillMaxSize().padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center,
    ) {
        Text(message, color = Color.White)
        if (actionLabel != null && onAction != null) {
            Text(
                text = actionLabel,
                color = Color.White,
                modifier = Modifier
                    .padding(top = 12.dp)
                    .clip(RoundedCornerShape(8.dp))
                    .clickable(onClick = onAction)
                    .padding(horizontal = 16.dp, vertical = 8.dp),
            )
        }
    }
}

@Composable
private fun CollectionCoverCard(baseUrl: String, item: VideoListItemDto, onClick: () -> Unit) {
    val thumbUrl = remember(baseUrl, item.thumbnailPath) { resolveShortFeedThumbnailUrl(baseUrl, item.thumbnailPath) }
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(8.dp))
            .clickable(onClick = onClick),
    ) {
        AsyncImage(
            model = thumbUrl,
            contentDescription = item.title,
            contentScale = ContentScale.Crop,
            modifier = Modifier.fillMaxWidth().aspectRatio(0.72f),
        )
        Box(modifier = Modifier.matchParentSize().background(Brush.verticalGradient(listOf(Color.Transparent, Color.Transparent, Color(0xBF090B11)))))
        Text(
            text = item.title,
            color = Color.White,
            style = MaterialTheme.typography.bodySmall,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.align(Alignment.BottomStart).padding(10.dp),
        )
    }
}
