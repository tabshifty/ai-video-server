package com.chee.videos.feature.tvauth

import com.chee.videos.core.util.UrlBuilder
import java.net.URLDecoder
import java.nio.charset.StandardCharsets

data class TvAuthDeepLink(
    val sessionId: String,
    val pairCode: String,
    val deviceName: String,
    val serverBaseUrl: String? = null,
)

object TvAuthDeepLinkParser {
    fun parse(raw: String?): TvAuthDeepLink? {
        val value = raw?.trim().orEmpty()
        if (value.isBlank()) {
            return null
        }
        val prefix = "cheevideos://tv-auth"
        if (!value.startsWith(prefix, ignoreCase = true)) {
            return null
        }
        val query = value.substringAfter('?', "")
        val params = query
            .split('&')
            .mapNotNull { part ->
                val key = part.substringBefore('=', "").trim()
                if (key.isBlank()) {
                    null
                } else {
                    key to URLDecoder.decode(part.substringAfter('=', ""), StandardCharsets.UTF_8)
                }
            }
            .toMap()
        val sessionId = params["session_id"].orEmpty().trim()
        val pairCode = params["pair_code"].orEmpty().trim()
        if (sessionId.isBlank() || pairCode.isBlank()) {
            return null
        }
        return TvAuthDeepLink(
            sessionId = sessionId,
            pairCode = pairCode,
            deviceName = params["device_name"].orEmpty().trim(),
            serverBaseUrl = params["server"]
                ?.trim()
                ?.takeIf { it.isNotBlank() }
                ?.let(UrlBuilder::normalizeBaseUrl),
        )
    }
}
