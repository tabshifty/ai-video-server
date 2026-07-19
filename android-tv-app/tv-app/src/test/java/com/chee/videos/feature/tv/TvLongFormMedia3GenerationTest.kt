package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvLongFormMedia3GenerationTest {

    @Test
    fun `media item id round trips retry generation`() {
        val mediaItemId = buildTvLongFormMedia3ItemId("video-1", 7)

        assertEquals("video-1|retry:7", mediaItemId)
        assertEquals(
            TvLongFormMedia3EventIdentity(mediaId = "video-1", retryKey = 7),
            parseTvLongFormMedia3EventIdentity(mediaItemId),
        )
    }

    @Test
    fun `invalid media item id does not invent retry generation`() {
        assertNull(parseTvLongFormMedia3EventIdentity("video-1"))
        assertNull(parseTvLongFormMedia3EventIdentity("video-1|retry:not-a-number"))
    }

    @Test
    fun `same retry key from different media remains distinct`() {
        val first = parseTvLongFormMedia3EventIdentity(buildTvLongFormMedia3ItemId("video-1", 0))
        val second = parseTvLongFormMedia3EventIdentity(buildTvLongFormMedia3ItemId("video-2", 0))

        org.junit.Assert.assertNotEquals(first, second)
    }
}
