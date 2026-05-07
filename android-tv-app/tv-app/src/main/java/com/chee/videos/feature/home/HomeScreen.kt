package com.chee.videos.feature.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.GridItemSpan
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
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
import com.chee.videos.core.model.resolveAvPosterUrl
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.homeContentTabs
import com.chee.videos.core.util.UrlBuilder
import com.chee.videos.feature.shorts.ShortFeedScreen
import com.chee.videos.feature.tv.TvCatalogScreen

private val tabs = homeContentTabs

@Composable
fun HomeScreen(
    baseUrl: String,
    accessToken: String,
    onOpenDetail: (String, String) -> Unit,
    onOpenTvSeries: (String) -> Unit,
    onOpenTvContinueWatching: (String, Int, Int) -> Unit,
    onOpenShortDiscover: (mode: String, value: String, title: String) -> Unit,
    onOpenImageCollectionViewer: (String) -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    var selectedTab by rememberSaveable { mutableIntStateOf(0) }

    LaunchedEffect(selectedTab) {
        val tab = tabs[selectedTab]
        if (tab.type == "movie" || tab.type == "av") {
            viewModel.loadCategory(tab.type)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .statusBarsPadding(),
    ) {
        HomeHeader(
            selectedTab = selectedTab,
            onSelectTab = { selectedTab = it },
        )

        Box(modifier = Modifier.fillMaxSize()) {
            when (tabs[selectedTab].type) {
                "short" -> {
                    ShortFeedScreen(
                        baseUrl = baseUrl,
                        accessToken = accessToken,
                        onOpenDiscover = onOpenShortDiscover,
                        onOpenImageCollectionViewer = onOpenImageCollectionViewer,
                    )
                }

                "movie" -> {
                    CategoryListSection(
                        baseUrl = baseUrl,
                        state = uiState.movie,
                        categoryTitle = tabs[selectedTab].title,
                        onRetry = { viewModel.loadCategory("movie", force = true) },
                        onOpenDetail = onOpenDetail,
                    )
                }

                "episode" -> {
                    TvCatalogScreen(
                        onOpenSeries = onOpenTvSeries,
                        onOpenContinueWatching = onOpenTvContinueWatching,
                        onOpenLongForm = onOpenDetail,
                    )
                }

                "av" -> {
                    AvCatalogSection(
                        baseUrl = baseUrl,
                        browseState = uiState.av,
                        searchState = uiState.avSearch,
                        onQueryChange = viewModel::updateAvQuery,
                        onRetry = viewModel::retryAvState,
                        onClearQuery = { viewModel.updateAvQuery("") },
                        onOpenDetail = onOpenDetail,
                    )
                }
            }
        }
    }
}

@Composable
private fun HomeHeader(
    selectedTab: Int,
    onSelectTab: (Int) -> Unit,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = Color.Transparent,
    ) {
        Surface(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 16.dp, vertical = 8.dp),
            color = AppChrome.Surface.copy(alpha = 0.92f),
            shape = AppChrome.SectionShape,
            tonalElevation = 0.dp,
            shadowElevation = 0.dp,
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(4.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp),
            ) {
                tabs.forEachIndexed { index, tab ->
                    val selected = index == selectedTab
                    Surface(
                        modifier = Modifier
                            .weight(1f)
                            .clip(RoundedCornerShape(14.dp))
                            .clickable { onSelectTab(index) },
                        color = if (selected) AppChrome.AccentSoft else Color.Transparent,
                        shape = RoundedCornerShape(14.dp),
                    ) {
                        Box(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 6.dp, vertical = 10.dp),
                            contentAlignment = Alignment.Center,
                        ) {
                            Text(
                                text = tab.title,
                                color = if (selected) AppChrome.TextPrimary else AppChrome.TextMuted,
                                style = MaterialTheme.typography.titleSmall,
                                fontWeight = FontWeight.SemiBold,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun AvCatalogSection(
    baseUrl: String,
    browseState: CategoryState,
    searchState: AvSearchState,
    onQueryChange: (String) -> Unit,
    onRetry: () -> Unit,
    onClearQuery: () -> Unit,
    onOpenDetail: (String, String) -> Unit,
) {
    val activeItems = if (searchState.isSearchMode) searchState.results else browseState.items
    val statusText = buildAvStatusText(browseState, searchState)

    LazyVerticalGrid(
        modifier = Modifier.fillMaxSize(),
        columns = GridCells.Adaptive(minSize = 156.dp),
        contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 6.dp, bottom = 20.dp),
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        item(span = { GridItemSpan(maxLineSpan) }) {
            AvCatalogHero(
                query = searchState.query,
                statusText = statusText,
                isSearching = searchState.loading,
                onQueryChange = onQueryChange,
                onClearQuery = onClearQuery,
            )
        }

        when {
            browseState.loading && browseState.items.isEmpty() && !searchState.isSearchMode -> {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    LoadingStateCard(label = "正在加载 AV 海报墙")
                }
            }

            searchState.loading && searchState.results.isEmpty() -> {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    LoadingStateCard(label = "正在远程搜索“${searchState.query.trim()}”")
                }
            }

            !searchState.errorMessage.isNullOrBlank() && searchState.results.isEmpty() -> {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    AvCatalogStateCard(
                        title = "搜索失败",
                        message = searchState.errorMessage.orEmpty(),
                        primaryLabel = "重试搜索",
                        onPrimaryClick = onRetry,
                        secondaryLabel = "清空搜索",
                        onSecondaryClick = onClearQuery,
                    )
                }
            }

            !browseState.errorMessage.isNullOrBlank() && browseState.items.isEmpty() -> {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    AvCatalogStateCard(
                        title = "加载失败",
                        message = browseState.errorMessage.orEmpty(),
                        primaryLabel = "重新加载",
                        onPrimaryClick = onRetry,
                    )
                }
            }

            searchState.isSearchMode && activeItems.isEmpty() -> {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    AvCatalogStateCard(
                        title = "没有找到结果",
                        message = "换个番号试试，或者清空搜索回到默认浏览。",
                        primaryLabel = "清空搜索",
                        onPrimaryClick = onClearQuery,
                    )
                }
            }

            activeItems.isEmpty() -> {
                item(span = { GridItemSpan(maxLineSpan) }) {
                    AvCatalogStateCard(
                        title = "暂无内容",
                        message = "当前还没有可展示的 AV 作品。",
                        primaryLabel = "重新加载",
                        onPrimaryClick = onRetry,
                    )
                }
            }

            else -> {
                items(activeItems, key = { it.id }) { item ->
                    AvPosterCard(
                        baseUrl = baseUrl,
                        item = item,
                        onOpenDetail = onOpenDetail,
                    )
                }
            }
        }
    }
}

@Composable
private fun AvCatalogHero(
    query: String,
    statusText: String,
    isSearching: Boolean,
    onQueryChange: (String) -> Unit,
    onClearQuery: () -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppChrome.CardShape)
            .background(AppChrome.HeroGradient)
            .padding(18.dp),
    ) {
        Column(
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = "AV 海报墙",
                style = MaterialTheme.typography.labelLarge,
                color = AppChrome.AccentWarm,
                fontWeight = FontWeight.SemiBold,
            )
            Text(
                text = "按番号 / 标题快速定位",
                style = MaterialTheme.typography.headlineSmall,
                color = AppChrome.TextPrimary,
                fontWeight = FontWeight.Bold,
            )
            OutlinedTextField(
                value = query,
                onValueChange = onQueryChange,
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                textStyle = MaterialTheme.typography.bodyLarge.copy(color = AppChrome.TextPrimary),
                placeholder = {
                    Text(
                        text = "输入番号或标题，直接远程搜索",
                        color = AppChrome.TextMuted,
                    )
                },
                leadingIcon = {
                    Icon(
                        imageVector = Icons.Filled.Search,
                        contentDescription = null,
                        tint = AppChrome.TextMuted,
                    )
                },
                trailingIcon = {
                    if (query.isNotBlank()) {
                        IconButton(onClick = onClearQuery) {
                            Icon(
                                imageVector = Icons.Filled.Close,
                                contentDescription = "清空搜索",
                                tint = AppChrome.TextMuted,
                            )
                        }
                    }
                },
                shape = RoundedCornerShape(18.dp),
                colors = OutlinedTextFieldDefaults.colors(
                    focusedContainerColor = AppChrome.Surface.copy(alpha = 0.92f),
                    unfocusedContainerColor = AppChrome.Surface.copy(alpha = 0.92f),
                    focusedBorderColor = AppChrome.AccentStrong,
                    unfocusedBorderColor = AppChrome.SurfaceStrong,
                    focusedTextColor = AppChrome.TextPrimary,
                    unfocusedTextColor = AppChrome.TextPrimary,
                    focusedPlaceholderColor = AppChrome.TextMuted,
                    unfocusedPlaceholderColor = AppChrome.TextMuted,
                    focusedLeadingIconColor = AppChrome.TextMuted,
                    unfocusedLeadingIconColor = AppChrome.TextMuted,
                ),
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                if (isSearching) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(16.dp),
                        strokeWidth = 2.dp,
                        color = AppChrome.AccentStrong,
                    )
                }
                Text(
                    text = statusText,
                    style = MaterialTheme.typography.bodyMedium,
                    color = AppChrome.TextSecondary,
                )
            }
        }
    }
}

@Composable
private fun LoadingStateCard(label: String) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 40.dp),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            CircularProgressIndicator(color = AppChrome.AccentStrong)
            Text(
                text = label,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.bodyMedium,
            )
        }
    }
}

@Composable
private fun AvCatalogStateCard(
    title: String,
    message: String,
    primaryLabel: String,
    onPrimaryClick: () -> Unit,
    secondaryLabel: String? = null,
    onSecondaryClick: (() -> Unit)? = null,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.CardShape,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp, vertical = 20.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = title,
                style = MaterialTheme.typography.titleMedium,
                color = AppChrome.TextPrimary,
                fontWeight = FontWeight.SemiBold,
            )
            Text(
                text = message,
                style = MaterialTheme.typography.bodyMedium,
                color = AppChrome.TextSecondary,
            )
            Row(
                horizontalArrangement = Arrangement.spacedBy(8.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Button(
                    onClick = onPrimaryClick,
                    colors = ButtonDefaults.buttonColors(
                        containerColor = AppChrome.Accent,
                        contentColor = Color.White,
                    ),
                ) {
                    Text(primaryLabel)
                }
                if (!secondaryLabel.isNullOrBlank() && onSecondaryClick != null) {
                    TextButton(onClick = onSecondaryClick) {
                        Text(
                            text = secondaryLabel,
                            color = AppChrome.TextSecondary,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun AvPosterCard(
    baseUrl: String,
    item: VideoListItemDto,
    onOpenDetail: (String, String) -> Unit,
) {
    val cardModel = buildAvCatalogCardModel(item)
    val thumb = resolveAvPosterUrl(baseUrl, item)

    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppChrome.CardShape)
            .clickable { onOpenDetail(item.id, item.type) },
        color = AppChrome.Surface,
        shape = AppChrome.CardShape,
        tonalElevation = 0.dp,
    ) {
        Column(
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(0.72f)
                    .background(AppChrome.CanvasRaised),
            ) {
                if (!thumb.isNullOrBlank()) {
                    AsyncImage(
                        model = thumb,
                        contentDescription = item.title,
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    Box(
                        modifier = Modifier.fillMaxSize(),
                        contentAlignment = Alignment.Center,
                    ) {
                        Icon(
                            imageVector = Icons.Filled.PlayArrow,
                            contentDescription = null,
                            tint = AppChrome.TextMuted,
                            modifier = Modifier.size(42.dp),
                        )
                    }
                }
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.verticalGradient(
                                colors = listOf(
                                    Color.Transparent,
                                    Color(0x66080A0F),
                                    Color(0xD90A0B0E),
                                ),
                            ),
                        ),
                )
                Surface(
                    modifier = Modifier
                        .align(Alignment.TopStart)
                        .padding(10.dp),
                    color = AppChrome.Surface.copy(alpha = 0.82f),
                    shape = AppChrome.PillShape,
                ) {
                    Text(
                        text = "AV",
                        color = AppChrome.AccentWarm,
                        style = MaterialTheme.typography.labelSmall,
                        fontWeight = FontWeight.Bold,
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                    )
                }
            }

            Column(
                modifier = Modifier.padding(start = 12.dp, end = 12.dp, bottom = 14.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = cardModel.primaryText,
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                    color = AppChrome.TextPrimary,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                cardModel.secondaryText?.let { secondaryText ->
                    Text(
                        text = secondaryText,
                        style = MaterialTheme.typography.bodyMedium,
                        color = AppChrome.TextSecondary,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
                Text(
                    text = cardModel.metaText,
                    style = MaterialTheme.typography.bodySmall,
                    color = AppChrome.TextMuted,
                )
            }
        }
    }
}

@Composable
private fun CategoryListSection(
    baseUrl: String,
    state: CategoryState,
    categoryTitle: String,
    onRetry: () -> Unit,
    onOpenDetail: (String, String) -> Unit,
) {
    when {
        state.loading && state.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator(color = AppChrome.AccentStrong)
            }
        }

        !state.errorMessage.isNullOrBlank() && state.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Surface(
                    color = AppChrome.SurfaceElevated,
                    shape = AppChrome.CardShape,
                ) {
                    Column(
                        modifier = Modifier.padding(horizontal = 20.dp, vertical = 18.dp),
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                        Text(categoryTitle, color = AppChrome.TextMuted, style = MaterialTheme.typography.labelLarge)
                        Text(state.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                        Button(
                            onClick = onRetry,
                            colors = ButtonDefaults.buttonColors(
                                containerColor = AppChrome.Accent,
                                contentColor = Color.White,
                            ),
                        ) {
                            Icon(Icons.Filled.Refresh, contentDescription = null)
                            Text("重试", modifier = Modifier.padding(start = 6.dp))
                        }
                    }
                }
            }
        }

        else -> {
            LazyColumn(
                modifier = Modifier.fillMaxSize(),
                contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 6.dp, bottom = 18.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                item(key = "hero") {
                    Surface(
                        modifier = Modifier.fillMaxWidth(),
                        color = AppChrome.Surface,
                        shape = AppChrome.CardShape,
                    ) {
                        Column(
                            modifier = Modifier.padding(horizontal = 18.dp, vertical = 16.dp),
                            verticalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            Text(
                                text = categoryTitle,
                                style = MaterialTheme.typography.labelLarge,
                                color = AppChrome.AccentWarm,
                                fontWeight = FontWeight.SemiBold,
                            )
                            Text(
                                text = "长内容列表改为统一深色排布，浏览与继续观看都更稳定。",
                                style = MaterialTheme.typography.bodyMedium,
                                color = AppChrome.TextSecondary,
                            )
                        }
                    }
                }
                items(state.items, key = { it.id }) { item ->
                    VideoCard(
                        baseUrl = baseUrl,
                        item = item,
                        onOpenDetail = onOpenDetail,
                    )
                }
            }
        }
    }
}

@Composable
private fun VideoCard(
    baseUrl: String,
    item: VideoListItemDto,
    onOpenDetail: (String, String) -> Unit,
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppChrome.CardShape)
            .clickable { onOpenDetail(item.id, item.type) },
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.CardShape,
        tonalElevation = 0.dp,
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            val thumb = resolveThumbnailUrl(baseUrl, item.thumbnailPath)
            if (!thumb.isNullOrBlank()) {
                AsyncImage(
                    model = thumb,
                    contentDescription = item.title,
                    modifier = Modifier
                        .size(width = 128.dp, height = 78.dp)
                        .clip(RoundedCornerShape(14.dp)),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(
                    modifier = Modifier
                        .size(width = 128.dp, height = 78.dp)
                        .clip(RoundedCornerShape(14.dp))
                        .background(AppChrome.SurfaceStrong),
                    contentAlignment = Alignment.Center,
                ) {
                    Icon(
                        imageVector = Icons.Filled.PlayArrow,
                        contentDescription = null,
                        tint = AppChrome.TextMuted,
                    )
                }
            }

            Column(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text(
                    text = item.title,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    color = AppChrome.TextPrimary,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = "${typeLabel(item.type)} · ${if (item.type == "av") formatDurationHms(item.duration) else "${item.duration} 秒"}",
                    style = MaterialTheme.typography.bodySmall,
                    color = AppChrome.TextMuted,
                )
                Spacer(modifier = Modifier.height(2.dp))
                Surface(
                    color = AppChrome.SurfaceStrong,
                    shape = AppChrome.PillShape,
                ) {
                    Text(
                        text = "打开详情",
                        color = AppChrome.TextSecondary,
                        style = MaterialTheme.typography.labelMedium,
                        modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                    )
                }
            }
        }
    }
}

private fun buildAvStatusText(
    browseState: CategoryState,
    searchState: AvSearchState,
): String {
    val searchLabel = searchState.query.trim().ifBlank { searchState.lastCompletedQuery }
    return when {
        searchState.loading && searchState.query.isNotBlank() -> "正在远程搜索“${searchState.query.trim()}”"
        !searchState.errorMessage.isNullOrBlank() && searchLabel.isNotBlank() -> "搜索“$searchLabel”失败"
        searchState.isSearchMode && searchState.results.isEmpty() && searchLabel.isNotBlank() -> "没有找到“$searchLabel”相关作品"
        searchState.isSearchMode -> "共找到 ${searchState.totalCount.coerceAtLeast(searchState.results.size)} 条结果"
        browseState.loaded -> "默认浏览 ${browseState.items.size} 部作品，直接按番号或标题定位"
        else -> "远程搜索当前服务器里的全部 AV 内容"
    }
}

private fun typeLabel(type: String): String {
    return when (type) {
        "short" -> "短视频"
        "movie" -> "电影"
        "episode" -> "电视剧"
        "av" -> "AV"
        else -> type
    }
}

private fun resolveThumbnailUrl(baseUrl: String, rawPath: String?): String? {
    val path = rawPath?.trim().orEmpty()
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

private fun formatDurationHms(totalSeconds: Int): String {
    val safeTotal = totalSeconds.coerceAtLeast(0)
    val hours = safeTotal / 3600
    val minutes = (safeTotal % 3600) / 60
    val seconds = safeTotal % 60
    return if (hours > 0) {
        "%02d:%02d:%02d".format(hours, minutes, seconds)
    } else {
        "%02d:%02d".format(minutes, seconds)
    }
}
