package com.chee.videos.feature.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Star
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
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
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.tvFocusableGlow
import com.chee.videos.feature.detail.DetailViewModel

@Composable
fun TvLongFormDetailScreen(
    onBack: () -> Unit,
    onPlay: (String, String) -> Unit,
    viewModel: DetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val playFocusRequester = remember { FocusRequester() }

    when {
        uiState.loading -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(AppChrome.PageGradient),
            ) {
                TvPageLoadingState(message = "正在加载详情")
            }
        }

        uiState.detail == null -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(AppChrome.PageGradient),
            ) {
                TvErrorState(
                    message = uiState.errorMessage ?: "详情加载失败",
                    onAction = viewModel::load,
                )
            }
        }

        else -> {
            val detail = uiState.detail!!
            val hero = remember(uiState.baseUrl, uiState.videoType, detail) {
                buildTvLongFormDetailHero(
                    baseUrl = uiState.baseUrl,
                    detail = detail,
                    videoType = uiState.videoType,
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

            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(AppChrome.PageGradient),
            ) {
                TvLongFormDetailBackground(hero = hero)

                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.verticalGradient(
                                colors = listOf(
                                    Color(0x6610151F),
                                    Color(0x3310151F),
                                    Color(0xDD070A10),
                                ),
                            ),
                        ),
                )

                TvRoundIconAction(
                    icon = Icons.AutoMirrored.Filled.ArrowBack,
                    contentDescription = "返回",
                    onClick = onBack,
                    modifier = Modifier
                        .align(Alignment.TopStart)
                        .padding(start = 28.dp, top = 28.dp),
                )

                Surface(
                    color = Color(0xD20B1018),
                    shape = RoundedCornerShape(topStart = 28.dp, topEnd = 28.dp),
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .fillMaxWidth(),
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(horizontal = 36.dp, vertical = 28.dp),
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                    ) {
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
                            maxLines = 3,
                            overflow = TextOverflow.Ellipsis,
                        )
                        if (hero.actors.isNotEmpty()) {
                            TvLongFormActorRow(actors = hero.actors)
                        }
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
                                onClick = { viewModel.toggleFavorite() },
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TvLongFormDetailBackground(hero: TvLongFormDetailHeroUiModel) {
    if (!hero.backdropUrl.isNullOrBlank()) {
        Box(modifier = Modifier.fillMaxSize()) {
            AsyncImage(
                model = hero.backdropUrl,
                contentDescription = hero.title,
                modifier = Modifier
                    .fillMaxSize()
                    .then(if (hero.usesPosterAsBackdropFallback) Modifier.blur(22.dp) else Modifier),
                contentScale = ContentScale.Crop,
            )
            if (hero.usesPosterAsBackdropFallback) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(Color(0xAA05070C)),
                )
            }
        }
        return
    }
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

@Composable
private fun TvLongFormActorRow(actors: List<TvLongFormDetailActorUiModel>) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(
            text = "演员",
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold,
        )
        LazyRow(horizontalArrangement = Arrangement.spacedBy(14.dp)) {
            items(actors.size) { index ->
                TvLongFormActorAvatar(actor = actors[index])
            }
        }
    }
}

@Composable
private fun TvLongFormActorAvatar(actor: TvLongFormDetailActorUiModel) {
    Column(
        modifier = Modifier.width(74.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        Box(
            modifier = Modifier
                .size(58.dp)
                .clip(CircleShape)
                .background(Color(0xFF253044)),
            contentAlignment = Alignment.Center,
        ) {
            if (actor.hasAvatar) {
                AsyncImage(
                    model = actor.avatarUrl,
                    contentDescription = actor.name,
                    modifier = Modifier.fillMaxSize(),
                    contentScale = ContentScale.Crop,
                )
            }
            if (!actor.hasAvatar) {
                Text(
                    text = actor.name.take(1),
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.titleMedium,
                    fontWeight = FontWeight.Bold,
                )
            }
        }
        Text(
            text = actor.name,
            color = AppChrome.TextSecondary,
            style = MaterialTheme.typography.labelSmall,
            maxLines = 1,
            overflow = TextOverflow.Ellipsis,
        )
    }
}

@Composable
private fun TvRoundIconAction(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = Color(0x66080B11),
        shape = CircleShape,
        modifier = modifier
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
        Row(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Icon(Icons.Filled.Star, contentDescription = null, tint = AppChrome.TextPrimary)
            Text(
                text = text,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleSmall,
                fontWeight = FontWeight.SemiBold,
            )
        }
    }
}
