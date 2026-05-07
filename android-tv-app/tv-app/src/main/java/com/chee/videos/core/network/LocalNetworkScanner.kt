package com.chee.videos.core.network

import android.content.Context
import android.net.ConnectivityManager
import dagger.hilt.android.qualifiers.ApplicationContext
import java.net.Inet4Address
import java.util.Collections
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.async
import kotlinx.coroutines.awaitAll
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.sync.Semaphore
import kotlinx.coroutines.sync.withPermit
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request

@Singleton
class LocalNetworkScanner @Inject constructor(
    @ApplicationContext private val context: Context,
    private val client: OkHttpClient,
) {
    private val probeClient: OkHttpClient = client.newBuilder()
        .connectTimeout(java.time.Duration.ofMillis(900))
        .readTimeout(java.time.Duration.ofMillis(900))
        .writeTimeout(java.time.Duration.ofMillis(900))
        .build()

    suspend fun discoverCandidateBaseUrls(
        ports: List<Int>,
        maxHost: Int = 254,
    ): List<String> = withContext(Dispatchers.IO) {
        val subnet = detectSubnetPrefix() ?: return@withContext emptyList()
        val found = Collections.synchronizedSet(mutableSetOf<String>())
        val semaphore = Semaphore(48)

        coroutineScope {
            (1..maxHost).map { host ->
                async {
                    val ip = "$subnet.$host"
                    for (port in ports) {
                        val baseUrl = "http://$ip:$port"
                        semaphore.withPermit {
                            if (probe(baseUrl)) {
                                found.add(baseUrl)
                                return@async
                            }
                        }
                    }
                }
            }.awaitAll()
        }

        return@withContext found.toList().sorted()
    }

    suspend fun probe(baseUrl: String): Boolean = withContext(Dispatchers.IO) {
        val req = Request.Builder()
            .url("${baseUrl.trimEnd('/')}/healthz")
            .get()
            .build()
        return@withContext try {
            probeClient.newCall(req).execute().use { resp ->
                if (!resp.isSuccessful) {
                    return@use false
                }
                val body = resp.body?.string().orEmpty()
                body.contains("ok", ignoreCase = true)
            }
        } catch (_: Exception) {
            false
        }
    }

    private fun detectSubnetPrefix(): String? {
        val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val activeNetwork = cm.activeNetwork ?: return null
        val linkProperties = cm.getLinkProperties(activeNetwork) ?: return null
        val ip = linkProperties.linkAddresses
            .mapNotNull { it.address }
            .firstOrNull { it is Inet4Address && !it.isLoopbackAddress }
            ?.hostAddress
            ?: return null

        val parts = ip.split('.')
        if (parts.size != 4) {
            return null
        }
        return parts.take(3).joinToString(".")
    }
}
