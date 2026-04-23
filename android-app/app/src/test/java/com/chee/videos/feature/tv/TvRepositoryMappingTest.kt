package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.google.gson.Gson
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRepositoryMappingTest {

    @Test
    fun mapSeriesDetail_handlesMissingCollections() {
        val uiModel = tvSeriesDetailToUiModel(
            TvSeriesDetailDto(
                id = "series-1",
                title = "雾城档案",
                overview = "调查悬案",
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
                            ),
                        ),
                    ),
                ),
            ),
        )

        assertEquals("雾城档案", uiModel.title)
        assertEquals(1, uiModel.seasons.size)
        assertFalse(uiModel.seasons.first().episodes.first().playable)
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
}
