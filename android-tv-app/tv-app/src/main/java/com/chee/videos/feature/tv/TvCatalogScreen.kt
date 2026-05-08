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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Search
import androidx.compose.material.icons.filled.Star
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
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
import com.chee.videos.core.util.UrlBuilder

@Composable
fun TvCatalogScreen(
    onOpenSeries: (String) -> Unit,
    onOpenContinueWatching: (String, Int, Int) -> Unit,
    onOpenLongForm: (String, String) -> Unit,
    onPlayLongForm: (String, String) -> Unit,
    viewModel: TvCatalogViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val isSearching = uiState.query.isNotBlank()
    val searchFocusRequester = remember { FocusRequester() }
    val featuredFocusRequester = remember { FocusRequester() }
    val continueFocusRequester = remember { FocusRequester() }
    val firstSectionItemFocusRequester = remember { FocusRequester() }
    val tvSeriesFocusRequester = remember { FocusRequester() }
    val movieFocusRequester = remember { FocusRequester() }
    val avFocusRequester = remember { FocusRequester() }
    val featuredContent = resolveTvFeaturedContent(
        continueWatching = uiState.continueWatching,
        sections = uiState.sections,
        tvSeries = uiState.tvSeries,
        movies = uiState.movies,
        av = uiState.av,
    )
    val initialFocusTarget = resolveTvCatalogInitialFocusTarget(
        hasFeaturedContent = featuredContent != null,
        hasContinueWatching = uiState.continueWatching != null,
        sectionItemCounts = uiState.sections.map { it.items.size },
        tvSeriesCount = uiState.tvSeries.size,
        movieCount = uiState.movies.size,
        avCount = uiState.av.size,
    )

    LaunchedEffect(uiState.loading, isSearching, initialFocusTarget) {
        if (uiState.loading || isSearching) return@LaunchedEffect
        when (initialFocusTarget) {
            TvCatalogInitialFocusTarget.FEATURED -> featuredFocusRequester.requestFocus()
            TvCatalogInitialFocusTarget.CONTINUE_WATCHING -> continueFocusRequester.requestFocus()
            TvCatalogInitialFocusTarget.FIRST_SECTION_ITEM -> firstSectionItemFocusRequester.requestFocus()
            TvCatalogInitialFocusTarget.TV_SERIES_ITEM -> tvSeriesFocusRequester.requestFocus()
            TvCatalogInitialFocusTarget.MOVIE_ITEM -> movieFocusRequester.requestFocus()
            TvCatalogInitialFocusTarget.AV_ITEM -> avFocusRequester.requestFocus()
            TvCatalogInitialFocusTarget.SEARCH -> searchFocusRequester.requestFocus()
        }
    }

    if (uiState.loading) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            CircularProgressIndicator(color = AppChrome.AccentStrong)
        }
        return
    }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        contentPadding = PaddingValues(start = 14.dp, end = 14.dp, top = 8.dp, bottom = 20.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        item(key = "search") {
            TvCatalogSearchBar(
                query = uiState.query,
                onQueryChanged = viewModel::updateQuery,
                modifier = Modifier
                    .focusRequester(searchFocusRequester)
                    .tvFocusableGlow(shape = RoundedCornerShape(18.dp), focusedScale = 1.01f),
            )
        }
        featuredContent?.let { featured ->
            item(key = "featured") {
                TvFeaturedHero(
                    baseUrl = uiState.baseUrl,
                    data = featured,
                    primaryModifier = Modifier.focusRequester(featuredFocusRequester),
                    onPrimaryAction = {
                        when {
                            featured.targetType == "tv" && featured.source == TvFeaturedContentSource.CONTINUE_WATCHING -> onOpenContinueWatching(
                                featured.targetId,
                                featured.seasonNumber.coerceAtLeast(1),
                                featured.episodeNumber.coerceAtLeast(1),
                            )

                            featured.targetType == "tv" -> onOpenSeries(featured.targetId)

                            else -> onPlayLongForm(
                                featured.videoId?.ifBlank { featured.targetId } ?: featured.targetId,
                                featured.targetType,
                            )
                        }
                    },
                    onSecondaryAction = {
                        if (featured.targetType == "tv") {
                            onOpenSeries(featured.targetId)
                        } else {
                            onOpenLongForm(featured.targetId, featured.targetType)
                        }
                    },
                )
            }
        }
        uiState.errorMessage?.let { message ->
            item(key = "error") {
                Surface(
                    color = AppChrome.SurfaceElevated,
                    shape = AppChrome.SectionShape,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text(
                        text = message,
                        color = MaterialTheme.colorScheme.error,
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 14.dp),
                    )
                }
            }
        }
        if (isSearching) {
            item(key = "search-header") {
                TvSearchResultHeader(resultCount = uiState.searchResults.size)
            }
            if (uiState.searchResults.isEmpty()) {
                item(key = "search-empty") {
                    TvSearchEmptyState()
                }
            } else {
                items(uiState.searchResults, key = { item -> item.id }) { item ->
                    TvSearchResultCard(
                        baseUrl = uiState.baseUrl,
                        item = item,
                        onClick = {
                            if (item.type == "tv") onOpenSeries(item.id) else onOpenLongForm(item.id, item.type)
                        },
                    )
                }
            }
            return@LazyColumn
        }
        uiState.continueWatching?.takeIf { featuredContent?.source != TvFeaturedContentSource.CONTINUE_WATCHING }?.let { continueWatching ->
            item(key = "continue-watching") {
                TvContinueWatchingBanner(
                    baseUrl = uiState.baseUrl,
                    data = continueWatching,
                    modifier = Modifier
                        .focusRequester(continueFocusRequester)
                        .tvFocusableGlow(shape = AppChrome.CardShape),
                    onClick = {
                        if (continueWatching.type == "tv") {
                            onOpenContinueWatching(
                                continueWatching.seriesId,
                                continueWatching.seasonNumber,
                                continueWatching.episodeNumber,
                            )
                        } else {
                            onPlayLongForm(continueWatching.videoId.orEmpty(), continueWatching.type)
                        }
                    },
                )
            }
        }
        items(uiState.sections, key = { section -> section.title }) { section ->
            TvCatalogSection(
                baseUrl = uiState.baseUrl,
                section = section,
                onOpenSeries = onOpenSeries,
                firstItemFocusRequester = firstSectionItemFocusRequester,
                requestInitialFocus = initialFocusTarget == TvCatalogInitialFocusTarget.FIRST_SECTION_ITEM,
            )
        }
        if (uiState.tvSeries.isNotEmpty()) {
            item(key = "tv-series-shelf") {
                TvHomeShelf(
                    title = "电视剧",
                    subtitle = "全部长剧集",
                    baseUrl = uiState.baseUrl,
                    items = uiState.tvSeries,
                    firstItemFocusRequester = tvSeriesFocusRequester,
                    requestInitialFocus = initialFocusTarget == TvCatalogInitialFocusTarget.TV_SERIES_ITEM,
                    onClick = { item -> onOpenSeries(item.id) },
                )
            }
        }
        if (uiState.movies.isNotEmpty()) {
            item(key = "movies-shelf") {
                TvHomeShelf(
                    title = "电影",
                    subtitle = "可直接播放的长片",
                    baseUrl = uiState.baseUrl,
                    items = uiState.movies,
                    firstItemFocusRequester = movieFocusRequester,
                    requestInitialFocus = initialFocusTarget == TvCatalogInitialFocusTarget.MOVIE_ITEM,
                    onClick = { item -> onOpenLongForm(item.id, item.type) },
                )
            }
        }
        if (uiState.av.isNotEmpty()) {
            item(key = "av-shelf") {
                TvHomeShelf(
                    title = "AV",
                    subtitle = "仅包含长视频",
                    baseUrl = uiState.baseUrl,
                    items = uiState.av,
                    firstItemFocusRequester = avFocusRequester,
                    requestInitialFocus = initialFocusTarget == TvCatalogInitialFocusTarget.AV_ITEM,
                    onClick = { item -> onOpenLongForm(item.id, item.type) },
                )
            }
        }
    }
}

@Composable
private fun TvCatalogSearchBar(
    query: String,
    onQueryChanged: (String) -> Unit,
    modifier: Modifier = Modifier,
) {
    OutlinedTextField(
        value = query,
        onValueChange = onQueryChanged,
        modifier = modifier.fillMaxWidth(),
        singleLine = true,
        placeholder = {
            Text("搜索剧名", color = AppChrome.TextMuted)
        },
        leadingIcon = {
            Icon(Icons.Filled.Search, contentDescription = null, tint = AppChrome.TextMuted)
        },
        trailingIcon = {
            if (query.isNotBlank()) {
                IconButton(onClick = { onQueryChanged("") }) {
                    Icon(Icons.Filled.Close, contentDescription = "清空搜索", tint = AppChrome.TextMuted)
                }
            }
        },
        shape = RoundedCornerShape(18.dp),
        colors = OutlinedTextFieldDefaults.colors(
            focusedTextColor = AppChrome.TextPrimary,
            unfocusedTextColor = AppChrome.TextPrimary,
            focusedBorderColor = AppChrome.AccentStrong,
            unfocusedBorderColor = AppChrome.Divider,
            cursorColor = AppChrome.AccentStrong,
            focusedContainerColor = AppChrome.Surface.copy(alpha = 0.92f),
            unfocusedContainerColor = AppChrome.Surface.copy(alpha = 0.82f),
        ),
    )
}

@Composable
private fun TvSearchResultHeader(resultCount: Int) {
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text(
            text = "搜索结果",
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
        )
        Text(
            text = if (resultCount > 0) "共找到 $resultCount 部剧集" else "没有匹配的剧集",
            color = AppChrome.TextMuted,
            style = MaterialTheme.typography.bodySmall,
        )
    }
}

@Composable
private fun TvSearchEmptyState() {
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.SectionShape,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 18.dp, vertical = 22.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Icon(
                imageVector = Icons.Filled.Search,
                contentDescription = null,
                tint = AppChrome.TextMuted,
                modifier = Modifier.size(24.dp),
            )
            Text(
                text = "没有找到相关剧集",
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
            )
            Text(
                text = "试试输入更完整的剧名关键词",
                color = AppChrome.TextMuted,
                style = MaterialTheme.typography.bodySmall,
            )
        }
    }
}

@Composable
private fun TvSearchResultCard(
    baseUrl: String,
    item: TvSearchResultUiModel,
    onClick: () -> Unit,
) {
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.SectionShape,
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppChrome.SectionShape)
            .tvFocusableGlow(shape = AppChrome.SectionShape, focusedScale = 1.02f)
            .clickable(onClick = onClick),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 12.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TvPosterArtwork(
                title = item.title,
                baseUrl = baseUrl,
                posterUrl = item.posterUrl,
                posterSeed = item.title.hashCode(),
                width = 84.dp,
                iconSize = 26.dp,
            )
            Column(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = item.title,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = tvTypeLabel(item.type),
                    color = AppChrome.TextSecondary,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = item.description,
                    color = AppChrome.TextMuted,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

@Composable
private fun TvFeaturedHero(
    baseUrl: String,
    data: TvFeaturedContentUiModel,
    primaryModifier: Modifier = Modifier,
    onPrimaryAction: () -> Unit,
    onSecondaryAction: () -> Unit,
) {
    val backdropUrl = resolveTvArtworkUrl(baseUrl, data.backdropUrl)
    val posterUrl = resolveTvArtworkUrl(baseUrl, data.posterUrl)

    Surface(
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.CardShape,
        modifier = Modifier
            .fillMaxWidth()
            .height(324.dp),
            shadowElevation = 12.dp,
    ) {
        Box(modifier = Modifier.fillMaxSize()) {
            if (!backdropUrl.isNullOrBlank()) {
                AsyncImage(
                    model = backdropUrl,
                    contentDescription = data.title,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.horizontalGradient(
                                colors = listOf(Color(0xFF211527), Color(0xFF0B101A)),
                            ),
                        ),
                )
            }
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.horizontalGradient(
                            colors = listOf(Color(0xE80B0F17), Color(0xC4111520), Color(0xAA111827)),
                        ),
                    ),
            )
            Row(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(horizontal = 28.dp, vertical = 24.dp),
                horizontalArrangement = Arrangement.spacedBy(24.dp),
                verticalAlignment = Alignment.Bottom,
            ) {
                TvFeaturedPoster(
                    title = data.title,
                    posterUrl = posterUrl,
                    posterSeed = data.title.hashCode(),
                )
                Column(
                    modifier = Modifier.weight(1f),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    Text(
                        text = data.eyebrow,
                        color = AppChrome.AccentWarm,
                        style = MaterialTheme.typography.labelLarge,
                        fontWeight = FontWeight.SemiBold,
                    )
                    Text(
                        text = data.title,
                        color = Color.White,
                        style = MaterialTheme.typography.headlineLarge,
                        fontWeight = FontWeight.Bold,
                        maxLines = 2,
                        overflow = TextOverflow.Ellipsis,
                    )
                    if (data.subtitle.isNotBlank()) {
                        Text(
                            text = data.subtitle,
                            color = AppChrome.TextSecondary,
                            style = MaterialTheme.typography.titleMedium,
                            maxLines = 1,
                            overflow = TextOverflow.Ellipsis,
                        )
                    }
                    Text(
                        text = data.description,
                        color = AppChrome.TextMuted,
                        style = MaterialTheme.typography.bodyMedium,
                        maxLines = 3,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                        TvHeroActionButton(
                            text = if (data.source == TvFeaturedContentSource.CONTINUE_WATCHING) "继续播放" else "立即播放",
                            icon = Icons.Filled.PlayArrow,
                            modifier = primaryModifier,
                            onClick = onPrimaryAction,
                        )
                        TvHeroSecondaryActionButton(
                            text = "查看详情",
                            onClick = onSecondaryAction,
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun TvFeaturedPoster(
    title: String,
    posterUrl: String?,
    posterSeed: Int,
) {
    if (!posterUrl.isNullOrBlank()) {
        AsyncImage(
            model = posterUrl,
            contentDescription = title,
            modifier = Modifier
                .width(186.dp)
                .aspectRatio(2f / 3f)
                .clip(RoundedCornerShape(24.dp)),
            contentScale = ContentScale.Crop,
        )
        return
    }
    Box(
        modifier = Modifier
            .width(186.dp)
            .aspectRatio(2f / 3f)
            .clip(RoundedCornerShape(24.dp))
            .background(tvPosterBrush(posterSeed)),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = Icons.Filled.Tv,
            contentDescription = null,
            tint = Color.White.copy(alpha = 0.84f),
            modifier = Modifier.size(46.dp),
        )
    }
}

@Composable
private fun TvHeroActionButton(
    text: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    Surface(
        color = AppChrome.Accent,
        shape = AppChrome.PillShape,
        modifier = modifier
            .tvFocusableGlow(shape = AppChrome.PillShape, focusedScale = 1.06f)
            .clickable(onClick = onClick),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Icon(icon, contentDescription = null, tint = Color.White)
            Text(text = text, color = Color.White, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
        }
    }
}

@Composable
private fun TvHeroSecondaryActionButton(
    text: String,
    onClick: () -> Unit,
) {
    Surface(
        color = Color(0x33070B13),
        shape = AppChrome.PillShape,
        modifier = Modifier
            .tvFocusableGlow(shape = AppChrome.PillShape, focusedScale = 1.05f)
            .clickable(onClick = onClick),
    ) {
        Text(
            text = text,
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
        )
    }
}

@Composable
private fun TvContinueWatchingBanner(
    baseUrl: String,
    data: TvContinueWatchingUiModel,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    val artworkUrl = resolveTvArtworkUrl(baseUrl, data.backdropUrl)
        ?: resolveTvArtworkUrl(baseUrl, data.posterUrl)
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.CardShape,
        modifier = modifier
            .fillMaxWidth()
            .clip(AppChrome.CardShape)
            .clickable(onClick = onClick),
    ) {
        Box(modifier = Modifier.fillMaxWidth()) {
            if (!artworkUrl.isNullOrBlank()) {
                AsyncImage(
                    model = artworkUrl,
                    contentDescription = data.seriesTitle,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            } else {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.horizontalGradient(
                                colors = listOf(Color(0xFF1A2031), Color(0xFF111827)),
                            ),
                        ),
                )
            }
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(
                        Brush.horizontalGradient(
                            colors = listOf(Color(0xE61A2031), Color(0xF0111827)),
                        ),
                    ),
            )
            Column(
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 14.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Text(
                    text = when (data.type) {
                        "movie" -> "继续看电影"
                        "av" -> "继续播放 AV"
                        else -> "继续追剧"
                    },
                    color = AppChrome.AccentWarm,
                    style = MaterialTheme.typography.labelLarge,
                    fontWeight = FontWeight.SemiBold,
                )
                Text(
                    text = data.seriesTitle,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleLarge,
                    fontWeight = FontWeight.Bold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = when (data.type) {
                        "movie", "av" -> data.episodeTitle.ifBlank { "继续播放" }
                        else -> "S${data.seasonNumber} · E${data.episodeNumber}  ${data.episodeTitle}"
                    },
                    color = AppChrome.TextSecondary,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                ) {
                    Surface(color = AppChrome.Accent, shape = AppChrome.PillShape) {
                        Row(
                            modifier = Modifier.padding(horizontal = 10.dp, vertical = 5.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(4.dp),
                        ) {
                            Icon(
                                imageVector = Icons.Filled.PlayArrow,
                                contentDescription = null,
                                tint = Color.White,
                                modifier = Modifier.size(14.dp),
                            )
                            Text("继续播放", color = Color.White, style = MaterialTheme.typography.labelMedium)
                        }
                    }
                    Text(
                        text = "已观看 ${data.progressPercent.coerceIn(0, 100)}%",
                        color = AppChrome.TextMuted,
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
            }
        }
    }
}

@Composable
private fun TvCatalogSection(
    baseUrl: String,
    section: TvCatalogSectionUiModel,
    onOpenSeries: (String) -> Unit,
    firstItemFocusRequester: FocusRequester,
    requestInitialFocus: Boolean,
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            text = section.title,
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
        )
        Text(
            text = section.subtitle,
            color = AppChrome.TextMuted,
            style = MaterialTheme.typography.bodySmall,
        )
        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            contentPadding = PaddingValues(top = 4.dp, bottom = 2.dp),
        ) {
            items(section.items, key = { item -> item.id }) { series ->
                TvSeriesPosterCard(
                    baseUrl = baseUrl,
                    series = series,
                    modifier = if (requestInitialFocus && section.items.firstOrNull()?.id == series.id) {
                        Modifier.focusRequester(firstItemFocusRequester).tvFocusableGlow(shape = RoundedCornerShape(16.dp))
                    } else {
                        Modifier.tvFocusableGlow(shape = RoundedCornerShape(16.dp))
                    },
                    onClick = { onOpenSeries(series.id) },
                )
            }
        }
    }
}

@Composable
private fun TvSeriesPosterCard(
    baseUrl: String,
    series: TvSeriesUiModel,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(16.dp),
        modifier = modifier
            .size(width = 146.dp, height = 256.dp)
            .clip(RoundedCornerShape(16.dp))
            .clickable(onClick = onClick),
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            TvPosterArtwork(
                title = series.title,
                baseUrl = baseUrl,
                posterUrl = series.posterUrl,
                posterSeed = series.posterSeed,
                width = 146.dp,
                iconSize = 34.dp,
            )
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = 10.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = series.title,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = series.updateText,
                    color = AppChrome.TextMuted,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                ) {
                    Icon(
                        imageVector = Icons.Filled.Star,
                        contentDescription = null,
                        tint = AppChrome.AccentWarm,
                        modifier = Modifier.size(14.dp),
                    )
                    Text(
                        text = series.ratingText,
                        color = AppChrome.TextSecondary,
                        style = MaterialTheme.typography.labelMedium,
                    )
                }
            }
        }
    }
}

@Composable
private fun TvHomeShelf(
    title: String,
    subtitle: String,
    baseUrl: String,
    items: List<TvHomeShelfItemUiModel>,
    firstItemFocusRequester: FocusRequester,
    requestInitialFocus: Boolean,
    onClick: (TvHomeShelfItemUiModel) -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            text = title,
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.titleMedium,
            fontWeight = FontWeight.SemiBold,
        )
        Text(
            text = subtitle,
            color = AppChrome.TextMuted,
            style = MaterialTheme.typography.bodySmall,
        )
        LazyRow(
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            contentPadding = PaddingValues(top = 4.dp, bottom = 2.dp),
        ) {
            items(items, key = { item -> "${item.type}-${item.id}" }) { item ->
                TvHomeShelfCard(
                    baseUrl = baseUrl,
                    item = item,
                    modifier = if (requestInitialFocus && items.firstOrNull()?.id == item.id) {
                        Modifier.focusRequester(firstItemFocusRequester).tvFocusableGlow(shape = RoundedCornerShape(16.dp))
                    } else {
                        Modifier.tvFocusableGlow(shape = RoundedCornerShape(16.dp))
                    },
                    onClick = { onClick(item) },
                )
            }
        }
    }
}

@Composable
private fun TvHomeShelfCard(
    baseUrl: String,
    item: TvHomeShelfItemUiModel,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(16.dp),
        modifier = modifier
            .size(width = 146.dp, height = 256.dp)
            .clip(RoundedCornerShape(16.dp))
            .clickable(onClick = onClick),
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            TvPosterArtwork(
                title = item.title,
                baseUrl = baseUrl,
                posterUrl = item.posterUrl,
                posterSeed = item.title.hashCode(),
                width = 146.dp,
                iconSize = 34.dp,
            )
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 10.dp, vertical = 10.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = item.title,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = tvTypeLabel(item.type),
                    color = AppChrome.TextSecondary,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                if (item.progressPercent > 0) {
                    Text(
                        text = "已观看 ${item.progressPercent.coerceIn(0, 100)}%",
                        color = AppChrome.TextMuted,
                        style = MaterialTheme.typography.labelMedium,
                    )
                }
            }
        }
    }
}

@Composable
private fun TvPosterArtwork(
    title: String,
    baseUrl: String,
    posterUrl: String?,
    posterSeed: Int,
    width: androidx.compose.ui.unit.Dp,
    iconSize: androidx.compose.ui.unit.Dp,
) {
    val resolvedPosterUrl = resolveTvArtworkUrl(baseUrl, posterUrl)
    if (!resolvedPosterUrl.isNullOrBlank()) {
        AsyncImage(
            model = resolvedPosterUrl,
            contentDescription = title,
            modifier = Modifier
                .width(width)
                .aspectRatio(2f / 3f)
                .clip(RoundedCornerShape(14.dp)),
            contentScale = ContentScale.Crop,
        )
        return
    }
    Box(
        modifier = Modifier
            .width(width)
            .aspectRatio(2f / 3f)
            .clip(RoundedCornerShape(14.dp))
            .background(tvPosterBrush(posterSeed)),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = Icons.Filled.Tv,
            contentDescription = null,
            tint = Color.White.copy(alpha = 0.82f),
            modifier = Modifier.size(iconSize),
        )
    }
}

private fun tvPosterBrush(seed: Int): Brush {
    return when (seed % 5) {
        0 -> Brush.verticalGradient(listOf(Color(0xFF2D1A48), Color(0xFF0B1220)))
        1 -> Brush.verticalGradient(listOf(Color(0xFF3C1D2E), Color(0xFF111827)))
        2 -> Brush.verticalGradient(listOf(Color(0xFF1F3A56), Color(0xFF111827)))
        3 -> Brush.verticalGradient(listOf(Color(0xFF3D3420), Color(0xFF121826)))
        else -> Brush.verticalGradient(listOf(Color(0xFF1C3D36), Color(0xFF101820)))
    }
}

private fun resolveTvArtworkUrl(baseUrl: String, rawUrl: String?): String? {
    val path = rawUrl?.trim().orEmpty()
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

private fun tvTypeLabel(type: String): String = when (type) {
    "tv" -> "电视剧"
    "movie" -> "电影"
    "av" -> "AV"
    else -> type.ifBlank { "长视频" }
}

internal enum class TvCatalogInitialFocusTarget {
    FEATURED,
    CONTINUE_WATCHING,
    FIRST_SECTION_ITEM,
    TV_SERIES_ITEM,
    MOVIE_ITEM,
    AV_ITEM,
    SEARCH,
}

internal fun resolveTvCatalogInitialFocusTarget(
    hasFeaturedContent: Boolean,
    hasContinueWatching: Boolean,
    sectionItemCounts: List<Int>,
    tvSeriesCount: Int,
    movieCount: Int,
    avCount: Int,
): TvCatalogInitialFocusTarget {
    if (hasFeaturedContent) {
        return TvCatalogInitialFocusTarget.FEATURED
    }
    if (hasContinueWatching) {
        return TvCatalogInitialFocusTarget.CONTINUE_WATCHING
    }
    if (sectionItemCounts.any { it > 0 }) {
        return TvCatalogInitialFocusTarget.FIRST_SECTION_ITEM
    }
    if (tvSeriesCount > 0) {
        return TvCatalogInitialFocusTarget.TV_SERIES_ITEM
    }
    if (movieCount > 0) {
        return TvCatalogInitialFocusTarget.MOVIE_ITEM
    }
    if (avCount > 0) {
        return TvCatalogInitialFocusTarget.AV_ITEM
    }
    return TvCatalogInitialFocusTarget.SEARCH
}
