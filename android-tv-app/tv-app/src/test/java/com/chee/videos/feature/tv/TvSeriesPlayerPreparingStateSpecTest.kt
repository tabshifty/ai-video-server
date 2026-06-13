package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesPlayerPreparingStateSpecTest {
    @Test
    fun `series player shows preparing state before unavailable error`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()
        val preparingBranch = source.indexOf("uiState.playbackPreparing")
        val preparingMessage = source.indexOf("正在准备当前分集")
        val unavailableMessage = source.indexOf("当前分集暂无可播放视频")

        assertTrue(preparingBranch >= 0)
        assertTrue(preparingMessage > preparingBranch)
        assertTrue(unavailableMessage > preparingMessage)
    }
}
