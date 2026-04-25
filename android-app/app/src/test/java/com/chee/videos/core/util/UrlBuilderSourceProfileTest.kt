package com.chee.videos.core.util

import org.junit.Assert.assertEquals
import org.junit.Test

class UrlBuilderSourceProfileTest {
    @Test
    fun source_defaultsToPrimaryPathWithoutProfileQuery() {
        val url = UrlBuilder.source("https://example.com", "video-1")

        assertEquals("https://example.com/api/v1/videos/video-1/source", url)
    }

    @Test
    fun source_appendsProfileQueryWhenProvided() {
        val url = UrlBuilder.source("https://example.com", "video-1", "compat")

        assertEquals("https://example.com/api/v1/videos/video-1/source?profile=compat", url)
    }
}
