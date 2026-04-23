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
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.chee.videos.core.ui.AppChrome

@Composable
fun TvSeriesDetailScreen(
    onBack: () -> Unit,
    onPlayEpisode: (seriesId: String, season: Int, episode: Int) -> Unit,
    viewModel: TvSeriesDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    if (uiState.loading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(AppChrome.PageGradient)
                .statusBarsPadding(),
            contentAlignment = Alignment.Center,
        ) {
            CircularProgressIndicator(color = AppChrome.AccentStrong)
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
            contentAlignment = Alignment.Center,
        ) {
            Text("电视剧占位数据不可用", color = AppChrome.TextSecondary)
        }
        return
    }
    val season = selectedDetailSeason(uiState)
    val episodes = season?.episodes.orEmpty()

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .statusBarsPadding(),
        contentPadding = PaddingValues(bottom = 24.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        item(key = "hero") {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(tvBackdropBrush(series.posterSeed))
                    .padding(horizontal = 14.dp, vertical = 14.dp),
            ) {
                Column(verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        Surface(color = Color(0x4DFFFFFF), shape = CircleShape) {
                            IconButton(onClick = onBack) {
                                Icon(
                                    imageVector = Icons.AutoMirrored.Filled.ArrowBack,
                                    contentDescription = "返回",
                                    tint = Color.White,
                                )
                            }
                        }
                        Text(
                            text = "电视剧详情",
                            color = Color.White,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                        )
                    }
                    Text(
                        text = series.title,
                        color = Color.White,
                        style = MaterialTheme.typography.headlineSmall,
                        fontWeight = FontWeight.Bold,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Text(
                        text = "${series.subtitle} · 评分 ${series.ratingText} · ${series.updateText}",
                        color = Color(0xFFE2E8F0),
                        style = MaterialTheme.typography.bodyMedium,
                    )
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Surface(
                            color = AppChrome.Accent,
                            shape = AppChrome.PillShape,
                            modifier = Modifier.clickable {
                                onPlayEpisode(
                                    series.id,
                                    uiState.selectedSeasonNumber,
                                    uiState.selectedEpisodeNumber,
                                )
                            },
                        ) {
                            Row(
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 7.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(4.dp),
                            ) {
                                Icon(
                                    imageVector = Icons.Filled.PlayArrow,
                                    contentDescription = null,
                                    tint = Color.White,
                                )
                                Text("立即播放", color = Color.White, style = MaterialTheme.typography.labelLarge)
                            }
                        }
                        Surface(
                            color = Color(0x33FFFFFF),
                            shape = AppChrome.PillShape,
                        ) {
                            Text(
                                text = "继续观看 S${uiState.selectedSeasonNumber}E${uiState.selectedEpisodeNumber}",
                                color = Color.White,
                                style = MaterialTheme.typography.labelMedium,
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 9.dp),
                            )
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
                        shape = RoundedCornerShape(12.dp),
                        modifier = Modifier.clickable { viewModel.selectSeason(item.number) },
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
                                shape = RoundedCornerShape(12.dp),
                                modifier = Modifier
                                    .weight(1f)
                                    .clickable {
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
