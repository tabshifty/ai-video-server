package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvIptvPlayerViewLayoutTest {
    @Test
    fun iptvPlayerViewUsesVlcVideoLayoutForCodecCompatibility() {
        val layoutPath = Path.of("src/main/res/layout/tv_iptv_player_view.xml")

        assertTrue("IPTV 播放器布局必须存在", layoutPath.exists())

        val xml = layoutPath.readText()
        assertTrue(
            "IPTV 播放器必须使用 LibVLC 的 VLCVideoLayout，提高直播源视频编码兼容性",
            xml.contains("org.videolan.libvlc.util.VLCVideoLayout"),
        )
        assertTrue(
            "IPTV 播放器不应继续使用 Media3 PlayerView，避免只依赖设备硬解导致只有声音没有画面",
            !xml.contains("androidx.media3.ui.PlayerView"),
        )
    }

    @Test
    fun iptvScreenAttachesVlcViewsWithTextureViewAndDiagnostics() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt")

        assertTrue("IPTV 播放页必须存在", screenPath.exists())

        val source = screenPath.readText()
        assertTrue(
            "IPTV 在 Compose AndroidView 中必须让 LibVLC 使用 TextureView 输出，避免 SurfaceView 黑屏",
            source.contains("attachViews(this, null, false, true)"),
        )
        assertTrue(
            "IPTV 播放页必须记录 LibVLC 事件和视频输出数量，方便区分无视频轨和输出层黑屏",
            source.contains("Log.i(IPTV_LOG_TAG"),
        )
        assertTrue(
            "IPTV 播放页必须记录 LibVLC 当前视频轨数量",
            source.contains("videoTracks=${'$'}{vlcPlayer.videoTracksCount}"),
        )
    }
}
