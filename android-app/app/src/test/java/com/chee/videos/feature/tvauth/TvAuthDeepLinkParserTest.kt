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

    @Test
    fun `resolve prefers scanned payload over launch deep link`() {
        val link = resolveTvAuthDeepLink(
            launchPayload = "cheevideos://tv-auth?session_id=launch-session&pair_code=LAUNCH1",
            scannedPayload = "cheevideos://tv-auth?session_id=scanned-session&pair_code=SCAN01",
        )

        requireNotNull(link)
        assertEquals("scanned-session", link.sessionId)
        assertEquals("SCAN01", link.pairCode)
    }

    @Test
    fun `resolve falls back to launch deep link when scanned payload is invalid`() {
        val link = resolveTvAuthDeepLink(
            launchPayload = "cheevideos://tv-auth?session_id=launch-session&pair_code=LAUNCH1",
            scannedPayload = "https://example.com/not-tv-auth",
        )

        requireNotNull(link)
        assertEquals("launch-session", link.sessionId)
        assertEquals("LAUNCH1", link.pairCode)
    }
}
