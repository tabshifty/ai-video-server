package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvRoutesTest {

    @Test
    fun buildTvSeriesRoute_returnsExpectedRoute() {
        assertEquals("tv/series/demo", buildTvSeriesRoute("demo"))
    }

    @Test
    fun buildTvSeriesRoute_encodesReservedCharactersInSeriesId() {
        assertEquals(
            "tv/series/series%2Falpha%3Fpart%3D1",
            buildTvSeriesRoute("series/alpha?part=1"),
        )
    }

    @Test
    fun buildTvPlayerRoute_clampsInvalidSeasonAndEpisode() {
        assertEquals(
            "tv/player/demo?season=1&episode=1",
            buildTvPlayerRoute("demo", season = 0, episode = -5),
        )
    }

    @Test
    fun buildTvPlayerRoute_encodesReservedCharactersInSeriesId() {
        assertEquals(
            "tv/player/series%2Falpha%231?season=2&episode=5",
            buildTvPlayerRoute("series/alpha#1", season = 2, episode = 5),
        )
    }

    @Test
    fun buildTvLongFormDetailRoute_includesVideoType() {
        assertEquals(
            "tv/detail/movie-1?videoType=movie",
            buildTvLongFormDetailRoute(videoId = "movie-1", videoType = "movie"),
        )
    }

    @Test
    fun buildTvLongFormDetailRoute_encodesReservedCharactersInVideoId() {
        assertEquals(
            "tv/detail/movie%2F2026%3Fcut%3Dtv?videoType=movie",
            buildTvLongFormDetailRoute(videoId = "movie/2026?cut=tv", videoType = "movie"),
        )
    }

    @Test
    fun buildTvLongFormPlayerRoute_defaultsToMovieForUnsupportedType() {
        assertEquals(
            "tv/long-form-player/demo?videoType=movie",
            buildTvLongFormPlayerRoute(videoId = "demo", videoType = "unknown"),
        )
    }

    @Test
    fun buildTvLongFormPlayerRoute_encodesReservedCharactersInVideoId() {
        assertEquals(
            "tv/long-form-player/av%2F2026%23scene-1?videoType=av",
            buildTvLongFormPlayerRoute(videoId = "av/2026#scene-1", videoType = "av"),
        )
    }
}
