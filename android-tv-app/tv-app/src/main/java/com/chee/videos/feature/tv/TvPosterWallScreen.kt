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
import com.chee.videos.core.ui.TvFocusSafeSpec
import com.chee.videos.core.ui.TvLayoutSpec
import com.chee.videos.core.ui.tvFocusableGlow
import kotlinx.coroutines.flow.distinctUntilChanged

private val tvPosterWallFocusSafeSpace = TvFocusSafeSpec.posterFocusSafeSpaceDp.dp

internal object TvPosterWallFocusLayoutSpec {
    const val gridHorizontalPaddingDp: Float = 24f
    const val gridTopPaddingDp: Float = 8f
    const val gridBottomPaddingDp: Float = TvLayoutSpec.scrollBottomSafePaddingDp
    const val gridItemSpacingDp: Float = 16f
    const val posterCardsUseFocusSafeContainer: Boolean = true
}

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
                    contentPadding = PaddingValues(
                        start = TvPosterWallFocusLayoutSpec.gridHorizontalPaddingDp.dp,
                        end = TvPosterWallFocusLayoutSpec.gridHorizontalPaddingDp.dp,
                        top = TvPosterWallFocusLayoutSpec.gridTopPaddingDp.dp,
                        bottom = TvPosterWallFocusLayoutSpec.gridBottomPaddingDp.dp,
                    ),
                    horizontalArrangement = Arrangement.spacedBy(TvPosterWallFocusLayoutSpec.gridItemSpacingDp.dp),
                    verticalArrangement = Arrangement.spacedBy(TvPosterWallFocusLayoutSpec.gridItemSpacingDp.dp),
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
    val cardContent = buildTvPosterWallCardContent(baseUrl, item)
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(18.dp),
        modifier = modifier
            .padding(tvPosterWallFocusSafeSpace)
            .aspectRatio(2f / 3f)
            .tvFocusableGlow(shape = RoundedCornerShape(18.dp), focusedScale = TvFocusSafeSpec.posterFocusedScale)
            .clickable(onClick = onClick),
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            Box(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .background(
                        Brush.verticalGradient(
                            colors = listOf(Color(0xFF20142D), Color(0xFF0C1018)),
                        ),
                    ),
                contentAlignment = Alignment.Center,
            ) {
                if (!cardContent.showPosterPlaceholder) {
                    AsyncImage(
                        model = cardContent.posterUrl,
                        contentDescription = "${cardContent.title} 海报",
                        modifier = Modifier
                            .fillMaxSize()
                            .padding(12.dp)
                            .clip(RoundedCornerShape(14.dp)),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        Icon(
                            imageVector = Icons.Filled.Tv,
                            contentDescription = null,
                            tint = AppChrome.TextMuted,
                            modifier = Modifier.size(30.dp),
                        )
                        Text(
                            text = "暂无海报",
                            color = AppChrome.TextMuted,
                            style = MaterialTheme.typography.bodySmall,
                            fontWeight = FontWeight.SemiBold,
                        )
                    }
                }
            }
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 10.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = cardContent.title,
                    color = Color.White,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = cardContent.description,
                    color = AppChrome.TextSecondary,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

internal data class TvPosterWallCardContent(
    val posterUrl: String?,
    val title: String,
    val description: String,
    val showPosterPlaceholder: Boolean,
)

internal fun buildTvPosterWallCardContent(
    baseUrl: String,
    item: TvCatalogWallItemUiModel,
): TvPosterWallCardContent {
    val posterUrl = resolveTvResourceUrl(baseUrl, item.posterUrl)
    return TvPosterWallCardContent(
        posterUrl = posterUrl,
        title = item.title,
        description = item.description,
        showPosterPlaceholder = posterUrl.isNullOrBlank(),
    )
}
