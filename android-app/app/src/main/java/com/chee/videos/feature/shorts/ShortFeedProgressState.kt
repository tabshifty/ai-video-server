package com.chee.videos.feature.shorts

internal fun shouldShowShortFeedProgressBar(
    currentVideoId: String?,
    detailSheetOpen: Boolean,
    pagerSettled: Boolean,
): Boolean {
    return !currentVideoId.isNullOrBlank() && !detailSheetOpen && pagerSettled
}
