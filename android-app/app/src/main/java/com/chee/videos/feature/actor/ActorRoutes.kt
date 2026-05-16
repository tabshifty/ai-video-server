package com.chee.videos.feature.actor

import java.net.URLEncoder
import java.nio.charset.StandardCharsets

const val ActorIdArg = "actorId"
const val ActorRoutePattern = "actor/{$ActorIdArg}"

fun buildActorRoute(actorId: String): String = "actor/${encodeRouteValue(actorId)}"

fun buildVideoDetailRoute(videoId: String, videoType: String): String =
    "detail/$videoId?type=${encodeRouteValue(videoType)}"

private fun encodeRouteValue(value: String): String =
    URLEncoder.encode(value, StandardCharsets.UTF_8.name()).replace("+", "%20")
