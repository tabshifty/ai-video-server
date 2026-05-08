package com.chee.videos.feature.tv

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
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.tvFocusableGlow
import kotlinx.coroutines.flow.distinctUntilChanged

@Composable
fun TvPosterWallScreen(
    baseUrl: String,
    onBack: () -> Unit,
    onOpenSeries: (String) -> Unit,
    onOpenLongForm: (String, String) -> Unit,
    viewModel: TvPosterWallViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val gridState = rememberLazyGridState()
    val firstItemFocusRequester = remember { FocusRequester() }
    val refreshFocusRequester = remember { FocusRequester() }
    val wallSpec = resolveTvCatalogWallSpec(uiState.kind, uiState.title)

    LaunchedEffect(uiState.items.isNotEmpty(), uiState.page, uiState.loading, uiState.refreshing) {
        if (uiState.items.isNotEmpty() && uiState.page == 1 && !uiState.loading && !uiState.refreshing) {
            firstItemFocusRequester.requestFocus()
        }
    }

    LaunchedEffect(gridState, uiState.items.size, uiState.loading, uiState.loadingMore, uiState.refreshing) {
        snapshotFlow { gridState.layoutInfo.visibleItemsInfo.lastOrNull()?.index ?: -1 }
            .distinctUntilChanged()
            .collect { viewModel.loadMoreIfNeeded(it) }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient),
    ) {
        TvPosterWallTopBar(
            title = wallSpec.title,
            subtitle = wallSpec.subtitle,
            onBack = onBack,
            onRefresh = viewModel::refresh,
            refreshFocusRequester = refreshFocusRequester,
        )

        when {
            uiState.loading && uiState.items.isEmpty() -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = AppChrome.AccentStrong)
                }
            }

            !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
                TvPosterWallEmptyState(
                    title = "加载失败",
                    message = uiState.errorMessage.orEmpty(),
                    actionLabel = "重试",
                    onAction = viewModel::refresh,
                )
            }

            uiState.items.isEmpty() -> {
                TvPosterWallEmptyState(
                    title = wallSpec.title,
                    message = "暂无可用内容",
                    actionLabel = "刷新",
                    onAction = viewModel::refresh,
                )
            }

            else -> {
                LazyVerticalGrid(
                    state = gridState,
                    columns = GridCells.Adaptive(minSize = 170.dp),
                    contentPadding = PaddingValues(start = 14.dp, end = 14.dp, top = 8.dp, bottom = 24.dp),
                    horizontalArrangement = Arrangement.spacedBy(12.dp),
                    verticalArrangement = Arrangement.spacedBy(14.dp),
                    modifier = Modifier.fillMaxSize(),
                ) {
                    items(uiState.items, key = { item -> item.id }) { item ->
                        TvPosterWallCard(
                            baseUrl = baseUrl,
                            item = item,
                            modifier = if (uiState.page == 1 && uiState.items.firstOrNull()?.id == item.id) {
                                Modifier.focusRequester(firstItemFocusRequester)
                            } else {
                                Modifier
                            },
                            onClick = {
                                if (item.type == "tv") {
                                    onOpenSeries(item.id)
                                } else {
                                    onOpenLongForm(item.id, item.type)
                                }
                            },
                        )
                    }
                    if (uiState.loadingMore) {
                        item(span = { androidx.compose.foundation.lazy.grid.GridItemSpan(maxLineSpan) }) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 8.dp),
                                horizontalArrangement = Arrangement.Center,
                                verticalAlignment = Alignment.CenterVertically,
                            ) {
                                CircularProgressIndicator(
                                    modifier = Modifier.size(18.dp),
                                    color = AppChrome.AccentStrong,
                                    strokeWidth = 2.dp,
                                )
                                Text(
                                    text = "正在加载更多",
                                    color = AppChrome.TextMuted,
                                    style = MaterialTheme.typography.bodySmall,
                                    modifier = Modifier.padding(start = 8.dp),
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TvPosterWallTopBar(
    title: String,
    subtitle: String,
    onBack: () -> Unit,
    onRefresh: () -> Unit,
    refreshFocusRequester: FocusRequester,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 14.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        IconButton(onClick = onBack, modifier = Modifier.tvFocusableGlow(shape = RoundedCornerShape(16.dp))) {
            Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回", tint = AppChrome.TextPrimary)
        }
        Column(modifier = Modifier.weight(1f), verticalArrangement = Arrangement.spacedBy(2.dp)) {
            Text(
                text = title,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Text(
                text = subtitle,
                color = AppChrome.TextMuted,
                style = MaterialTheme.typography.bodySmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        Surface(
            color = AppChrome.SurfaceElevated,
            shape = AppChrome.PillShape,
            modifier = Modifier
                .focusRequester(refreshFocusRequester)
                .tvFocusableGlow(shape = AppChrome.PillShape, focusedScale = 1.04f)
                .clickable(onClick = onRefresh),
        ) {
            Row(
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 10.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Icon(Icons.Filled.Refresh, contentDescription = null, tint = AppChrome.TextPrimary, modifier = Modifier.size(18.dp))
                Text(
                    text = "刷新",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                )
            }
        }
    }
}

@Composable
private fun TvPosterWallEmptyState(
    title: String,
    message: String,
    actionLabel: String,
    onAction: () -> Unit,
) {
    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
        Surface(
            color = AppChrome.SurfaceElevated,
            shape = AppChrome.CardShape,
            modifier = Modifier
                .padding(horizontal = 20.dp)
                .fillMaxWidth()
                .height(180.dp)
                .tvFocusableGlow(shape = AppChrome.CardShape)
                .clickable(onClick = onAction),
        ) {
            Column(
                modifier = Modifier.fillMaxSize().padding(horizontal = 20.dp, vertical = 22.dp),
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Icon(
                    imageVector = Icons.Filled.Tv,
                    contentDescription = null,
                    tint = AppChrome.TextMuted,
                    modifier = Modifier.size(28.dp),
                )
                Text(
                    text = title,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                )
                Text(
                    text = message,
                    color = AppChrome.TextMuted,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = actionLabel,
                    color = AppChrome.AccentStrong,
                    style = MaterialTheme.typography.labelLarge,
                    fontWeight = FontWeight.SemiBold,
                )
            }
        }
    }
}

@Composable
private fun TvPosterWallCard(
    baseUrl: String,
    item: TvCatalogWallItemUiModel,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    val posterUrl = resolveTvResourceUrl(baseUrl, item.posterUrl)
    val backdropUrl = resolveTvResourceUrl(baseUrl, item.backdropUrl)
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(18.dp),
        modifier = modifier
            .aspectRatio(0.72f)
            .clip(RoundedCornerShape(18.dp))
            .tvFocusableGlow(shape = RoundedCornerShape(18.dp), focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            if (!backdropUrl.isNullOrBlank()) {
                AsyncImage(
                    model = backdropUrl,
                    contentDescription = item.title,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.verticalGradient(
                                colors = listOf(Color(0xFF20142D), Color(0xFF0C1018)),
                            ),
                        ),
                )
            }
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.verticalGradient(
                            colors = listOf(Color(0x00000000), Color(0xD90B0F17)),
                        ),
                    ),
            )
            if (!posterUrl.isNullOrBlank()) {
                AsyncImage(
                    model = posterUrl,
                    contentDescription = null,
                    modifier = Modifier
                        .align(Alignment.TopCenter)
                        .fillMaxWidth(0.86f)
                        .padding(top = 14.dp)
                        .aspectRatio(2f / 3f)
                        .clip(RoundedCornerShape(14.dp)),
                    contentScale = ContentScale.Crop,
                )
            }
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .align(Alignment.BottomStart)
                    .padding(horizontal = 12.dp, vertical = 12.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = item.title,
                    color = Color.White,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = item.description,
                    color = AppChrome.TextSecondary,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}
