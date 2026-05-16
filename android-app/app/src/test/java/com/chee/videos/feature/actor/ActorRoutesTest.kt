package com.chee.videos.feature.actor

import org.junit.Assert.assertEquals
import org.junit.Test

class ActorRoutesTest {
    @Test
    fun buildActorRoute_encodesActorId() {
        assertEquals(
            "actor/a%2Fb%20c",
            buildActorRoute("a/b c"),
        )
    }

    @Test
    fun buildVideoDetailRoute_encodesVideoType() {
        assertEquals(
            "detail/video-1?type=movie%2Fav",
            buildVideoDetailRoute(videoId = "video-1", videoType = "movie/av"),
        )
    }
}
