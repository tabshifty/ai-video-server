package com.chee.videos.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvDeviceIdentityTest {
    @Test
    fun buildTvDeviceIdFromParts_prefersAndroidId() {
        val actual = buildTvDeviceIdFromParts(
            androidId = "abc123",
            fallbackId = "fallback-1",
        )

        assertEquals("android-tv-abc123", actual)
    }

    @Test
    fun buildTvDeviceIdFromParts_fallsBackToStoredFallbackId() {
        val actual = buildTvDeviceIdFromParts(
            androidId = "   ",
            fallbackId = "fallback-1",
        )

        assertEquals("android-tv-fallback-1", actual)
    }

    @Test
    fun buildLegacyTvDeviceId_keepsLegacyModelFingerprint() {
        val actual = buildLegacyTvDeviceId(
            device = "AFTKA",
            model = "Fire TV",
        )

        assertEquals("android-tv-aftka-fire tv", actual)
    }
}
