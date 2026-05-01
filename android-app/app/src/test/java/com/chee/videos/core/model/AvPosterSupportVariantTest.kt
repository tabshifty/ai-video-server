package com.chee.videos.core.model

import org.junit.Assert.assertEquals
import org.junit.Test

class AvPosterSupportVariantTest {

    @Test
    fun `list item prefers cropped poster path before original and thumbnail fallback`() {
        val item = VideoListItemDto(
            id = "av-10",
            title = "SSIS-010",
            type = "av",
            thumbnailPath = "/api/v1/videos/av-10/thumbnail",
            metadata = mapOf(
                "poster_cropped_path" to "/storage/posters/av-10-cropped.jpg",
                "poster_original_path" to "/storage/posters/av-10-original.jpg",
                "poster_variant" to "cropped",
            ),
        )

        val poster = resolveAvPoster(item)

        assertEquals("/storage/posters/av-10-cropped.jpg", poster.rawUrl)
    }

    @Test
    fun `detail prefers original poster path before cropped and thumbnail fallback`() {
        val detail = VideoDetailDto(
            id = "av-11",
            title = "SSIS-011",
            thumbnailPath = "/api/v1/videos/av-11/thumbnail",
            metadata = mapOf(
                "poster_cropped_path" to "/storage/posters/av-11-cropped.jpg",
                "poster_original_path" to "/storage/posters/av-11-original.jpg",
                "poster_variant" to "cropped",
            ),
        )

        val poster = resolveAvPoster(detail)

        assertEquals("/storage/posters/av-11-original.jpg", poster.rawUrl)
    }

    @Test
    fun `list item falls back to original poster path when cropped asset is absent`() {
        val item = VideoListItemDto(
            id = "av-12",
            title = "SSIS-012",
            type = "av",
            thumbnailPath = "/api/v1/videos/av-12/thumbnail",
            metadata = mapOf(
                "poster_original_path" to "/storage/posters/av-12-original.jpg",
                "poster_variant" to "original",
            ),
        )

        val poster = resolveAvPoster(item)

        assertEquals("/storage/posters/av-12-original.jpg", poster.rawUrl)
    }

    @Test
    fun `detail falls back to cropped poster path when original asset is absent`() {
        val detail = VideoDetailDto(
            id = "av-13",
            title = "SSIS-013",
            thumbnailPath = "/api/v1/videos/av-13/thumbnail",
            metadata = mapOf(
                "poster_cropped_path" to "/storage/posters/av-13-cropped.jpg",
                "poster_variant" to "cropped",
            ),
        )

        val poster = resolveAvPoster(detail)

        assertEquals("/storage/posters/av-13-cropped.jpg", poster.rawUrl)
    }
}
