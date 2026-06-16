package com.chee.videos.feature.connection

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test
import java.nio.file.Path

class ConnectionScreenLoadingSpecTest {
    @Test
    fun `server scan loading uses compact inline indicator`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt")
        val source = sourcePath.toFile().readText()

        assertTrue("服务器扫描 loading 应限制为小型行内尺寸", source.contains("ConnectionScanLoadingIndicatorSize"))
        assertTrue("服务器扫描 loading 应显式限制宽高，避免默认进度环过大", source.contains(".size(ConnectionScanLoadingIndicatorSize)"))
        assertTrue("服务器扫描 loading 应使用较细线宽", source.contains("strokeWidth = 2.dp"))
        assertFalse("服务器扫描 loading 不应只限制高度", source.contains("CircularProgressIndicator(modifier = Modifier.height(20.dp)"))
    }

    @Test
    fun `connection page uses tv panels and shared focus actions`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt")
        val source = sourcePath.toFile().readText()

        assertTrue("连接页区块应使用 TV 深色面板组件", source.contains("ConnectionPanel("))
        assertTrue("连接页操作应使用共享 TV 焦点视觉", source.contains("ConnectionActionButton("))
        assertTrue("连接页操作按钮内部必须接入 tvFocusableGlow", source.contains(".tvFocusableGlow("))
        assertFalse("连接页不应继续使用默认 Material Card 作为主要区块", source.contains("import androidx.compose.material3.Card"))
        assertFalse("连接页不应继续使用默认 Material Button 作为主要操作", source.contains("import androidx.compose.material3.Button"))
        assertFalse("连接页不应继续使用默认 Material TextButton 作为主要操作", source.contains("import androidx.compose.material3.TextButton"))
        assertFalse("连接页不应继续使用默认 Material IconButton 作为主要操作", source.contains("import androidx.compose.material3.IconButton"))
    }

    @Test
    fun `endpoint rows stay stable while connecting and with long urls`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt")
        val source = sourcePath.toFile().readText()

        assertTrue("发现服务入口连接中必须禁用，避免重复触发网络探测", source.contains("actionsEnabled = !uiState.connecting"))
        assertTrue("EndpointList 必须把 actionsEnabled 传给使用按钮", source.contains("enabled = actionsEnabled"))
        assertTrue("历史地址长 URL 必须限制单行", source.contains("maxLines = 1"))
        assertTrue("历史地址长 URL 必须省略，避免挤压 TV 操作按钮", source.contains("overflow = TextOverflow.Ellipsis"))
    }
}
