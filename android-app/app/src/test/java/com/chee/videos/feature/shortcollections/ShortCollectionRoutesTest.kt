package com.chee.videos.feature.shortcollections

import org.junit.Assert.assertEquals
import org.junit.Test

class ShortCollectionRoutesTest {
    @Test
    fun buildShortCollectionContentRoute_encodesIdAndName() {
        assertEquals(
            "short-collections/a%2Fb?collectionName=%E7%83%AD%E9%97%A8%20%E5%90%88%E9%9B%86",
            buildShortCollectionContentRoute("a/b", "热门 合集"),
        )
    }
}
