package com.chee.videos.feature.tv

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
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
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
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
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvIconActionButton
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableGlow
import com.chee.videos.core.ui.tvSharedSeriesPoster

private object TvSeriesDetailTokens {
    val EpisodeCardShape = RoundedCornerShape(8.dp)
    val PrimaryActionShape = RoundedCornerShape(8.dp)
    val SeasonChipShape = RoundedCornerShape(8.dp)
    val QualityBadgeShape = RoundedCornerShape(8.dp)
    val MetaChipShape = AppChrome.ChipShape
    const val EpisodePaneWidthDp = 620
    const val EpisodeThumbWidthDp = 216
    const val EpisodeCardHeightDp = 136
}

@Composable
fun TvSeriesDetailScreen(
    onBack: () -> Unit,
    onPlayEpisode: (seriesId: String, season: Int, episode: Int) -> Unit,
    viewModel: TvSeriesDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val playFocusRequester = remember { FocusRequester() }

    if (uiState.loading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(AppChrome.PageGradient)
                .statusBarsPadding(),
        ) {
            TvPageLoadingState(message = "正在加载电视剧详情")
        }
        return
    }
    val series = uiState.series
    if (series == null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(AppChrome.PageGradient)
                .statusBarsPadding(),
        ) {
            TvErrorState(
                message = uiState.errorMessage ?: "电视剧详情加载失败",
                onAction = viewModel::retry,
            )
        }
        return
    }

    val season = selectedDetailSeason(uiState)
    val currentEpisode = selectedDetailEpisode(uiState)
    val episodes = season?.episodes.orEmpty()
    val backdropUrl = resolveTvResourceUrl(uiState.baseUrl, series.backdropUrl)
    val posterUrl = resolveTvResourceUrl(uiState.baseUrl, series.posterUrl)

    LaunchedTvInitialFocus(series.id) {
        playFocusRequester.tryRequestFocus()
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.Canvas),
    ) {
        TvSeriesDetailBackdrop(
            series = series,
            backdropUrl = backdropUrl,
            posterUrl = posterUrl,
        )
        TvSeriesDetailScrim()

        TvIconActionButton(
            icon = Icons.AutoMirrored.Filled.ArrowBack,
            contentDescription = "返回",
            onClick = onBack,
            modifier = Modifier
                .align(Alignment.TopStart)
                .statusBarsPadding()
                .padding(start = 34.dp, top = 28.dp),
            size = 52.dp,
            iconSize = 24.dp,
            containerColor = Color(0x66070A10),
            contentColor = Color.White,
            focusedScale = 1.08f,
        )

        TvSeriesQualityBadge(
            modifier = Modifier
                .align(Alignment.TopEnd)
                .statusBarsPadding()
                .padding(top = 30.dp, end = 42.dp),
        )

        Row(
            modifier = Modifier
                .fillMaxSize()
                .statusBarsPadding()
                .navigationBarsPadding()
                .padding(start = 92.dp, top = 96.dp, end = 48.dp, bottom = 50.dp),
            horizontalArrangement = Arrangement.spacedBy(50.dp),
            verticalAlignment = Alignment.Top,
        ) {
            TvSeriesHeroPane(
                series = series,
                currentEpisode = currentEpisode,
                selectedSeasonNumber = uiState.selectedSeasonNumber,
                playFocusRequester = playFocusRequester,
                modifier = Modifier
                    .weight(1f)
                    .fillMaxHeight(),
                onPlay = {
                    val episode = currentEpisode ?: return@TvSeriesHeroPane
                    if (episode.playable) {
                        onPlayEpisode(series.id, uiState.selectedSeasonNumber, uiState.selectedEpisodeNumber)
                    }
                },
            )

            TvSeriesEpisodePane(
                series = series,
                season = season,
                episodes = episodes,
                baseUrl = uiState.baseUrl,
                selectedSeasonNumber = uiState.selectedSeasonNumber,
                selectedEpisodeNumber = uiState.selectedEpisodeNumber,
                modifier = Modifier
                    .width(TvSeriesDetailTokens.EpisodePaneWidthDp.dp)
                    .fillMaxHeight(),
                onSelectSeason = viewModel::selectSeason,
                onPlayEpisode = { episode ->
                    viewModel.selectEpisode(episode.number)
                    onPlayEpisode(series.id, uiState.selectedSeasonNumber, episode.number)
                },
            )
        }
    }
}

@Composable
private fun TvSeriesDetailBackdrop(
    series: TvSeriesUiModel,
    backdropUrl: String?,
    posterUrl: String?,
) {
    when {
        !backdropUrl.isNullOrBlank() -> {
            AsyncImage(
                model = backdropUrl,
                contentDescription = series.title,
                modifier = Modifier
                    .fillMaxSize()
                    .tvSharedSeriesPoster(series.id),
                contentScale = ContentScale.Crop,
            )
        }

        !posterUrl.isNullOrBlank() -> {
            AsyncImage(
                model = posterUrl,
                contentDescription = series.title,
                modifier = Modifier
                    .fillMaxSize()
                    .tvSharedSeriesPoster(series.id),
                contentScale = ContentScale.Crop,
            )
        }

        else -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .tvSharedSeriesPoster(series.id)
                    .background(tvBackdropBrush(series.posterSeed)),
            )
        }
    }
}

@Composable
private fun TvSeriesDetailScrim() {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.horizontalGradient(
                    colorStops = arrayOf(
                        0.00f to Color(0xF704070B),
                        0.30f to Color(0xD8060A10),
                        0.58f to Color(0x4D080A0F),
                        0.74f to Color(0xC006080C),
                        1.00f to Color(0xFA040508),
                    ),
                ),
            ),
    )
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colorStops = arrayOf(
                        0.00f to Color(0x8A020407),
                        0.52f to Color.Transparent,
                        1.00f to Color(0xF8040508),
                    ),
                ),
            ),
    )
}

@Composable
private fun TvSeriesQualityBadge(modifier: Modifier = Modifier) {
    Surface(
        modifier = modifier,
        color = Color(0x26000000),
        shape = TvSeriesDetailTokens.QualityBadgeShape,
        border = BorderStroke(1.dp, AppChrome.AccentWarm.copy(alpha = 0.78f)),
    ) {
        Text(
            text = "4K",
            color = AppChrome.AccentWarm,
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.Black,
            modifier = Modifier.padding(horizontal = 11.dp, vertical = 4.dp),
        )
    }
}

@Composable
private fun TvSeriesHeroPane(
    series: TvSeriesUiModel,
    currentEpisode: TvEpisodeUiModel?,
    selectedSeasonNumber: Int,
    playFocusRequester: FocusRequester,
    modifier: Modifier = Modifier,
    onPlay: () -> Unit,
) {
    Column(
        modifier = modifier.padding(top = 48.dp),
        verticalArrangement = Arrangement.spacedBy(18.dp),
    ) {
        TvSeriesTitleBlock(title = series.title)
        TvSeriesMetaRow(series = series)
        TvSeriesRatingRow(series = series)
        Text(
            text = series.description,
            color = Color(0xDDE7ECF5),
            style = MaterialTheme.typography.bodyLarge,
            lineHeight = 28.sp,
            maxLines = 3,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.width(640.dp),
        )
        TvSeriesReferenceActionRow(
            primaryText = buildTvSeriesPlayButtonLabel(currentEpisode),
            enabled = currentEpisode?.playable == true,
            seasonNumber = selectedSeasonNumber,
            episode = currentEpisode,
            playFocusRequester = playFocusRequester,
            onPlay = onPlay,
        )
        if (series.cast.isNotEmpty()) {
            Spacer(modifier = Modifier.height(18.dp))
            TvSeriesCastRow(cast = series.cast)
        }
    }
}

@Composable
private fun TvSeriesTitleBlock(title: String) {
    Column(
        verticalArrangement = Arrangement.spacedBy(2.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.width(650.dp),
    ) {
        Text(
            text = "剧 集",
            color = Color.White.copy(alpha = 0.88f),
            fontSize = 30.sp,
            lineHeight = 36.sp,
            letterSpacing = 14.sp,
            fontWeight = FontWeight.Bold,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            textAlign = TextAlign.Center,
        )
        Text(
            text = formatTvSeriesReferenceTitle(title),
            color = Color.White,
            fontSize = 76.sp,
            lineHeight = 86.sp,
            letterSpacing = 9.sp,
            fontWeight = FontWeight.Black,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
private fun TvSeriesMetaRow(series: TvSeriesUiModel) {
    val items = remember(series) { buildTvSeriesDetailMetaItems(series) }
    Row(
        horizontalArrangement = Arrangement.spacedBy(10.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        items.forEachIndexed { index, item ->
            TvSeriesMetaChip(
                text = item,
                emphasized = index == 0,
            )
        }
    }
}

@Composable
private fun TvSeriesMetaChip(
    text: String,
    emphasized: Boolean,
) {
    Surface(
        color = if (emphasized) Color(0x29E9BE62) else Color(0x26000000),
        shape = TvSeriesDetailTokens.MetaChipShape,
        border = BorderStroke(
            1.dp,
            if (emphasized) AppChrome.AccentWarm.copy(alpha = 0.48f) else Color.White.copy(alpha = 0.12f),
        ),
    ) {
        Text(
            text = text,
            color = if (emphasized) Color(0xFFFFE1A0) else Color(0xD6E8EDF6),
            style = MaterialTheme.typography.labelLarge,
            fontWeight = if (emphasized) FontWeight.SemiBold else FontWeight.Normal,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 6.dp),
        )
    }
}

@Composable
private fun TvSeriesRatingRow(series: TvSeriesUiModel) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(12.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(3.dp)) {
            repeat(5) { index ->
                Icon(
                    imageVector = Icons.Filled.Star,
                    contentDescription = null,
                    modifier = Modifier.size(25.dp),
                    tint = AppChrome.AccentWarm.copy(alpha = if (index < 4) 1f else 0.48f),
                )
            }
        }
        Surface(
            color = Color(0x28E9BE62),
            shape = TvSeriesDetailTokens.MetaChipShape,
            border = BorderStroke(1.dp, AppChrome.AccentWarm.copy(alpha = 0.44f)),
        ) {
            Text(
                text = buildTvSeriesAvailabilityText(series),
                color = Color(0xFFFFD783),
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.Black,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.padding(horizontal = 12.dp, vertical = 5.dp),
            )
        }
    }
}

@Composable
private fun TvSeriesReferenceActionRow(
    primaryText: String,
    enabled: Boolean,
    seasonNumber: Int,
    episode: TvEpisodeUiModel?,
    playFocusRequester: FocusRequester,
    onPlay: () -> Unit,
) {
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        Row(
            horizontalArrangement = Arrangement.spacedBy(30.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TvSeriesPrimaryActionButton(
                text = primaryText,
                enabled = enabled,
                modifier = Modifier.focusRequester(playFocusRequester),
                onClick = onPlay,
            )
            TvSeriesSecondaryActionButton(text = "我的片单")
        }
        Text(
            text = buildTvSeriesSelectedEpisodeStatus(
                seasonNumber = seasonNumber,
                episode = episode,
            ),
            color = Color(0xA9D5DCE8),
            style = MaterialTheme.typography.bodySmall,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

@Composable
private fun TvSeriesPrimaryActionButton(
    text: String,
    enabled: Boolean,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    Surface(
        color = if (enabled) Color(0xFFEFC463) else Color(0xFF2A2F36),
        shape = TvSeriesDetailTokens.PrimaryActionShape,
        shadowElevation = if (enabled) 14.dp else 0.dp,
        modifier = modifier
            .width(306.dp)
            .tvFocusableGlow(
                enabled = enabled,
                shape = TvSeriesDetailTokens.PrimaryActionShape,
                focusedScale = 1.04f,
            )
            .clickable(enabled = enabled, onClick = onClick),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 24.dp, vertical = 17.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(18.dp),
        ) {
            Box(
                modifier = Modifier
                    .size(42.dp)
                    .clip(CircleShape)
                    .background(Color(0xFF06080B)),
                contentAlignment = Alignment.Center,
            ) {
                Icon(
                    imageVector = Icons.Filled.PlayArrow,
                    contentDescription = null,
                    tint = Color.White,
                    modifier = Modifier.size(28.dp),
                )
            }
            Text(
                text = text,
                color = if (enabled) Color(0xFF070A0D) else Color(0xFFBAC2CE),
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Black,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun TvSeriesSecondaryActionButton(text: String) {
    Surface(
        color = Color(0x2C141820),
        shape = TvSeriesDetailTokens.PrimaryActionShape,
        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.16f)),
        modifier = Modifier
            .width(178.dp)
            .tvFocusableGlow(shape = TvSeriesDetailTokens.PrimaryActionShape, focusedScale = 1.04f)
            .clickable(onClick = {}),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 22.dp, vertical = 19.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Icon(
                imageVector = Icons.Filled.Add,
                contentDescription = null,
                tint = Color.White.copy(alpha = 0.9f),
                modifier = Modifier.size(24.dp),
            )
            Text(
                text = text,
                color = Color.White.copy(alpha = 0.92f),
                style = MaterialTheme.typography.titleSmall,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun TvSeriesCastRow(cast: List<String>) {
    Column(verticalArrangement = Arrangement.spacedBy(18.dp)) {
        Text(
            text = "演员",
            color = Color(0xBFE8EDF6),
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(28.dp)) {
            cast.take(5).forEachIndexed { index, name ->
                TvSeriesCastItem(index = index, name = name)
            }
        }
    }
}

@Composable
private fun TvSeriesCastItem(index: Int, name: String) {
    Column(
        modifier = Modifier.width(104.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Box(
            modifier = Modifier
                .size(76.dp)
                .clip(CircleShape)
                .background(tvCastAvatarBrush(index)),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = name.take(1).ifBlank { "演" },
                color = Color.White,
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold,
            )
        }
        Text(
            text = name,
            color = Color(0xFFECD38F),
            style = MaterialTheme.typography.labelLarge,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Text(
            text = "主演",
            color = Color(0xA9D5DCE8),
            style = MaterialTheme.typography.labelMedium,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

@Composable
private fun TvSeriesEpisodePane(
    series: TvSeriesUiModel,
    season: TvSeasonUiModel?,
    episodes: List<TvEpisodeUiModel>,
    baseUrl: String,
    selectedSeasonNumber: Int,
    selectedEpisodeNumber: Int,
    modifier: Modifier = Modifier,
    onSelectSeason: (Int) -> Unit,
    onPlayEpisode: (TvEpisodeUiModel) -> Unit,
) {
    Column(
        modifier = modifier.padding(top = 32.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = "剧集",
                color = AppChrome.AccentWarm,
                fontSize = 32.sp,
                lineHeight = 38.sp,
                letterSpacing = 3.sp,
                fontWeight = FontWeight.Black,
            )
            Row(
                modifier = Modifier.horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                series.seasons.forEach { item ->
                    TvSeriesSeasonChip(
                        season = item,
                        selected = item.number == selectedSeasonNumber,
                        onClick = { onSelectSeason(item.number) },
                    )
                }
            }
        }

        Text(
            text = season?.title?.ifBlank { "第 $selectedSeasonNumber 季" } ?: "暂无季信息",
            color = Color(0xBFE8EDF6),
            style = MaterialTheme.typography.titleSmall,
            letterSpacing = 1.sp,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )

        if (episodes.isEmpty()) {
            TvSeriesEmptyEpisodes()
            return@Column
        }

        LazyColumn(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(16.dp),
            contentPadding = PaddingValues(bottom = 10.dp),
        ) {
            items(episodes, key = { it.id }) { episode ->
                TvSeriesEpisodeCard(
                    episode = episode,
                    baseUrl = baseUrl,
                    selected = episode.number == selectedEpisodeNumber,
                    onClick = { onPlayEpisode(episode) },
                )
            }
        }
    }
}

@Composable
private fun TvSeriesSeasonChip(
    season: TvSeasonUiModel,
    selected: Boolean,
    onClick: () -> Unit,
) {
    Surface(
        color = if (selected) Color(0x42E9BE62) else Color(0x22000000),
        shape = TvSeriesDetailTokens.SeasonChipShape,
        border = BorderStroke(
            1.dp,
            if (selected) AppChrome.AccentWarm.copy(alpha = 0.68f) else Color.White.copy(alpha = 0.14f),
        ),
        modifier = Modifier
            .tvFocusableGlow(shape = TvSeriesDetailTokens.SeasonChipShape, focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Text(
            text = season.title.ifBlank { "第 ${season.number} 季" },
            color = if (selected) Color.White else Color(0xBEDFE6F0),
            style = MaterialTheme.typography.labelLarge,
            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 9.dp),
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

@Composable
private fun TvSeriesEpisodeCard(
    episode: TvEpisodeUiModel,
    baseUrl: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    val stillUrl = remember(baseUrl, episode.stillUrl) {
        resolveTvResourceUrl(baseUrl, episode.stillUrl)
    }
    Surface(
        color = if (selected) Color(0x4DE9BE62) else Color(0x8A10161F),
        shape = TvSeriesDetailTokens.EpisodeCardShape,
        border = BorderStroke(
            width = if (selected) 2.dp else 1.dp,
            color = if (selected) AppChrome.AccentWarm else Color.White.copy(alpha = 0.12f),
        ),
        modifier = Modifier
            .fillMaxWidth()
            .height(TvSeriesDetailTokens.EpisodeCardHeightDp.dp)
            .tvFocusableGlow(
                enabled = episode.playable,
                shape = TvSeriesDetailTokens.EpisodeCardShape,
                focusedScale = 1.025f,
            )
            .clickable(enabled = episode.playable, onClick = onClick),
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TvSeriesEpisodeStill(
                episode = episode,
                stillUrl = stillUrl,
                selected = selected,
                playable = episode.playable,
            )
            Column(
                modifier = Modifier
                    .weight(1f)
                    .padding(horizontal = 18.dp, vertical = 14.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = "第 ${episode.number} 集  ${episode.title.ifBlank { "未命名" }}",
                    color = if (episode.playable) Color.White else Color(0x8FE8EDF6),
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Black,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = episode.durationLabel,
                    color = if (selected) Color(0xFFFFE2A0) else Color(0xBBC7CFDA),
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = buildTvEpisodeCardSummary(episode),
                    color = Color(0xA9D5DCE8),
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
            Box(
                modifier = Modifier
                    .padding(end = 18.dp)
                    .size(46.dp)
                    .clip(CircleShape)
                    .background(
                        if (selected && episode.playable) {
                            AppChrome.AccentWarm
                        } else if (episode.playable) {
                            Color(0x26E9BE62)
                        } else {
                            Color(0x1FFFFFFF)
                        },
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Icon(
                    imageVector = Icons.Filled.PlayArrow,
                    contentDescription = null,
                    tint = when {
                        selected && episode.playable -> Color(0xFF080A0D)
                        episode.playable -> Color(0xFFE9BE62)
                        else -> Color(0x80FFFFFF)
                    },
                    modifier = Modifier.size(28.dp),
                )
            }
        }
    }
}

@Composable
private fun TvSeriesEpisodeStill(
    episode: TvEpisodeUiModel,
    stillUrl: String?,
    selected: Boolean,
    playable: Boolean,
) {
    Box(
        modifier = Modifier
            .width(TvSeriesDetailTokens.EpisodeThumbWidthDp.dp)
            .fillMaxHeight()
            .background(tvEpisodeStillBrush(episode.number)),
        contentAlignment = Alignment.Center,
    ) {
        if (!stillUrl.isNullOrBlank()) {
            AsyncImage(
                model = stillUrl,
                contentDescription = episode.title,
                modifier = Modifier.fillMaxSize(),
                contentScale = ContentScale.Crop,
            )
        }
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.horizontalGradient(
                        colors = listOf(Color.Transparent, Color(0xAA06080C)),
                    ),
                ),
        )
        if (stillUrl.isNullOrBlank()) {
            Text(
                text = "第 ${episode.number} 集",
                color = Color.White.copy(alpha = if (selected) 0.92f else 0.72f),
                style = MaterialTheme.typography.labelLarge,
                fontWeight = FontWeight.SemiBold,
            )
        }
        Surface(
            color = if (selected && playable) AppChrome.AccentWarm else Color(0x7A05070A),
            shape = CircleShape,
            border = BorderStroke(
                1.dp,
                if (selected) AppChrome.AccentWarm else Color.White.copy(alpha = 0.18f),
            ),
            modifier = Modifier
                .align(Alignment.BottomStart)
                .padding(start = 14.dp, bottom = 12.dp),
        ) {
            Row(
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.spacedBy(5.dp),
                modifier = Modifier.padding(start = 8.dp, top = 5.dp, end = 10.dp, bottom = 5.dp),
            ) {
                Icon(
                    imageVector = Icons.Filled.PlayArrow,
                    contentDescription = null,
                    tint = if (selected && playable) Color(0xFF080A0D) else Color.White.copy(alpha = 0.9f),
                    modifier = Modifier.size(16.dp),
                )
                Text(
                    text = episode.number.toString().padStart(2, '0'),
                    color = if (selected && playable) Color(0xFF080A0D) else Color.White.copy(alpha = 0.88f),
                    style = MaterialTheme.typography.labelMedium,
                    fontWeight = FontWeight.Black,
                )
            }
        }
    }
}

@Composable
private fun TvSeriesEmptyEpisodes() {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = Color(0x7010161F),
        shape = TvSeriesDetailTokens.EpisodeCardShape,
        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.12f)),
    ) {
        Text(
            text = "暂无可显示剧集",
            color = Color(0xBEDFE6F0),
            style = MaterialTheme.typography.titleSmall,
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 22.dp),
        )
    }
}

private fun buildTvSeriesDetailMetaItems(series: TvSeriesUiModel): List<String> {
    val items = mutableListOf<String>()
    series.subtitle
        .split("·")
        .map { it.trim() }
        .filter { it.isNotBlank() }
        .take(3)
        .forEach(items::add)
    if (series.tags.isNotEmpty()) {
        items += series.tags.take(2)
    }
    items += if (series.playableEpisodes > 0) {
        "${series.playableEpisodes} 集可播"
    } else {
        "待绑定视频"
    }
    return items
}

private fun formatTvSeriesReferenceTitle(title: String): String {
    val normalized = title.trim()
    if (normalized.isBlank()) {
        return "未 命 名 剧 集"
    }
    val withoutExtraSpaces = normalized.replace(Regex("\\s+"), " ")
    return withoutExtraSpaces.map { char ->
        if (char == ' ') "   " else char.toString()
    }.joinToString(" ")
}

private fun buildTvSeriesAvailabilityText(series: TvSeriesUiModel): String {
    return if (series.playableEpisodes > 0 && series.totalEpisodes > 0) {
        "${series.playableEpisodes}/${series.totalEpisodes} 集可播"
    } else {
        series.ratingText
    }
}

private fun buildTvSeriesPlayButtonLabel(episode: TvEpisodeUiModel?): String {
    return if (episode?.playable == true) {
        "播放第 ${episode.number} 集"
    } else {
        "暂无片源"
    }
}

private fun buildTvSeriesSelectedEpisodeStatus(
    seasonNumber: Int,
    episode: TvEpisodeUiModel?,
): String {
    if (episode == null) {
        return "当前未选择分集"
    }
    val base = "第 $seasonNumber 季 · 第 ${episode.number} 集"
    return when {
        !episode.playable -> "$base · 待绑定"
        episode.progressPercent > 0 -> "$base · 已观看 ${episode.progressPercent}%"
        else -> base
    }
}

private fun buildTvEpisodeCardSummary(episode: TvEpisodeUiModel): String {
    return when {
        !episode.playable -> "待绑定 / 未就绪"
        episode.summary.isNotBlank() -> episode.summary
        else -> "暂无剧情简介"
    }
}

private fun tvBackdropBrush(seed: Int): Brush {
    return when (seed % 5) {
        0 -> Brush.verticalGradient(listOf(Color(0xFF1C1532), Color(0xFF06080F)))
        1 -> Brush.verticalGradient(listOf(Color(0xFF2A1D2F), Color(0xFF060A13)))
        2 -> Brush.verticalGradient(listOf(Color(0xFF1A3142), Color(0xFF060B12)))
        3 -> Brush.verticalGradient(listOf(Color(0xFF2D2A1A), Color(0xFF070A11)))
        else -> Brush.verticalGradient(listOf(Color(0xFF19372E), Color(0xFF060A10)))
    }
}

private fun tvCastAvatarBrush(index: Int): Brush {
    val colors = listOf(
        Color(0xFF6D4B24) to Color(0xFF20150B),
        Color(0xFF314958) to Color(0xFF0D151C),
        Color(0xFF4A344D) to Color(0xFF17101B),
        Color(0xFF4C4230) to Color(0xFF15120D),
        Color(0xFF2E4A3B) to Color(0xFF0D1711),
    )
    val pair = colors[index % colors.size]
    return Brush.verticalGradient(listOf(pair.first, pair.second))
}

private fun tvEpisodeStillBrush(number: Int): Brush {
    return when (number % 5) {
        0 -> Brush.horizontalGradient(listOf(Color(0xFF17222F), Color(0xFF080B10)))
        1 -> Brush.horizontalGradient(listOf(Color(0xFF2B2E25), Color(0xFF080B10)))
        2 -> Brush.horizontalGradient(listOf(Color(0xFF24322D), Color(0xFF080B10)))
        3 -> Brush.horizontalGradient(listOf(Color(0xFF2F2731), Color(0xFF080B10)))
        else -> Brush.horizontalGradient(listOf(Color(0xFF1E303A), Color(0xFF080B10)))
    }
}
