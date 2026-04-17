package com.chee.videos.feature.auth

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle

@Composable
fun LoginScreen(
    onSwitchServer: () -> Unit,
    viewModel: LoginViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colorScheme.surface)
            .padding(horizontal = 24.dp, vertical = 48.dp),
        verticalArrangement = Arrangement.spacedBy(14.dp),
    ) {
        Text(
            text = "登录家用视频",
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold,
        )
        Text(
            text = "请输入账号（用户名或邮箱）和密码",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
        )

        OutlinedTextField(
            value = uiState.identity,
            onValueChange = viewModel::updateIdentity,
            modifier = Modifier.fillMaxWidth(),
            label = { Text("账号") },
            singleLine = true,
            enabled = !uiState.loading,
        )
        OutlinedTextField(
            value = uiState.password,
            onValueChange = viewModel::updatePassword,
            modifier = Modifier.fillMaxWidth(),
            label = { Text("密码") },
            visualTransformation = PasswordVisualTransformation(),
            singleLine = true,
            enabled = !uiState.loading,
        )

        if (!uiState.errorMessage.isNullOrBlank()) {
            Text(
                text = uiState.errorMessage.orEmpty(),
                color = MaterialTheme.colorScheme.error,
                style = MaterialTheme.typography.bodySmall,
            )
        }

        Button(
            onClick = viewModel::login,
            enabled = !uiState.loading,
            modifier = Modifier.fillMaxWidth(),
        ) {
            Text(if (uiState.loading) "登录中..." else "登录")
        }

        TextButton(onClick = onSwitchServer) {
            Text("更换服务器地址")
        }
    }
}
