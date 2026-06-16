package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesCastAvatarPaletteTest {
    @Test
    fun `cast avatar palette index handles negative hash seeds`() {
        assertEquals(1, tvCastAvatarPaletteIndex(seed = -4, paletteSize = 5))
        assertEquals(2, tvCastAvatarPaletteIndex(seed = Int.MIN_VALUE, paletteSize = 5))
    }

    @Test
    fun `cast avatar palette index always stays inside palette bounds`() {
        val seeds = listOf(
            Int.MIN_VALUE,
            -10_001,
            -5,
            -4,
            -1,
            0,
            1,
            4,
            5,
            10_001,
            Int.MAX_VALUE,
        )

        seeds.forEach { seed ->
            val index = tvCastAvatarPaletteIndex(seed = seed, paletteSize = 5)
            assertTrue("seed=$seed should resolve to a valid palette index", index in 0 until 5)
        }
    }

    @Test(expected = IllegalArgumentException::class)
    fun `cast avatar palette index rejects empty palette`() {
        tvCastAvatarPaletteIndex(seed = 0, paletteSize = 0)
    }
}
