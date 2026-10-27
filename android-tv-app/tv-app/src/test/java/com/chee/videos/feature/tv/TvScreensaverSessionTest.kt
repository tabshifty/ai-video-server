package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvScreensaverImage
import com.chee.videos.core.model.TvScreensaverSettings
import kotlin.random.Random
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.NonCancellable
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.test.advanceTimeBy
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.withContext
import org.junit.Assert.*
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class TvScreensaverSessionTest {
    private fun image(id: String) = TvScreensaverImage(id, "/api/v1/images/$id/view", "revision-1")
    private fun ready() = TvScreensaverRuntime(sessionKey = "服务器甲/账号甲", foreground = true)

    @Test
    fun rapidBackgroundOrPlaybackRoundTripStillRestartsWait() = runTest {
        val controller = TvScreensaverController()
        controller.setSession("账号甲")
        controller.setForeground(true)
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(fetchImages = { listOf(image("a")) }, loadImage = { it.id }, onFrame = frames::add)
                .run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), controller.runtime)
        }
        runCurrent()
        advanceTimeBy(59_000)
        controller.setForeground(false)
        controller.setForeground(true)
        runCurrent()
        advanceTimeBy(1_000)
        runCurrent()
        assertNull(frames.last())
        advanceTimeBy(58_000)
        val player = Any()
        controller.setPlaybackBusy(player, true)
        controller.setPlaybackBusy(player, false)
        runCurrent()
        advanceTimeBy(1_000)
        runCurrent()
        assertNull(frames.last())
        advanceTimeBy(59_000)
        runCurrent()
        assertNotNull(frames.last())
    }

    @Test
    fun wakeBeforeCollectorProcessesInputCannotReopenOldFrame() = runTest {
        val controller = TvScreensaverController()
        controller.setSession("账号甲")
        controller.setForeground(true)
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = { listOf(image("a")) },
                loadImage = { controller.userActivity(); it.id }, onFrame = frames::add,
            ).run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), controller.runtime)
        }
        runCurrent()
        advanceTimeBy(60_000)
        runCurrent()
        assertFalse(frames.any { it != null })
    }

    @Test
    fun defaultsWaitTenMinutesAndOnlyShowAfterFirstDecode() = runTest {
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        var fetches = 0
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = { fetches++; listOf(image("a")) },
                loadImage = { delay(2_000); it.id }, onFrame = frames::add,
            ).run(MutableStateFlow(TvScreensaverSettings()), MutableStateFlow(ready()))
        }
        runCurrent()
        advanceTimeBy(599_999)
        runCurrent()
        assertEquals(0, fetches)
        advanceTimeBy(1)
        runCurrent()
        assertFalse(frames.any { it != null })
        advanceTimeBy(2_000)
        runCurrent()
        assertEquals("a", frames.last()?.content)
    }

    @Test
    fun activityPlaybackBackgroundAndSessionChangesRestartFullWait() = runTest {
        val runtime = MutableStateFlow(ready())
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(fetchImages = { listOf(image("a")) }, loadImage = { it.id }, onFrame = frames::add)
                .run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), runtime)
        }
        runCurrent()
        advanceTimeBy(59_000)
        runtime.value = runtime.value.copy(interactionSequence = 1)
        runCurrent()
        advanceTimeBy(59_000)
        runCurrent()
        assertNull(frames.lastOrNull())
        advanceTimeBy(1_000)
        runCurrent()
        assertNotNull(frames.last())

        for (blocked in listOf(
            ready().copy(playbackBusy = true),
            ready().copy(foreground = false),
            ready().copy(sessionKey = null),
        )) {
            runtime.value = blocked
            runCurrent()
            assertNull(frames.last())
            val displayed = frames.count { it != null }
            advanceTimeBy(600_000)
            runCurrent()
            assertEquals(displayed, frames.count { it != null })
            runtime.value = ready()
            runCurrent()
            advanceTimeBy(59_999)
            runCurrent()
            assertNull(frames.last())
            advanceTimeBy(1)
            runCurrent()
            assertNotNull(frames.last())
        }
    }

    @Test
    fun preloadsNextButWaitsFullDwellAndSlowLoadsDoNotShortenNextSlide() = runTest {
        val shown = mutableListOf<Pair<String, Long>>()
        val loaded = mutableListOf<String>()
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = { listOf(image("a"), image("b")) },
                loadImage = {
                    loaded += it.id
                    delay(if (loaded.size == 1) 2_000 else 8_000)
                    it.id
                },
                onFrame = { if (it != null) shown += it.content to testScheduler.currentTime },
                random = Random(4),
            ).run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), MutableStateFlow(ready()))
        }
        runCurrent()
        advanceTimeBy(62_000)
        runCurrent()
        assertEquals(1, shown.size)
        assertEquals(2, loaded.size)
        advanceTimeBy(5_000)
        runCurrent()
        assertEquals(1, shown.size)
        advanceTimeBy(3_000)
        runCurrent()
        assertEquals(2, shown.size)
        assertNotEquals(shown[0].first, shown[1].first)
        assertEquals(8_000, shown[1].second - shown[0].second)
        advanceTimeBy(4_999)
        runCurrent()
        assertEquals(2, shown.size)
    }

    @Test
    fun eachRoundHasNoDuplicatesAndDoesNotRepeatPreviousLast() {
        val images = listOf(image("a"), image("b"), image("a"), image("c"))
        for (seed in 0..30) {
            val deck = shuffleScreensaverImages(images, "b", Random(seed))
            assertEquals(setOf("a", "b", "c"), deck.map { it.id }.toSet())
            assertEquals(3, deck.size)
            assertNotEquals("b", deck.first().id)
        }
    }

    @Test
    fun oneImageStaysVisibleWithoutReloadAndStillChecksForNewImages() = runTest {
        var fetches = 0
        var loads = 0
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = { fetches++; listOf(image("a")) },
                loadImage = { loads++; it.id }, onFrame = frames::add,
            ).run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), MutableStateFlow(ready()))
        }
        runCurrent()
        advanceTimeBy(75_000)
        runCurrent()
        assertTrue(fetches >= 3)
        assertEquals(1, loads)
        assertTrue(frames.filterNotNull().all { it.content == "a" })
        assertNotNull(frames.last())
    }

    @Test
    fun emptyOrAllBrokenImagesUseBackoffAndNeverCoverOriginalPage() = runTest {
        var fetches = 0
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession<String>(
                fetchImages = { fetches++; if (fetches == 1) emptyList() else listOf(image("bad")) },
                loadImage = { null }, onFrame = frames::add,
            ).run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), MutableStateFlow(ready()))
        }
        runCurrent()
        advanceTimeBy(60_000)
        runCurrent()
        assertEquals(1, fetches)
        advanceTimeBy(59_999)
        runCurrent()
        assertEquals(1, fetches)
        advanceTimeBy(1)
        runCurrent()
        assertEquals(2, fetches)
        assertTrue(frames.all { it == null })
    }

    @Test
    fun brokenNextImageAndFailedRefreshKeepCurrentUsableImage() = runTest {
        var fetches = 0
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = {
                    fetches++
                    if (fetches > 1) error("离线")
                    listOf(image("good"), image("bad"))
                },
                loadImage = { it.id.takeIf { id -> id == "good" } }, onFrame = frames::add,
            ).run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), MutableStateFlow(ready()))
        }
        runCurrent()
        advanceTimeBy(80_000)
        runCurrent()
        assertTrue(frames.filterNotNull().isNotEmpty())
        assertTrue(frames.filterNotNull().all { it.content == "good" })
        assertNotNull(frames.last())
    }

    @Test
    fun successfulEmptyRefreshRemovesLastImageAndExits() = runTest {
        var fetches = 0
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = { fetches++; if (fetches == 1) listOf(image("a")) else emptyList() },
                loadImage = { it.id }, onFrame = frames::add,
            ).run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), MutableStateFlow(ready()))
        }
        runCurrent()
        advanceTimeBy(60_000)
        runCurrent()
        assertNotNull(frames.last())
        advanceTimeBy(5_000)
        runCurrent()
        assertNull(frames.last())
    }

    @Test
    fun cancelledOldImageCannotAppearAfterChangingSessionEvenIfLoaderIgnoresCancellation() = runTest {
        val runtime = MutableStateFlow(ready())
        val release = CompletableDeferred<Unit>()
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = { listOf(image("old")) },
                loadImage = { withContext(NonCancellable) { release.await(); "旧会话图片" } },
                onFrame = frames::add,
            ).run(MutableStateFlow(TvScreensaverSettings(idleMinutes = 1)), runtime)
        }
        runCurrent()
        advanceTimeBy(60_000)
        runCurrent()
        runtime.value = ready().copy(sessionKey = "服务器乙/账号乙")
        runCurrent()
        release.complete(Unit)
        runCurrent()
        assertFalse(frames.any { it?.content == "旧会话图片" })
    }

    @Test
    fun disabledNeverLoadsAndEnablingOrChangingDurationStartsNewWait() = runTest {
        var fetches = 0
        val config = MutableStateFlow(TvScreensaverSettings(enabled = false, idleMinutes = 1))
        val frames = mutableListOf<TvScreensaverFrame<String>?>()
        backgroundScope.launch {
            TvScreensaverSession(
                fetchImages = { fetches++; listOf(image("a")) }, loadImage = { it.id }, onFrame = frames::add,
            ).run(config, MutableStateFlow(ready()))
        }
        runCurrent()
        advanceTimeBy(600_000)
        runCurrent()
        assertEquals(0, fetches)
        config.value = config.value.copy(enabled = true)
        runCurrent()
        advanceTimeBy(59_000)
        config.value = config.value.copy(idleMinutes = 3)
        runCurrent()
        advanceTimeBy(179_999)
        runCurrent()
        assertNull(frames.last())
        advanceTimeBy(1)
        runCurrent()
        assertNotNull(frames.last())
        config.value = config.value.copy(enabled = false)
        runCurrent()
        assertNull(frames.last())
    }
}
