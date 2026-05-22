package com.chee.videos.feature.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.RowScope
import androidx.compose.foundation.layout.aspectRatio
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
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
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.TvLayoutSpec
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvIconActionButton
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableGlow
import com.chee.videos.core.ui.tvSharedSeriesPoster

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

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .statusBarsPadding(),
        contentPadding = PaddingValues(bottom = TvLayoutSpec.scrollBottomSafePaddingDp.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        item(key = "hero") {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(468.dp),
            ) {
                if (!backdropUrl.isNullOrBlank()) {
                    AsyncImage(
                        model = backdropUrl,
                        contentDescription = series.title,
                        modifier = Modifier.fillMaxSize(),
                        contentScale = ContentScale.Crop,
                    )
                } else {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(tvBackdropBrush(series.posterSeed)),
                    )
                }
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.verticalGradient(
                                colors = listOf(Color(0x88070A10), Color(0xD3080C13), AppChrome.Canvas),
                            ),
                        ),
                )
                Row(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(horizontal = 24.dp, vertical = 24.dp),
                    horizontalArrangement = Arrangement.spacedBy(24.dp),
                    verticalAlignment = Alignment.Bottom,
                ) {
                    if (!posterUrl.isNullOrBlank()) {
                        AsyncImage(
                            model = posterUrl,
                            contentDescription = "${series.title}海报",
                            modifier = Modifier
                                .width(228.dp)
                                .aspectRatio(2f / 3f)
                                .tvSharedSeriesPoster(series.id)
                                .clip(AppChrome.SurfaceShape),
                            contentScale = ContentScale.Crop,
                        )
                    } else {
                        Box(
                            modifier = Modifier
                                .width(228.dp)
                                .aspectRatio(2f / 3f)
                                .tvSharedSeriesPoster(series.id)
                                .clip(AppChrome.SurfaceShape)
                                .background(tvBackdropBrush(series.posterSeed)),
                            contentAlignment = Alignment.Center,
                        ) {
                            Icon(
                                imageVector = Icons.Filled.PlayArrow,
                                contentDescription = null,
                                tint = Color.White.copy(alpha = 0.82f),
                            )
                        }
                    }
                    Column(
                        modifier = Modifier.weight(1f),
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
                        TvIconActionButton(
                            icon = Icons.AutoMirrored.Filled.ArrowBack,
                            contentDescription = "返回",
                            onClick = onBack,
                            size = 52.dp,
                            iconSize = 24.dp,
                            containerColor = Color(0x44090C12),
                            contentColor = Color.White,
                        )
                        Text(
                            text = "电视剧详情",
                            color = AppChrome.AccentWarm,
                            style = MaterialTheme.typography.labelLarge,
                            fontWeight = FontWeight.SemiBold,
                        )
                        Text(
                            text = series.title,
                            color = Color.White,
                            style = MaterialTheme.typography.headlineLarge,
                            fontWeight = FontWeight.Bold,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis,
                        )
                        Text(
                            text = "${series.subtitle} · ${series.updateText}",
                            color = Color(0xFFE2E8F0),
                            style = MaterialTheme.typography.titleMedium,
                        )
                        Text(
                            text = series.description,
                            color = AppChrome.TextMuted,
                            style = MaterialTheme.typography.bodyLarge,
                            maxLines = 4,
                            overflow = TextOverflow.Ellipsis,
                        )
                        Row(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                            Surface(
                                color = AppChrome.Accent,
                                shape = AppChrome.PillShape,
                                modifier = Modifier
                                    .focusRequester(playFocusRequester)
                                    .tvFocusableGlow(shape = AppChrome.PillShape, focusedScale = 1.06f)
                                    .clickable(enabled = currentEpisode?.playable == true) {
                                        onPlayEpisode(
                                            series.id,
                                            uiState.selectedSeasonNumber,
                                            uiState.selectedEpisodeNumber,
                                        )
                                    },
                            ) {
                                Row(
                                    modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
                                    verticalAlignment = Alignment.CenterVertically,
                                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                                ) {
                                    Icon(
                                        imageVector = Icons.Filled.PlayArrow,
                                        contentDescription = null,
                                        tint = Color.White,
                                    )
                                    Text("立即播放", color = Color.White, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                                }
                            }
                            Surface(
                                color = Color(0x33080B11),
                                shape = AppChrome.PillShape,
                            ) {
                                Text(
                                    text = if (currentEpisode?.playable == true) {
                                        "继续观看 S${uiState.selectedSeasonNumber}E${uiState.selectedEpisodeNumber}"
                                    } else {
                                        "当前分集暂无可播放视频"
                                    },
                                    color = Color.White,
                                    style = MaterialTheme.typography.labelLarge,
                                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                                )
                            }
                        }
                    }
                }
            }
        }

        item(key = "tags") {
            Column(
                modifier = Modifier.padding(horizontal = 14.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                Text("剧情简介", color = AppChrome.TextPrimary, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                Text(series.description, color = AppChrome.TextSecondary, style = MaterialTheme.typography.bodyMedium)
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.horizontalScroll(rememberScrollState())) {
                    series.tags.forEach { tag ->
                        Surface(color = AppChrome.SurfaceStrong, shape = AppChrome.PillShape) {
                            Text(
                                text = tag,
                                color = AppChrome.TextSecondary,
                                style = MaterialTheme.typography.labelMedium,
                                modifier = Modifier.padding(horizontal = 10.dp, vertical = 6.dp),
                            )
                        }
                    }
                }
                Text(
                    text = "主演：${series.cast.joinToString(" · ")}",
                    color = AppChrome.TextMuted,
                    style = MaterialTheme.typography.bodySmall,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }

        item(key = "seasons") {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState())
                    .padding(horizontal = 14.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                series.seasons.forEach { item ->
                    val selected = item.number == uiState.selectedSeasonNumber
                    Surface(
                        color = if (selected) AppChrome.AccentSoft else AppChrome.Surface,
                        shape = AppChrome.ChipShape,
                        modifier = Modifier
                            .tvFocusableGlow(shape = AppChrome.ChipShape, focusedScale = 1.04f)
                            .clickable { viewModel.selectSeason(item.number) },
                    ) {
                        Text(
                            text = item.title,
                            color = if (selected) AppChrome.TextPrimary else AppChrome.TextMuted,
                            style = MaterialTheme.typography.labelLarge,
                            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
                            modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                        )
                    }
                }
            }
        }

        item(key = "episodes") {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 14.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Text(
                    text = "选集播放 · ${season?.title.orEmpty()}",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                )
                episodes.chunked(4).forEach { rowEpisodes ->
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                        rowEpisodes.forEach { episode ->
                            val selected = episode.number == uiState.selectedEpisodeNumber
                            Surface(
                                color = if (selected) AppChrome.AccentSoft else AppChrome.SurfaceElevated,
                                shape = AppChrome.SurfaceShape,
                                modifier = Modifier
                                    .weight(1f)
                                    .tvFocusableGlow(shape = AppChrome.SurfaceShape, focusedScale = 1.03f)
                                    .clickable(enabled = episode.playable) {
                                        viewModel.selectEpisode(episode.number)
                                        onPlayEpisode(
                                            series.id,
                                            uiState.selectedSeasonNumber,
                                            episode.number,
                                        )
                                    },
                            ) {
                                Column(
                                    modifier = Modifier.padding(horizontal = 10.dp, vertical = 10.dp),
                                    verticalArrangement = Arrangement.spacedBy(4.dp),
                                ) {
                                    Text(
                                        text = "E${episode.number}",
                                        color = if (selected) AppChrome.TextPrimary else AppChrome.TextSecondary,
                                        style = MaterialTheme.typography.labelLarge,
                                        fontWeight = FontWeight.SemiBold,
                                    )
                                    Text(
                                        text = episode.durationLabel,
                                        color = AppChrome.TextMuted,
                                        style = MaterialTheme.typography.bodySmall,
                                    )
                                    Text(
                                        text = if (episode.playable) "可播放" else "待绑定",
                                        color = if (episode.playable) AppChrome.AccentWarm else AppChrome.TextMuted,
                                        style = MaterialTheme.typography.bodySmall,
                                    )
                                }
                            }
                        }
                        repeat(4 - rowEpisodes.size) {
                            SpacerCell()
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun RowScope.SpacerCell() {
    Box(modifier = Modifier.width(0.dp).weight(1f))
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
