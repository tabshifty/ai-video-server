package com.chee.videos.feature.actor

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
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.staggeredgrid.LazyVerticalStaggeredGrid
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridCells
import androidx.compose.foundation.lazy.staggeredgrid.StaggeredGridItemSpan
import androidx.compose.foundation.lazy.staggeredgrid.itemsIndexed
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.ExperimentalMaterialApi
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.pullrefresh.PullRefreshIndicator
import androidx.compose.material.pullrefresh.pullRefresh
import androidx.compose.material.pullrefresh.rememberPullRefreshState
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
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
import com.chee.videos.core.model.ActorDetailDto
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.model.resolveAvPosterUrl
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.util.UrlBuilder

@OptIn(ExperimentalMaterialApi::class)
@Composable
fun ActorDetailScreen(
    actorId: String,
    baseUrl: String,
    onBack: () -> Unit,
    onOpenDetail: (String, String) -> Unit,
    viewModel: ActorDetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val pullRefreshState = rememberPullRefreshState(
        refreshing = uiState.refreshing,
        onRefresh = viewModel::refresh,
    )

    LaunchedEffect(actorId) {
        viewModel.initialize(actorId)
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .pullRefresh(pullRefreshState),
    ) {
        LazyVerticalStaggeredGrid(
            modifier = Modifier.fillMaxSize(),
            columns = StaggeredGridCells.Fixed(2),
            verticalItemSpacing = 14.dp,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
            contentPadding = PaddingValues(start = 16.dp, end = 16.dp, top = 12.dp, bottom = 24.dp),
        ) {
            item(span = StaggeredGridItemSpan.FullLine) {
                ActorHeader(
                    baseUrl = baseUrl,
                    actor = uiState.actor,
                    loading = uiState.loading,
                    totalCount = uiState.totalCount,
                    onBack = onBack,
                )
            }

            when {
                uiState.loading && uiState.items.isEmpty() -> {
                    item(span = StaggeredGridItemSpan.FullLine) {
                        ActorStateCard(message = "正在加载演员作品", showProgress = true)
                    }
                }

                !uiState.errorMessage.isNullOrBlank() && uiState.items.isEmpty() -> {
                    item(span = StaggeredGridItemSpan.FullLine) {
                        ActorErrorCard(
                            message = uiState.errorMessage.orEmpty(),
                            onRetry = viewModel::retry,
                        )
                    }
                }

                uiState.items.isEmpty() -> {
                    item(span = StaggeredGridItemSpan.FullLine) {
                        ActorStateCard(message = "暂无可展示作品", showProgress = false)
                    }
                }

                else -> {
                    itemsIndexed(uiState.items, key = { _, item -> item.id }) { index, item ->
                        if (index >= uiState.items.lastIndex - 5) {
                            LaunchedEffect(index, uiState.items.size, uiState.loadingMore, uiState.hasMore) {
                                viewModel.loadMoreIfNeeded(index)
                            }
                        }
                        ActorWorkCard(
                            baseUrl = baseUrl,
                            item = item,
                            onClick = { onOpenDetail(item.id, item.type) },
                        )
                    }
                    item(span = StaggeredGridItemSpan.FullLine) {
                        ActorFooter(
                            loadingMore = uiState.loadingMore,
                            hasMore = uiState.hasMore,
                            errorMessage = uiState.loadMoreErrorMessage,
                            onRetry = { viewModel.loadMoreIfNeeded(uiState.items.lastIndex) },
                        )
                    }
                }
            }
        }

        PullRefreshIndicator(
            refreshing = uiState.refreshing,
            state = pullRefreshState,
            modifier = Modifier.align(Alignment.TopCenter),
            contentColor = AppChrome.AccentStrong,
            backgroundColor = AppChrome.SurfaceElevated,
        )
    }
}

@Composable
private fun ActorHeader(
    baseUrl: String,
    actor: ActorDetailDto?,
    loading: Boolean,
    totalCount: Int,
    onBack: () -> Unit,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = AppChrome.Surface.copy(alpha = 0.94f),
        shape = AppChrome.SectionShape,
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 16.dp),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            IconButton(onClick = onBack) {
                Icon(Icons.AutoMirrored.Filled.ArrowBack, contentDescription = "返回", tint = AppChrome.TextPrimary)
            }
            Row(
                horizontalArrangement = Arrangement.spacedBy(14.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                ActorAvatar(baseUrl = baseUrl, actor = actor)
                Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Text(
                        text = actor?.name?.trim()?.ifBlank { "演员" } ?: if (loading) "加载中" else "演员",
                        style = MaterialTheme.typography.headlineSmall,
                        color = AppChrome.TextPrimary,
                        fontWeight = FontWeight.Bold,
                    )
                    Text(
                        text = if (totalCount > 0) "$totalCount 部作品" else "作品",
                        style = MaterialTheme.typography.bodyMedium,
                        color = AppChrome.TextSecondary,
                    )
                }
            }
            actor?.let {
                val rows = actorProfileRows(it)
                if (rows.isNotEmpty()) {
                    Column(verticalArrangement = Arrangement.spacedBy(6.dp)) {
                        rows.forEach { row ->
                            Text(
                                text = row,
                                style = MaterialTheme.typography.bodyMedium,
                                color = AppChrome.TextSecondary,
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ActorAvatar(baseUrl: String, actor: ActorDetailDto?) {
    val avatarUrl = resolveActorAvatarUrl(baseUrl, actor?.avatarUrl)
    Box(
        modifier = Modifier
            .size(86.dp)
            .clip(RoundedCornerShape(22.dp))
            .background(AppChrome.CanvasRaised),
        contentAlignment = Alignment.Center,
    ) {
        if (!avatarUrl.isNullOrBlank()) {
            AsyncImage(
                model = avatarUrl,
                contentDescription = "${actor?.name.orEmpty()}头像",
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize(),
            )
        } else {
            Text(
                text = actor?.name?.take(1).orEmpty(),
                style = MaterialTheme.typography.titleLarge,
                color = AppChrome.TextSecondary,
                fontWeight = FontWeight.Bold,
            )
        }
    }
}

private fun actorProfileRows(actor: ActorDetailDto): List<String> = buildList {
    actor.aliases.filter { it.isNotBlank() }.takeIf { it.isNotEmpty() }?.let {
        add("别名：${it.joinToString(" / ")}")
    }
    listOfNotNull(
        actor.gender?.trim()?.takeIf { it.isNotBlank() }?.let { "性别：$it" },
        actor.country?.trim()?.takeIf { it.isNotBlank() }?.let { "国家：$it" },
        actor.birthDate?.trim()?.takeIf { it.isNotBlank() }?.let { "生日：$it" },
    ).forEach(::add)
}

@Composable
private fun ActorWorkCard(
    baseUrl: String,
    item: VideoListItemDto,
    onClick: () -> Unit,
) {
    val posterUrl = resolveAvPosterUrl(baseUrl, item) ?: resolveThumbnailUrl(baseUrl, item.thumbnailPath)
    val primaryText = item.metadata?.get("av_code") as? String

    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .clip(AppChrome.CardShape)
            .clickable(onClick = onClick),
        color = AppChrome.Surface,
        shape = AppChrome.CardShape,
    ) {
        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .aspectRatio(if (item.type == "av") 0.72f else 1.35f)
                    .background(AppChrome.CanvasRaised),
            ) {
                if (!posterUrl.isNullOrBlank()) {
                    AsyncImage(
                        model = posterUrl,
                        contentDescription = item.title,
                        contentScale = ContentScale.Crop,
                        modifier = Modifier.fillMaxSize(),
                    )
                } else {
                    Icon(
                        imageVector = Icons.Filled.PlayArrow,
                        contentDescription = null,
                        tint = AppChrome.TextMuted,
                        modifier = Modifier
                            .size(42.dp)
                            .align(Alignment.Center),
                    )
                }
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.verticalGradient(
                                colors = listOf(Color.Transparent, Color(0x55080A0F), Color(0xD90A0B0E)),
                            ),
                        ),
                )
            }
            Column(
                modifier = Modifier.padding(start = 12.dp, end = 12.dp, bottom = 14.dp),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = primaryText?.trim()?.takeIf { it.isNotBlank() } ?: item.title,
                    style = MaterialTheme.typography.titleMedium,
                    color = AppChrome.TextPrimary,
                    fontWeight = FontWeight.Bold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                Text(
                    text = actorWorkTypeLabel(item.type),
                    style = MaterialTheme.typography.bodySmall,
                    color = AppChrome.TextMuted,
                )
            }
        }
    }
}

@Composable
private fun ActorStateCard(message: String, showProgress: Boolean) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 40.dp),
        contentAlignment = Alignment.Center,
    ) {
        Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
            if (showProgress) {
                CircularProgressIndicator(color = AppChrome.AccentStrong)
            }
            Text(message, color = AppChrome.TextSecondary, style = MaterialTheme.typography.bodyMedium)
        }
    }
}

@Composable
private fun ActorErrorCard(message: String, onRetry: () -> Unit) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.CardShape,
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 20.dp, vertical = 20.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(message, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodyMedium)
            Button(
                onClick = onRetry,
                colors = ButtonDefaults.buttonColors(containerColor = AppChrome.Accent, contentColor = Color.White),
            ) {
                Text("重试")
            }
        }
    }
}

@Composable
private fun ActorFooter(
    loadingMore: Boolean,
    hasMore: Boolean,
    errorMessage: String?,
    onRetry: () -> Unit,
) {
    when {
        loadingMore -> Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 8.dp),
            horizontalArrangement = Arrangement.Center,
            verticalAlignment = Alignment.CenterVertically,
        ) {
            CircularProgressIndicator(modifier = Modifier.size(18.dp), color = AppChrome.AccentStrong, strokeWidth = 2.dp)
            Text("正在加载更多", color = AppChrome.TextSecondary, modifier = Modifier.padding(start = 8.dp))
        }

        !errorMessage.isNullOrBlank() -> Text(
            text = "加载失败，点击重试",
            color = MaterialTheme.colorScheme.error,
            modifier = Modifier
                .fillMaxWidth()
                .clickable(onClick = onRetry)
                .padding(vertical = 8.dp),
        )

        !hasMore -> Text(
            text = "没有更多了",
            color = AppChrome.TextMuted,
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 8.dp),
        )
    }
}

private fun resolveActorAvatarUrl(baseUrl: String, rawUrl: String?): String? = resolveResourceUrl(baseUrl, rawUrl)

private fun actorWorkTypeLabel(type: String): String {
    return when (type.trim().lowercase()) {
        "av" -> "AV"
        "movie" -> "电影"
        "episode" -> "剧集"
        "short" -> "短视频"
        else -> "视频"
    }
}

private fun resolveThumbnailUrl(baseUrl: String, rawPath: String?): String? {
    return resolveResourceUrl(baseUrl, rawPath)
}

private fun resolveResourceUrl(baseUrl: String, rawPath: String?): String? {
    val path = rawPath?.trim().orEmpty()
    if (path.isBlank()) {
        return null
    }
    if (path.startsWith("http://") || path.startsWith("https://")) {
        return path
    }
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    return if (path.startsWith("/")) "$normalizedBase$path" else "$normalizedBase/$path"
}
