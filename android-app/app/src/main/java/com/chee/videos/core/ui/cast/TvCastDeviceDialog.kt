package com.chee.videos.core.ui.cast

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import com.chee.videos.core.model.TvDeviceDto

/**
 * 投屏设备选择对话框。搜索结果页与短视频合集页共用，由各自 ViewModel 驱动状态。
 */
@Composable
fun TvCastDeviceDialog(
    devices: List<TvDeviceDto>,
    loading: Boolean,
    launching: Boolean,
    selectedDeviceId: String?,
    errorMessage: String?,
    onDismiss: () -> Unit,
    onSelectDevice: (String) -> Unit,
    onConfirm: () -> Unit,
) {
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("选择电视") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
                when {
                    loading -> {
                        Row(horizontalArrangement = Arrangement.spacedBy(10.dp), verticalAlignment = Alignment.CenterVertically) {
                            CircularProgressIndicator(strokeWidth = 2.dp)
                            Text("正在加载已授权电视")
                        }
                    }

                    devices.isEmpty() -> {
                        Text("当前账号还没有已授权的电视设备")
                    }

                    else -> {
                        devices.forEach { device ->
                            Row(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .clip(RoundedCornerShape(12.dp))
                                    .clickable { onSelectDevice(device.deviceId) }
                                    .padding(vertical = 6.dp),
                                verticalAlignment = Alignment.CenterVertically,
                                horizontalArrangement = Arrangement.spacedBy(10.dp),
                            ) {
                                RadioButton(
                                    selected = device.deviceId == selectedDeviceId,
                                    onClick = { onSelectDevice(device.deviceId) },
                                )
                                Column {
                                    Text(device.deviceName.ifBlank { device.deviceId })
                                    Text(
                                        text = if (device.isOnline) "当前可接收投放" else "当前不可接收投放",
                                        color = if (device.isOnline) Color(0xFF2FA36B) else Color(0xFFB18828),
                                        style = MaterialTheme.typography.bodySmall,
                                    )
                                }
                            }
                        }
                    }
                }
                if (!errorMessage.isNullOrBlank()) {
                    Text(errorMessage, color = MaterialTheme.colorScheme.error)
                }
            }
        },
        confirmButton = {
            TextButton(
                enabled = !loading && !launching && selectedDeviceId != null,
                onClick = onConfirm,
            ) {
                Text(if (launching) "投放中..." else "开始投放")
            }
        },
        dismissButton = {
            OutlinedButton(
                enabled = !launching,
                onClick = onDismiss,
            ) {
                Text("取消")
            }
        },
    )
}

