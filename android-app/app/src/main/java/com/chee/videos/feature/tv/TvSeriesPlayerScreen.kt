package com.chee.videos.feature.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.SlowMotionVideo
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.ExperimentalMaterial3Api
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

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TvSeriesPlayerScreen(
    onBack: () -> Unit,
    viewModel: TvSeriesPlayerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    if (uiState.loading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
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
                .background(Color.Black),
            contentAlignment = Alignment.Center,
        ) {
            Text("播放器占位数据不存在", color = AppChrome.TextSecondary)
        }
        return
    }
    val currentEpisode = selectedEpisode(uiState)
    val currentSeason = selectedSeason(uiState)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(
                    Brush.verticalGradient(
                        colors = listOf(Color(0xFF111827), Color(0xFF070B13), Color.Black),
                    ),
                ),
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 8.dp, vertical = 8.dp),
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
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = series.title,
                        color = Color.White,
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Text(
                        text = "S${uiState.selectedSeasonNumber} · E${uiState.selectedEpisodeNumber}  ${currentEpisode?.title.orEmpty()}",
                        color = AppChrome.TextSecondary,
                        style = MaterialTheme.typography.bodySmall,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
            }

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(240.dp)
                    .padding(horizontal = 12.dp)
                    .clip(RoundedCornerShape(18.dp))
                    .background(
                        Brush.verticalGradient(
                            colors = listOf(Color(0xFF2B1F4A), Color(0xFF0D111A)),
                        ),
                    ),
                contentAlignment = Alignment.Center,
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Icon(
                        imageVector = Icons.Filled.Tv,
                        contentDescription = null,
                        tint = Color.White.copy(alpha = 0.82f),
                        modifier = Modifier.size(40.dp),
                    )
                    Text("电视剧播放占位区域", color = Color.White, style = MaterialTheme.typography.bodyMedium)
                    Text(
                        text = currentEpisode?.durationLabel ?: "45 分钟",
                        color = AppChrome.TextMuted,
                        style = MaterialTheme.typography.bodySmall,
                    )
                }
            }

            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 14.dp, vertical = 14.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Surface(color = AppChrome.SurfaceElevated, shape = RoundedCornerShape(14.dp)) {
                    Text(
                        text = currentEpisode?.summary ?: "当前集剧情占位文案",
                        color = AppChrome.TextSecondary,
                        style = MaterialTheme.typography.bodyMedium,
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 12.dp),
                    )
                }
                Row(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                    TvPlayerActionButton(
                        label = if (uiState.isPlaying) "暂停" else "播放",
                        icon = Icons.Filled.PlayArrow,
                        modifier = Modifier.weight(1f),
                        onClick = viewModel::togglePlay,
                    )
                    TvPlayerActionButton(
                        label = "${uiState.playbackSpeed}x",
                        icon = Icons.Filled.SlowMotionVideo,
                        modifier = Modifier.weight(1f),
                        onClick = viewModel::cycleSpeed,
                    )
                    TvPlayerActionButton(
                        label = "下一集",
                        icon = Icons.Filled.SkipNext,
                        modifier = Modifier.weight(1f),
                        onClick = viewModel::nextEpisode,
                    )
                }
                Surface(
                    color = AppChrome.AccentSoft,
                    shape = AppChrome.PillShape,
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable { viewModel.setSelectorVisible(true) },
                ) {
                    Text(
                        text = "打开选集抽屉（${currentSeason?.title.orEmpty()}）",
                        color = AppChrome.TextPrimary,
                        style = MaterialTheme.typography.labelLarge,
                        fontWeight = FontWeight.SemiBold,
                        modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
                    )
                }
            }
        }
    }

    if (uiState.selectorVisible) {
        ModalBottomSheet(
            onDismissRequest = { viewModel.setSelectorVisible(false) },
            containerColor = AppChrome.Surface,
            contentColor = AppChrome.TextPrimary,
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .navigationBarsPadding(),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                Text(
                    text = "选集播放",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.SemiBold,
                    modifier = Modifier.padding(horizontal = 16.dp),
                )
                Row(
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.padding(horizontal = 16.dp),
                ) {
                    series.seasons.forEach { season ->
                        val selected = season.number == uiState.selectedSeasonNumber
                        Surface(
                            color = if (selected) AppChrome.AccentSoft else AppChrome.SurfaceElevated,
                            shape = RoundedCornerShape(10.dp),
                            modifier = Modifier.clickable { viewModel.selectSeason(season.number) },
                        ) {
                            Text(
                                text = season.title,
                                color = if (selected) AppChrome.TextPrimary else AppChrome.TextMuted,
                                style = MaterialTheme.typography.labelLarge,
                                modifier = Modifier.padding(horizontal = 12.dp, vertical = 8.dp),
                            )
                        }
                    }
                }
                LazyColumn(
                    contentPadding = PaddingValues(start = 16.dp, end = 16.dp, bottom = 18.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp),
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    items(selectedSeason(uiState)?.episodes.orEmpty(), key = { episode -> episode.id }) { episode ->
                        val selected = episode.number == uiState.selectedEpisodeNumber
                        Surface(
                            color = if (selected) AppChrome.AccentSoft else AppChrome.SurfaceElevated,
                            shape = RoundedCornerShape(12.dp),
                            modifier = Modifier
                                .fillMaxWidth()
                                .clickable {
                                    viewModel.selectEpisode(episode.number)
                                    viewModel.setSelectorVisible(false)
                                },
                        ) {
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = 12.dp, vertical = 10.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(8.dp),
                            ) {
                                Text(
                                    text = "E${episode.number}",
                                    color = if (selected) AppChrome.TextPrimary else AppChrome.TextSecondary,
                                    style = MaterialTheme.typography.labelLarge,
                                    fontWeight = FontWeight.SemiBold,
                                    modifier = Modifier.width(42.dp),
                                )
                                Column(
                                    modifier = Modifier.weight(1f),
                                    verticalArrangement = Arrangement.spacedBy(2.dp),
                                ) {
                                    Text(
                                        text = episode.title,
                                        color = AppChrome.TextPrimary,
                                        style = MaterialTheme.typography.bodyMedium,
                                        maxLines = 1,
                                        overflow = TextOverflow.Ellipsis,
                                    )
                                    Text(
                                        text = episode.durationLabel,
                                        color = AppChrome.TextMuted,
                                        style = MaterialTheme.typography.bodySmall,
                                    )
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TvPlayerActionButton(
    label: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    modifier: Modifier = Modifier,
    onClick: () -> Unit,
) {
    Surface(
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(12.dp),
        modifier = modifier.clickable(onClick = onClick),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 10.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.Center,
        ) {
            Icon(icon, contentDescription = null, tint = AppChrome.TextSecondary, modifier = Modifier.size(16.dp))
            Text(
                text = label,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.labelLarge,
                modifier = Modifier.padding(start = 6.dp),
            )
        }
    }
}
