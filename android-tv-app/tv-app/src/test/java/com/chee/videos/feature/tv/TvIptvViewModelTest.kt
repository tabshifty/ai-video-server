package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvIptvChannelDto
import com.chee.videos.core.model.TvIptvPayload
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest

@OptIn(ExperimentalCoroutinesApi::class)
class TvIptvViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun initLoadsChannelsAndSelectsFirstChannel() = runTest {
        val viewModel = TvIptvViewModel(
            repository = FakeTvRepository(
                iptvPayload = TvIptvPayload(
                    channelCount = 2,
                    channels = listOf(
                        TvIptvChannelDto(id = "c1", name = "新闻一台", url = "https://example.com/news.m3u8", group = "新闻", sortOrder = 10),
                        TvIptvChannelDto(id = "c2", name = "电影一台", url = "https://example.com/movie.m3u8", group = "电影", sortOrder = 20),
                    ),
                ),
            ),
        )

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("c1", state.currentChannel?.id)
        assertEquals("https://example.com/news.m3u8", state.currentChannel?.url)
        assertEquals(listOf("新闻", "电影"), state.groups.map { it.group })
    }

    @Test
    fun emptyPayloadShowsChineseEmptyState() = runTest {
        val viewModel = TvIptvViewModel(repository = FakeTvRepository(iptvPayload = TvIptvPayload()))

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("暂无可播放的 IPTV 频道", state.statusMessage)
        assertTrue(state.channels.isEmpty())
    }

    @Test
    fun initFiltersAudioOnlyChannelsBeforeSelectingDefault() = runTest {
        val viewModel = TvIptvViewModel(
            repository = FakeTvRepository(
                iptvPayload = TvIptvPayload(
                    channels = listOf(
                        TvIptvChannelDto(id = "audio", name = "CCTV-1 音频", url = "https://piccpndali.v.myalicdn.com/audio/cctv1_2.m3u8", group = "Audio", sortOrder = 0),
                        TvIptvChannelDto(id = "video", name = "CCTV-1 综合", url = "https://live.example/cctv1.m3u8", group = "央视频道", sortOrder = 1),
                    ),
                ),
            ),
        )

        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals(listOf("video"), state.channels.map { it.id })
        assertEquals("video", state.currentChannel?.id)
        assertEquals(listOf("央视频道"), state.groups.map { it.group })
    }

    @Test
    fun reload_withExistingChannelsKeepsCurrentChannelAndDoesNotReturnToPageLoading() = runTest {
        val repository = DelayedIptvTvRepository()
        val viewModel = TvIptvViewModel(repository = repository)

        repository.completeIptv(
            TvIptvPayload(
                channels = listOf(
                    TvIptvChannelDto(id = "c1", name = "新闻一台", url = "https://example.com/news.m3u8", group = "新闻", sortOrder = 10),
                    TvIptvChannelDto(id = "c2", name = "电影一台", url = "https://example.com/movie.m3u8", group = "电影", sortOrder = 20),
                ),
            ),
        )
        advanceUntilIdle()
        viewModel.selectChannel("c2")

        viewModel.reload()
        runCurrent()

        var state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("c2", state.currentChannel?.id)
        assertEquals(listOf("c1", "c2"), state.channels.map { it.id })

        repository.completeIptv(
            TvIptvPayload(
                channels = listOf(
                    TvIptvChannelDto(id = "c1", name = "新闻一台（新）", url = "https://example.com/news-new.m3u8", group = "新闻", sortOrder = 10),
                    TvIptvChannelDto(id = "c2", name = "电影一台（新）", url = "https://example.com/movie-new.m3u8", group = "电影", sortOrder = 20),
                ),
            ),
        )
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals("c2", state.currentChannel?.id)
        assertEquals("电影一台（新）", state.currentChannel?.name)
        assertEquals("https://example.com/movie-new.m3u8", state.currentChannel?.url)
        assertEquals(listOf("c1", "c2"), state.channels.map { it.id })
    }

    @Test
    fun staleReloadResponse_doesNotOverrideNewerChannelListState() = runTest {
        val repository = DelayedIptvTvRepository()
        val viewModel = TvIptvViewModel(repository = repository)

        repository.completeIptv(
            TvIptvPayload(
                channels = listOf(
                    TvIptvChannelDto(id = "c1", name = "新闻一台", url = "https://example.com/news.m3u8", group = "新闻", sortOrder = 10),
                ),
            ),
        )
        advanceUntilIdle()

        viewModel.reload()
        runCurrent()
        viewModel.reload()
        runCurrent()

        repository.completeLatestIptv(
            TvIptvPayload(
                channels = listOf(
                    TvIptvChannelDto(id = "c2", name = "电影一台", url = "https://example.com/movie.m3u8", group = "电影", sortOrder = 20),
                ),
            ),
        )
        advanceUntilIdle()

        var state = viewModel.uiState.value
        assertEquals(listOf("c2"), state.channels.map { it.id })
        assertEquals("c2", state.currentChannel?.id)

        repository.completeIptv(
            TvIptvPayload(
                channels = listOf(
                    TvIptvChannelDto(id = "stale", name = "旧频道", url = "https://example.com/stale.m3u8", group = "旧", sortOrder = 30),
                ),
            ),
        )
        advanceUntilIdle()

        state = viewModel.uiState.value
        assertEquals(listOf("c2"), state.channels.map { it.id })
        assertEquals("c2", state.currentChannel?.id)
    }

    @Test
    fun reloadFailure_withExistingChannelsKeepsPlaybackStateAndShowsInlineMessage() = runTest {
        val repository = DelayedIptvTvRepository()
        val viewModel = TvIptvViewModel(repository = repository)

        repository.completeIptv(
            TvIptvPayload(
                channels = listOf(
                    TvIptvChannelDto(id = "c1", name = "新闻一台", url = "https://example.com/news.m3u8", group = "新闻", sortOrder = 10),
                    TvIptvChannelDto(id = "c2", name = "电影一台", url = "https://example.com/movie.m3u8", group = "电影", sortOrder = 20),
                ),
            ),
        )
        advanceUntilIdle()
        viewModel.selectChannel("c2")

        viewModel.reload()
        runCurrent()
        repository.completeIptvFailure()
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertEquals(listOf("c1", "c2"), state.channels.map { it.id })
        assertEquals("c2", state.currentChannel?.id)
        assertEquals("IPTV 频道加载失败，请重试", state.statusMessage)
    }
}
