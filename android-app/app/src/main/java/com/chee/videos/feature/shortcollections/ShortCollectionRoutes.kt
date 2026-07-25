package com.chee.videos.feature.shortcollections

import java.net.URLEncoder
import java.nio.charset.StandardCharsets

const val ShortCollectionIdArg = "collectionId"
const val ShortCollectionNameArg = "collectionName"
const val ShortCollectionContentRoutePattern =
    "short-collections/{$ShortCollectionIdArg}?$ShortCollectionNameArg={$ShortCollectionNameArg}"

fun buildShortCollectionContentRoute(collectionId: String, collectionName: String): String {
    return "short-collections/${encodeShortCollectionRouteValue(collectionId)}" +
        "?$ShortCollectionNameArg=${encodeShortCollectionRouteValue(collectionName)}"
}

private fun encodeShortCollectionRouteValue(value: String): String =
    URLEncoder.encode(value, StandardCharsets.UTF_8.name()).replace("+", "%20")
