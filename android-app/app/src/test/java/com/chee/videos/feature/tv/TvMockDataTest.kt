package com.chee.videos.feature.tv

import org.junit.Assert.assertTrue
import org.junit.Test

class TvMockDataTest {

    @Test
    fun everySeriesHasAtLeastOneSeasonAndEpisode() {
        val all = TvMockData.allSeries()
        assertTrue(all.isNotEmpty())
        assertTrue(
            all.all { series ->
                series.seasons.isNotEmpty() && series.seasons.all { season -> season.episodes.isNotEmpty() }
            },
        )
    }
}
