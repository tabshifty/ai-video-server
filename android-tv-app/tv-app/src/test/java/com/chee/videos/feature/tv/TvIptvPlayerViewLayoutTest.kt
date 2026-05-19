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

    @Test
    fun iptvScreenRendersChannelLogosWithFallbackIcon() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt")

        assertTrue("IPTV 播放页必须存在", screenPath.exists())

        val source = screenPath.readText()
        assertTrue(
            "IPTV 播放页必须使用 Coil AsyncImage 加载频道 logoUrl",
            source.contains("AsyncImage(") && source.contains("model = channel.logoUrl"),
        )
        assertTrue(
            "IPTV 频道台标加载失败或缺失时必须保留 TV 图标回退",
            source.contains("Icons.Filled.Tv"),
        )
    }

    @Test
    fun iptvTopChannelHintIsTransientCompactAndHiddenByChannelList() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt")

        assertTrue("IPTV 播放页必须存在", screenPath.exists())

        val source = screenPath.readText()
        assertTrue(
            "IPTV 顶部频道信息必须通过 shouldShowIptvChannelHint 门控，频道列表打开或异常状态时不显示",
            source.contains("shouldShowIptvChannelHint("),
        )
        assertTrue(
            "IPTV 顶部频道提示应使用 LaunchedEffect 监听当前频道并延迟 3 秒关闭",
            source.contains("delay(3_000)") && source.contains("showChannelHint = false"),
        )
        assertTrue(
            "IPTV 顶部提示不应继续作为全宽常驻条显示",
            !source.contains(".fillMaxWidth(),"),
        )
    }

    @Test
    fun iptvChannelListInitializesNearCurrentChannelWithoutOpeningAnimation() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt")

        assertTrue("IPTV 播放页必须存在", screenPath.exists())

        val source = screenPath.readText()
        assertTrue(
            "IPTV 频道列表打开时必须传入打开序号，用于按当前频道重新初始化 LazyListState",
            source.contains("channelListOpenNonce"),
        )
        assertTrue(
            "IPTV 频道列表初次显示必须使用 initialFirstVisibleItemIndex 直接定位，避免从顶部动画滚动",
            source.contains("initialFirstVisibleItemIndex = initialFirstVisibleItemIndex"),
        )
        assertTrue(
            "IPTV 频道列表打开后焦点上下移动仍应使用短动画跟随焦点",
            source.contains("listState.animateScrollToItem(itemIndex)"),
        )
    }
}
