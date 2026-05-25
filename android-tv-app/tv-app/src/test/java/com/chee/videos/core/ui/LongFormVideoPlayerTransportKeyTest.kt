package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class LongFormVideoPlayerTransportKeyTest {
    @Test
    fun pendingSeekStartsFromCurrentPositionAndDefersCommit() {
        assertEquals(
            TvPendingStepSeekUpdate(
                anchorPositionMs = 10_000L,
                targetPositionMs = 20_000L,
                accumulatedDeltaMs = 10_000L,
                shouldCommitImmediately = false,
            ),
            resolveTvPendingStepSeek(
                previous = null,
                currentPositionMs = 10_000L,
                durationMs = 120_000L,
                deltaMs = 10_000L,
            ),
        )
    }

    @Test
    fun pendingSeekAccumulatesRepeatedPressesBeforeCommit() {
        val first = resolveTvPendingStepSeek(
            previous = null,
            currentPositionMs = 10_000L,
            durationMs = 120_000L,
            deltaMs = 10_000L,
        )

        assertEquals(
            TvPendingStepSeekUpdate(
                anchorPositionMs = 10_000L,
                targetPositionMs = 30_000L,
                accumulatedDeltaMs = 20_000L,
                shouldCommitImmediately = false,
            ),
            resolveTvPendingStepSeek(
                previous = first,
                currentPositionMs = 10_000L,
                durationMs = 120_000L,
                deltaMs = 10_000L,
            ),
        )
    }

    @Test
    fun pendingSeekDirectionChangeContinuesFromAccumulatedTarget() {
        val pending = TvPendingStepSeekUpdate(
            anchorPositionMs = 10_000L,
            targetPositionMs = 30_000L,
            accumulatedDeltaMs = 20_000L,
            shouldCommitImmediately = false,
        )

        assertEquals(
            TvPendingStepSeekUpdate(
                anchorPositionMs = 10_000L,
                targetPositionMs = 20_000L,
                accumulatedDeltaMs = 10_000L,
                shouldCommitImmediately = false,
            ),
            resolveTvPendingStepSeek(
                previous = pending,
                currentPositionMs = 10_000L,
                durationMs = 120_000L,
                deltaMs = -10_000L,
            ),
        )
    }

    @Test
    fun pendingSeekClampsTargetToDuration() {
        assertEquals(
            TvPendingStepSeekUpdate(
                anchorPositionMs = 95_000L,
                targetPositionMs = 100_000L,
                accumulatedDeltaMs = 5_000L,
                shouldCommitImmediately = false,
            ),
            resolveTvPendingStepSeek(
                previous = null,
                currentPositionMs = 95_000L,
                durationMs = 100_000L,
                deltaMs = 10_000L,
            ),
        )
    }

    @Test
    fun playerUpdatesSeekPreviewImmediatelyButCommitsSeekAfterDebounce() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        assertTrue("长视频播放器必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        val debouncedSeekSource = source
            .substringAfter("fun performDebouncedStepSeek(")
            .substringBefore("fun handleTvRemoteKeyAction(")

        assertTrue(debouncedSeekSource.contains("updateSeekPreview("))
        assertTrue(debouncedSeekSource.contains("delay(TvStepSeekDebounceMillis)"))
        assertTrue(debouncedSeekSource.contains("player.time = pending.targetPositionMs"))
    }

    @Test
    fun tvControlBarBackButtonsRequestPlaybackExitConfirmation() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        assertTrue("长视频播放器必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("播放器应暴露请求退出播放回调", source.contains("onRequestExitPlayback: (() -> Unit)? = null"))
        assertTrue("TV 控制条返回详情应走退出确认", source.contains("onRequestExitPlayback?.invoke() ?: onBack()"))
        assertTrue("TV 控制条退出播放应走退出确认", source.contains("onRequestExitPlayback?.invoke() ?: exitPlayback()"))
    }
}
