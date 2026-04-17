package com.chee.videos.core.repository

import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.AppException
import com.chee.videos.core.model.LoginRequest
import com.chee.videos.core.model.SessionTokens
import com.chee.videos.core.network.ApiService
import com.chee.videos.core.util.UrlBuilder
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class AuthRepository @Inject constructor(
    private val api: ApiService,
    private val store: AppPreferencesStore,
) {
    suspend fun login(identity: String, password: String): Result<Unit> {
        val baseUrl = store.readActiveBaseUrl()
            ?: return Result.failure(AppException("请先配置服务器地址"))
        if (identity.isBlank() || password.isBlank()) {
            return Result.failure(AppException("请输入账号和密码"))
        }

        return runCatching {
            val resp = api.login(UrlBuilder.login(baseUrl), LoginRequest(identity.trim(), password))
            if (resp.code != 0 || resp.data == null) {
                throw AppException(resp.msg.ifBlank { "登录失败(code=${resp.code})" })
            }
            store.saveTokens(
                SessionTokens(
                    accessToken = resp.data.accessToken,
                    refreshToken = resp.data.refreshToken,
                )
            )
            Unit
        }
    }

    suspend fun logoutLocal() {
        store.clearTokens()
    }

    suspend fun refreshTokenIfPossible(): Boolean {
        val baseUrl = store.readActiveBaseUrl() ?: return false
        val refreshToken = store.readRefreshToken() ?: return false

        return try {
            val resp = api.refresh(UrlBuilder.refresh(baseUrl), "Bearer $refreshToken")
            if (resp.code != 0 || resp.data == null) {
                store.clearTokens()
                false
            } else {
                store.saveTokens(
                    SessionTokens(
                        accessToken = resp.data.accessToken,
                        refreshToken = resp.data.refreshToken,
                    )
                )
                true
            }
        } catch (_: Exception) {
            store.clearTokens()
            false
        }
    }
}
