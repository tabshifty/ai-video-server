package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvRemoteSessionDto
import com.google.gson.Gson
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRemotePlaybackAutoplayNextTest {
    @Test
    fun autoplayNextRequiresEnabledSessionAndNextItem() {
        assertTrue(
            shouldTvRemoteAutoplayNext(
                hasNext = true,
                actionLoading = false,
            ),
        )
    }

    @Test
    fun autoplayNextRequestStopsWhenAtEndOrBusy() {
        assertFalse(
            shouldTvRemoteAutoplayNext(
                hasNext = false,
                actionLoading = false,
            ),
        )
        assertFalse(
            shouldTvRemoteAutoplayNext(
                hasNext = true,
                actionLoading = true,
            ),
        )
        assertFalse(
            shouldTvRemoteAutoplayNext(
                hasNext = true,
                actionLoading = false,
                loading = true,
            ),
        )
    }

    @Test
    fun autoplayNextDeduplicatesCurrentEndedEvent() {
        assertFalse(
            shouldTvRemoteAutoplayNext(
                hasNext = true,
                actionLoading = false,
                currentAutoplayKey = "session:1:video",
                lastAutoplayAdvancedKey = "session:1:video",
            ),
        )
    }

    @Test
    fun missingAutoplayNextFieldDefaultsToEnabled() {
        assertTrue(
            TvRemoteSessionDto(
                sessionId = "session",
                deviceId = "device",
                autoplayNextEnabled = null,
            ).isAutoplayNextEnabled,
        )
        val parsed = Gson().fromJson(
            """
            {
              "session_id": "session",
              "device_id": "device"
            }
            """.trimIndent(),
            TvRemoteSessionDto::class.java,
        )
        assertTrue(parsed.isAutoplayNextEnabled)
        assertFalse(
            TvRemoteSessionDto(
                sessionId = "session",
                deviceId = "device",
                autoplayNextEnabled = false,
            ).isAutoplayNextEnabled,
        )
    }
}
