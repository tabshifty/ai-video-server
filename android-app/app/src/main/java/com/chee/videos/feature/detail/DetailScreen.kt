package com.chee.videos.feature.detail

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ExperimentalLayoutApi
import androidx.compose.foundation.layout.FlowRow
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.CenterAlignedTopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FilterChip
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle

@OptIn(ExperimentalMaterial3Api::class, ExperimentalLayoutApi::class)
@Composable
fun DetailScreen(
    onBack: () -> Unit,
    viewModel: DetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Scaffold(
        topBar = {
            CenterAlignedTopAppBar(
                title = { Text("视频详情") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "返回")
                    }
                },
            )
        },
    ) { innerPadding ->
        when {
            uiState.loading -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(innerPadding),
                    contentAlignment = Alignment.Center,
                ) {
                    CircularProgressIndicator()
                }
            }

            !uiState.errorMessage.isNullOrBlank() && uiState.detail == null -> {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(innerPadding),
                    contentAlignment = Alignment.Center,
                ) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(12.dp)) {
                        Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                        Button(onClick = viewModel::load) { Text("重试") }
                    }
                }
            }

            uiState.detail != null -> {
                val detail = uiState.detail
                Column(
                    modifier = Modifier
                        .fillMaxSize()
                        .padding(innerPadding)
                        .background(Color(0xFFF7F8FA))
                        .verticalScroll(rememberScrollState())
                        .padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    Text(detail!!.title, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.Bold)
                    Text(
                        text = detail.description.orEmpty().ifBlank { "暂无简介" },
                        style = MaterialTheme.typography.bodyMedium,
                    )

                    Text("播放入口（占位）：${detail.playUrl.orEmpty().ifBlank { "将使用 /videos/:id/source" }}")

                    Text("播放数据")
                    Text("时长：${detail.duration} 秒")
                    Text("播放：${detail.viewsCount}  点赞：${detail.likesCount}  收藏：${detail.favoritesCount}")

                    if (detail.tags.isNotEmpty()) {
                        FlowRow(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            detail.tags.forEach { tag ->
                                FilterChip(
                                    selected = false,
                                    onClick = {},
                                    label = { Text(tag) },
                                )
                            }
                        }
                    }

                    Column(verticalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
                        Button(onClick = viewModel::toggleLike, modifier = Modifier.fillMaxWidth()) {
                            Text(if (detail.userState.isLiked) "取消点赞" else "点赞")
                        }
                        Button(onClick = viewModel::toggleFavorite, modifier = Modifier.fillMaxWidth()) {
                            Text(if (detail.userState.isFavorited) "取消收藏" else "收藏")
                        }
                        Button(onClick = viewModel::toggleDislike, modifier = Modifier.fillMaxWidth()) {
                            Text(if (detail.userState.isDisliked) "取消不喜欢" else "不喜欢")
                        }
                    }

                    if (!uiState.errorMessage.isNullOrBlank()) {
                        Text(uiState.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                    }
                }
            }
        }
    }
}
