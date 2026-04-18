package com.chee.videos.core.model

import com.google.gson.annotations.SerializedName

data class ApiEnvelope<T>(
    @SerializedName("code") val code: Int = -1,
    @SerializedName("msg") val msg: String = "",
    @SerializedName("data") val data: T? = null,
)

data class HealthPayload(
    @SerializedName("status") val status: String? = null,
)

data class LoginRequest(
    @SerializedName("username") val username: String,
    @SerializedName("password") val password: String,
)

data class LoginPayload(
    @SerializedName("user_id") val userId: String,
    @SerializedName("access_token") val accessToken: String,
    @SerializedName("refresh_token") val refreshToken: String,
)

data class RefreshPayload(
    @SerializedName("access_token") val accessToken: String,
    @SerializedName("refresh_token") val refreshToken: String,
)

data class FeedPayload(
    @SerializedName("items") val items: List<FeedVideoDto> = emptyList(),
)

data class FeedVideoDto(
    @SerializedName("id") val id: String,
    @SerializedName("title") val title: String,
    @SerializedName("type") val type: String,
    @SerializedName("transcoded_path") val transcodedPath: String? = null,
    @SerializedName("thumbnail_path") val thumbnailPath: String? = null,
    @SerializedName("duration_seconds") val durationSeconds: Int? = null,
    @SerializedName("duration") val duration: Int? = null,
)

data class SearchPayload(
    @SerializedName("items") val items: List<VideoListItemDto> = emptyList(),
    @SerializedName("total_count") val totalCount: Int = 0,
    @SerializedName("page") val page: Int = 1,
    @SerializedName("page_size") val pageSize: Int = 20,
)

data class ContinueHistoryPayload(
    @SerializedName("items") val items: List<HistoryItemDto> = emptyList(),
    @SerializedName("total_count") val totalCount: Int = 0,
    @SerializedName("page") val page: Int = 1,
    @SerializedName("page_size") val pageSize: Int = 20,
)

data class HistoryItemDto(
    @SerializedName("video_id") val videoId: String,
    @SerializedName("title") val title: String,
    @SerializedName("thumbnail_path") val thumbnailPath: String? = null,
    @SerializedName("duration") val duration: Int = 0,
    @SerializedName("watch_seconds") val watchSeconds: Int = 0,
    @SerializedName("progress") val progress: Float = 0f,
    @SerializedName("last_watched_at") val lastWatchedAt: String? = null,
)

data class VideoListItemDto(
    @SerializedName("id") val id: String,
    @SerializedName("title") val title: String,
    @SerializedName("type") val type: String,
    @SerializedName("thumbnail_path") val thumbnailPath: String? = null,
    @SerializedName("duration") val duration: Int = 0,
    @SerializedName("created_at") val createdAt: String? = null,
)

data class VideoDetailDto(
    @SerializedName("id") val id: String,
    @SerializedName("title") val title: String,
    @SerializedName("description") val description: String? = null,
    @SerializedName("play_url") val playUrl: String? = null,
    @SerializedName("thumbnail_path") val thumbnailPath: String? = null,
    @SerializedName("duration") val duration: Int = 0,
    @SerializedName("views_count") val viewsCount: Long = 0,
    @SerializedName("likes_count") val likesCount: Long = 0,
    @SerializedName("favorites_count") val favoritesCount: Long = 0,
    @SerializedName("tags") val tags: List<String>? = emptyList(),
    @SerializedName("actors") val actors: List<VideoActorDto>? = emptyList(),
    @SerializedName("collections") val collections: List<VideoCollectionDto>? = emptyList(),
    @SerializedName("metadata") val metadata: Map<String, Any?>? = emptyMap(),
    @SerializedName("user_state") val userState: UserStateDto = UserStateDto(),
)

data class VideoActorDto(
    @SerializedName("id") val id: String,
    @SerializedName("name") val name: String,
    @SerializedName("avatar_url") val avatarUrl: String? = null,
)

data class VideoCollectionDto(
    @SerializedName("id") val id: String,
    @SerializedName("name") val name: String,
    @SerializedName("cover_url") val coverUrl: String? = null,
)

data class UserStateDto(
    @SerializedName("is_liked") val isLiked: Boolean = false,
    @SerializedName("is_favorited") val isFavorited: Boolean = false,
    @SerializedName("is_disliked") val isDisliked: Boolean = false,
    @SerializedName("watch_seconds") val watchSeconds: Int = 0,
)

data class ActionTogglePayload(
    @SerializedName("action") val action: String,
    @SerializedName("enabled") val enabled: Boolean,
)

data class RecordHistoryRequest(
    @SerializedName("video_id") val videoId: String,
    @SerializedName("watch_seconds") val watchSeconds: Int,
    @SerializedName("completed") val completed: Boolean,
)

data class ServerEndpoint(
    val baseUrl: String,
    val lastSuccessAt: Long,
)

data class SessionTokens(
    val accessToken: String,
    val refreshToken: String,
)

data class UserProfileDto(
    @SerializedName("username") val username: String = "",
    @SerializedName("email") val email: String? = null,
    @SerializedName("role") val role: String = "",
    @SerializedName("created_at") val createdAt: String? = null,
)

sealed class AppRootState {
    data object Loading : AppRootState()
    data object NeedServer : AppRootState()
    data object NeedLogin : AppRootState()
    data class Ready(val baseUrl: String, val accessToken: String) : AppRootState()
}

class AppException(message: String) : RuntimeException(message)

class AuthExpiredException : RuntimeException("登录已失效，请重新登录")
