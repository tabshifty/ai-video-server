package com.chee.videos.feature.tvauth

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.chee.videos.core.ui.AppChrome

@Composable
fun TvAuthApprovalScreen(
    deepLink: TvAuthDeepLink,
    onFinished: () -> Unit,
    viewModel: TvAuthApprovalViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(deepLink) {
        viewModel.bind(deepLink)
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(AppChrome.PageGradient)
            .statusBarsPadding()
            .padding(horizontal = 20.dp, vertical = 24.dp),
        contentAlignment = Alignment.Center,
    ) {
        Surface(
            modifier = Modifier.fillMaxWidth(),
            color = AppChrome.Surface,
            shape = AppChrome.CardShape,
            shadowElevation = 12.dp,
        ) {
            Column(
                modifier = Modifier.padding(horizontal = 22.dp, vertical = 24.dp),
                verticalArrangement = Arrangement.spacedBy(14.dp),
            ) {
                Text(
                    text = "确认 TV 登录",
                    style = MaterialTheme.typography.headlineSmall,
                    fontWeight = FontWeight.Bold,
                    color = AppChrome.TextPrimary,
                )
                Text(
                    text = "设备：${uiState.deviceName.ifBlank { "Android TV" }}",
                    color = AppChrome.TextPrimary,
                )
                Text(
                    text = "配对码：${uiState.pairCode}",
                    color = AppChrome.TextSecondary,
                )
                Text(
                    text = "服务器：${uiState.serverBaseUrl.ifBlank { "当前已连接服务器" }}",
                    color = AppChrome.TextSecondary,
                )
                if (!uiState.message.isNullOrBlank()) {
                    Text(
                        text = uiState.message.orEmpty(),
                        color = if (uiState.isError) MaterialTheme.colorScheme.error else AppChrome.AccentStrong,
                    )
                }

                Button(
                    onClick = viewModel::approve,
                    enabled = !uiState.submitting,
                    modifier = Modifier.fillMaxWidth(),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = AppChrome.Accent,
                        contentColor = AppChrome.TextPrimary,
                    ),
                ) {
                    Text(if (uiState.submitting) "授权中..." else "确认登录这台 TV")
                }
                OutlinedButton(
                    onClick = {
                        viewModel.deny()
                        onFinished()
                    },
                    enabled = !uiState.submitting,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    Text("拒绝")
                }
                OutlinedButton(onClick = onFinished, modifier = Modifier.fillMaxWidth()) {
                    Text("稍后处理")
                }
            }
        }
    }
}
