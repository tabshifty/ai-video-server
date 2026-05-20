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
}
