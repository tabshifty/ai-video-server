package com.chee.videos.feature.tv

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.KeyboardArrowDown
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.Icon
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
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
import com.chee.videos.core.ui.tvSharedSeriesPoster

private object TvSeriesDetailTokens {
    val EpisodeCardShape = RoundedCornerShape(8.dp)
    val PrimaryActionShape = RoundedCornerShape(8.dp)
    val SeasonChipShape = RoundedCornerShape(8.dp)
    val QualityBadgeShape = RoundedCornerShape(8.dp)
    val MetaChipShape = AppChrome.ChipShape
    const val HeroInfoWidthDp = 348
    const val EpisodePaneWidthDp = 430
    const val EpisodeThumbWidthDp = 116
    const val EpisodeCardHeightDp = 84
    val ReferenceGold = Color(0xFFE8B85B)
    val ReferenceGlass = Color(0x7810161F)
    val ReferenceGlassFocused = Color(0x98241C11)
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
            size = 42.dp,
            iconSize = 20.dp,
            containerColor = Color(0x66070A10),
            contentColor = Color.White,
            focusedScale = 1.06f,
        )

        TvSeriesQualityBadge(
            modifier = Modifier
                .align(Alignment.TopEnd)
                .statusBarsPadding()
                .padding(top = 30.dp, end = 34.dp),
        )

        Row(
            modifier = Modifier
                .fillMaxSize()
                .statusBarsPadding()
                .navigationBarsPadding()
                .padding(start = 72.dp, top = 72.dp, end = 24.dp, bottom = 26.dp),
            horizontalArrangement = Arrangement.spacedBy(22.dp),
            verticalAlignment = Alignment.Top,
        ) {
            TvSeriesHeroPane(
                series = series,
                currentEpisode = currentEpisode,
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
        border = BorderStroke(1.dp, TvSeriesDetailTokens.ReferenceGold.copy(alpha = 0.78f)),
    ) {
        Text(
            text = "4K",
            color = TvSeriesDetailTokens.ReferenceGold,
            fontSize = 15.sp,
            lineHeight = 18.sp,
            fontWeight = FontWeight.Black,
            modifier = Modifier.padding(horizontal = 8.dp, vertical = 3.dp),
        )
    }
}

@Composable
private fun TvSeriesHeroPane(
    series: TvSeriesUiModel,
    currentEpisode: TvEpisodeUiModel?,
    playFocusRequester: FocusRequester,
    modifier: Modifier = Modifier,
    onPlay: () -> Unit,
) {
    Column(
        modifier = modifier.padding(top = 18.dp),
        verticalArrangement = Arrangement.spacedBy(9.dp),
    ) {
        TvSeriesTitleBlock(title = series.title)
        TvSeriesMetaRow(series = series)
        TvSeriesRatingRow(series = series)
        Text(
            text = series.description,
            color = Color(0xDDE7ECF5),
            fontSize = 13.sp,
            lineHeight = 18.sp,
            maxLines = 2,
            overflow = TextOverflow.Ellipsis,
            modifier = Modifier.width(TvSeriesDetailTokens.HeroInfoWidthDp.dp),
        )
        TvSeriesReferenceActionRow(
            primaryText = buildTvSeriesPlayButtonLabel(currentEpisode),
            enabled = currentEpisode?.playable == true,
            playFocusRequester = playFocusRequester,
            onPlay = onPlay,
        )
        if (series.cast.isNotEmpty()) {
            Spacer(modifier = Modifier.height(2.dp))
            TvSeriesCastRow(cast = series.cast)
        }
    }
}

@Composable
private fun TvSeriesTitleBlock(title: String) {
    val titleParts = remember(title) { splitTvSeriesReferenceTitle(title) }
    val mainFontSize = when {
        titleParts.stretched -> 44.sp
        titleParts.main.length <= 2 -> 46.sp
        titleParts.main.length <= 6 -> 42.sp
        else -> 36.sp
    }
    val mainLineHeight = when {
        titleParts.stretched -> 50.sp
        titleParts.main.length <= 2 -> 52.sp
        titleParts.main.length <= 6 -> 48.sp
        else -> 42.sp
    }
    Column(
        verticalArrangement = Arrangement.spacedBy(if (titleParts.stretched) 0.dp else 2.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        modifier = Modifier.width(TvSeriesDetailTokens.HeroInfoWidthDp.dp),
    ) {
        Text(
            text = formatTvSeriesReferenceTitle(titleParts.eyebrow, titleParts.stretched),
            color = Color.White.copy(alpha = 0.88f),
            fontSize = if (titleParts.stretched) 16.sp else 14.sp,
            lineHeight = 19.sp,
            letterSpacing = if (titleParts.stretched) 6.sp else 1.sp,
            fontWeight = FontWeight.Bold,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
            textAlign = TextAlign.Center,
        )
        Text(
            text = formatTvSeriesReferenceTitle(titleParts.main, titleParts.stretched),
            color = Color.White,
            fontSize = mainFontSize,
            lineHeight = mainLineHeight,
            letterSpacing = if (titleParts.stretched) 4.sp else 0.sp,
            fontWeight = FontWeight.Black,
            maxLines = if (titleParts.stretched) 2 else 1,
            overflow = TextOverflow.Ellipsis,
            textAlign = TextAlign.Center,
        )
    }
}

@Composable
private fun TvSeriesMetaRow(series: TvSeriesUiModel) {
    val items = remember(series) { buildTvSeriesDetailMetaItems(series) }
    Row(
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        items.take(4).forEachIndexed { index, item ->
            if (index > 0) {
                Text(
                    text = "•",
                    color = Color.White.copy(alpha = 0.64f),
                    fontSize = 12.sp,
                    lineHeight = 15.sp,
                )
            }
            Text(
                text = item,
                color = Color.White.copy(alpha = 0.72f),
                fontSize = 12.sp,
                lineHeight = 15.sp,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
        TvSeriesAgeBadge()
    }
}

@Composable
private fun TvSeriesAgeBadge() {
    Surface(
        color = Color.Transparent,
        shape = TvSeriesDetailTokens.MetaChipShape,
        border = BorderStroke(1.dp, Color.White.copy(alpha = 0.25f)),
    ) {
        Text(
            text = "18+",
            color = Color.White.copy(alpha = 0.78f),
            fontSize = 10.sp,
            lineHeight = 12.sp,
            fontWeight = FontWeight.Medium,
            modifier = Modifier.padding(horizontal = 5.dp, vertical = 1.dp),
        )
    }
}

@Composable
private fun TvSeriesRatingRow(series: TvSeriesUiModel) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(7.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Row(horizontalArrangement = Arrangement.spacedBy(2.dp)) {
            repeat(5) { index ->
                Icon(
                    imageVector = Icons.Filled.Star,
                    contentDescription = null,
                    modifier = Modifier.size(14.dp),
                    tint = TvSeriesDetailTokens.ReferenceGold.copy(alpha = if (index < 4) 1f else 0.58f),
                )
            }
        }
        Text(
            text = series.ratingText.ifBlank { "暂无评分" },
            color = Color.White,
            fontSize = 13.sp,
            lineHeight = 16.sp,
            fontWeight = FontWeight.Medium,
            maxLines = 1,
        )
        Surface(
            color = Color(0x22100A02),
            shape = TvSeriesDetailTokens.MetaChipShape,
            border = BorderStroke(1.dp, TvSeriesDetailTokens.ReferenceGold.copy(alpha = 0.82f)),
        ) {
            Text(
                text = "IMDb",
                color = TvSeriesDetailTokens.ReferenceGold,
                fontSize = 9.sp,
                lineHeight = 11.sp,
                fontWeight = FontWeight.Black,
                modifier = Modifier.padding(horizontal = 5.dp, vertical = 1.dp),
            )
        }
    }
}

@Composable
private fun TvSeriesReferenceActionRow(
    primaryText: String,
    enabled: Boolean,
    playFocusRequester: FocusRequester,
    onPlay: () -> Unit,
) {
    Row(
        horizontalArrangement = Arrangement.spacedBy(10.dp),
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
}

@Composable
private fun TvSeriesPrimaryActionButton(
    text: String,
    enabled: Boolean,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    var focused by remember { mutableStateOf(false) }
    Surface(
        color = if (enabled) Color(0xFFEFC463) else Color(0xFF2A2F36),
        shape = TvSeriesDetailTokens.PrimaryActionShape,
        border = BorderStroke(
            width = if (focused) 2.dp else 0.dp,
            color = if (focused) Color.White.copy(alpha = 0.86f) else Color.Transparent,
        ),
        shadowElevation = if (focused && enabled) 18.dp else if (enabled) 8.dp else 0.dp,
        modifier = modifier
            .width(150.dp)
            .height(42.dp)
            .onFocusChanged { focused = it.isFocused }
            .clickable(enabled = enabled, onClick = onClick),
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Box(
                modifier = Modifier
                    .size(26.dp)
                    .clip(CircleShape)
                    .background(Color(0xFF06080B)),
                contentAlignment = Alignment.Center,
            ) {
                Icon(
                    imageVector = Icons.Filled.PlayArrow,
                    contentDescription = null,
                    tint = Color.White,
                    modifier = Modifier.size(17.dp),
                )
            }
            Text(
                text = text,
                color = if (enabled) Color(0xFF070A0D) else Color(0xFFBAC2CE),
                fontSize = 13.sp,
                lineHeight = 16.sp,
                fontWeight = FontWeight.Black,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun TvSeriesSecondaryActionButton(text: String) {
    var focused by remember { mutableStateOf(false) }
    Surface(
        color = if (focused) Color(0x362B2419) else Color(0x31141820),
        shape = TvSeriesDetailTokens.PrimaryActionShape,
        border = BorderStroke(
            width = if (focused) 2.dp else 1.dp,
            color = if (focused) TvSeriesDetailTokens.ReferenceGold else Color.White.copy(alpha = 0.16f),
        ),
        modifier = Modifier
            .width(108.dp)
            .height(40.dp)
            .onFocusChanged { focused = it.isFocused }
            .clickable(onClick = {}),
    ) {
        Row(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 11.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(6.dp),
        ) {
            Icon(
                imageVector = Icons.Filled.Add,
                contentDescription = null,
                tint = Color.White.copy(alpha = 0.9f),
                modifier = Modifier.size(14.dp),
            )
            Text(
                text = text,
                color = Color.White.copy(alpha = 0.92f),
                fontSize = 12.sp,
                lineHeight = 15.sp,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun TvSeriesCastRow(cast: List<String>) {
    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
        Text(
            text = "演员",
            color = Color(0xBFE8EDF6),
            fontSize = 12.sp,
            lineHeight = 15.sp,
            fontWeight = FontWeight.Normal,
        )
        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
            cast.take(4).forEachIndexed { index, name ->
                TvSeriesCastItem(index = index, name = name)
            }
        }
    }
}

@Composable
private fun TvSeriesCastItem(index: Int, name: String) {
    Column(
        modifier = Modifier.width(66.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(3.dp),
    ) {
        Box(
            modifier = Modifier
                .size(42.dp)
                .clip(CircleShape)
                .background(tvCastAvatarBrush(index)),
            contentAlignment = Alignment.Center,
        ) {
            Text(
                text = name.take(1).ifBlank { "演" },
                color = Color.White,
                fontSize = 17.sp,
                lineHeight = 21.sp,
                fontWeight = FontWeight.Bold,
            )
        }
        Text(
            text = name,
            color = Color(0xFFECD38F),
            fontSize = 10.sp,
            lineHeight = 12.sp,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Text(
            text = "主演",
            color = Color(0xA9D5DCE8),
            fontSize = 9.sp,
            lineHeight = 11.sp,
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
        modifier = modifier.padding(top = 12.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = "剧集",
                color = TvSeriesDetailTokens.ReferenceGold,
                fontSize = 19.sp,
                lineHeight = 23.sp,
                letterSpacing = 0.sp,
                fontWeight = FontWeight.Black,
            )
            TvSeriesSeasonSelector(
                label = season?.title?.ifBlank { "第 $selectedSeasonNumber 季" } ?: "第 $selectedSeasonNumber 季",
                onClick = {
                    nextTvSeriesSeasonNumber(series.seasons, selectedSeasonNumber)?.let(onSelectSeason)
                },
            )
        }

        if (episodes.isEmpty()) {
            TvSeriesEmptyEpisodes()
            return@Column
        }

        LazyColumn(
            modifier = Modifier.weight(1f),
            verticalArrangement = Arrangement.spacedBy(10.dp),
            contentPadding = PaddingValues(bottom = 8.dp),
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
private fun TvSeriesSeasonSelector(
    label: String,
    onClick: () -> Unit,
) {
    var focused by remember { mutableStateOf(false) }
    Row(
        modifier = Modifier
            .clip(TvSeriesDetailTokens.SeasonChipShape)
            .border(
                BorderStroke(
                    width = if (focused) 1.dp else 0.dp,
                    color = if (focused) TvSeriesDetailTokens.ReferenceGold else Color.Transparent,
                ),
                TvSeriesDetailTokens.SeasonChipShape,
            )
            .onFocusChanged { focused = it.isFocused }
            .clickable(onClick = onClick)
            .padding(horizontal = 8.dp, vertical = 4.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        verticalAlignment = Alignment.CenterVertically,
    ) {
        Text(
            text = label,
            color = Color.White.copy(alpha = 0.7f),
            fontSize = 13.sp,
            lineHeight = 16.sp,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
        Icon(
            imageVector = Icons.Filled.KeyboardArrowDown,
            contentDescription = null,
            tint = Color.White.copy(alpha = 0.58f),
            modifier = Modifier.size(16.dp),
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
    var focused by remember { mutableStateOf(false) }
    val highlighted = selected || focused
    Surface(
        color = if (highlighted) TvSeriesDetailTokens.ReferenceGlassFocused else TvSeriesDetailTokens.ReferenceGlass,
        shape = TvSeriesDetailTokens.EpisodeCardShape,
        border = BorderStroke(
            width = if (highlighted) 2.dp else 1.dp,
            color = if (highlighted) TvSeriesDetailTokens.ReferenceGold else Color.White.copy(alpha = 0.12f),
        ),
        shadowElevation = if (focused) 8.dp else 0.dp,
        modifier = Modifier
            .fillMaxWidth()
            .height(TvSeriesDetailTokens.EpisodeCardHeightDp.dp)
            .onFocusChanged { focused = it.isFocused }
            .clickable(enabled = episode.playable, onClick = onClick),
    ) {
        Row(
            verticalAlignment = Alignment.CenterVertically,
        ) {
            TvSeriesEpisodeStill(
                episode = episode,
                stillUrl = stillUrl,
                selected = selected,
            )
            Column(
                modifier = Modifier
                    .weight(1f)
                    .padding(horizontal = 13.dp, vertical = 10.dp),
                verticalArrangement = Arrangement.spacedBy(3.dp),
            ) {
                Text(
                    text = "${episode.number}. ${episode.title.ifBlank { "未命名" }}",
                    color = if (episode.playable) Color.White else Color(0x8FE8EDF6),
                    fontSize = 14.sp,
                    lineHeight = 17.sp,
                    fontWeight = FontWeight.Black,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = episode.durationLabel,
                    color = if (selected) Color(0xFFFFE2A0) else Color(0xBBC7CFDA),
                    fontSize = 12.sp,
                    lineHeight = 15.sp,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = buildTvEpisodeCardSummary(episode),
                    color = Color(0xA9D5DCE8),
                    fontSize = 11.sp,
                    lineHeight = 14.sp,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
            Box(
                modifier = Modifier
                    .padding(end = 14.dp)
                    .size(34.dp)
                    .clip(CircleShape)
                    .background(
                        if (highlighted && episode.playable) {
                            TvSeriesDetailTokens.ReferenceGold
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
                        episode.playable -> TvSeriesDetailTokens.ReferenceGold
                        else -> Color(0x80FFFFFF)
                    },
                    modifier = Modifier.size(21.dp),
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
                fontSize = 10.sp,
                lineHeight = 12.sp,
                fontWeight = FontWeight.SemiBold,
            )
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
            fontSize = 13.sp,
            lineHeight = 16.sp,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 16.dp),
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

private data class TvSeriesReferenceTitleParts(
    val eyebrow: String,
    val main: String,
    val stretched: Boolean,
)

private fun splitTvSeriesReferenceTitle(title: String): TvSeriesReferenceTitleParts {
    val normalized = title.trim().replace(Regex("\\s+"), " ")
    if (normalized.isBlank()) {
        return TvSeriesReferenceTitleParts(
            eyebrow = "剧集",
            main = "未命名",
            stretched = false,
        )
    }
    val words = normalized.split(" ").filter { it.isNotBlank() }
    val shouldStretch = words.size >= 2 &&
        words.all { word -> word.any { char -> char.isLetterOrDigit() } } &&
        normalized.none { char -> char.code > 127 }
    if (shouldStretch) {
        return TvSeriesReferenceTitleParts(
            eyebrow = words.first().uppercase(),
            main = words.drop(1).joinToString(" ").uppercase(),
            stretched = true,
        )
    }
    return TvSeriesReferenceTitleParts(
        eyebrow = "剧集",
        main = normalized,
        stretched = false,
    )
}

internal fun formatTvSeriesReferenceTitle(title: String, stretched: Boolean): String {
    val normalized = title.trim()
    if (normalized.isBlank()) {
        return if (stretched) "未 命 名 剧 集" else "未命名剧集"
    }
    val withoutExtraSpaces = normalized.replace(Regex("\\s+"), " ")
    if (!stretched) {
        return withoutExtraSpaces
    }
    return withoutExtraSpaces.map { char ->
        if (char == ' ') "   " else char.toString()
    }.joinToString(" ")
}

private fun nextTvSeriesSeasonNumber(
    seasons: List<TvSeasonUiModel>,
    selectedSeasonNumber: Int,
): Int? {
    if (seasons.isEmpty()) {
        return null
    }
    val index = seasons.indexOfFirst { it.number == selectedSeasonNumber }
    return seasons[((index.takeIf { it >= 0 } ?: 0) + 1) % seasons.size].number
}

private fun buildTvSeriesPlayButtonLabel(episode: TvEpisodeUiModel?): String {
    return if (episode?.playable == true) {
        "播放第 ${episode.number} 集"
    } else {
        "暂无片源"
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
