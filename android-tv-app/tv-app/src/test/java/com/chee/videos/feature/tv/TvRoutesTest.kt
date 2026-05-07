package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvRoutesTest {

    @Test
    fun buildTvSeriesRoute_returnsExpectedRoute() {
        assertEquals("tv/series/demo", buildTvSeriesRoute("demo"))
    }

    @Test
    fun buildTvPlayerRoute_clampsInvalidSeasonAndEpisode() {
        assertEquals(
            "tv/player/demo?season=1&episode=1",
            buildTvPlayerRoute("demo", season = 0, episode = -5),
        )
    }
}
