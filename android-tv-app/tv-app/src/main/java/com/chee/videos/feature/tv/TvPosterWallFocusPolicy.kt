package com.chee.videos.feature.tv

internal fun resolveTvPosterWallRestoreIndex(itemIds: List<String>, lastFocusedItemId: String?): Int =
    itemIds.indexOf(lastFocusedItemId).coerceAtLeast(0)
