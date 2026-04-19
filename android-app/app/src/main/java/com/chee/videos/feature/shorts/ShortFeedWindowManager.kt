package com.chee.videos.feature.shorts

import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.VideoDetailDto
import kotlin.math.min

data class ShortFeedWindowSnapshot(
    val items: List<FeedVideoDto>,
    val detailByVideoId: Map<String, VideoDetailDto> = emptyMap(),
    val detailLoadingVideoIds: Set<String> = emptySet(),
    val actionBusyVideoIds: Set<String> = emptySet(),
    val pausedByUserVideoIds: Set<String> = emptySet(),
    val detailSheetVideoId: String? = null,
)

data class ShortFeedWindowMergeResult(
    val items: List<FeedVideoDto>,
    val detailByVideoId: Map<String, VideoDetailDto>,
    val detailLoadingVideoIds: Set<String>,
    val actionBusyVideoIds: Set<String>,
    val pausedByUserVideoIds: Set<String>,
    val detailSheetVideoId: String?,
    val trimmedHeadCount: Int,
    val anchorPageAfterTrim: Int?,
)

object ShortFeedWindowManager {
    const val LoadBatchSize = 20
    const val MaxWindowSize = 60
    const val LoadMoreThreshold = 5
    private const val KeepItemsBeforeAnchor = 2

    fun shouldLoadMore(
        currentIndex: Int,
        totalCount: Int,
        loadingMore: Boolean,
    ): Boolean {
        if (loadingMore || totalCount <= 0 || currentIndex < 0) {
            return false
        }
        return currentIndex >= (totalCount - LoadMoreThreshold).coerceAtLeast(0)
    }

    fun mergeAppend(
        snapshot: ShortFeedWindowSnapshot,
        incoming: List<FeedVideoDto>,
        anchorVideoId: String?,
    ): ShortFeedWindowMergeResult {
        val currentIDs = snapshot.items.map { it.id }.toSet()
        val uniqueBatch = incoming.distinctBy { it.id }
        val appendBatch = uniqueBatch.filterNot { it.id in currentIDs }
            .ifEmpty { uniqueBatch }

        var mergedItems = snapshot.items + appendBatch
        var trimmedHeadCount = 0
        var anchorPageAfterTrim: Int? = null

        if (
            mergedItems.size > MaxWindowSize &&
            !anchorVideoId.isNullOrBlank()
        ) {
            val extraItems = mergedItems.size - MaxWindowSize
            val anchorIndex = mergedItems.indexOfFirst { it.id == anchorVideoId }
            if (anchorIndex >= 0) {
                val maxTrim = (anchorIndex - KeepItemsBeforeAnchor).coerceAtLeast(0)
                trimmedHeadCount = min(extraItems, maxTrim)
                if (trimmedHeadCount > 0) {
                    mergedItems = mergedItems.drop(trimmedHeadCount)
                    anchorPageAfterTrim = (anchorIndex - trimmedHeadCount).coerceAtLeast(0)
                }
            }
        }

        val keptIDs = mergedItems.map { it.id }.toSet()
        return ShortFeedWindowMergeResult(
            items = mergedItems,
            detailByVideoId = snapshot.detailByVideoId.filterKeys { it in keptIDs },
            detailLoadingVideoIds = snapshot.detailLoadingVideoIds.filterTo(mutableSetOf()) { it in keptIDs },
            actionBusyVideoIds = snapshot.actionBusyVideoIds.filterTo(mutableSetOf()) { it in keptIDs },
            pausedByUserVideoIds = snapshot.pausedByUserVideoIds.filterTo(mutableSetOf()) { it in keptIDs },
            detailSheetVideoId = snapshot.detailSheetVideoId?.takeIf { it in keptIDs },
            trimmedHeadCount = trimmedHeadCount,
            anchorPageAfterTrim = anchorPageAfterTrim,
        )
    }
}
