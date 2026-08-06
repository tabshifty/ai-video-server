package com.chee.videos.core.model

private val phoneSupportedVideoTypes = setOf("short", "av")

internal fun isPhoneSupportedVideoType(rawType: String?): Boolean {
    return rawType?.trim()?.lowercase() in phoneSupportedVideoTypes
}

internal data class PhoneContentSourcePage<T>(
    val items: List<T>,
    val hasMore: Boolean,
)

internal data class PhoneContentBatch<T>(
    val items: List<T>,
    val lastLoadedPage: Int,
    val hasMore: Boolean,
    val trailingError: Throwable? = null,
)

internal suspend fun <T> loadPhoneContentBatch(
    firstPage: Int,
    minimumVisibleItems: Int,
    typeOf: (T) -> String,
    fetchPage: suspend (Int) -> Result<PhoneContentSourcePage<T>>,
): Result<PhoneContentBatch<T>> {
    require(firstPage >= 1) { "firstPage must be at least 1" }
    require(minimumVisibleItems >= 1) { "minimumVisibleItems must be at least 1" }

    val visibleItems = mutableListOf<T>()
    var nextPage = firstPage
    var lastLoadedPage = firstPage - 1
    var hasMore = true

    while (visibleItems.size < minimumVisibleItems && hasMore) {
        val sourceResult = fetchPage(nextPage)
        if (sourceResult.isFailure) {
            val error = checkNotNull(sourceResult.exceptionOrNull())
            if (lastLoadedPage < firstPage) {
                return Result.failure(error)
            }
            return Result.success(
                PhoneContentBatch(
                    items = visibleItems,
                    lastLoadedPage = lastLoadedPage,
                    hasMore = true,
                    trailingError = error,
                ),
            )
        }

        val sourcePage = sourceResult.getOrThrow()
        visibleItems += sourcePage.items.filter { item -> isPhoneSupportedVideoType(typeOf(item)) }
        lastLoadedPage = nextPage
        hasMore = sourcePage.hasMore
        nextPage += 1
    }

    return Result.success(
        PhoneContentBatch(
            items = visibleItems,
            lastLoadedPage = lastLoadedPage,
            hasMore = hasMore,
        ),
    )
}

internal fun <T> resolvePhoneContentStartIndex(
    items: List<T>,
    startItemId: String,
    idOf: (T) -> String,
): Int? {
    val normalizedId = startItemId.trim()
    if (normalizedId.isBlank()) {
        return null
    }
    return items.indexOfFirst { item -> idOf(item) == normalizedId }.takeIf { it >= 0 }
}
