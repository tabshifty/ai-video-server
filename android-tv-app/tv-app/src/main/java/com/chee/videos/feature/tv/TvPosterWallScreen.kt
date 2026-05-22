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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.itemsIndexed
import androidx.compose.foundation.lazy.grid.rememberLazyGridState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.Icon
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
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.TvFocusSafeSpec
import com.chee.videos.core.ui.TvLayoutSpec
import com.chee.videos.core.ui.TvEmptyState
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvIconActionButton
import com.chee.videos.core.ui.TvInlineLoadingState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableGlow
import com.chee.videos.core.ui.tvFocusableScaleOnly
import com.chee.videos.core.ui.tvSharedSeriesPoster
import com.chee.videos.core.ui.tvStaggerEntry
import kotlinx.coroutines.flow.distinctUntilChanged

private val tvPosterWallFocusSafeSpace = TvFocusSafeSpec.posterFocusSafeSpaceDp.dp
private val TvPosterWallCardShape = AppChrome.SurfaceShape
private val TvPosterWallTitleBackground = Color(0xE80A0E15)

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

    LaunchedTvInitialFocus(uiState.items.isNotEmpty(), uiState.page, uiState.loading, uiState.refreshing) {
        if (uiState.items.isNotEmpty() && uiState.page == 1 && !uiState.loading && !uiState.refreshing) {
            firstItemFocusRequester.tryRequestFocus()
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
            .background(AppChrome.PageGradient)
            .statusBarsPadding(),
    ) {
        TvPosterWallTopBar(
            title = wallSpec.title,
            subtitle = wallSpec.subtitle,
            sortBy = uiState.sortBy,
            sortOrder = uiState.sortOrder,
            onBack = onBack,
            onRefresh = viewModel::refresh,
            onChangeSort = viewModel::changeSort,
            refreshFocusRequester = refreshFocusRequester,
        )

        when {
            uiState.loading && uiState.items.isEmpty() -> {
                TvPageLoadingState(message = "正在加载${wallSpec.title}")
            }

            !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
                TvErrorState(
                    message = uiState.errorMessage.orEmpty(),
                    onAction = viewModel::refresh,
                )
            }

            uiState.items.isEmpty() -> {
                TvEmptyState(
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
                    itemsIndexed(
                        uiState.items,
                        key = { _, item -> item.id },
                    ) { index, item ->
                        val focusModifier = if (uiState.page == 1 && uiState.items.firstOrNull()?.id == item.id) {
                            Modifier.focusRequester(firstItemFocusRequester)
                        } else {
                            Modifier
                        }
                        TvPosterWallCard(
                            baseUrl = baseUrl,
                            item = item,
                            modifier = focusModifier.tvStaggerEntry(index = index),
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
                            TvInlineLoadingState(
                                message = "正在加载更多",
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 8.dp),
                            )
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
    sortBy: String,
    sortOrder: String,
    onBack: () -> Unit,
    onRefresh: () -> Unit,
    onChangeSort: (String, String) -> Unit,
    refreshFocusRequester: FocusRequester,
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(horizontal = 14.dp, vertical = 12.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        TvIconActionButton(
            icon = Icons.AutoMirrored.Filled.ArrowBack,
            contentDescription = "返回",
            onClick = onBack,
            shape = AppChrome.SurfaceShape,
            containerColor = AppChrome.SurfaceElevated,
            contentColor = AppChrome.TextPrimary,
        )
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
        TvPosterWallSortButton(
            label = tvPosterWallSortByLabel(sortBy),
            onClick = { onChangeSort(if (sortBy == "added") "release" else "added", sortOrder) },
        )
        TvPosterWallSortButton(
            label = tvPosterWallSortOrderLabel(sortOrder),
            onClick = { onChangeSort(sortBy, if (sortOrder == "desc") "asc" else "desc") },
        )
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
private fun TvPosterWallSortButton(
    label: String,
    onClick: () -> Unit,
) {
    Surface(
        color = AppChrome.SurfaceElevated.copy(alpha = 0.9f),
        shape = AppChrome.PillShape,
        modifier = Modifier
            .tvFocusableGlow(shape = AppChrome.PillShape, focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Text(
            text = label,
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

internal fun tvPosterWallSortByLabel(sortBy: String): String {
    return when (sortBy) {
        "release" -> "发售时间"
        else -> "添加时间"
    }
}

internal fun tvPosterWallSortOrderLabel(sortOrder: String): String {
    return when (sortOrder) {
        "asc" -> "正序"
        else -> "倒序"
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
        color = Color.Transparent,
        shape = TvPosterWallCardShape,
        modifier = modifier
            .padding(tvPosterWallFocusSafeSpace)
            .tvFocusableScaleOnly(shape = TvPosterWallCardShape, focusedScale = TvFocusSafeSpec.posterFocusedScale)
            .clickable(onClick = onClick),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .clip(TvPosterWallCardShape),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(9f / 16f)
                    .then(
                        if (item.type == "tv") {
                            Modifier.tvSharedSeriesPoster(item.id)
                        } else {
                            Modifier
                        },
                    )
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
                        modifier = Modifier.fillMaxSize(),
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
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(TvPosterWallTitleBackground)
                    .padding(horizontal = 10.dp, vertical = 9.dp),
            ) {
                Text(
                    text = cardContent.title,
                    color = Color.White,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.Bold,
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
    val showDescription: Boolean,
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
        showDescription = false,
        showPosterPlaceholder = posterUrl.isNullOrBlank(),
    )
}
