package com.chee.videos.core.model

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class AvPosterSupportTest {

    @Test
    fun `prefers scraped poster metadata over thumbnail fallback for av list item`() {
        val item = VideoListItemDto(
            id = "av-1",
            title = "SSIS-001",
            type = "av",
            thumbnailPath = "/api/v1/videos/av-1/thumbnail",
            metadata = mapOf(
                "scrape_source" to "javdb",
                "poster_source" to "video_cover",
                "poster_quality" to "primary",
                "javdb" to mapOf(
                    "poster_url" to "https://img.example/poster.jpg",
                    "poster_source" to "video_cover",
                ),
            ),
        )

        val poster = resolveAvPoster(item)

        assertEquals("https://img.example/poster.jpg", poster.rawUrl)
        assertTrue(poster.isScrapedPoster)
    }

    @Test
    fun `falls back to thumbnail when scraped poster is absent`() {
        val item = VideoListItemDto(
            id = "av-2",
            title = "SSIS-002",
            type = "av",
            thumbnailPath = "/api/v1/videos/av-2/thumbnail",
            metadata = mapOf("av_code" to "SSIS-002"),
        )

        val poster = resolveAvPoster(item)

        assertEquals("/api/v1/videos/av-2/thumbnail", poster.rawUrl)
        assertFalse(poster.isScrapedPoster)
    }

    @Test
    fun `resolves same poster source for av list item and detail`() {
        val metadata = mapOf(
            "scrape_source" to "javdb",
            "poster_source" to "video_cover",
            "poster_quality" to "primary",
            "poster_decision" to "primary_selected",
            "javdb" to mapOf(
                "poster_url" to "https://img.example/poster.jpg",
            ),
        )
        val item = VideoListItemDto(
            id = "av-3",
            title = "SSIS-003",
            type = "av",
            thumbnailPath = "/api/v1/videos/av-3/thumbnail",
            metadata = metadata,
        )
        val detail = VideoDetailDto(
            id = "av-3",
            title = "SSIS-003",
            thumbnailPath = "/api/v1/videos/av-3/thumbnail",
            metadata = metadata,
        )

        val listPoster = resolveAvPoster(item)
        val detailPoster = resolveAvPoster(detail)

        assertEquals(listPoster.rawUrl, detailPoster.rawUrl)
        assertEquals(listPoster.isScrapedPoster, detailPoster.isScrapedPoster)
        assertEquals(listPoster.posterDecision, detailPoster.posterDecision)
    }

    @Test
    fun `keeps thumbnail fallback when poster decision is invalid`() {
        val item = VideoListItemDto(
            id = "av-4",
            title = "SSIS-004",
            type = "av",
            thumbnailPath = "/api/v1/videos/av-4/thumbnail",
            metadata = mapOf(
                "scrape_source" to "javdb",
                "poster_decision" to "invalid_keep_old",
                "javdb" to mapOf(
                    "poster_url" to "/relative-invalid.jpg",
                ),
            ),
        )

        val poster = resolveAvPoster(item)

        assertEquals("/api/v1/videos/av-4/thumbnail", poster.rawUrl)
        assertFalse(poster.isScrapedPoster)
    }
}
