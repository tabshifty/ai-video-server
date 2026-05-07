package com.chee.videos.feature.shorts

import org.junit.Assert.assertEquals
import org.junit.Test

class ShortFeedTimeFormatTest {

    @Test
    fun `format zero milliseconds as full hms`() {
        assertEquals("00:00:00", formatShortPlaybackTimeHms(0))
    }

    @Test
    fun `format under one minute`() {
        assertEquals("00:00:59", formatShortPlaybackTimeHms(59_000))
    }

    @Test
    fun `format over one minute`() {
        assertEquals("00:01:01", formatShortPlaybackTimeHms(61_000))
    }

    @Test
    fun `format over one hour`() {
        assertEquals("01:00:01", formatShortPlaybackTimeHms(3_601_000))
    }
}
