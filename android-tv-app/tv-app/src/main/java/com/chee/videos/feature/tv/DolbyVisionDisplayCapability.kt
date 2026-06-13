package com.chee.videos.feature.tv

import android.content.Context
import android.hardware.display.DisplayManager
import android.view.Display

internal const val HDR_TYPE_DOLBY_VISION: Int = Display.HdrCapabilities.HDR_TYPE_DOLBY_VISION
internal const val HDR_TYPE_HDR10: Int = Display.HdrCapabilities.HDR_TYPE_HDR10
internal const val HDR_TYPE_HLG: Int = Display.HdrCapabilities.HDR_TYPE_HLG
internal const val HDR_TYPE_HDR10_PLUS: Int = Display.HdrCapabilities.HDR_TYPE_HDR10_PLUS

internal enum class DolbyVisionDisplayCapabilityStatus {
    SUPPORTED,
    UNSUPPORTED,
    UNKNOWN,
}

internal enum class DolbyVisionDisplayCapabilityReason(val code: String) {
    DOLBY_VISION_PRESENT("dolby_vision_present"),
    DOLBY_VISION_MISSING("dolby_vision_missing"),
    NO_DISPLAY("no_display"),
    NO_HDR_CAPABILITIES("no_hdr_capabilities"),
    API_ERROR("api_error"),
}

internal data class DolbyVisionDisplayCapability(
    val status: DolbyVisionDisplayCapabilityStatus,
    val reason: DolbyVisionDisplayCapabilityReason,
    val supportedHdrTypeNames: List<String>,
)

internal interface DisplayHdrCapabilityReader {
    fun readCurrentDefaultDisplay(): DisplayHdrCapabilityReadResult
}

internal sealed class DisplayHdrCapabilityReadResult {
    data class Available(val supportedHdrTypes: List<Int>) : DisplayHdrCapabilityReadResult()
    object NoDisplay : DisplayHdrCapabilityReadResult()
    object NoHdrCapabilities : DisplayHdrCapabilityReadResult()
}

internal class AndroidDisplayHdrCapabilityReader(
    private val context: Context,
) : DisplayHdrCapabilityReader {
    override fun readCurrentDefaultDisplay(): DisplayHdrCapabilityReadResult {
        val displayManager = context.getSystemService(DisplayManager::class.java)
        val display = displayManager?.getDisplay(Display.DEFAULT_DISPLAY)
            ?: return DisplayHdrCapabilityReadResult.NoDisplay
        val hdrCapabilities = display.hdrCapabilities
            ?: return DisplayHdrCapabilityReadResult.NoHdrCapabilities
        @Suppress("DEPRECATION")
        val supportedHdrTypes = hdrCapabilities.supportedHdrTypes.toList()
        return DisplayHdrCapabilityReadResult.Available(
            supportedHdrTypes = supportedHdrTypes,
        )
    }
}

internal fun evaluateDolbyVisionDisplayCapability(
    reader: DisplayHdrCapabilityReader,
): DolbyVisionDisplayCapability =
    try {
        when (val result = reader.readCurrentDefaultDisplay()) {
            is DisplayHdrCapabilityReadResult.Available -> evaluateHdrTypes(result.supportedHdrTypes)
            DisplayHdrCapabilityReadResult.NoDisplay -> DolbyVisionDisplayCapability(
                status = DolbyVisionDisplayCapabilityStatus.UNKNOWN,
                reason = DolbyVisionDisplayCapabilityReason.NO_DISPLAY,
                supportedHdrTypeNames = emptyList(),
            )
            DisplayHdrCapabilityReadResult.NoHdrCapabilities -> DolbyVisionDisplayCapability(
                status = DolbyVisionDisplayCapabilityStatus.UNKNOWN,
                reason = DolbyVisionDisplayCapabilityReason.NO_HDR_CAPABILITIES,
                supportedHdrTypeNames = emptyList(),
            )
        }
    } catch (_: RuntimeException) {
        DolbyVisionDisplayCapability(
            status = DolbyVisionDisplayCapabilityStatus.UNKNOWN,
            reason = DolbyVisionDisplayCapabilityReason.API_ERROR,
            supportedHdrTypeNames = emptyList(),
        )
    }

private fun evaluateHdrTypes(supportedHdrTypes: List<Int>): DolbyVisionDisplayCapability {
    val typeNames = normalizeHdrTypeNames(supportedHdrTypes)
        ?: return DolbyVisionDisplayCapability(
            status = DolbyVisionDisplayCapabilityStatus.UNKNOWN,
            reason = DolbyVisionDisplayCapabilityReason.API_ERROR,
            supportedHdrTypeNames = emptyList(),
        )

    return if (HDR_TYPE_DOLBY_VISION in supportedHdrTypes) {
        DolbyVisionDisplayCapability(
            status = DolbyVisionDisplayCapabilityStatus.SUPPORTED,
            reason = DolbyVisionDisplayCapabilityReason.DOLBY_VISION_PRESENT,
            supportedHdrTypeNames = typeNames,
        )
    } else {
        DolbyVisionDisplayCapability(
            status = DolbyVisionDisplayCapabilityStatus.UNSUPPORTED,
            reason = DolbyVisionDisplayCapabilityReason.DOLBY_VISION_MISSING,
            supportedHdrTypeNames = typeNames,
        )
    }
}

private fun normalizeHdrTypeNames(supportedHdrTypes: List<Int>): List<String>? {
    val uniqueTypes = supportedHdrTypes.toSet()
    if (uniqueTypes.any { it !in HdrTypeNames }) {
        return null
    }
    return HdrTypeOrder
        .filter { it in uniqueTypes }
        .mapNotNull { HdrTypeNames[it] }
}

private val HdrTypeOrder = listOf(
    HDR_TYPE_DOLBY_VISION,
    HDR_TYPE_HDR10,
    HDR_TYPE_HDR10_PLUS,
    HDR_TYPE_HLG,
)

private val HdrTypeNames = mapOf(
    HDR_TYPE_DOLBY_VISION to "DOLBY_VISION",
    HDR_TYPE_HDR10 to "HDR10",
    HDR_TYPE_HDR10_PLUS to "HDR10_PLUS",
    HDR_TYPE_HLG to "HLG",
)
