package com.chee.videos.feature.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.homeContentTabs
import com.chee.videos.core.util.UrlBuilder
import com.chee.videos.feature.shorts.ShortFeedScreen

private val tabs = homeContentTabs

@Composable
fun HomeScreen(
    baseUrl: String,
    accessToken: String,
    onOpenDetail: (String, String) -> Unit,
    onOpenShortDiscover: (mode: String, value: String, title: String) -> Unit,
    onOpenImageCollectionViewer: (String) -> Unit,
    viewModel: HomeViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    var selectedTab by rememberSaveable { mutableIntStateOf(0) }

    LaunchedEffect(selectedTab) {
        val tab = tabs[selectedTab]
        if (tab.type != "short") {
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
                    CategoryListSection(
                        baseUrl = baseUrl,
                        state = uiState.episode,
                        categoryTitle = tabs[selectedTab].title,
                        onRetry = { viewModel.loadCategory("episode", force = true) },
                        onOpenDetail = onOpenDetail,
                    )
                }

                "av" -> {
                    CategoryListSection(
                        baseUrl = baseUrl,
                        state = uiState.av,
                        categoryTitle = tabs[selectedTab].title,
                        onRetry = { viewModel.loadCategory("av", force = true) },
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
