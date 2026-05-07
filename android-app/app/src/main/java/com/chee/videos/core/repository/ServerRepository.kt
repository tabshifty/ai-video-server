package com.chee.videos.core.repository

import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ServerEndpoint
import com.chee.videos.core.network.LocalNetworkScanner
import com.chee.videos.core.util.UrlBuilder
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow

@Singleton
class ServerRepository @Inject constructor(
    private val scanner: LocalNetworkScanner,
    private val store: AppPreferencesStore,
) {
    val endpointsFlow: Flow<List<ServerEndpoint>> = store.endpointsFlow
    val activeBaseUrlFlow: Flow<String?> = store.activeBaseUrlFlow

    suspend fun scanLocalNetwork(
        ports: List<Int> = listOf(8080, 80, 3000, 5000),
    ): List<ServerEndpoint> {
        val discovered = scanner.discoverCandidateBaseUrls(ports)
        val now = System.currentTimeMillis()
        return discovered.map { ServerEndpoint(baseUrl = it, lastSuccessAt = now) }
    }

    suspend fun testEndpoint(rawInput: String): Boolean {
        val baseUrl = UrlBuilder.normalizeBaseUrl(rawInput)
        if (baseUrl.isBlank()) {
            return false
        }
        return scanner.probe(baseUrl)
    }

    suspend fun activateEndpoint(baseUrl: String, clearTokens: Boolean = false) {
        val endpoint = store.upsertEndpoint(baseUrl)
        if (clearTokens) {
            store.clearTokens()
        }
        store.setActiveBaseUrl(endpoint.baseUrl)
    }

    suspend fun removeEndpoint(baseUrl: String) {
        store.removeEndpoint(baseUrl)
    }
}
