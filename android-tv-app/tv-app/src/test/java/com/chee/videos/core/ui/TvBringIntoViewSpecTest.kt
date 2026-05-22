@file:OptIn(ExperimentalFoundationApi::class)

package com.chee.videos.core.ui

import androidx.compose.foundation.ExperimentalFoundationApi
import org.junit.Assert.assertEquals
import org.junit.Test

class TvBringIntoViewSpecTest {

    @Test
    fun `returns zero when item fully visible at top of container`() {
        val distance = calculateTvMinimalBringIntoViewScrollDistance(
            offset = 0f,
            size = 324f,
            containerSize = 2160f,
        )
        assertEquals(0f, distance, 0f)
    }

    @Test
    fun `returns zero when focused button inside hero already visible`() {
        // TvFeaturedHero (324dp) sits at top of LazyColumn after 18dp contentPadding.
        // The hero's play button (Pill ~44dp) bottom edge is well inside a 4K viewport.
        // PivotBringIntoViewSpec would still return non-zero pulling it to 30%; this spec must not.
        val distance = calculateTvMinimalBringIntoViewScrollDistance(
            offset = 270f, // play button leading edge in container coords (px)
            size = 44f,
            containerSize = 2160f,
        )
        assertEquals(0f, distance, 0f)
    }

    @Test
    fun `returns zero when item spans larger than container and currently covers it`() {
        val distance = calculateTvMinimalBringIntoViewScrollDistance(
            offset = -50f,
            size = 4000f,
            containerSize = 2160f,
        )
        assertEquals(0f, distance, 0f)
    }

    @Test
    fun `scrolls up when item leading edge above container`() {
        val distance = calculateTvMinimalBringIntoViewScrollDistance(
            offset = -120f,
            size = 200f,
            containerSize = 2160f,
        )
        assertEquals(-120f, distance, 0f)
    }

    @Test
    fun `scrolls down when item trailing edge below container`() {
        val distance = calculateTvMinimalBringIntoViewScrollDistance(
            offset = 2100f,
            size = 200f,
            containerSize = 2160f,
        )
        assertEquals(140f, distance, 0f)
    }

    @Test
    fun `picks nearest edge when item exceeds container on both sides falls back to minimum`() {
        // Item partially off-screen on bottom; leading visible.
        val distance = calculateTvMinimalBringIntoViewScrollDistance(
            offset = 2000f,
            size = 500f,
            containerSize = 2160f,
        )
        // trailingEdge - containerSize = 2500 - 2160 = 340
        // leadingEdge = 2000; trailing distance 340 is smaller, expect 340
        assertEquals(340f, distance, 0f)
    }

    @Test
    fun `spec instance delegates to scroll distance calculator`() {
        val direct = calculateTvMinimalBringIntoViewScrollDistance(
            offset = -200f,
            size = 100f,
            containerSize = 1000f,
        )
        val viaSpec = TvMinimalBringIntoViewSpec.calculateScrollDistance(
            offset = -200f,
            size = 100f,
            containerSize = 1000f,
        )
        assertEquals(direct, viaSpec, 0f)
        assertEquals(-200f, viaSpec, 0f)
    }
}
