package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvIptvNavigationPolicyTest {
    private val channels = listOf(
        TvIptvChannelUiModel(id = "c1", name = "新闻一台", url = "https://example.com/news.m3u8", group = "新闻", sortOrder = 10),
        TvIptvChannelUiModel(id = "c2", name = "电影一台", url = "https://example.com/movie.m3u8", group = "电影", sortOrder = 20),
        TvIptvChannelUiModel(id = "c3", name = "电影二台", url = "https://example.com/movie2.m3u8", group = "电影", sortOrder = 30),
    )

    @Test
    fun defaultChannelUsesFirstChannelInOrder() {
        assertEquals("c1", resolveDefaultIptvChannel(channels)?.id)
        assertEquals(null, resolveDefaultIptvChannel(emptyList()))
    }

    @Test
    fun audioOnlyUrlsAreNotPlayableChannels() {
        val mixed = listOf(
            TvIptvChannelUiModel(id = "audio", name = "CCTV-1 音频", url = "https://piccpndali.v.myalicdn.com/audio/cctv1_2.m3u8", group = "Audio", sortOrder = 0),
            TvIptvChannelUiModel(id = "video", name = "CCTV-1 综合", url = "https://live.example/cctv1.m3u8", group = "央视频道", sortOrder = 1),
        )

        assertEquals(false, isPlayableIptvVideoChannel(mixed[0]))
        assertEquals(true, isPlayableIptvVideoChannel(mixed[1]))
        assertEquals("video", resolveDefaultIptvChannel(mixed)?.id)
        assertEquals(listOf("video"), filterPlayableIptvVideoChannels(mixed).map { it.id })
    }

    @Test
    fun upAndDownCycleThroughChannels() {
        assertEquals("c2", resolveIptvChannelAfterStep(channels, currentChannelId = "c1", step = 1)?.id)
        assertEquals("c3", resolveIptvChannelAfterStep(channels, currentChannelId = "c1", step = -1)?.id)
        assertEquals("c1", resolveIptvChannelAfterStep(channels, currentChannelId = "c3", step = 1)?.id)
    }

    @Test
    fun rightAndBackRespectChannelListState() {
        assertEquals(TvIptvRemoteAction.ShowChannelList, resolveIptvPlaybackRemoteAction(TvIptvRemoteKey.Right))
        assertEquals(TvIptvRemoteAction.ExitPage, resolveIptvPlaybackRemoteAction(TvIptvRemoteKey.Back))
        assertEquals(TvIptvRemoteAction.CloseChannelList, resolveIptvChannelListRemoteAction(TvIptvRemoteKey.Back))
        assertEquals(TvIptvRemoteAction.SelectFocusedChannel, resolveIptvChannelListRemoteAction(TvIptvRemoteKey.Ok))
    }

    @Test
    fun groupsPreserveBackendOrderAndUseFallbackLabel() {
        val grouped = groupIptvChannels(
            channels + TvIptvChannelUiModel(id = "c4", name = "未分组", url = "https://example.com/other.m3u8", group = "", sortOrder = 40),
        )

        assertEquals(listOf("新闻", "电影", "未分组"), grouped.map { it.group })
        assertEquals(listOf("c2", "c3"), grouped[1].channels.map { it.id })
    }
}
