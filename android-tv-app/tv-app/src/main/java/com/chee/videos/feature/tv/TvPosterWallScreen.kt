package com.chee.videos.feature.tv

import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxHeight
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
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.shadow
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shadow
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.TvFocusMotionTokens
import com.chee.videos.core.ui.TvFocusSafeSpec
import com.chee.videos.core.ui.TvLayoutSpec
import com.chee.videos.core.ui.TvEmptyState
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvIconActionButton
import com.chee.videos.core.ui.TvInlineLoadingState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableScaleOnly
import com.chee.videos.core.ui.tvSharedSeriesPoster
import com.chee.videos.core.ui.tvStaggerEntry
import kotlinx.coroutines.flow.distinctUntilChanged

private val tvPosterWallFocusSafeSpace = TvFocusSafeSpec.posterFocusSafeSpaceDp.dp
private val TvPosterWallCardShape = AppChrome.SurfaceShape
private val TvPosterWallPlaceholderBrush = Brush.verticalGradient(
    colors = listOf(AppChrome.SurfaceStrong, AppChrome.Canvas),
)
// D-轻：标题融进海报底部渐变遮罩。遮罩盖卡片底部约 45%，自上而下由透明渐到接近不透明，
// 保证两行标题（尤其上行）都有足够深的底色，再叠加文字阴影兜底，替代旧的独立深色标题色块。
private val TvPosterWallTitleScrimBrush = Brush.verticalGradient(
    colors = listOf(
        Color.Transparent,
        AppChrome.Canvas.copy(alpha = 0.55f),
        AppChrome.Canvas.copy(alpha = 0.88f),
    ),
)

internal object TvPosterWallFocusLayoutSpec {
    const val gridColumnCount: Int = 6
    const val gridHorizontalPaddingDp: Float = 24f
    const val gridTopPaddingDp: Float = 26f
    const val gridBottomPaddingDp: Float = TvLayoutSpec.scrollBottomSafePaddingDp
    // 卡片本身已带 8dp 焦点安全外容器，网格 gutter 维持 8dp 即可把可见卡缝收紧到约 24dp。
    const val gridItemSpacingDp: Float = 8f
    // 海报墙想要比首页/目录页更醒目的聚焦放大，但 TvFocusSafeSpec.posterFocusedScale 是首页/目录页共用的
    // 共享 token（恒 1.04f），改它会连带改其它页面。海报墙专属的 1.08f 收口在这里，显式传给卡片。
    const val posterWallFocusedScale: Float = 1.08f
    // 海报墙是 [[TV 焦点反馈只缩放]] 全仓口径的允许例外：聚焦时叠加一道中性（黑色）落影抬升，
    // 让卡片从网格里「浮」起来。仅海报墙本地内联，不抽共享修饰器、不动 tvFocusableScaleOnly。
    const val posterWallFocusedShadowElevationDp: Float = 10f
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
    var initialFocusRequested by remember { mutableStateOf(false) }
    val wallSpec = resolveTvCatalogWallSpec(uiState.kind, uiState.title)

    LaunchedTvInitialFocus(uiState.items.firstOrNull()?.id, uiState.page, uiState.loading) {
        if (!initialFocusRequested && uiState.items.isNotEmpty() && uiState.page == 1 && !uiState.loading && !uiState.refreshing) {
            if (firstItemFocusRequester.tryRequestFocus()) {
                initialFocusRequested = true
            }
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
                    columns = GridCells.Fixed(TvPosterWallFocusLayoutSpec.gridColumnCount),
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
                    if (uiState.refreshing) {
                        item(span = { androidx.compose.foundation.lazy.grid.GridItemSpan(maxLineSpan) }) {
                            TvInlineLoadingState(
                                message = "正在更新${wallSpec.title}",
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 8.dp),
                            )
                        }
                    }
                    if (!uiState.errorMessage.isNullOrBlank()) {
                        item(span = { androidx.compose.foundation.lazy.grid.GridItemSpan(maxLineSpan) }) {
                            TvPosterWallInlineError(
                                message = uiState.errorMessage.orEmpty(),
                                onAction = viewModel::refresh,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(vertical = 8.dp),
                            )
                        }
                    }
                    itemsIndexed(
                        uiState.items,
                        key = { _, item -> item.id },
                        // N 张海报卡显式同型，避免与 refreshing/error/loadingMore 头尾项混用组合槽
                        contentType = { _, _ -> "poster" },
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
private fun TvPosterWallInlineError(
    message: String,
    onAction: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.SurfaceShape,
        modifier = modifier,
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                imageVector = Icons.Filled.Warning,
                contentDescription = null,
                tint = AppChrome.Error,
                modifier = Modifier.size(20.dp),
            )
            Text(
                text = message,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.bodyMedium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f),
            )
            Surface(
                color = AppChrome.SurfaceStrong,
                shape = AppChrome.PillShape,
                modifier = Modifier
                    .tvFocusableScaleOnly(focusedScale = 1.04f)
                    .clickable(onClick = onAction),
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 14.dp, vertical = 8.dp),
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Icon(
                        imageVector = Icons.Filled.Refresh,
                        contentDescription = null,
                        tint = AppChrome.TextPrimary,
                        modifier = Modifier.size(16.dp),
                    )
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
                .tvFocusableScaleOnly(focusedScale = 1.04f)
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
            .tvFocusableScaleOnly(focusedScale = 1.04f)
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
    // 海报墙例外：聚焦时叠加中性（黑色）落影抬升。elevation 跟随焦点动画，非聚焦时归零。
    var isCardFocused by remember { mutableStateOf(false) }
    val shadowElevation by animateDpAsState(
        targetValue = if (isCardFocused) TvPosterWallFocusLayoutSpec.posterWallFocusedShadowElevationDp.dp else 0.dp,
        animationSpec = spring(
            dampingRatio = TvFocusMotionTokens.ScaleDampingRatio,
            stiffness = TvFocusMotionTokens.ScaleStiffness,
        ),
        label = "tvPosterWallCardShadow",
    )
    Surface(
        color = Color.Transparent,
        shape = TvPosterWallCardShape,
        modifier = modifier
            .padding(tvPosterWallFocusSafeSpace)
            .onFocusChanged { state -> isCardFocused = state.isFocused || state.hasFocus }
            .tvFocusableScaleOnly(focusedScale = TvPosterWallFocusLayoutSpec.posterWallFocusedScale)
            .shadow(
                elevation = shadowElevation,
                shape = TvPosterWallCardShape,
                clip = false,
                ambientColor = Color.Black,
                spotColor = Color.Black,
            )
            .clickable(onClick = onClick),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .aspectRatio(9f / 16f)
                .clip(TvPosterWallCardShape)
                .background(TvPosterWallPlaceholderBrush),
            contentAlignment = Alignment.Center,
        ) {
            // shared-element 只绑在海报图片节点上（与详情页 TvSeriesDetailBackdrop 的纯背景图节点对应），
            // scrim/标题作为兄弟覆盖层，不进入 shared-element 子树，避免过渡时把标题一起「飞」进详情页。
            val sharedModifier = if (item.type == "tv") {
                Modifier.tvSharedSeriesPoster(item.id)
            } else {
                Modifier
            }
            if (!cardContent.showPosterPlaceholder) {
                AsyncImage(
                    model = cardContent.posterUrl,
                    contentDescription = "${cardContent.title} 海报",
                    modifier = Modifier.fillMaxSize().then(sharedModifier),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.fillMaxSize().then(sharedModifier),
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
            // 底部渐变遮罩 + 标题压在遮罩之上，消除独立标题色块的「胶布感」。
            // 占位卡（无海报）的「暂无海报」图标+标签由外层 Box 的 contentAlignment=Center 居中，
            // 标题压在底部遮罩上，两者空间分离、不堆叠；占位卡同样保留标题，避免无海报时无法识别剧集。
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .fillMaxHeight(0.55f)
                    .align(Alignment.BottomCenter)
                    .background(TvPosterWallTitleScrimBrush),
            )
            Text(
                text = cardContent.title,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleSmall.copy(
                    shadow = Shadow(
                        color = AppChrome.Canvas.copy(alpha = 0.9f),
                        offset = Offset(0f, 1f),
                        blurRadius = 4f,
                    ),
                ),
                fontWeight = FontWeight.Bold,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .fillMaxWidth()
                    .padding(horizontal = 8.dp, vertical = 8.dp),
            )
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
