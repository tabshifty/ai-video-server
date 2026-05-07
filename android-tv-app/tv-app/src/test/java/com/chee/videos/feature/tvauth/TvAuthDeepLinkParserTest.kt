package com.chee.videos.feature.tvauth

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvAuthDeepLinkParserTest {
    @Test
    fun `parse extracts session pair code and server`() {
        val link = TvAuthDeepLinkParser.parse(
            "cheevideos://tv-auth?session_id=abc-123&pair_code=ZXCV12&device_name=%E5%AE%A2%E5%8E%85%E7%94%B5%E8%A7%86&server=http%3A%2F%2F192.168.1.8%3A8080",
        )

        requireNotNull(link)
        assertEquals("abc-123", link.sessionId)
        assertEquals("ZXCV12", link.pairCode)
        assertEquals("客厅电视", link.deviceName)
        assertEquals("http://192.168.1.8:8080", link.serverBaseUrl)
    }

    @Test
    fun `parse rejects unrelated uri`() {
        assertNull(TvAuthDeepLinkParser.parse("https://example.com"))
        assertNull(TvAuthDeepLinkParser.parse("cheevideos://tv-auth?pair_code=ONLYCODE"))
    }
}
