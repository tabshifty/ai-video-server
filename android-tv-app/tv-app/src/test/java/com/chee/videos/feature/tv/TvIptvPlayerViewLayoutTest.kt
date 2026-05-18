package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvIptvPlayerViewLayoutTest {
    @Test
    fun iptvPlayerViewUsesTextureViewSurfaceForComposeRendering() {
        val layoutPath = Path.of("src/main/res/layout/tv_iptv_player_view.xml")

        assertTrue("IPTV 播放器布局必须存在", layoutPath.exists())

        val xml = layoutPath.readText()
        assertTrue(
            "IPTV 播放器必须使用 texture_view，避免部分 TV 设备在 Compose + SurfaceView 下只有声音没有画面",
            xml.contains("app:surface_type=\"texture_view\""),
        )
        assertTrue(
            "IPTV 播放器必须使用 Media3 PlayerView",
            xml.contains("androidx.media3.ui.PlayerView"),
        )
    }
}
