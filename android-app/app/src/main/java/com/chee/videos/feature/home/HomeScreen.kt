package com.chee.videos.feature.home

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.setValue
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.util.UrlBuilder
import com.chee.videos.feature.shorts.ShortFeedScreen

private data class ContentTab(
    val title: String,
    val type: String,
)

private val tabs = listOf(
    ContentTab("短视频", "short"),
    ContentTab("电影", "movie"),
    ContentTab("电视剧", "episode"),
    ContentTab("AV", "av"),
)

@Composable
fun HomeScreen(
    baseUrl: String,
    accessToken: String,
    onOpenDetail: (String, String) -> Unit,
    onOpenShortDiscover: (mode: String, value: String, title: String) -> Unit,
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
            .background(Color.Black)
            .statusBarsPadding(),
    ) {
        TabRow(
            selectedTabIndex = selectedTab,
            containerColor = Color.Black,
            contentColor = Color.White,
        ) {
            tabs.forEachIndexed { index, tab ->
                val selected = selectedTab == index
                Tab(
                    selected = selected,
                    onClick = { selectedTab = index },
                    selectedContentColor = Color.White,
                    unselectedContentColor = Color.White.copy(alpha = 0.78f),
                    text = { Text(tab.title) },
                )
            }
        }

        when (tabs[selectedTab].type) {
            "short" -> {
                ShortFeedScreen(
                    baseUrl = baseUrl,
                    accessToken = accessToken,
                    onOpenDiscover = onOpenShortDiscover,
                )
            }

            "movie" -> {
                CategoryListSection(
                    baseUrl = baseUrl,
                    state = uiState.movie,
                    darkStyle = false,
                    onRetry = { viewModel.loadCategory("movie", force = true) },
                    onOpenDetail = onOpenDetail,
                )
            }

            "episode" -> {
                CategoryListSection(
                    baseUrl = baseUrl,
                    state = uiState.episode,
                    darkStyle = false,
                    onRetry = { viewModel.loadCategory("episode", force = true) },
                    onOpenDetail = onOpenDetail,
                )
            }

            "av" -> {
                CategoryListSection(
                    baseUrl = baseUrl,
                    state = uiState.av,
                    darkStyle = true,
                    onRetry = { viewModel.loadCategory("av", force = true) },
                    onOpenDetail = onOpenDetail,
                )
            }
        }
    }
}

@Composable
private fun CategoryListSection(
    baseUrl: String,
    state: CategoryState,
    darkStyle: Boolean,
    onRetry: () -> Unit,
    onOpenDetail: (String, String) -> Unit,
) {
    when {
        state.loading && state.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        }

        !state.errorMessage.isNullOrBlank() && state.items.isEmpty() -> {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                    Text(state.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                    Button(onClick = onRetry) {
                        Icon(Icons.Filled.Refresh, contentDescription = null)
                        Text("重试", modifier = Modifier.padding(start = 6.dp))
                    }
                }
            }
        }

        else -> {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .background(if (darkStyle) Color(0xFF090A0D) else Color(0xFFF7F8FA))
                    .padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                items(state.items, key = { it.id }) { item ->
                    VideoCard(
                        baseUrl = baseUrl,
                        item = item,
                        darkStyle = darkStyle,
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
    darkStyle: Boolean,
    onOpenDetail: (String, String) -> Unit,
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onOpenDetail(item.id, item.type) },
        colors = if (darkStyle) {
            CardDefaults.cardColors(containerColor = Color(0xFF161920))
        } else {
            CardDefaults.cardColors()
        },
    ) {
        Row(modifier = Modifier.padding(12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            val thumb = resolveThumbnailUrl(baseUrl, item.thumbnailPath)
            if (!thumb.isNullOrBlank()) {
                AsyncImage(
                    model = thumb,
                    contentDescription = item.title,
                    modifier = Modifier
                        .padding(top = 2.dp)
                        .weight(0.35f),
                )
            }
            Column(modifier = Modifier.weight(0.65f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                val titleColor = if (darkStyle) Color(0xFFEAECEF) else Color.Unspecified
                val subColor = if (darkStyle) Color(0xFFB6BECC) else MaterialTheme.colorScheme.onSurfaceVariant
                Text(
                    item.title,
                    style = MaterialTheme.typography.titleSmall,
                    fontWeight = FontWeight.SemiBold,
                    color = titleColor,
                )
                Text(
                    text = "类型：${typeLabel(item.type)}",
                    style = MaterialTheme.typography.bodySmall,
                    color = subColor,
                )
                Text(
                    text = "时长：${if (item.type == "av") formatDurationHms(item.duration) else "${item.duration} 秒"}",
                    style = MaterialTheme.typography.bodySmall,
                    color = subColor,
                )
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
    val safeSeconds = totalSeconds.coerceAtLeast(0)
    val hours = safeSeconds / 3600
    val minutes = (safeSeconds % 3600) / 60
    val seconds = safeSeconds % 60
    return String.format("%02d:%02d:%02d", hours, minutes, seconds)
}
