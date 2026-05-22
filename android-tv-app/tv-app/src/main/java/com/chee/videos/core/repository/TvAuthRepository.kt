package com.chee.videos.core.repository

import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.AppException
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.SessionTokens
import com.chee.videos.core.model.TvAuthSessionCreatePayload
import com.chee.videos.core.model.TvAuthSessionCreateRequest
import com.chee.videos.core.model.TvAuthSessionStatusPayload
import com.chee.videos.core.network.ApiService
import com.chee.videos.core.util.UrlBuilder
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.CancellationException

@Singleton
class TvAuthRepository @Inject constructor(
    private val api: ApiService,
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
) {
    suspend fun createSession(deviceId: String, deviceName: String): Result<TvAuthSessionCreatePayload> {
        val baseUrl = store.readActiveBaseUrl()
            ?: return Result.failure(AppException("请先配置服务器地址"))
        val normalizedDeviceId = deviceId.trim()
        val normalizedDeviceName = deviceName.trim()
        if (normalizedDeviceId.isBlank() || normalizedDeviceName.isBlank()) {
            return Result.failure(AppException("缺少 TV 设备信息"))
        }
        return try {
            val resp = api.createTvAuthSession(
                url = UrlBuilder.tvAuthSessions(baseUrl),
                body = TvAuthSessionCreateRequest(
                    deviceId = normalizedDeviceId,
                    deviceName = normalizedDeviceName,
                ),
            )
            Result.success(requireEnvelope(resp))
        } catch (e: CancellationException) {
            throw e
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun approve(sessionId: String): Result<Unit> {
        return callWithAuth(sessionId) { baseUrl, bearer, normalizedSessionId ->
            api.approveTvAuthSession(UrlBuilder.tvAuthApprove(baseUrl, normalizedSessionId), bearer)
        }
    }

    suspend fun deny(sessionId: String): Result<Unit> {
        return callWithAuth(sessionId) { baseUrl, bearer, normalizedSessionId ->
            api.denyTvAuthSession(UrlBuilder.tvAuthDeny(baseUrl, normalizedSessionId), bearer)
        }
    }

    suspend fun fetchSession(sessionId: String, explicitBaseUrl: String? = null): Result<TvAuthSessionStatusPayload> {
        val baseUrl = explicitBaseUrl?.let(UrlBuilder::normalizeBaseUrl)
            ?.takeIf { it.isNotBlank() }
            ?: store.readActiveBaseUrl()
            ?: return Result.failure(AppException("请先配置服务器地址"))
        val normalizedSessionId = sessionId.trim()
        if (normalizedSessionId.isBlank()) {
            return Result.failure(AppException("缺少配对会话"))
        }
        return runCatching {
            val resp = api.getTvAuthSession(UrlBuilder.tvAuthSession(baseUrl, normalizedSessionId))
            requireEnvelope(resp)
        }
    }

    suspend fun saveApprovedSession(payload: TvAuthSessionStatusPayload) {
        val accessToken = payload.accessToken?.trim().orEmpty()
        val refreshToken = payload.refreshToken?.trim().orEmpty()
        if (accessToken.isBlank() || refreshToken.isBlank()) {
            return
        }
        store.saveTokens(
            SessionTokens(
                accessToken = accessToken,
                refreshToken = refreshToken,
            ),
        )
    }

    private suspend fun callWithAuth(
        sessionId: String,
        block: suspend (baseUrl: String, authorization: String, sessionId: String) -> ApiEnvelope<Map<String, Boolean>>,
    ): Result<Unit> {
        val baseUrl = store.readActiveBaseUrl()
            ?: return Result.failure(AppException("请先配置服务器地址"))
        val normalizedSessionId = sessionId.trim()
        if (normalizedSessionId.isBlank()) {
            return Result.failure(AppException("缺少配对会话"))
        }

        var accessToken = store.readAccessToken()
            ?: return Result.failure(AuthExpiredException())

        var resp = runCatching {
            block(baseUrl, "Bearer $accessToken", normalizedSessionId)
        }.getOrElse { return Result.failure(it) }

        if (resp.code == 401) {
            val refreshed = authRepository.refreshTokenIfPossible()
            if (!refreshed) {
                return Result.failure(AuthExpiredException())
            }
            accessToken = store.readAccessToken()
                ?: return Result.failure(AuthExpiredException())
            resp = runCatching {
                block(baseUrl, "Bearer $accessToken", normalizedSessionId)
            }.getOrElse { return Result.failure(it) }
        }

        return runCatching {
            requireEnvelope(resp)
            Unit
        }
    }

    private fun <T> requireEnvelope(resp: ApiEnvelope<T>): T {
        val data = resp.data
        if (resp.code != 0 || data == null) {
            throw AppException(resp.msg.ifBlank { "TV 授权请求失败(code=${resp.code})" })
        }
        return data
    }
}
