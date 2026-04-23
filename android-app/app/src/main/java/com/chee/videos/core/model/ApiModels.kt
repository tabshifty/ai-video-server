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

data class ImageCollectionsPayload(
    @SerializedName("items") val items: List<ImageCollectionListItemDto> = emptyList(),
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

data class TvHomePayload(
    @SerializedName("continue_watching") val continueWatching: TvContinueWatchingDto? = null,
    @SerializedName("sections") val sections: List<TvSectionDto> = emptyList(),
    @SerializedName("search_results") val searchResults: List<TvSeriesSummaryDto> = emptyList(),
    @SerializedName("page") val page: Int = 1,
    @SerializedName("page_size") val pageSize: Int = 20,
)

data class TvContinueWatchingDto(
    @SerializedName("series_id") val seriesId: String,
    @SerializedName("series_title") val seriesTitle: String,
    @SerializedName("season_number") val seasonNumber: Int = 1,
    @SerializedName("episode_number") val episodeNumber: Int = 1,
    @SerializedName("episode_title") val episodeTitle: String = "",
    @SerializedName("poster_url") val posterUrl: String? = null,
    @SerializedName("backdrop_url") val backdropUrl: String? = null,
    @SerializedName("watch_seconds") val watchSeconds: Int = 0,
    @SerializedName("duration_seconds") val durationSeconds: Int = 0,
    @SerializedName("progress_percent") val progressPercent: Int = 0,
)

data class TvSectionDto(
    @SerializedName("title") val title: String = "",
    @SerializedName("subtitle") val subtitle: String = "",
    @SerializedName("items") val items: List<TvSeriesSummaryDto> = emptyList(),
)

data class TvSeriesSummaryDto(
    @SerializedName("id") val id: String,
    @SerializedName("title") val title: String,
    @SerializedName("overview") val overview: String? = null,
    @SerializedName("poster_url") val posterUrl: String? = null,
    @SerializedName("backdrop_url") val backdropUrl: String? = null,
    @SerializedName("first_air_date") val firstAirDate: String? = null,
    @SerializedName("total_seasons") val totalSeasons: Int = 0,
    @SerializedName("total_episodes") val totalEpisodes: Int = 0,
    @SerializedName("playable_episodes") val playableEpisodes: Int = 0,
    @SerializedName("latest_episode_air_date") val latestEpisodeAirDate: String? = null,
)

data class TvSeriesDetailDto(
    @SerializedName("id") val id: String,
    @SerializedName("title") val title: String,
    @SerializedName("overview") val overview: String? = null,
    @SerializedName("poster_url") val posterUrl: String? = null,
    @SerializedName("backdrop_url") val backdropUrl: String? = null,
    @SerializedName("first_air_date") val firstAirDate: String? = null,
    @SerializedName("total_seasons") val totalSeasons: Int = 0,
    @SerializedName("total_episodes") val totalEpisodes: Int = 0,
    @SerializedName("playable_episodes") val playableEpisodes: Int = 0,
    @SerializedName("tags") val tags: List<String> = emptyList(),
    @SerializedName("cast") val cast: List<String> = emptyList(),
    @SerializedName("seasons") val seasons: List<TvSeasonDto> = emptyList(),
)

data class TvSeasonDto(
    @SerializedName("id") val id: String,
    @SerializedName("season_number") val seasonNumber: Int = 1,
    @SerializedName("title") val title: String = "",
    @SerializedName("overview") val overview: String? = null,
    @SerializedName("poster_url") val posterUrl: String? = null,
    @SerializedName("air_date") val airDate: String? = null,
    @SerializedName("episodes") val episodes: List<TvEpisodeDto> = emptyList(),
)

data class TvEpisodeDto(
    @SerializedName("id") val id: String,
    @SerializedName("episode_number") val episodeNumber: Int = 1,
    @SerializedName("title") val title: String = "",
    @SerializedName("overview") val overview: String? = null,
    @SerializedName("runtime") val runtime: Int = 0,
    @SerializedName("air_date") val airDate: String? = null,
    @SerializedName("still_url") val stillUrl: String? = null,
    @SerializedName("video_id") val videoId: String = "",
    @SerializedName("video_title") val videoTitle: String? = null,
    @SerializedName("video_status") val videoStatus: String? = null,
    @SerializedName("watch_seconds") val watchSeconds: Int = 0,
    @SerializedName("progress_percent") val progressPercent: Int = 0,
    @SerializedName("last_watched_at") val lastWatchedAt: String? = null,
    @SerializedName("playable") val playable: Boolean = false,
)

data class HistoryItemDto(
    @SerializedName("video_id") val videoId: String,
    @SerializedName("title") val title: String,
    @SerializedName("type") val type: String = "",
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
    @SerializedName("image_collection") val imageCollection: VideoImageCollectionDto? = null,
    @SerializedName("metadata") val metadata: Map<String, Any?>? = emptyMap(),
    @SerializedName("user_state") val userState: UserStateDto = UserStateDto(),
)

data class ImageCollectionListItemDto(
    @SerializedName("id") val id: String,
    @SerializedName("name") val name: String,
    @SerializedName("description") val description: String? = null,
    @SerializedName("cover_url") val coverUrl: String? = null,
    @SerializedName("image_count") val imageCount: Int = 0,
    @SerializedName("created_at") val createdAt: String? = null,
    @SerializedName("updated_at") val updatedAt: String? = null,
)

data class ImageCollectionDetailDto(
    @SerializedName("id") val id: String,
    @SerializedName("name") val name: String,
    @SerializedName("description") val description: String? = null,
    @SerializedName("cover_url") val coverUrl: String? = null,
    @SerializedName("image_count") val imageCount: Int = 0,
    @SerializedName("created_at") val createdAt: String? = null,
    @SerializedName("updated_at") val updatedAt: String? = null,
    @SerializedName("images") val images: List<ImageCollectionImageDto> = emptyList(),
)

data class ImageCollectionImageDto(
    @SerializedName("id") val id: String,
    @SerializedName("title") val title: String,
    @SerializedName("description") val description: String? = null,
    @SerializedName("thumbnail_url") val thumbnailUrl: String? = null,
    @SerializedName("view_url") val viewUrl: String? = null,
    @SerializedName("width") val width: Int = 0,
    @SerializedName("height") val height: Int = 0,
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

data class VideoImageCollectionDto(
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
