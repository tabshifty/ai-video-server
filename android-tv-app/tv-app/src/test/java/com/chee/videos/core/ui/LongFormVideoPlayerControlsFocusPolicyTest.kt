package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class LongFormVideoPlayerControlsFocusPolicyTest {
    @Test
    fun seriesControlsWrapAcrossThreeActions() {
        val targets = listOf(
            TvControlFocusTarget.PlayPause,
            TvControlFocusTarget.Subtitle,
            TvControlFocusTarget.AudioTrack,
        )

        assertEquals(
            TvControlFocusTarget.AudioTrack,
            resolveTvControlHorizontalFocusTarget(
                current = TvControlFocusTarget.PlayPause,
                direction = TvControlFocusDirection.Left,
                targets = targets,
            ),
        )
        assertEquals(
            TvControlFocusTarget.Subtitle,
            resolveTvControlHorizontalFocusTarget(
                current = TvControlFocusTarget.PlayPause,
                direction = TvControlFocusDirection.Right,
                targets = targets,
            ),
        )
        assertEquals(
            TvControlFocusTarget.PlayPause,
            resolveTvControlHorizontalFocusTarget(
                current = TvControlFocusTarget.AudioTrack,
                direction = TvControlFocusDirection.Right,
                targets = targets,
            ),
        )
    }

    @Test
    fun missingOrEmptyFocusChainReturnsNull() {
        assertNull(
            resolveTvControlHorizontalFocusTarget(
                current = TvControlFocusTarget.PlayPause,
                direction = TvControlFocusDirection.Right,
                targets = emptyList(),
            ),
        )
        assertNull(
            resolveTvControlHorizontalFocusTarget(
                current = TvControlFocusTarget.NextEpisode,
                direction = TvControlFocusDirection.Right,
                targets = listOf(TvControlFocusTarget.PlayPause),
            ),
        )
    }
}
