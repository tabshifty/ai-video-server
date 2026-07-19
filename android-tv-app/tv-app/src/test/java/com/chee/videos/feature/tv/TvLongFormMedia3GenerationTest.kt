package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvLongFormMedia3GenerationTest {

    @Test
    fun `media item id round trips retry generation`() {
        val mediaItemId = buildTvLongFormMedia3ItemId("video-1", 7)

        assertEquals("video-1|retry:7", mediaItemId)
        assertEquals(7, parseTvLongFormMedia3RetryKey(mediaItemId))
    }

    @Test
    fun `invalid media item id does not invent retry generation`() {
        assertNull(parseTvLongFormMedia3RetryKey("video-1"))
        assertNull(parseTvLongFormMedia3RetryKey("video-1|retry:not-a-number"))
    }
}
