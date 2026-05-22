package com.chee.videos.core.network

import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.ContinueHistoryPayload
import com.chee.videos.core.model.FeedPayload
import com.chee.videos.core.model.ImageCollectionDetailDto
import com.chee.videos.core.model.ImageCollectionsPayload
import com.chee.videos.core.model.LoginPayload
import com.chee.videos.core.model.LoginRequest
import com.chee.videos.core.model.RecordHistoryRequest
import com.chee.videos.core.model.RefreshPayload
import com.chee.videos.core.model.SearchPayload
import com.chee.videos.core.model.TvAuthCreateEnvelope
import com.chee.videos.core.model.TvAuthSessionCreateRequest
import com.chee.videos.core.model.TvAuthStatusEnvelope
import com.chee.videos.core.model.TvCatalogWallPayload
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvIptvPayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.UserProfileDto
import com.chee.videos.core.model.VideoDetailDto
import retrofit2.http.Body
import retrofit2.http.GET
import retrofit2.http.Header
import retrofit2.http.POST
import retrofit2.http.Query
import retrofit2.http.Url

interface ApiService {
    @POST
    suspend fun login(
        @Url url: String,
        @Body body: LoginRequest,
    ): ApiEnvelope<LoginPayload>

    @POST
    suspend fun refresh(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<RefreshPayload>

    @GET
    suspend fun recommend(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<FeedPayload>

    @GET
    suspend fun randomShort(
        @Url url: String,
        @Query("page_size") pageSize: Int,
        @Query("exclude_ids") excludeIds: String? = null,
    ): ApiEnvelope<FeedPayload>

    @GET
    suspend fun shortDiscover(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("mode") mode: String,
        @Query("tag") tag: String? = null,
        @Query("collection_id") collectionID: String? = null,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<SearchPayload>

    @GET
    suspend fun tvHome(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("kind") kind: String? = null,
        @Query("q") keyword: String? = null,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<TvHomePayload>

    @GET
    suspend fun tvSearch(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("q") keyword: String,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<TvSearchPayload>

    @GET
    suspend fun tvCatalogWall(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("kind") kind: String,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
        @Query("sort_by") sortBy: String,
        @Query("sort_order") sortOrder: String,
    ): ApiEnvelope<TvCatalogWallPayload>

    @GET
    suspend fun tvIptvChannels(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<TvIptvPayload>

    @GET
    suspend fun tvSeriesDetail(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<TvSeriesDetailDto>

    @POST
    suspend fun createTvAuthSession(
        @Url url: String,
        @Body body: TvAuthSessionCreateRequest,
    ): TvAuthCreateEnvelope

    @GET
    suspend fun getTvAuthSession(
        @Url url: String,
    ): TvAuthStatusEnvelope

    @POST
    suspend fun approveTvAuthSession(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<Map<String, Boolean>>

    @POST
    suspend fun denyTvAuthSession(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<Map<String, Boolean>>

    @GET
    suspend fun imageCollections(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("q") keyword: String? = null,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<ImageCollectionsPayload>

    @GET
    suspend fun imageCollectionDetail(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<ImageCollectionDetailDto>

    @GET
    suspend fun search(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("q") keyword: String,
        @Query("type") type: String,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<SearchPayload>

    @GET
    suspend fun detail(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<VideoDetailDto>

    @POST
    suspend fun toggleLike(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<ActionTogglePayload>

    @POST
    suspend fun toggleFavorite(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<ActionTogglePayload>

    @POST
    suspend fun toggleDislike(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<ActionTogglePayload>

    @POST
    suspend fun recordHistory(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Body body: RecordHistoryRequest,
    ): ApiEnvelope<Map<String, Boolean>>

    @GET
    suspend fun continueHistory(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("page") page: Int,
        @Query("limit") limit: Int,
    ): ApiEnvelope<ContinueHistoryPayload>

    @GET
    suspend fun likedVideos(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<SearchPayload>

    @GET
    suspend fun favoritedVideos(
        @Url url: String,
        @Header("Authorization") authorization: String,
        @Query("page") page: Int,
        @Query("page_size") pageSize: Int,
    ): ApiEnvelope<SearchPayload>

    @GET
    suspend fun userProfile(
        @Url url: String,
        @Header("Authorization") authorization: String,
    ): ApiEnvelope<UserProfileDto>
}
