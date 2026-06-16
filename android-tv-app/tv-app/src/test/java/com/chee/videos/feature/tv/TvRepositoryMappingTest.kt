package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.VideoActorDto
import com.google.gson.Gson
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRepositoryMappingTest {
    @Test
    fun mapContinueWatching_preservesArtworkAndResumeState() {
        val uiModel = tvContinueWatchingToUiModel(
            TvContinueWatchingDto(
                seriesId = "series-1",
                seriesTitle = "雾城档案",
                seasonNumber = 2,
                episodeNumber = 4,
                episodeTitle = "暗线浮现",
                posterUrl = "/poster.jpg",
                backdropUrl = "/backdrop.jpg",
                watchSeconds = 512,
                progressPercent = 64,
            ),
        )

        assertEquals("/poster.jpg", uiModel.posterUrl)
        assertEquals("/backdrop.jpg", uiModel.backdropUrl)
        assertEquals(512, uiModel.watchSeconds)
        assertEquals(64, uiModel.progressPercent)
    }

    @Test
    fun mapContinueWatching_fallsBackWhenTypeIsNull() {
        val dto = Gson().fromJson(
            """
            {
              "type": null,
              "series_id": "series-1",
              "series_title": "雾城档案",
              "season_number": 2,
              "episode_number": 4,
              "episode_title": "暗线浮现",
              "progress_percent": 64
            }
            """.trimIndent(),
            TvContinueWatchingDto::class.java,
        )

        val uiModel = tvContinueWatchingToUiModel(dto)

        assertEquals("tv", uiModel.type)
        assertEquals("series-1", uiModel.seriesId)
    }

    @Test
    fun mapSeriesDetail_handlesMissingCollections() {
        val uiModel = tvSeriesDetailToUiModel(
            TvSeriesDetailDto(
                id = "series-1",
                title = "雾城档案",
                overview = "调查悬案",
                cast = listOf(
                    VideoActorDto(id = "actor-1", name = "林舟", avatarUrl = "/api/v1/actors/actor-1/avatar"),
                    VideoActorDto(id = "actor-2", name = "周岚"),
                ),
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            TvEpisodeDto(
                                id = "e1",
                                episodeNumber = 1,
                                title = "第1集",
                                videoId = "",
                                videoStatus = "",
                                runtime = 45,
                                watchSeconds = 93,
                                stillUrl = "/still/e1.jpg",
                            ),
                        ),
                    ),
                ),
            ),
        )

        assertEquals("雾城档案", uiModel.title)
        assertEquals(1, uiModel.seasons.size)
        assertEquals(2, uiModel.cast.size)
        assertEquals("林舟", uiModel.cast.first().name)
        assertEquals("/api/v1/actors/actor-1/avatar", uiModel.cast.first().avatarUrl)
        assertFalse(uiModel.seasons.first().episodes.first().playable)
        assertEquals(93, uiModel.seasons.first().episodes.first().watchSeconds)
        assertEquals("/still/e1.jpg", uiModel.seasons.first().episodes.first().stillUrl)
    }

    @Test
    fun mapSeriesDetail_preservesEpisodePlaybackCompatibilityMetadata() {
        val metadata = mapOf(
            "playback_compat" to mapOf(
                "version" to 1,
                "status" to "probe_failed",
            ),
        )
        val uiModel = tvSeriesDetailToUiModel(
            TvSeriesDetailDto(
                id = "series-1",
                title = "雾城档案",
                seasons = listOf(
                    TvSeasonDto(
                        id = "s1",
                        seasonNumber = 1,
                        title = "第一季",
                        episodes = listOf(
                            TvEpisodeDto(
                                id = "e1",
                                episodeNumber = 1,
                                title = "第1集",
                                videoId = "video-1",
                                videoStatus = "ready",
                                metadata = metadata,
                            ),
                        ),
                    ),
                ),
            ),
        )

        assertEquals(metadata, uiModel.seasons.first().episodes.first().metadata)
    }

    @Test
    fun mapSeriesDetail_handlesNullArrayFields() {
        val dto = Gson().fromJson(
            """
            {
              "id": "series-1",
              "title": "雾城档案",
              "overview": "调查悬案",
              "tags": null,
              "cast": null,
              "seasons": null
            }
            """.trimIndent(),
            TvSeriesDetailDto::class.java,
        )

        val uiModel = tvSeriesDetailToUiModel(dto)

        assertTrue(uiModel.tags.isEmpty())
        assertTrue(uiModel.cast.isEmpty())
        assertTrue(uiModel.seasons.isEmpty())
    }

    @Test
    fun mapSeriesDetail_acceptsActorObjectsFromJson() {
        val dto = Gson().fromJson(
            """
            {
              "id": "series-1",
              "title": "雾城档案",
              "cast": [
                {"id": "actor-1", "name": "林舟", "avatar_url": "/api/v1/actors/actor-1/avatar"},
                {"id": "actor-2", "name": "周岚"}
              ],
              "seasons": []
            }
            """.trimIndent(),
            TvSeriesDetailDto::class.java,
        )

        val uiModel = tvSeriesDetailToUiModel(dto)

        assertEquals(2, uiModel.cast.size)
        assertEquals("actor-1", uiModel.cast[0].id)
        assertEquals("/api/v1/actors/actor-1/avatar", uiModel.cast[0].avatarUrl)
        assertEquals("周岚", uiModel.cast[1].name)
    }
}
