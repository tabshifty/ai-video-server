package com.chee.videos.feature.tv

import androidx.media3.exoplayer.hls.HlsMediaSource
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Test
import org.videolan.libvlc.LibVLC

class TvIptvPlaybackDependencyTest {
    @Test
    fun hlsMediaSourceFactoryIsPackagedForM3u8Channels() {
        assertNotNull(HlsMediaSource.Factory::class.java)
    }

    @Test
    fun libVlcIsPackagedForIptvCodecCompatibility() {
        assertEquals("org.videolan.libvlc.LibVLC", LibVLC::class.java.name)
    }
}
