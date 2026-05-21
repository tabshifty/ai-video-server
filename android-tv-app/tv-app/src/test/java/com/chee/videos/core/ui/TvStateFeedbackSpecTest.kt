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
    }
}
