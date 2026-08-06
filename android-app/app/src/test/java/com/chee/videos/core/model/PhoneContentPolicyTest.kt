package com.chee.videos.core.model

import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class PhoneContentPolicyTest {
    @Test
    fun `phone content accepts only short and av video types`() {
        assertTrue(isPhoneSupportedVideoType("short"))
        assertTrue(isPhoneSupportedVideoType(" SHORT "))
        assertTrue(isPhoneSupportedVideoType("av"))
        assertTrue(isPhoneSupportedVideoType(" Av "))

        listOf("movie", "episode", "tv", "unknown", "", "   ").forEach { type ->
            assertFalse("不应支持类型 $type", isPhoneSupportedVideoType(type))
        }
        assertFalse(isPhoneSupportedVideoType(null))
    }

    @Test
    fun `phone content batch skips excluded source pages until supported items are found`() = runTest {
        val requestedPages = mutableListOf<Int>()
        val sourcePages = mapOf(
            1 to PhoneContentSourcePage(
                items = listOf(item("movie-1", "movie"), item("episode-1", "episode")),
                hasMore = true,
            ),
            2 to PhoneContentSourcePage(
                items = listOf(item("short-1", "short"), item("av-1", "av")),
                hasMore = true,
            ),
        )

        val batch = loadPhoneContentBatch(
            firstPage = 1,
            minimumVisibleItems = 2,
            typeOf = PhoneItem::type,
            fetchPage = { page ->
                requestedPages += page
                Result.success(checkNotNull(sourcePages[page]))
            },
        ).getOrThrow()

        assertEquals(listOf(1, 2), requestedPages)
        assertEquals(listOf("short-1", "av-1"), batch.items.map { it.id })
        assertEquals(2, batch.lastLoadedPage)
        assertTrue(batch.hasMore)
        assertNull(batch.trailingError)
    }

    @Test
    fun `phone content batch reaches ordinary empty state only after source exhaustion`() = runTest {
        val batch = loadPhoneContentBatch(
            firstPage = 3,
            minimumVisibleItems = 20,
            typeOf = PhoneItem::type,
            fetchPage = {
                Result.success(
                    PhoneContentSourcePage(
                        items = listOf(item("movie-3", "movie")),
                        hasMore = false,
                    ),
                )
            },
        ).getOrThrow()

        assertTrue(batch.items.isEmpty())
        assertEquals(3, batch.lastLoadedPage)
        assertFalse(batch.hasMore)
        assertNull(batch.trailingError)
    }

    @Test
    fun `phone content batch keeps partial supported items when a later page fails`() = runTest {
        val failure = IllegalStateException("下一页失败")

        val batch = loadPhoneContentBatch(
            firstPage = 1,
            minimumVisibleItems = 2,
            typeOf = PhoneItem::type,
            fetchPage = { page ->
                when (page) {
                    1 -> Result.success(
                        PhoneContentSourcePage(
                            items = listOf(item("short-1", "short")),
                            hasMore = true,
                        ),
                    )
                    else -> Result.failure(failure)
                }
            },
        ).getOrThrow()

        assertEquals(listOf("short-1"), batch.items.map { it.id })
        assertEquals(1, batch.lastLoadedPage)
        assertTrue(batch.hasMore)
        assertEquals(failure, batch.trailingError)
    }

    @Test
    fun `phone content batch fails when its first source page fails`() = runTest {
        val failure = IllegalStateException("首屏失败")

        val result = loadPhoneContentBatch(
            firstPage = 1,
            minimumVisibleItems = 20,
            typeOf = PhoneItem::type,
            fetchPage = { Result.failure(failure) },
        )

        assertTrue(result.isFailure)
        assertEquals(failure, result.exceptionOrNull())
    }

    @Test
    fun `player start index rejects a filtered or missing start item`() {
        val items = listOf(item("short-1", "short"), item("av-1", "av"))

        assertEquals(1, resolvePhoneContentStartIndex(items, "av-1", PhoneItem::id))
        assertNull(resolvePhoneContentStartIndex(items, "movie-1", PhoneItem::id))
        assertNull(resolvePhoneContentStartIndex(items, "", PhoneItem::id))
    }

    private fun item(id: String, type: String): PhoneItem = PhoneItem(id = id, type = type)

    private data class PhoneItem(
        val id: String,
        val type: String,
    )
}
