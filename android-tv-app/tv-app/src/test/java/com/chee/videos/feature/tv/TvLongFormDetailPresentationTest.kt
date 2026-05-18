package com.chee.videos.feature.tv

import com.chee.videos.core.model.VideoDetailDto
import org.junit.Assert.assertEquals
import org.junit.Test

class TvLongFormDetailPresentationTest {

    @Test
    fun `builds cinematic hero with resolved backdrop poster and play labels`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "movie-1",
                title = "午夜列车",
                description = "一部悬疑长片",
                duration = 7260,
                playUrl = "/api/v1/videos/movie-1/source",
                thumbnailPath = "/poster/movie-1.jpg",
                tags = listOf("悬疑", "惊悚"),
            ),
            videoType = "movie",
        )

        assertEquals("电影", hero.eyebrow)
        assertEquals("午夜列车", hero.title)
        assertEquals("立即播放", hero.primaryActionLabel)
        assertEquals("更多信息", hero.secondaryActionLabel)
        assertEquals("https://media.example.com/poster/movie-1.jpg", hero.posterUrl)
        assertEquals("https://media.example.com/poster/movie-1.jpg", hero.backdropUrl)
    }

    @Test
    fun `builds av hero with adult eyebrow`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "av-1",
                title = "SNIS-001",
                description = "18+作品",
                duration = 5400,
                thumbnailPath = "/poster/av-1.jpg",
            ),
            videoType = "av",
        )

        assertEquals("18+", hero.eyebrow)
        assertEquals("SNIS-001", hero.title)
    }
}
