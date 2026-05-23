package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesAutoplaySpecTest {
    @Test
    fun resolveNextPlayableEpisode_sortsByNumberAndSkipsUnplayableEpisodes() {
        val series = tvSeriesDetailToUiModel(
            TvSeriesDetailDto(
                id = "series-1",
                title = "雾城档案",
                seasons = listOf(
                    TvSeasonDto(
                        id = "s2",
                        seasonNumber = 2,
                        title = "第二季",
                        episodes = listOf(
                            tvEpisodeDto(id = "s2e1", number = 1, title = "第1集", playable = true, videoId = "video-s2e1"),
                        ),
                    ),
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisodeDto(id = "s1e3", number = 3, title = "第3集", playable = false),
                            tvEpisodeDto(id = "s1e2", number = 2, title = "第2集", playable = true, videoId = "video-s1e2"),
                            tvEpisodeDto(id = "s1e1", number = 1, title = "第1集", playable = true, videoId = "video-s1e1"),
                        ),
                    ),
                ),
            ),
        )

        val next = resolveNextPlayableEpisode(series, currentSeasonNumber = 1, currentEpisodeNumber = 1)

        assertEquals(1, next?.seasonNumber)
        assertEquals(2, next?.episodeNumber)
        assertEquals("第2集", next?.title)
    }

    @Test
    fun resolveNextPlayableEpisode_crossesToNextSeasonAndReturnsNullAtSeriesEnd() {
        val series = tvSeriesDetailToUiModel(
            TvSeriesDetailDto(
                id = "series-1",
                title = "雾城档案",
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            tvEpisodeDto(id = "s1e2", number = 2, title = "第2集", playable = true, videoId = "video-s1e2"),
                            tvEpisodeDto(id = "s1e1", number = 1, title = "第1集", playable = true, videoId = "video-s1e1"),
                        ),
                    ),
                    TvSeasonDto(
                        id = "s2",
                        seasonNumber = 2,
                        title = "第二季",
                        episodes = listOf(
                            tvEpisodeDto(id = "s2e2", number = 2, title = "第2集", playable = false),
                            tvEpisodeDto(id = "s2e1", number = 1, title = "第1集", playable = true, videoId = "video-s2e1"),
                        ),
                    ),
                ),
            ),
        )

        val crossSeasonNext = resolveNextPlayableEpisode(series, currentSeasonNumber = 1, currentEpisodeNumber = 2)
        val noNext = resolveNextPlayableEpisode(series, currentSeasonNumber = 2, currentEpisodeNumber = 1)

        assertEquals(2, crossSeasonNext?.seasonNumber)
        assertEquals(1, crossSeasonNext?.episodeNumber)
        assertEquals("第1集", crossSeasonNext?.title)
        assertNull(noNext)
    }

    @Test
    fun autoplayCountdownTickRemaining_usesWholeSecondsFromRemainingMillis() {
        assertEquals(10, autoplayCountdownTickRemaining(10_000L))
        assertEquals(10, autoplayCountdownTickRemaining(9_999L))
        assertEquals(5, autoplayCountdownTickRemaining(5_000L))
        assertEquals(1, autoplayCountdownTickRemaining(1L))
        assertEquals(0, autoplayCountdownTickRemaining(0L))
    }

    @Test
    fun tvSeriesAutoplaySetting_defaultsMissingPreferenceToEnabled() {
        assertTrue(TvSeriesAutoplaySetting.parse(null))
        assertTrue(TvSeriesAutoplaySetting.parse(true))
        assertFalse(TvSeriesAutoplaySetting.parse(false))
    }

    @Test
    fun shouldShowAutoplayPromptCard_requiresAllowedStateAndPositiveRemainingWindow() {
        val allowed = AutoplayPromptGuardInput(
            isPlaying = true,
            autoplayEnabled = true,
            hasNextEpisode = true,
            isPlayerError = false,
            isSelectorVisible = false,
            isBackConfirmVisible = false,
            isLoading = false,
            isCanceledForCurrentEpisode = false,
            remainingMs = 5_000L,
            durationMs = 120_000L,
        )

        val blockedBySelector = allowed.copy(isSelectorVisible = true)
        val blockedByCancel = allowed.copy(isCanceledForCurrentEpisode = true)

        assertTrue(shouldShowAutoplayPromptCard(allowed))
        assertFalse(shouldShowAutoplayPromptCard(blockedBySelector))
        assertFalse(shouldShowAutoplayPromptCard(blockedByCancel))
    }
}

private fun tvEpisodeDto(
    id: String,
    number: Int,
    title: String,
    playable: Boolean,
    videoId: String = "",
): TvEpisodeDto = TvEpisodeDto(
    id = id,
    episodeNumber = number,
    title = title,
    playable = playable,
    videoId = videoId,
)
