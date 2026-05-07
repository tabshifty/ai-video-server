package com.chee.videos.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvQrCodeEncoderTest {

    @Test
    fun encode_createsSquareMatrixWithDarkPixels() {
        val matrix = TvQrCodeEncoder.encode(
            content = "cheevideos://tv-auth?session_id=session-1&code=ABCD12",
            size = 320,
        )

        assertEquals(320, matrix.width)
        assertEquals(320, matrix.height)
        assertTrue(matrix.rows.any { row -> row.any { it } })
    }
}
