package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvIptvPlaybackConfigTest {
    @Test
    fun iptvPlaybackConfigUsesSmoothFirstLiveCaching() {
        assertEquals(4_000, TV_IPTV_NETWORK_CACHING_MS)

        val args = buildIptvLibVlcArgs()
        val options = buildIptvMediaOptions()

        assertTrue(args.contains("--network-caching=4000"))
        assertTrue(options.contains(":network-caching=4000"))
    }

    @Test
    fun iptvPlaybackConfigDoesNotForceLowLatencyClockOptions() {
        val allOptions = buildIptvLibVlcArgs() + buildIptvMediaOptions()

        assertFalse(allOptions.any { it.contains("clock-jitter=0") })
        assertFalse(allOptions.any { it.contains("clock-synchro=0") })
    }
}

