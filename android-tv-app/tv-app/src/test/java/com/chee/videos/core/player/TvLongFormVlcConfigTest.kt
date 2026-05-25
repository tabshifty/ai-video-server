package com.chee.videos.core.player

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormVlcConfigTest {
    @Test
    fun buildSharedLibVlcArgs_keepFramesAndUseRtspTcp() {
        val args = buildTvSharedVlcArgs()

        assertTrue(args.contains("--no-drop-late-frames"))
        assertTrue(args.contains("--no-skip-frames"))
        assertTrue(args.contains("--rtsp-tcp"))
    }

    @Test
    fun buildLongFormMediaOptions_useLongFormCachingAndHardwareDecode() {
        val options = buildLongFormVlcMediaOptions()

        assertTrue(options.contains(":file-caching=1500"))
        assertTrue(options.contains(":network-caching=2000"))
        assertEquals(TvLongFormHwDecoderConfig(enabled = true, force = true), longFormHwDecoderConfig())
    }
}
