package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvStateFeedbackSpecTest {
    @Test
    fun `shared tv state feedback components exist`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/core/ui/TvStateFeedback.kt")
        assertTrue("TV 端应提供统一状态反馈组件", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue(source.contains("fun TvPageLoadingState("))
        assertTrue(source.contains("fun TvInlineLoadingState("))
        assertTrue(source.contains("fun TvEmptyState("))
        assertTrue(source.contains("fun TvErrorState("))
        assertTrue(source.contains("tvFocusableGlow"))
        assertTrue("状态页加载指示应使用参考图暖金 accent", source.contains("color = AppChrome.Accent"))
        assertTrue("状态页面板应使用暗玻璃 elevated surface", source.contains("color = AppChrome.SurfaceElevated"))
        assertTrue("状态页操作按钮应使用暖金柔和底色", source.contains("color = AppChrome.AccentSoft"))
        assertTrue("状态页不应残留旧蓝青操作底色", !source.contains("0x2239D7E8"))
    }
}
