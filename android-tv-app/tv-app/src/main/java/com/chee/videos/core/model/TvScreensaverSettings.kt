package com.chee.videos.core.model

import com.google.gson.annotations.SerializedName

data class TvScreensaverSettings(
    val enabled: Boolean = true,
    val idleMinutes: Int = 10,
    val slideSeconds: Int = 5,
) {
    companion object {
        val idleMinuteOptions = listOf(1, 3, 5, 10, 15, 30)
        val slideSecondOptions = listOf(5, 10, 15, 30, 60)

        fun normalizeIdleMinutes(value: Int?): Int = value?.takeIf { it in idleMinuteOptions } ?: 10
        fun normalizeSlideSeconds(value: Int?): Int = value?.takeIf { it in slideSecondOptions } ?: 5
    }
}

data class TvScreensaverImage(
    @SerializedName("id") val id: String,
    @SerializedName("view_url") val viewUrl: String,
    @SerializedName("updated_at") val updatedAt: String = "",
)

data class TvScreensaverImagesPayload(
    @SerializedName("items") val items: List<TvScreensaverImage> = emptyList(),
    @SerializedName("next_cursor") val nextCursor: String = "",
)
