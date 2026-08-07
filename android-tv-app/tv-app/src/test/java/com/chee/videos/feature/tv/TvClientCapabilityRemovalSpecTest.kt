package com.chee.videos.feature.tv

import java.io.File
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Test

class TvClientCapabilityRemovalSpecTest {
    private val mainRoot = File("src/main")
    private val tvFeatureRoot = File(mainRoot, "java/com/chee/videos/feature/tv")

    @Test
    fun `tv app removes iptv menu route client api and dedicated implementation`() {
        assertEquals(
            listOf("电视剧", "电影", "18+", "短视频", "搜索", "设置"),
            TvHomeMenuItem.defaults().map { it.label },
        )

        listOf(
            "TvIptvModels.kt",
            "TvIptvPlaybackConfig.kt",
            "TvIptvScreen.kt",
            "TvIptvViewModel.kt",
        ).forEach { fileName ->
            assertFalse("TV IPTV 专属源码仍存在：$fileName", File(tvFeatureRoot, fileName).exists())
        }
        assertFalse(
            "TV IPTV 专属播放器布局仍存在",
            File(mainRoot, "res/layout/tv_iptv_player_view.xml").exists(),
        )

        val forbiddenTokens = listOf(
            "TvIptv",
            "TvHomeMenuItem.Iptv",
            "tv/iptv",
            "tvIptvChannels",
            "fetchIptvChannels",
            "onOpenIptv",
            "tv_iptv_player_view",
        )
        mainRoot.walkTopDown()
            .filter { it.isFile }
            .forEach { source ->
                val content = source.readText()
                forbiddenTokens.forEach { token ->
                    assertFalse("${source.path} 仍包含已下线 TV IPTV 能力：$token", content.contains(token))
                }
            }
    }
}
