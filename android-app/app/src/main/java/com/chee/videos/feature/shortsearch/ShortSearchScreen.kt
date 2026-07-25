package com.chee.videos.feature.shortsearch

import androidx.activity.compose.BackHandler
import androidx.compose.foundation.ExperimentalFoundationApi
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
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
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
import com.chee.videos.core.ui.cast.TvCastDeviceDialog
import com.chee.videos.core.ui.shorts.ShortVerticalFeedPlayer
import com.chee.videos.core.util.UrlBuilder

@OptIn(ExperimentalFoundationApi::class)
@Composable
fun ShortSearchScreen(
    baseUrl: String,
    accessToken: String,
    onFullscreenChange: (Boolean) -> Unit = {},
    onOpenRemoteControl: (String) -> Unit = {},
    viewModel: ShortSearchViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val playerStartIndex = remember(uiState.playingVideoId, uiState.items) {
        val id = uiState.playingVideoId ?: return@remember -1
        uiState.items.indexOfFirst { it.id == id }
    }

    BackHandler(enabled = uiState.playingVideoId != null) { viewModel.closePlayer() }

    LaunchedEffect(uiState.pendingRemoteSessionId) {
        val sessionId = uiState.pendingRemoteSessionId ?: return@LaunchedEffect
        onOpenRemoteControl(sessionId)
        viewModel.consumePendingRemoteSession()
    }

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
            onConfirm = viewModel::startCastFromCurrentSearch,
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

private fun resolveThumbnailUrl(baseUrl: String, rawPath: String?): String? {
    val path = rawPath?.trim().orEmpty()
    if (path.isBlank()) return null
    if (path.startsWith("http://") || path.startsWith("https://")) return path
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) return null
    return if (path.startsWith('/')) "$normalizedBase$path" else "$normalizedBase/$path"
}
