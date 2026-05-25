package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test
import org.videolan.libvlc.LibVLC

class TvIptvPlaybackDependencyTest {
    @Test
    fun libVlcIsPackagedForIptvCodecCompatibility() {
        assertEquals("org.videolan.libvlc.LibVLC", LibVLC::class.java.name)
    }
}
