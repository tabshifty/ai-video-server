package com.chee.videos.core.repository

import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ServerEndpoint
import com.chee.videos.core.network.LocalNetworkScanner
import com.chee.videos.core.util.UrlBuilder
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow

interface ConnectionServerRepository {
    val endpointsFlow: Flow<List<ServerEndpoint>>

    suspend fun scanLocalNetwork(ports: List<Int>): List<ServerEndpoint>

    suspend fun testEndpoint(rawInput: String): Boolean

    suspend fun activateEndpoint(baseUrl: String, clearTokens: Boolean)

    suspend fun removeEndpoint(baseUrl: String)
}

@Singleton
class ServerRepository @Inject constructor(
    private val scanner: LocalNetworkScanner,
    private val store: AppPreferencesStore,
) : ConnectionServerRepository {
    override val endpointsFlow: Flow<List<ServerEndpoint>> = store.endpointsFlow
    val activeBaseUrlFlow: Flow<String?> = store.activeBaseUrlFlow

    override suspend fun scanLocalNetwork(ports: List<Int>): List<ServerEndpoint> {
        val discovered = scanner.discoverCandidateBaseUrls(ports)
        val now = System.currentTimeMillis()
        return discovered.map { ServerEndpoint(baseUrl = it, lastSuccessAt = now) }
    }

    override suspend fun testEndpoint(rawInput: String): Boolean {
        val baseUrl = UrlBuilder.normalizeBaseUrl(rawInput)
        if (baseUrl.isBlank()) {
            return false
        }
        return scanner.probe(baseUrl)
    }

    override suspend fun activateEndpoint(baseUrl: String, clearTokens: Boolean) {
        val endpoint = store.upsertEndpoint(baseUrl)
        if (clearTokens) {
            store.clearTokens()
        }
        store.setActiveBaseUrl(endpoint.baseUrl)
    }

    override suspend fun removeEndpoint(baseUrl: String) {
        store.removeEndpoint(baseUrl)
    }
}
