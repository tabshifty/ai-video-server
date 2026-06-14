package com.chee.videos.feature.tv

import androidx.lifecycle.SavedStateHandle
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvTrackPreference
import kotlinx.coroutines.ExperimentalCoroutinesApi
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest

@OptIn(ExperimentalCoroutinesApi::class)
class TvSeriesPlayerViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun init_prefersRouteSeasonAndEpisodeWhenPlayable() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                                tvEpisode(id = "e4", number = 4, title = "第4集", videoId = "video-4", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(
                mapOf(
                    TvSeriesIdArg to "series-1",
                    TvSeasonArg to 1,
                    TvEpisodeArg to 4,
                ),
            ),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals(1, state.selectedSeasonNumber)
        assertEquals(4, state.selectedEpisodeNumber)
        assertEquals("video-4", state.currentVideoId)
    }

    @Test
    fun init_exposesActiveBaseUrlForSubtitleResolution() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                baseUrl = "https://media.example.com",
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals("https://media.example.com", viewModel.uiState.value.baseUrl)
    }

    @Test
    fun retry_afterErrorRequestsDetailAgain() = runTest {
        val repository = FakeTvRepository(detailError = IllegalStateException("播放页失败"))
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        viewModel.awaitIdle()
        viewModel.retry()
        viewModel.awaitIdle()

        assertEquals(listOf("series-1", "series-1"), repository.detailRequests)
    }

    @Test
    fun init_exposesEpisodeWatchSecondsForResume() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e1",
                                    number = 1,
                                    title = "第1集",
                                    videoId = "video-1",
                                    videoStatus = "ready",
                                    watchSeconds = 187,
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals(187, selectedEpisode(viewModel.uiState.value)?.watchSeconds)
    }

    @Test
    fun init_readsTvSeekStepSeconds() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                tvSeekStepSeconds = 20,
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals(20, viewModel.uiState.value.tvSeekStepSeconds)
    }

    @Test
    fun init_restoresSavedSubtitleSelectionForCurrentVideo() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                subtitlePreferences = mapOf("video-1" to TvTrackPreference(language = "zh-CN", type = "default")),
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e1",
                                    number = 1,
                                    title = "第1集",
                                    videoId = "video-1",
                                    videoStatus = "ready",
                                    subtitleTracks = listOf(
                                        SubtitleTrackDto(
                                            id = "subtitle-zh",
                                            languageCode = "zh-CN",
                                            format = "ass",
                                            url = "/subtitles/subtitle-zh/file",
                                            isDefault = true,
                                            available = true,
                                        ),
                                    ),
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals("subtitle-zh", viewModel.uiState.value.selectedSubtitleTrackId)
    }

    @Test
    fun init_restoresSavedAudioSelectionForCurrentVideo() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                audioPreferences = mapOf("video-1" to TvTrackPreference(language = "zh", type = "default")),
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e1",
                                    number = 1,
                                    title = "第1集",
                                    videoId = "video-1",
                                    videoStatus = "ready",
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        assertEquals(null, viewModel.uiState.value.selectedAudioTrackId)
        assertEquals(TvTrackPreference(language = "zh", type = "default"), viewModel.uiState.value.selectedAudioPreference)
    }

    @Test
    fun selectAudioTrack_savesCurrentVideoPreference() = runTest {
        val repository = FakeTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(
                                id = "e1",
                                number = 1,
                                title = "第1集",
                                videoId = "video-1",
                                videoStatus = "ready",
                            ),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val preference = TvTrackPreference(language = "zh", type = "default")
        viewModel.selectAudioTrack("audio-zh-51", preference)
        advanceUntilIdle()

        assertEquals("audio-zh-51", viewModel.uiState.value.selectedAudioTrackId)
        assertEquals(preference, repository.readTvAudioPreference("video-1"))
    }

    @Test
    fun init_withoutRouteEpisode_prefersMostRecentlyWatchedPlayableEpisode() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e1",
                                    number = 1,
                                    title = "第1集",
                                    videoId = "video-1",
                                    videoStatus = "ready",
                                    watchSeconds = 90,
                                    lastWatchedAt = "2026-04-25T08:00:00Z",
                                ),
                            ),
                        ),
                        TvSeasonDto(
                            id = "s2",
                            seasonNumber = 2,
                            title = "第二季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e2",
                                    number = 3,
                                    title = "第3集",
                                    videoId = "video-2",
                                    videoStatus = "ready",
                                    watchSeconds = 45,
                                    lastWatchedAt = "2026-04-26T09:30:00Z",
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedSeasonNumber)
        assertEquals(3, state.selectedEpisodeNumber)
        assertEquals("video-2", state.currentVideoId)
    }

    @Test
    fun init_withoutRouteEpisode_skipsPlaybackCompatibilityBlockedEpisode() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(
                                    id = "e1",
                                    number = 1,
                                    title = "第1集",
                                    videoId = "video-1",
                                    videoStatus = "ready",
                                    watchSeconds = 120,
                                    lastWatchedAt = "2026-04-26T09:30:00Z",
                                    metadata = probeFailedPlaybackCompatibilityMetadata(),
                                ),
                                tvEpisode(
                                    id = "e2",
                                    number = 2,
                                    title = "第2集",
                                    videoId = "video-2",
                                    videoStatus = "ready",
                                    watchSeconds = 30,
                                    lastWatchedAt = "2026-04-25T08:00:00Z",
                                ),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals(1, state.selectedSeasonNumber)
        assertEquals(2, state.selectedEpisodeNumber)
        assertEquals("video-2", state.currentVideoId)
        assertTrue(state.canPlayCurrentEpisode)
    }

    @Test
    fun init_marksPlaybackPreparingUntilInitialSourceReady() = runTest {
        val repository = DelayedSourceTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        advanceUntilIdle()

        val preparingState = viewModel.uiState.value
        assertEquals(1, preparingState.selectedSeasonNumber)
        assertEquals(1, preparingState.selectedEpisodeNumber)
        assertEquals("", preparingState.currentVideoId)
        assertEquals("", preparingState.currentSourceUrl)
        assertFalse(preparingState.canPlayCurrentEpisode)
        assertTrue(preparingState.playbackPreparing)
        assertEquals(null, preparingState.playbackBlockedMessage)

        repository.completeSourceUrl("video-1")
        advanceUntilIdle()

        val readyState = viewModel.uiState.value
        assertEquals("video-1", readyState.currentVideoId)
        assertEquals("https://example.com/video-1.m3u8", readyState.currentSourceUrl)
        assertTrue(readyState.canPlayCurrentEpisode)
        assertFalse(readyState.playbackPreparing)
    }

    @Test
    fun nextEpisode_skipsToFollowingEpisodeAndUpdatesVideoId() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                                tvEpisode(id = "e2", number = 2, title = "第2集", videoId = "video-2", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(
                mapOf(TvSeriesIdArg to "series-1"),
            ),
        )
        viewModel.awaitIdle()
        viewModel.nextEpisode()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedEpisodeNumber)
        assertEquals("video-2", state.currentVideoId)
    }

    @Test
    fun nextEpisode_skipsUnplayableEpisodesAndCrossesSeasonBoundary() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                                tvEpisode(id = "e2", number = 2, title = "第2集"),
                            ),
                        ),
                        TvSeasonDto(
                            id = "s2",
                            seasonNumber = 2,
                            title = "第二季",
                            episodes = listOf(
                                tvEpisode(id = "e3", number = 1, title = "第1集", videoId = "video-3", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        viewModel.nextEpisode()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedSeasonNumber)
        assertEquals(1, state.selectedEpisodeNumber)
        assertEquals("video-3", state.currentVideoId)
    }

    @Test
    fun nextEpisode_skipsPlaybackCompatibilityBlockedEpisodes() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                                tvEpisode(
                                    id = "e2",
                                    number = 2,
                                    title = "第2集",
                                    videoId = "video-2",
                                    videoStatus = "ready",
                                    metadata = probeFailedPlaybackCompatibilityMetadata(),
                                ),
                                tvEpisode(id = "e3", number = 3, title = "第3集", videoId = "video-3", videoStatus = "ready"),
                            ),
                        ),
                    ),
                ),
            ),
            SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        viewModel.nextEpisode()

        val state = viewModel.uiState.value
        assertEquals(3, state.selectedEpisodeNumber)
        assertEquals("video-3", state.currentVideoId)
    }

    @Test
    fun unboundEpisode_cannotPlay() = runTest {
        val viewModel = TvSeriesPlayerViewModel(
            repository = FakeTvRepository(
                detailPayload = tvSeriesDetail(
                    seasons = listOf(
                        TvSeasonDto(
                            id = "s1",
                            seasonNumber = 1,
                            title = "第一季",
                            episodes = listOf(
                                tvEpisode(id = "e1", number = 1, title = "第1集"),
                            ),
                        ),
                    ),
                ),
            ),
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertFalse(state.canPlayCurrentEpisode)
        assertFalse(state.playbackPreparing)
        assertEquals("", state.currentVideoId)
    }

    @Test
    fun playbackCompatibilityBlockedEpisode_doesNotBuildSourceUrl() = runTest {
        val repository = FakeTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(
                                id = "e1",
                                number = 1,
                                title = "第1集",
                                videoId = "video-1",
                                videoStatus = "ready",
                                metadata = probeFailedPlaybackCompatibilityMetadata(),
                            ),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals("video-1", state.currentVideoId)
        assertEquals("", state.currentSourceUrl)
        assertFalse(state.canPlayCurrentEpisode)
        assertFalse(state.playbackPreparing)
        assertEquals("播放兼容信息不完整，暂不能确认安全播放", state.playbackBlockedMessage)
        assertTrue(repository.sourceUrlRequests.isEmpty())
    }

    @Test
    fun outputDolbyVisionEpisode_buildsSourceUrlForDedicatedRouteDecision() = runTest {
        val repository = FakeTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(
                                id = "e1",
                                number = 1,
                                title = "第1集",
                                videoId = "video-1",
                                videoStatus = "ready",
                                metadata = playbackCompatibilityMetadata(
                                    sourceDolbyVision = true,
                                    outputDolbyVision = true,
                                ),
                            ),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals("video-1", state.currentVideoId)
        assertEquals("https://example.com/video-1.m3u8", state.currentSourceUrl)
        assertTrue(state.canPlayCurrentEpisode)
        assertFalse(state.playbackPreparing)
        assertEquals(null, state.playbackBlockedMessage)
        assertEquals(listOf("video-1"), repository.sourceUrlRequests)
    }

    @Test
    fun dolbyVisionTranscodedOutput_blocksWithoutSourceUrl() = runTest {
        val repository = FakeTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(
                                id = "e1",
                                number = 1,
                                title = "第1集",
                                videoId = "video-1",
                                videoStatus = "ready",
                                metadata = playbackCompatibilityMetadata(
                                    sourceDolbyVision = true,
                                    outputDolbyVision = false,
                                ),
                            ),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        val state = viewModel.uiState.value
        assertEquals("video-1", state.currentVideoId)
        assertEquals("", state.currentSourceUrl)
        assertFalse(state.canPlayCurrentEpisode)
        assertFalse(state.playbackPreparing)
        assertEquals("该视频来源为杜比视界，当前压缩结果可能无法安全播放", state.playbackBlockedMessage)
        assertEquals(emptyList<String>(), repository.sourceUrlRequests)
    }

    @Test
    fun selectEpisode_marksPlaybackPreparingUntilNewSourceReady() = runTest {
        val repository = DelayedSourceTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                            tvEpisode(id = "e2", number = 2, title = "第2集", videoId = "video-2", videoStatus = "ready"),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeSourceUrl("video-1")
        advanceUntilIdle()
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)

        viewModel.selectEpisode(2)
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(2, state.selectedEpisodeNumber)
        assertEquals("", state.currentVideoId)
        assertEquals("", state.currentSourceUrl)
        assertFalse(state.canPlayCurrentEpisode)
        assertTrue(state.playbackPreparing)
        assertEquals(null, state.playbackBlockedMessage)

        repository.completeSourceUrl("video-2")
        advanceUntilIdle()

        val readyState = viewModel.uiState.value
        assertEquals("video-2", readyState.currentVideoId)
        assertEquals("https://example.com/video-2.m3u8", readyState.currentSourceUrl)
        assertTrue(readyState.canPlayCurrentEpisode)
        assertFalse(readyState.playbackPreparing)
    }

    @Test
    fun selectEpisode_ignoresLateSourceUrlFromPreviousEpisode() = runTest {
        val repository = DelayedSourceTvRepository(
            detailPayload = tvSeriesDetail(
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisode(id = "e1", number = 1, title = "第1集", videoId = "video-1", videoStatus = "ready"),
                            tvEpisode(id = "e2", number = 2, title = "第2集", videoId = "video-2", videoStatus = "ready"),
                        ),
                    ),
                ),
            ),
        )
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )

        repository.completeSourceUrl("video-1")
        advanceUntilIdle()
        assertEquals(1, viewModel.uiState.value.selectedEpisodeNumber)
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)

        viewModel.selectEpisode(2)
        advanceUntilIdle()
        viewModel.selectEpisode(1)
        advanceUntilIdle()
        assertEquals(1, viewModel.uiState.value.selectedEpisodeNumber)
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)

        repository.completeSourceUrl("video-2")
        advanceUntilIdle()
        assertEquals(1, viewModel.uiState.value.selectedEpisodeNumber)
        assertEquals("video-1", viewModel.uiState.value.currentVideoId)
        assertEquals("https://example.com/video-1.m3u8", viewModel.uiState.value.currentSourceUrl)
        assertTrue(viewModel.uiState.value.canPlayCurrentEpisode)
    }

    @Test
    fun reportHistory_withPositiveProgress_callsRepository() = runTest {
        val repository = FakeTvRepository()
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        viewModel.reportHistory("video-1", watchSeconds = 38, completed = true)
        advanceUntilIdle()

        assertEquals(listOf(TvHistoryReport("video-1", 38, true)), repository.historyReports)
    }

    @Test
    fun reportHistory_ignoresInvalidInput() = runTest {
        val repository = FakeTvRepository()
        val viewModel = TvSeriesPlayerViewModel(
            repository = repository,
            savedStateHandle = SavedStateHandle(mapOf(TvSeriesIdArg to "series-1")),
        )
        viewModel.awaitIdle()

        viewModel.reportHistory("", watchSeconds = 38, completed = false)
        viewModel.reportHistory("video-1", watchSeconds = 0, completed = false)
        advanceUntilIdle()

        assertTrue(repository.historyReports.isEmpty())
    }
}

private fun probeFailedPlaybackCompatibilityMetadata(): Map<String, Any?> =
    mapOf(
        "playback_compat" to mapOf(
            "version" to 1,
            "status" to "probe_failed",
        ),
    )

private fun playbackCompatibilityMetadata(
    sourceDolbyVision: Boolean,
    outputDolbyVision: Boolean,
): Map<String, Any?> =
    mapOf(
        "playback_compat" to mapOf(
            "version" to 1,
            "status" to "ok",
            "source" to mapOf("dolby_vision" to sourceDolbyVision),
            "output" to mapOf("dolby_vision" to outputDolbyVision),
        ),
    )
