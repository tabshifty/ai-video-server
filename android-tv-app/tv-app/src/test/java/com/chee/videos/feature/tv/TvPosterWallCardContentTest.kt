package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class TvPosterWallCardContentTest {
    @Test
    fun buildTvPosterWallCardContent_prefersPosterUrlAndKeepsDescriptionBelowArtwork() {
        val content = buildTvPosterWallCardContent(
            baseUrl = "https://media.example.com",
            item = TvCatalogWallItemUiModel(
                id = "movie-1",
                type = "movie",
                title = "午夜列车",
                description = "悬疑长片",
                posterUrl = "/poster.jpg",
                backdropUrl = "/backdrop.jpg",
            ),
        )

        assertEquals("https://media.example.com/poster.jpg", content.posterUrl)
        assertEquals("午夜列车", content.title)
        assertEquals("悬疑长片", content.description)
        assertFalse(content.showPosterPlaceholder)
    }

    @Test
    fun buildTvPosterWallCardContent_usesPlaceholderWhenPosterUrlIsMissing() {
        val content = buildTvPosterWallCardContent(
            baseUrl = "https://media.example.com",
            item = TvCatalogWallItemUiModel(
                id = "movie-2",
                type = "movie",
                title = "无海报条目",
                description = "只有文本信息",
                posterUrl = null,
                backdropUrl = "/backdrop.jpg",
            ),
        )

        assertNull(content.posterUrl)
        assertEquals("无海报条目", content.title)
        assertEquals("只有文本信息", content.description)
        assertTrue(content.showPosterPlaceholder)
    }
}
