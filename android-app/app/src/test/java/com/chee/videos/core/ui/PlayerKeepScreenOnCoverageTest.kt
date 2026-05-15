package com.chee.videos.core.ui

import java.io.File
import org.junit.Assert.assertTrue
import org.junit.Test

class PlayerKeepScreenOnCoverageTest {
    @Test
    fun phonePlayerScreens_keepScreenOnWhileVideoIsPlaying() {
        val sourceRoot = File("src/main/java/com/chee/videos")
        val playerFiles = sourceRoot
            .walkTopDown()
            .filter { file -> file.isFile && file.extension == "kt" }
            .filter { file -> file.readText().contains("ExoPlayer.Builder") }
            .toList()

        val missingKeepScreenOnEffect = playerFiles
            .filterNot { file -> file.readText().contains("KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)") }
            .map { file -> file.relativeTo(sourceRoot).path }

        assertTrue(
            "播放器页面必须在实际播放时启用屏幕常亮，缺失文件：$missingKeepScreenOnEffect",
            missingKeepScreenOnEffect.isEmpty(),
        )
    }
}
