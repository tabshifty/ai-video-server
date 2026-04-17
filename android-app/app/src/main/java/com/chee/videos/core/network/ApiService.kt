package com.chee.videos.core.network

import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.FeedPayload
import com.chee.videos.core.model.LoginPayload
import com.chee.videos.core.model.LoginRequest
import com.chee.videos.core.model.RecordHistoryRequest
import com.chee.videos.core.model.RefreshPayload
import com.chee.videos.core.model.SearchPayload
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
    ): ApiEnvelope<FeedPayload>

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
}
