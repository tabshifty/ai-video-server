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
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
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
import com.chee.videos.feature.detail.DetailViewModel
import kotlinx.coroutines.launch

@Composable
fun TvLongFormDetailScreen(
    onBack: () -> Unit,
    onPlay: (String, String) -> Unit,
    viewModel: DetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val listState = rememberLazyListState()
    val scope = rememberCoroutineScope()
    val playFocusRequester = remember { FocusRequester() }

    when {
        uiState.loading -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(AppChrome.PageGradient),
                contentAlignment = Alignment.Center,
            ) {
                CircularProgressIndicator(color = AppChrome.AccentStrong)
            }
        }

        uiState.detail == null -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(AppChrome.PageGradient),
                contentAlignment = Alignment.Center,
            ) {
                Text(uiState.errorMessage ?: "详情加载失败", color = AppChrome.TextSecondary)
            }
        }

        else -> {
            val detail = uiState.detail!!
            val hero = remember(uiState.baseUrl, uiState.videoType, detail) {
                buildTvLongFormDetailHero(
                    baseUrl = uiState.baseUrl,
                    videoType = uiState.videoType,
                    detail = detail,
                )
            }
            val canPlay = resolveTvLongFormPlayUrl(
                baseUrl = uiState.baseUrl,
                detail = detail,
                preferredPlaybackProfile = uiState.preferredPlaybackProfile,
            ) != null

            LaunchedEffect(detail.id) {
                playFocusRequester.requestFocus()
            }

            LazyColumn(
                state = listState,
                modifier = Modifier
                    .fillMaxSize()
                    .background(AppChrome.PageGradient),
                contentPadding = PaddingValues(bottom = 28.dp),
                verticalArrangement = Arrangement.spacedBy(18.dp),
            ) {
                item(key = "hero") {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(468.dp),
                    ) {
                        if (!hero.backdropUrl.isNullOrBlank()) {
                            AsyncImage(
                                model = hero.backdropUrl,
                                contentDescription = hero.title,
                                modifier = Modifier.fillMaxSize(),
                                contentScale = ContentScale.Crop,
                            )
                        } else {
                            Box(
                                modifier = Modifier
                                    .fillMaxSize()
                                    .background(
                                        Brush.horizontalGradient(
                                            colors = listOf(Color(0xFF23131F), Color(0xFF090C12)),
                                        ),
                                    ),
                            )
                        }
                        Box(
                            modifier = Modifier
                                .fillMaxSize()
                                .background(
                                    Brush.verticalGradient(
                                        colors = listOf(Color(0x88060A10), Color(0xD4090B10), AppChrome.Canvas),
                                    ),
                                ),
                        )
                        Row(
                            modifier = Modifier
                                .fillMaxSize()
                                .padding(horizontal = 28.dp, vertical = 28.dp),
                            horizontalArrangement = Arrangement.spacedBy(28.dp),
                            verticalAlignment = Alignment.Bottom,
                        ) {
                            TvLongFormPoster(
                                title = hero.title,
                                posterUrl = hero.posterUrl,
                                videoType = uiState.videoType,
                            )
                            Column(
                                modifier = Modifier.weight(1f),
                                verticalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                TvRoundIconAction(
                                    icon = Icons.AutoMirrored.Filled.ArrowBack,
                                    contentDescription = "返回",
                                    onClick = onBack,
                                )
                                Text(
                                    text = hero.eyebrow,
                                    color = AppChrome.AccentWarm,
                                    style = MaterialTheme.typography.labelLarge,
                                    fontWeight = FontWeight.SemiBold,
                                )
                                Text(
                                    text = hero.title,
                                    color = AppChrome.TextPrimary,
                                    style = MaterialTheme.typography.headlineLarge,
                                    fontWeight = FontWeight.Bold,
                                    maxLines = 2,
                                    overflow = TextOverflow.Ellipsis,
                                )
                                if (hero.metaLine.isNotBlank()) {
                                    Text(
                                        text = hero.metaLine,
                                        color = AppChrome.TextSecondary,
                                        style = MaterialTheme.typography.titleMedium,
                                    )
                                }
                                Text(
                                    text = hero.summary,
                                    color = AppChrome.TextMuted,
                                    style = MaterialTheme.typography.bodyLarge,
                                    maxLines = 4,
                                    overflow = TextOverflow.Ellipsis,
                                )
                                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                                    TvDetailPrimaryActionButton(
                                        text = if (canPlay) hero.primaryActionLabel else "暂无片源",
                                        modifier = Modifier.focusRequester(playFocusRequester),
                                        enabled = canPlay,
                                        onClick = {
                                            if (canPlay) {
                                                onPlay(detail.id, uiState.videoType)
                                            }
                                        },
                                    )
                                    TvDetailSecondaryActionButton(
                                        text = hero.secondaryActionLabel,
                                        onClick = {
                                            scope.launch { listState.animateScrollToItem(1) }
                                        },
                                    )
                                }
                            }
                        }
                    }
                }

                item(key = "summary") {
                    TvDetailInfoCard(
                        title = "剧情简介",
                        content = detail.description.orEmpty().ifBlank { "暂无简介" },
                    )
                }

                if (detail.tags.orEmpty().isNotEmpty()) {
                    item(key = "tags") {
                        Surface(
                            color = AppChrome.Surface,
                            shape = AppChrome.CardShape,
                            modifier = Modifier.padding(horizontal = 18.dp),
                        ) {
                            Column(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(horizontal = 18.dp, vertical = 18.dp),
                                verticalArrangement = Arrangement.spacedBy(12.dp),
                            ) {
                                Text(
                                    text = "标签",
                                    color = AppChrome.TextPrimary,
                                    style = MaterialTheme.typography.titleMedium,
                                    fontWeight = FontWeight.SemiBold,
                                )
                                LazyRow(horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                                    items(detail.tags.orEmpty().size) { index ->
                                        Surface(
                                            color = AppChrome.SurfaceStrong,
                                            shape = AppChrome.PillShape,
                                        ) {
                                            Text(
                                                text = detail.tags.orEmpty()[index],
                                                color = AppChrome.TextSecondary,
                                                style = MaterialTheme.typography.labelLarge,
                                                modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                }

                item(key = "stats") {
                    Surface(
                        color = AppChrome.Surface,
                        shape = AppChrome.CardShape,
                        modifier = Modifier.padding(horizontal = 18.dp),
                    ) {
                        Row(
                            modifier = Modifier
                                .fillMaxWidth()
                                .padding(horizontal = 18.dp, vertical = 18.dp),
                            horizontalArrangement = Arrangement.spacedBy(12.dp),
                        ) {
                            TvMetricCard(
                                modifier = Modifier.weight(1f),
                                label = "播放",
                                value = detail.viewsCount.toString(),
                            )
                            TvMetricCard(
                                modifier = Modifier.weight(1f),
                                label = "点赞",
                                value = detail.likesCount.toString(),
                            )
                            TvMetricCard(
                                modifier = Modifier.weight(1f),
                                label = "收藏",
                                value = detail.favoritesCount.toString(),
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TvLongFormPoster(
    title: String,
    posterUrl: String?,
    videoType: String,
) {
    if (!posterUrl.isNullOrBlank()) {
        AsyncImage(
            model = posterUrl,
            contentDescription = title,
            modifier = Modifier
                .width(248.dp)
                .aspectRatio(2f / 3f)
                .clip(RoundedCornerShape(28.dp)),
            contentScale = ContentScale.Crop,
        )
        return
    }
    Box(
        modifier = Modifier
            .width(248.dp)
            .aspectRatio(2f / 3f)
            .clip(RoundedCornerShape(28.dp))
            .background(
                Brush.verticalGradient(
                    colors = if (normalizeTvLongFormVideoType(videoType) == "av") {
                        listOf(Color(0xFF402030), Color(0xFF111827))
                    } else {
                        listOf(Color(0xFF1F3457), Color(0xFF111827))
                    },
                ),
            ),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = Icons.Filled.Tv,
            contentDescription = null,
            tint = Color.White.copy(alpha = 0.82f),
            modifier = Modifier.size(56.dp),
        )
    }
}

@Composable
private fun TvRoundIconAction(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
) {
    Surface(
        color = Color(0x44080B11),
        shape = CircleShape,
        modifier = Modifier
            .size(52.dp)
            .tvFocusableGlow(shape = CircleShape, focusedScale = 1.08f)
            .clickable(onClick = onClick),
    ) {
        Box(contentAlignment = Alignment.Center) {
            Icon(icon, contentDescription = contentDescription, tint = Color.White)
        }
    }
}

@Composable
private fun TvDetailPrimaryActionButton(
    text: String,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    onClick: () -> Unit,
) {
    Surface(
        color = if (enabled) AppChrome.Accent else AppChrome.SurfaceStrong,
        shape = AppChrome.PillShape,
        modifier = modifier
            .tvFocusableGlow(shape = AppChrome.PillShape, focusedScale = 1.06f)
            .clickable(enabled = enabled, onClick = onClick),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Icon(Icons.Filled.PlayArrow, contentDescription = null, tint = Color.White)
            Text(
                text = text,
                color = Color.White,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
            )
        }
    }
}

@Composable
private fun TvDetailSecondaryActionButton(
    text: String,
    onClick: () -> Unit,
) {
    Surface(
        color = Color(0x33090C13),
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
private fun TvDetailInfoCard(
    title: String,
    content: String,
) {
    Surface(
        color = AppChrome.Surface,
        shape = AppChrome.CardShape,
        modifier = Modifier.padding(horizontal = 18.dp),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 18.dp, vertical = 18.dp),
            verticalArrangement = Arrangement.spacedBy(10.dp),
        ) {
            Text(
                text = title,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
            )
            Text(
                text = content,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.bodyLarge,
            )
        }
    }
}

@Composable
private fun TvMetricCard(
    modifier: Modifier = Modifier,
    label: String,
    value: String,
) {
    Surface(
        modifier = modifier,
        color = AppChrome.SurfaceElevated,
        shape = RoundedCornerShape(18.dp),
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 10.dp, vertical = 14.dp),
            verticalArrangement = Arrangement.spacedBy(4.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Text(
                text = value,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold,
            )
            Text(
                text = label,
                color = AppChrome.TextMuted,
                style = MaterialTheme.typography.labelSmall,
            )
        }
    }
}
