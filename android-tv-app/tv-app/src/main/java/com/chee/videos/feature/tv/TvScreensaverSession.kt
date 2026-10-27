package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvScreensaverImage
import com.chee.videos.core.model.TvScreensaverSettings
import kotlin.random.Random
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.async
import kotlinx.coroutines.coroutineScope
import kotlinx.coroutines.currentCoroutineContext
import kotlinx.coroutines.delay
import kotlinx.coroutines.ensureActive
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.isActive
import kotlinx.coroutines.withTimeoutOrNull

data class TvScreensaverRuntime(
    val sessionKey: String? = null,
    val foreground: Boolean = false,
    val playbackBusy: Boolean = false,
    val interactionSequence: Long = 0L,
)

internal data class TvScreensaverFrame<out T>(val image: TvScreensaverImage, val content: T)

/** 与 Android 渲染解耦：任何输入或会话变化都会先取消旧加载，再从零等待。 */
internal class TvScreensaverSession<T>(
    private val fetchImages: suspend () -> List<TvScreensaverImage>,
    private val loadImage: suspend (TvScreensaverImage) -> T?,
    private val onFrame: (TvScreensaverFrame<T>?) -> Unit,
    private val random: Random = Random.Default,
) {
    suspend fun run(settings: Flow<TvScreensaverSettings>, runtime: StateFlow<TvScreensaverRuntime>) {
        combine(settings.distinctUntilChanged(), runtime) { config, state -> config to state }
            .collectLatest { (config, state) ->
                try {
                    if (!config.enabled || state.sessionKey == null || !state.foreground || state.playbackBusy) {
                        return@collectLatest
                    }
                    delay(config.idleMinutes * 60_000L)
                    while (currentCoroutineContext().isActive) {
                        play(config.slideSeconds * 1_000L) { frame ->
                            // 输入已同步唤醒但 collectLatest 尚未调度取消时，也不能重新覆盖原页面。
                            if (runtime.value == state) onFrame(frame)
                        }
                        onFrame(null)
                        // 空库与初次失败有界重试，不在每次重新组合或后台轮询时发起请求。
                        delay(60_000L)
                    }
                } finally {
                    onFrame(null)
                }
            }
    }

    private suspend fun play(dwellMillis: Long, display: (TvScreensaverFrame<T>?) -> Unit) = coroutineScope {
        var pool = emptyList<TvScreensaverImage>()
        var deck = ArrayDeque<TvScreensaverImage>()
        var current: TvScreensaverFrame<T>? = null

        suspend fun nextFrame(): TvScreensaverFrame<T>? {
            val next = withTimeoutOrNull(60_000L) {
                val attempted = mutableSetOf<String>()
                var refreshed = false
                while (currentCoroutineContext().isActive) {
                    if (deck.isEmpty()) {
                        if (refreshed) return@withTimeoutOrNull null
                        val refreshedPool = safely { fetchImages() }
                        if (refreshedPool != null) {
                            pool = refreshedPool.distinctBy { it.id }
                        }
                        deck = ArrayDeque(shuffleScreensaverImages(pool, current?.image?.id, random))
                        refreshed = true
                        if (deck.isEmpty()) return@withTimeoutOrNull null
                    }
                    val candidate = deck.removeFirst()
                    if (!attempted.add(candidate.id)) continue
                    val previous = current
                    if (previous != null && previous.image == candidate) return@withTimeoutOrNull previous
                    val loaded = safely { loadImage(candidate) } ?: continue
                    currentCoroutineContext().ensureActive()
                    return@withTimeoutOrNull TvScreensaverFrame(candidate, loaded)
                }
                null
            }
            // 超时并不证明当前图已失效；只要最新成功候选仍包含它，就保留已显示的图。
            return next ?: current?.takeIf { shown -> pool.any { it == shown.image } }
        }

        current = nextFrame() ?: return@coroutineScope
        display(current)
        while (currentCoroutineContext().isActive) {
            val pending = async { nextFrame() }
            delay(dwellMillis)
            val next = pending.await() ?: return@coroutineScope
            currentCoroutineContext().ensureActive()
            current = next
            display(next)
        }
    }

    private suspend fun <R> safely(block: suspend () -> R?): R? = try {
        withTimeoutOrNull(20_000L) { block() }
    } catch (err: CancellationException) {
        throw err
    } catch (_: Exception) {
        null
    }
}

internal fun shuffleScreensaverImages(
    images: List<TvScreensaverImage>,
    lastImageId: String?,
    random: Random,
): List<TvScreensaverImage> {
    val shuffled = images.distinctBy { it.id }.shuffled(random).toMutableList()
    if (shuffled.size > 1 && shuffled.first().id == lastImageId) {
        val first = shuffled[0]
        shuffled[0] = shuffled[1]
        shuffled[1] = first
    }
    return shuffled
}
