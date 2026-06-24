package com.chee.videos.feature.tv

import com.chee.videos.core.model.FeedVideoDto
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class TvShortFeedViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun load_readsSeekStepAndStartsFromFirstItem() = runTest {
        val repository = FakeTvRepository(
            tvSeekStepSeconds = 20,
            shortFeedPages = listOf(
                listOf(
                    shortFeedItem("short-1"),
                    shortFeedItem("short-2"),
                ),
            ),
        )
        val viewModel = TvShortFeedViewModel(repository)

        viewModel.load()
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals(20, state.seekStepSeconds)
        assertEquals(listOf("short-1", "short-2"), state.items.map { it.id })
        assertEquals(0, state.currentIndex)
        assertTrue(repository.shortFeedRequests.first().excludeIds.isEmpty())
    }

    @Test
    fun ensureMoreLoaded_excludesSeenIdsAndAppendsNewBatch() = runTest {
        val repository = FakeTvRepository(
            shortFeedPages = listOf(
                (1..6).map { shortFeedItem("short-$it") },
                listOf(shortFeedItem("short-7"), shortFeedItem("short-8")),
            ),
        )
        val viewModel = TvShortFeedViewModel(repository)

        viewModel.load()
        advanceUntilIdle()
        repeat(4) { viewModel.moveNext() }
        viewModel.ensureMoreLoaded(currentIndex = 4)
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals((1..8).map { "short-$it" }, state.items.map { it.id })
        assertEquals((1..6).map { "short-$it" }, repository.shortFeedRequests[1].excludeIds)
    }

    @Test
    fun moveNextAndPrevious_doNotWrapAtFeedEdges() = runTest {
        val repository = FakeTvRepository(
            shortFeedPages = listOf(
                listOf(shortFeedItem("short-1"), shortFeedItem("short-2")),
            ),
        )
        val viewModel = TvShortFeedViewModel(repository)

        viewModel.load()
        advanceUntilIdle()

        assertFalse(viewModel.movePrevious())
        assertTrue(viewModel.moveNext())
        assertFalse(viewModel.moveNext())
        assertEquals(1, viewModel.uiState.value.currentIndex)
    }

    @Test
    fun ensureMoreLoaded_acceptsDuplicateFallbackWhenPoolIsTooSmall() = runTest {
        val repository = FakeTvRepository(
            shortFeedPages = listOf(
                listOf(shortFeedItem("short-1"), shortFeedItem("short-2")),
                listOf(shortFeedItem("short-1")),
            ),
        )
        val viewModel = TvShortFeedViewModel(repository)

        viewModel.load()
        advanceUntilIdle()
        viewModel.moveNext()
        viewModel.ensureMoreLoaded(currentIndex = 1)
        advanceUntilIdle()

        assertEquals(listOf("short-1", "short-2", "short-1"), viewModel.uiState.value.items.map { it.id })
    }

    private fun shortFeedItem(id: String): FeedVideoDto = FeedVideoDto(
        id = id,
        title = "标题 $id",
        type = "short",
        thumbnailPath = "/thumb/$id.jpg",
        durationSeconds = 30,
    )
}
