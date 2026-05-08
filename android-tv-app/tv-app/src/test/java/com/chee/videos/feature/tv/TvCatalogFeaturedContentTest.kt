package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvCatalogFeaturedContentTest {

    @Test
    fun `continue watching is promoted to featured hero first`() {
        val featured = resolveTvFeaturedContent(
            continueWatching = TvContinueWatchingUiModel(
                type = "tv",
                seriesId = "series-1",
                seriesTitle = "雾城档案",
                seasonNumber = 2,
                episodeNumber = 4,
                episodeTitle = "暗线浮现",
                backdropUrl = "/continue-backdrop.jpg",
                posterUrl = "/continue-poster.jpg",
                progressPercent = 64,
            ),
            sections = listOf(
                TvCatalogSectionUiModel(
                    title = "最近更新",
                    subtitle = "按播出时间排序",
                    items = listOf(
                        TvSeriesUiModel(
                            id = "series-2",
                            title = "静默轨道",
                            subtitle = "2026 · 1 季",
                            ratingText = "可播 8",
                            updateText = "8 集可播",
                            description = "悬疑调查",
                            tags = emptyList(),
                            cast = emptyList(),
                            seasons = emptyList(),
                            posterSeed = 1,
                        ),
                    ),
                ),
            ),
            tvSeries = emptyList(),
            movies = emptyList(),
            av = emptyList(),
        )

        requireNotNull(featured)
        assertEquals(TvFeaturedContentSource.CONTINUE_WATCHING, featured.source)
        assertEquals("series-1", featured.targetId)
        assertEquals("tv", featured.targetType)
        assertEquals("继续追剧", featured.eyebrow)
        assertEquals("/continue-backdrop.jpg", featured.backdropUrl)
    }

    @Test
    fun `falls back from section to shelf content when continue watching is missing`() {
        val sectionFeatured = resolveTvFeaturedContent(
            continueWatching = null,
            sections = listOf(
                TvCatalogSectionUiModel(
                    title = "最近更新",
                    subtitle = "按播出时间排序",
                    items = listOf(
                        TvSeriesUiModel(
                            id = "series-9",
                            title = "静默轨道",
                            subtitle = "2026 · 1 季",
                            ratingText = "可播 8",
                            updateText = "8 集可播",
                            description = "悬疑调查",
                            tags = emptyList(),
                            cast = emptyList(),
                            seasons = emptyList(),
                            posterSeed = 1,
                            backdropUrl = "/section-backdrop.jpg",
                        ),
                    ),
                ),
            ),
            tvSeries = listOf(
                TvHomeShelfItemUiModel(
                    id = "tv-1",
                    type = "tv",
                    title = "雾城档案",
                    description = "城市悬疑",
                ),
            ),
            movies = emptyList(),
            av = emptyList(),
        )
        requireNotNull(sectionFeatured)
        assertEquals(TvFeaturedContentSource.SECTION, sectionFeatured.source)
        assertEquals("series-9", sectionFeatured.targetId)

        val shelfFeatured = resolveTvFeaturedContent(
            continueWatching = null,
            sections = emptyList(),
            tvSeries = emptyList(),
            movies = listOf(
                TvHomeShelfItemUiModel(
                    id = "movie-7",
                    type = "movie",
                    title = "午夜列车",
                    description = "悬疑长片",
                    backdropUrl = "/movie-backdrop.jpg",
                ),
            ),
            av = emptyList(),
        )
        requireNotNull(shelfFeatured)
        assertEquals(TvFeaturedContentSource.SHELF, shelfFeatured.source)
        assertEquals("movie-7", shelfFeatured.targetId)
        assertEquals("movie", shelfFeatured.targetType)
        assertEquals("电影精选", shelfFeatured.eyebrow)
    }

    @Test
    fun `returns null when tv catalog is completely empty`() {
        assertNull(
            resolveTvFeaturedContent(
                continueWatching = null,
                sections = emptyList(),
                tvSeries = emptyList(),
                movies = emptyList(),
                av = emptyList(),
            ),
        )
    }
}
