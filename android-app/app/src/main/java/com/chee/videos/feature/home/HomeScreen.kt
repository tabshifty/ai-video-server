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
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.Button
import androidx.compose.material3.Card
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
    onOpenDetail: (String) -> Unit,
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
            .background(Color.Black),
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
                )
            }

            "movie" -> {
                CategoryListSection(
                    state = uiState.movie,
                    onRetry = { viewModel.loadCategory("movie", force = true) },
                    onOpenDetail = onOpenDetail,
                )
            }

            "episode" -> {
                CategoryListSection(
                    state = uiState.episode,
                    onRetry = { viewModel.loadCategory("episode", force = true) },
                    onOpenDetail = onOpenDetail,
                )
            }

            "av" -> {
                CategoryListSection(
                    state = uiState.av,
                    onRetry = { viewModel.loadCategory("av", force = true) },
                    onOpenDetail = onOpenDetail,
                )
            }
        }
    }
}

@Composable
private fun CategoryListSection(
    state: CategoryState,
    onRetry: () -> Unit,
    onOpenDetail: (String) -> Unit,
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
                    .background(Color(0xFFF7F8FA))
                    .padding(12.dp),
                verticalArrangement = Arrangement.spacedBy(10.dp),
            ) {
                items(state.items, key = { it.id }) { item ->
                    VideoCard(item = item, onOpenDetail = onOpenDetail)
                }
            }
        }
    }
}

@Composable
private fun VideoCard(
    item: VideoListItemDto,
    onOpenDetail: (String) -> Unit,
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onOpenDetail(item.id) },
    ) {
        Row(modifier = Modifier.padding(12.dp), horizontalArrangement = Arrangement.spacedBy(12.dp)) {
            val thumb = item.thumbnailPath.orEmpty().trim()
            if (thumb.startsWith("http://") || thumb.startsWith("https://")) {
                AsyncImage(
                    model = thumb,
                    contentDescription = item.title,
                    modifier = Modifier
                        .padding(top = 2.dp)
                        .weight(0.35f),
                )
            }
            Column(modifier = Modifier.weight(0.65f), verticalArrangement = Arrangement.spacedBy(6.dp)) {
                Text(item.title, style = MaterialTheme.typography.titleSmall, fontWeight = FontWeight.SemiBold)
                Text(
                    text = "类型：${typeLabel(item.type)}",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                )
                Text(
                    text = "时长：${item.duration} 秒",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
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
