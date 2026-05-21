package com.chee.videos.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvPairingConnectionExperienceTest {
    @Test
    fun pairingActionsUseSharedTvFocusAndSingleFocusableTarget() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvPairingScreen.kt").readText()

        assertTrue("TV 配对页操作应接入共享 TV 焦点视觉", source.contains("tvFocusableGlow"))
        assertFalse(
            "TV 配对页不应裸用 foundation focusable；共享 tvFocusableGlow 已经提供唯一焦点目标",
            source.contains("import androidx.compose.foundation.focusable"),
        )
        assertFalse(
            "TV 配对页操作不能在 tvFocusableGlow 之外再叠加 .focusable()，否则遥控确认键可能落到重复焦点层",
            source.contains(".focusable()"),
        )
    }

    @Test
    fun shellLoadingUsesSharedTvStateFeedback() {
        val source = Path.of("src/main/java/com/chee/videos/tv/TvShellApp.kt").readText()
        val loadingBranch = source.substringAfter("AppRootState.Loading ->")
            .substringBefore("AppRootState.NeedServer")

        assertTrue("TV 根启动 loading 应使用共享页面级状态组件", loadingBranch.contains("TvPageLoadingState("))
        assertFalse("TV 根启动不应继续直接裸放默认进度环", loadingBranch.contains("CircularProgressIndicator("))
    }

    @Test
    fun serverScanLoadingRemainsCompactInlineFeedback() {
        val source = Path.of("src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt").readText()

        assertTrue("服务器自动嗅探仍应保留小型行内 loading 尺寸常量", source.contains("ConnectionScanLoadingIndicatorSize"))
        assertTrue("服务器自动嗅探 loading 应显式限制为小型尺寸", source.contains(".size(ConnectionScanLoadingIndicatorSize)"))
        assertFalse("服务器自动嗅探不应升级成页面级 loading", source.contains("TvPageLoadingState("))
    }
}
