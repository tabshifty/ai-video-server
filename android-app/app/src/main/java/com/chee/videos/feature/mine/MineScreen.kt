package com.chee.videos.feature.mine

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Logout
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Router
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import coil.compose.AsyncImage
import com.chee.videos.core.util.UrlBuilder

private val MineBackground = Color(0xFF090A0D)
private val MineSectionBackground = Color(0xFF14171D)
private val MineCardBackground = Color(0xFF181B22)
private val MineAccent = Color(0xFFFF5A7A)

@Composable
fun MineScreen(
    baseUrl: String,
    onOpenPlayer: (source: String, videoId: String) -> Unit,
    onSwitchServer: () -> Unit,
    onLogout: () -> Unit,
    viewModel: MineViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val currentList = when (uiState.selectedSection) {
        MineSection.HISTORY -> uiState.history
        MineSection.FAVORITE -> uiState.favorite
        MineSection.LIKE -> uiState.like
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MineBackground)
            .padding(horizontal = 14.dp, vertical = 10.dp),
    ) {
        ProfileHeader(
            baseUrl = baseUrl,
            username = uiState.profile?.username.orEmpty(),
            role = uiState.profile?.role.orEmpty(),
            email = uiState.profile?.email,
            profileLoading = uiState.profileLoading,
            profileError = uiState.profileErrorMessage,
            onRefreshProfile = viewModel::refreshProfile,
            onSwitchServer = onSwitchServer,
            onLogout = onLogout,
        )

        Spacer(modifier = Modifier.height(14.dp))

        SectionTabs(
            selected = uiState.selectedSection,
            onSelect = viewModel::selectSection,
        )

        Spacer(modifier = Modifier.height(12.dp))

        when {
            currentList.loading && currentList.items.isEmpty() -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator(color = Color.White)
                }
            }

            !currentList.errorMessage.isNullOrBlank() && currentList.items.isEmpty() -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(10.dp)) {
                        Text(currentList.errorMessage.orEmpty(), color = MaterialTheme.colorScheme.error)
                        Button(onClick = viewModel::refreshCurrentSection) {
                            Text("重试")
                        }
                    }
                }
            }

            currentList.items.isEmpty() -> {
                Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                    Text("暂无内容", color = Color(0xFFB9BEC9))
                }
            }

            else -> {
                LazyColumn(
                    modifier = Modifier.fillMaxSize(),
                    verticalArrangement = Arrangement.spacedBy(10.dp),
                ) {
                    items(currentList.items, key = { item -> item.videoId }) { item ->
                        MineVideoCard(
                            baseUrl = baseUrl,
                            item = item,
                            onClick = {
                                onOpenPlayer(uiState.selectedSection.source, item.videoId)
                            },
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun ProfileHeader(
    baseUrl: String,
    username: String,
    role: String,
    email: String?,
    profileLoading: Boolean,
    profileError: String?,
    onRefreshProfile: () -> Unit,
    onSwitchServer: () -> Unit,
    onLogout: () -> Unit,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(18.dp),
        color = MineSectionBackground,
    ) {
        Column(modifier = Modifier.fillMaxWidth().padding(14.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically,
            ) {
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                    val name = username.ifBlank { "家用用户" }
                    Box(
                        modifier = Modifier
                            .size(44.dp)
                            .clip(CircleShape)
                            .background(MineAccent.copy(alpha = 0.25f)),
                        contentAlignment = Alignment.Center,
                    ) {
                        Text(
                            text = name.take(1),
                            color = Color.White,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.Bold,
                        )
                    }

                    Column(verticalArrangement = Arrangement.spacedBy(2.dp)) {
                        Text(
                            text = name,
                            color = Color.White,
                            style = MaterialTheme.typography.titleMedium,
                            fontWeight = FontWeight.SemiBold,
                        )
                        Text(
                            text = role.ifBlank { "member" },
                            color = Color(0xFFE2E6EF),
                            style = MaterialTheme.typography.labelMedium,
                        )
                    }
                }

                Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
                    IconButton(onClick = onRefreshProfile) {
                        Icon(Icons.Filled.Refresh, contentDescription = "刷新资料", tint = Color.White)
                    }
                    IconButton(onClick = onSwitchServer) {
                        Icon(Icons.Filled.Router, contentDescription = "切换服务器", tint = Color.White)
                    }
                    IconButton(onClick = onLogout) {
                        Icon(Icons.Filled.Logout, contentDescription = "退出登录", tint = Color.White)
                    }
                }
            }

            val endpoint = UrlBuilder.normalizeBaseUrl(baseUrl)
                .removePrefix("http://")
                .removePrefix("https://")
            Text(
                text = "服务地址：$endpoint",
                color = Color(0xFFABB3C2),
                style = MaterialTheme.typography.bodySmall,
            )

            if (!email.isNullOrBlank()) {
                Text(
                    text = "邮箱：$email",
                    color = Color(0xFFABB3C2),
                    style = MaterialTheme.typography.bodySmall,
                )
            }

            if (profileLoading) {
                LinearProgressIndicator(
                    modifier = Modifier.fillMaxWidth(),
                    color = MineAccent,
                    trackColor = Color(0xFF252A34),
                )
            }

            if (!profileError.isNullOrBlank()) {
                Text(
                    text = profileError,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall,
                )
            }
        }
    }
}

@Composable
private fun SectionTabs(
    selected: MineSection,
    onSelect: (MineSection) -> Unit,
) {
    Surface(
        modifier = Modifier.fillMaxWidth(),
        shape = RoundedCornerShape(14.dp),
        color = MineSectionBackground,
    ) {
        Row(modifier = Modifier.fillMaxWidth().padding(6.dp), horizontalArrangement = Arrangement.spacedBy(6.dp)) {
            MineSection.values().forEach { section ->
                val isSelected = section == selected
                val bg = if (isSelected) MineAccent.copy(alpha = 0.22f) else Color.Transparent
                val textColor = if (isSelected) Color.White else Color(0xFF9DA5B5)
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .clip(RoundedCornerShape(12.dp))
                        .background(bg)
                        .clickable { onSelect(section) }
                        .padding(vertical = 10.dp),
                    contentAlignment = Alignment.Center,
                ) {
                    Text(
                        text = section.title,
                        color = textColor,
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = if (isSelected) FontWeight.SemiBold else FontWeight.Normal,
                    )
                }
            }
        }
    }
}

@Composable
private fun MineVideoCard(
    baseUrl: String,
    item: MineVideoItem,
    onClick: () -> Unit,
) {
    Surface(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(16.dp))
            .clickable { onClick() },
        shape = RoundedCornerShape(16.dp),
        color = MineCardBackground,
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(10.dp),
            horizontalArrangement = Arrangement.spacedBy(10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            val thumbUrl = resolveThumbnailUrl(baseUrl, item.thumbnailPath)
            if (!thumbUrl.isNullOrBlank()) {
                AsyncImage(
                    model = thumbUrl,
                    contentDescription = item.title,
                    modifier = Modifier
                        .size(width = 120.dp, height = 68.dp)
                        .clip(RoundedCornerShape(10.dp)),
                )
            } else {
                Box(
                    modifier = Modifier
                        .size(width = 120.dp, height = 68.dp)
                        .clip(RoundedCornerShape(10.dp))
                        .background(Color(0xFF242A33)),
                )
            }

            Column(
                modifier = Modifier.weight(1f),
                verticalArrangement = Arrangement.spacedBy(6.dp),
            ) {
                Text(
                    text = item.title,
                    color = Color.White,
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                )
                if (!item.subtitle.isNullOrBlank()) {
                    Text(
                        text = item.subtitle,
                        color = Color(0xFFADB4C4),
                        style = MaterialTheme.typography.bodySmall,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
                item.progress?.let { progress ->
                    LinearProgressIndicator(
                        progress = progress.coerceIn(0f, 1f),
                        modifier = Modifier
                            .fillMaxWidth()
                            .height(3.dp)
                            .clip(CircleShape),
                        color = MineAccent,
                        trackColor = Color(0xFF2B3040),
                    )
                }
            }
        }
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
