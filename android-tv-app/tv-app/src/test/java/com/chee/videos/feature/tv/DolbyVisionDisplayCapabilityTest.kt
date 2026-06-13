package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class DolbyVisionDisplayCapabilityTest {
    @Test
    fun containsDolbyVision_reportsSupported() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(
                DisplayHdrCapabilityReadResult.Available(
                    supportedHdrTypes = listOf(HDR_TYPE_HDR10, HDR_TYPE_DOLBY_VISION),
                ),
            ),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.SUPPORTED, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.DOLBY_VISION_PRESENT, result.reason)
        assertEquals(listOf("DOLBY_VISION", "HDR10"), result.supportedHdrTypeNames)
    }

    @Test
    fun emptyHdrTypeList_reportsUnsupported() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(
                DisplayHdrCapabilityReadResult.Available(supportedHdrTypes = emptyList()),
            ),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.UNSUPPORTED, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.DOLBY_VISION_MISSING, result.reason)
        assertEquals(emptyList<String>(), result.supportedHdrTypeNames)
    }

    @Test
    fun hdr10AndHlgOnly_reportsUnsupported() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(
                DisplayHdrCapabilityReadResult.Available(
                    supportedHdrTypes = listOf(HDR_TYPE_HLG, HDR_TYPE_HDR10),
                ),
            ),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.UNSUPPORTED, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.DOLBY_VISION_MISSING, result.reason)
        assertEquals(listOf("HDR10", "HLG"), result.supportedHdrTypeNames)
    }

    @Test
    fun missingDisplay_reportsUnknown() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(DisplayHdrCapabilityReadResult.NoDisplay),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.UNKNOWN, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.NO_DISPLAY, result.reason)
        assertEquals(emptyList<String>(), result.supportedHdrTypeNames)
    }

    @Test
    fun missingHdrCapabilities_reportsUnknown() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(DisplayHdrCapabilityReadResult.NoHdrCapabilities),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.UNKNOWN, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.NO_HDR_CAPABILITIES, result.reason)
        assertEquals(emptyList<String>(), result.supportedHdrTypeNames)
    }

    @Test
    fun apiException_reportsUnknown() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(error = IllegalStateException("display service failed")),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.UNKNOWN, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.API_ERROR, result.reason)
        assertEquals(emptyList<String>(), result.supportedHdrTypeNames)
    }

    @Test
    fun unparseableHdrType_reportsUnknown() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(
                DisplayHdrCapabilityReadResult.Available(
                    supportedHdrTypes = listOf(HDR_TYPE_DOLBY_VISION, -99),
                ),
            ),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.UNKNOWN, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.API_ERROR, result.reason)
        assertEquals(emptyList<String>(), result.supportedHdrTypeNames)
    }

    @Test
    fun duplicateAndUnorderedHdrTypes_areNormalized() {
        val result = evaluateDolbyVisionDisplayCapability(
            FakeDisplayHdrCapabilityReader(
                DisplayHdrCapabilityReadResult.Available(
                    supportedHdrTypes = listOf(
                        HDR_TYPE_HLG,
                        HDR_TYPE_DOLBY_VISION,
                        HDR_TYPE_HDR10,
                        HDR_TYPE_DOLBY_VISION,
                        HDR_TYPE_HLG,
                    ),
                ),
            ),
        )

        assertEquals(DolbyVisionDisplayCapabilityStatus.SUPPORTED, result.status)
        assertEquals(DolbyVisionDisplayCapabilityReason.DOLBY_VISION_PRESENT, result.reason)
        assertEquals(listOf("DOLBY_VISION", "HDR10", "HLG"), result.supportedHdrTypeNames)
    }
}

private class FakeDisplayHdrCapabilityReader(
    private val result: DisplayHdrCapabilityReadResult? = null,
    private val error: RuntimeException? = null,
) : DisplayHdrCapabilityReader {
    override fun readCurrentDefaultDisplay(): DisplayHdrCapabilityReadResult {
        error?.let { throw it }
        return checkNotNull(result)
    }
}
