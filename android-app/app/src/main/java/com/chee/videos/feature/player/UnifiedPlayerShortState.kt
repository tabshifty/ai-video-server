package com.chee.videos.feature.player

internal fun shouldShowUnifiedPlayerShortFitToggle(type: String): Boolean {
    return type == "short"
}

internal data class UnifiedPlayerShortUtilityRailLayout(
    val showFitModeToggle: Boolean,
    val showPlaybackModeToggle: Boolean,
    val stackVertically: Boolean,
    val spacingDp: Int,
)

internal fun buildUnifiedPlayerShortUtilityRailLayout(
    showFitModeToggle: Boolean,
): UnifiedPlayerShortUtilityRailLayout {
    return UnifiedPlayerShortUtilityRailLayout(
        showFitModeToggle = showFitModeToggle,
        showPlaybackModeToggle = true,
        stackVertically = showFitModeToggle,
        spacingDp = 12,
    )
}
