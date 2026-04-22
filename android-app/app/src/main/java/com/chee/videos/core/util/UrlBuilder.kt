package com.chee.videos.core.util

object UrlBuilder {
    fun normalizeBaseUrl(raw: String): String {
        var value = raw.trim()
        if (value.isBlank()) {
            return ""
        }
        if (!value.startsWith("http://") && !value.startsWith("https://")) {
            value = "http://$value"
        }
        return value.trimEnd('/')
    }

    fun health(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/healthz"

    fun login(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/auth/login"

    fun refresh(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/auth/refresh"

    fun recommend(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/recommend"

    fun randomShort(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/short/random"

    fun shortDiscover(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/short/discover"

    fun imageCollections(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/image-collections"

    fun imageCollectionDetail(baseUrl: String, collectionId: String): String =
        "${normalizeBaseUrl(baseUrl)}/api/v1/image-collections/$collectionId"

    fun appImageView(baseUrl: String, imageId: String): String =
        "${normalizeBaseUrl(baseUrl)}/api/v1/images/$imageId/view"

    fun search(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/search"

    fun detail(baseUrl: String, videoId: String): String =
        "${normalizeBaseUrl(baseUrl)}/api/v1/videos/$videoId"

    fun source(baseUrl: String, videoId: String): String =
        "${normalizeBaseUrl(baseUrl)}/api/v1/videos/$videoId/source"

    fun toggleLike(baseUrl: String, videoId: String): String =
        "${normalizeBaseUrl(baseUrl)}/api/v1/videos/$videoId/like"

    fun toggleFavorite(baseUrl: String, videoId: String): String =
        "${normalizeBaseUrl(baseUrl)}/api/v1/videos/$videoId/favorite"

    fun toggleDislike(baseUrl: String, videoId: String): String =
        "${normalizeBaseUrl(baseUrl)}/api/v1/videos/$videoId/dislike"

    fun history(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/history"

    fun historyContinue(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/history/continue"

    fun likedVideos(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/user/liked-videos"

    fun favoritedVideos(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/user/favorited-videos"

    fun userProfile(baseUrl: String): String = "${normalizeBaseUrl(baseUrl)}/api/v1/user/profile"
}
