package com.chee.videos.core.model

import org.junit.Assert.assertEquals
import org.junit.Test

class AvPosterSupportVariantTest {

    @Test
    fun `prefers cropped poster path before original and thumbnail fallback`() {
        val detail = VideoDetailDto(
            id = "av-10",
            title = "SSIS-010",
            thumbnailPath = "/api/v1/videos/av-10/thumbnail",
            metadata = mapOf(
                "poster_cropped_path" to "/storage/posters/av-10-cropped.jpg",
                "poster_original_path" to "/storage/posters/av-10-original.jpg",
                "poster_variant" to "cropped",
            ),
        )

        val poster = resolveAvPoster(detail)

        assertEquals("/storage/posters/av-10-cropped.jpg", poster.rawUrl)
    }

    @Test
    fun `falls back to original poster path when cropped asset is absent`() {
        val item = VideoListItemDto(
            id = "av-11",
            title = "SSIS-011",
            type = "av",
            thumbnailPath = "/api/v1/videos/av-11/thumbnail",
            metadata = mapOf(
                "poster_original_path" to "/storage/posters/av-11-original.jpg",
                "poster_variant" to "original",
            ),
        )

        val poster = resolveAvPoster(item)

        assertEquals("/storage/posters/av-11-original.jpg", poster.rawUrl)
    }
}
