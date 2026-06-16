package com.chee.videos.feature.connection

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.HorizontalDivider
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.chee.videos.core.model.ServerEndpoint
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.TvLayoutSpec
import com.chee.videos.core.ui.tvFocusableGlow

private val ConnectionScanLoadingIndicatorSize = 14.dp

@Composable
fun ConnectionScreen(
    viewModel: ConnectionViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) {
        viewModel.scanLan()
    }

    LazyColumn(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .statusBarsPadding()
            .padding(16.dp),
        contentPadding = PaddingValues(bottom = TvLayoutSpec.scrollBottomSafePaddingDp.dp),
        verticalArrangement = Arrangement.spacedBy(12.dp),
    ) {
        item {
            Text(
                text = "连接家用服务器",
                style = MaterialTheme.typography.headlineSmall,
                fontWeight = FontWeight.Bold,
                color = AppChrome.TextPrimary,
            )
            Text(
                text = "先自动扫描局域网服务，找不到再手动填写 IP 和端口。",
                style = MaterialTheme.typography.bodyMedium,
                color = AppChrome.TextMuted,
            )
        }

        item {
            ConnectionPanel {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Row(
                        modifier = Modifier.fillMaxWidth(),
                        horizontalArrangement = Arrangement.SpaceBetween,
                        verticalAlignment = Alignment.CenterVertically,
                    ) {
                        Text("自动嗅探", style = MaterialTheme.typography.titleMedium, color = AppChrome.TextPrimary)
                        ConnectionActionButton(
                            text = if (uiState.scanning) "扫描中..." else "重新扫描",
                            enabled = !uiState.scanning && !uiState.connecting,
                            onClick = { viewModel.scanLan() },
                        )
                    }
                    if (uiState.scanning) {
                        Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            CircularProgressIndicator(
                                modifier = Modifier.size(ConnectionScanLoadingIndicatorSize),
                                color = AppChrome.AccentStrong,
                                strokeWidth = 2.dp,
                            )
                            Text("正在扫描同网段可用服务", color = AppChrome.TextSecondary)
                        }
                    }
                    if (!uiState.message.isNullOrBlank()) {
                        Text(
                            text = uiState.message.orEmpty(),
                            color = if (uiState.messageIsError) MaterialTheme.colorScheme.error else AppChrome.AccentStrong,
                            style = MaterialTheme.typography.bodyMedium,
                        )
                    }
                    if (uiState.discoveredEndpoints.isNotEmpty()) {
                        EndpointList(
                            title = "发现的服务",
                            endpoints = uiState.discoveredEndpoints,
                            useLabel = "使用",
                            actionsEnabled = !uiState.connecting,
                            onUse = { viewModel.useEndpoint(it.baseUrl) },
                            onDelete = null,
                        )
                    }
                }
            }
        }

        item {
            ConnectionPanel {
                Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                    Text("手动填写", style = MaterialTheme.typography.titleMedium, color = AppChrome.TextPrimary)
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp), verticalAlignment = Alignment.Top) {
                        OutlinedTextField(
                            value = uiState.hostInput,
                            onValueChange = viewModel::updateHostInput,
                            label = { Text("IP/域名") },
                            modifier = Modifier.weight(1f),
                            singleLine = true,
                            enabled = !uiState.connecting,
                        )
                        OutlinedTextField(
                            value = uiState.portInput,
                            onValueChange = viewModel::updatePortInput,
                            label = { Text("端口") },
                            modifier = Modifier.weight(0.5f),
                            singleLine = true,
                            enabled = !uiState.connecting,
                        )
                    }
                    ConnectionActionButton(
                        text = if (uiState.connecting) "连接中..." else "测试并保存",
                        enabled = !uiState.connecting,
                        modifier = Modifier.fillMaxWidth(),
                        onClick = viewModel::manualConnect,
                    )
                }
            }
        }

        if (uiState.savedEndpoints.isNotEmpty()) {
            item {
                ConnectionPanel {
                    EndpointList(
                        title = "历史地址",
                        endpoints = uiState.savedEndpoints,
                        useLabel = "连接",
                        actionsEnabled = !uiState.connecting,
                        onUse = { viewModel.useEndpoint(it.baseUrl) },
                        onDelete = { viewModel.removeSavedEndpoint(it.baseUrl) },
                    )
                }
            }
        }
    }
}

@Composable
private fun ConnectionPanel(
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit,
) {
    Surface(
        color = AppChrome.Surface,
        shape = AppChrome.SurfaceShape,
        modifier = modifier.fillMaxWidth(),
    ) {
        Column(
            modifier = Modifier.padding(14.dp),
            content = { content() },
        )
    }
}

@Composable
private fun ConnectionActionButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
    primary: Boolean = true,
) {
    Surface(
        color = when {
            !enabled -> AppChrome.SurfaceElevated.copy(alpha = 0.45f)
            primary -> AppChrome.Accent
            else -> AppChrome.SurfaceElevated.copy(alpha = 0.9f)
        },
        shape = AppChrome.PillShape,
        modifier = modifier
            .tvFocusableGlow(enabled = enabled, shape = AppChrome.PillShape, focusedScale = 1.04f)
            .clickable(enabled = enabled, onClick = onClick),
    ) {
        val contentColor = when {
            !enabled -> AppChrome.TextMuted
            primary -> AppChrome.Canvas
            else -> AppChrome.TextPrimary
        }
        Text(
            text = text,
            color = contentColor,
            style = MaterialTheme.typography.titleSmall,
            fontWeight = FontWeight.SemiBold,
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 9.dp),
        )
    }
}

@Composable
private fun EndpointList(
    title: String,
    endpoints: List<ServerEndpoint>,
    useLabel: String,
    actionsEnabled: Boolean,
    onUse: (ServerEndpoint) -> Unit,
    onDelete: ((ServerEndpoint) -> Unit)?,
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text(title, style = MaterialTheme.typography.titleMedium, color = AppChrome.TextPrimary)
        endpoints.forEachIndexed { index, endpoint ->
            Row(
                modifier = Modifier.fillMaxWidth(),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween,
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = endpoint.baseUrl,
                        style = MaterialTheme.typography.bodyLarge,
                        color = AppChrome.TextPrimary,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                    Text(
                        text = "最近成功时间：${java.text.SimpleDateFormat("MM-dd HH:mm", java.util.Locale.getDefault()).format(java.util.Date(endpoint.lastSuccessAt))}",
                        style = MaterialTheme.typography.labelSmall,
                        color = AppChrome.TextMuted,
                    )
                }
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    ConnectionActionButton(
                        text = useLabel,
                        primary = true,
                        enabled = actionsEnabled,
                        onClick = { onUse(endpoint) },
                    )
                    if (onDelete != null) {
                        ConnectionActionButton(
                            text = "删除",
                            primary = false,
                            enabled = actionsEnabled,
                            onClick = { onDelete(endpoint) },
                        )
                    }
                }
            }
            if (index != endpoints.lastIndex) {
                HorizontalDivider(color = AppChrome.Divider)
            }
        }
    }
}
