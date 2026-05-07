package com.chee.videos.feature.shorts

import com.chee.videos.core.model.VideoImageCollectionDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class ShortFeedImageCollectionTest {

    @Test
    fun `shows short detail image collection card when related image collection exists`() {
        assertTrue(
            shouldShowShortDetailImageCollection(
                VideoImageCollectionDto(
                    id = "collection-1",
                    name = "剧照合集",
                ),
            ),
        )
    }

    @Test
    fun `hides short detail image collection card when related image collection missing`() {
        assertFalse(shouldShowShortDetailImageCollection(null))
        assertFalse(
            shouldShowShortDetailImageCollection(
                VideoImageCollectionDto(
                    id = "   ",
                    name = "剧照合集",
                ),
            ),
        )
    }

    @Test
    fun `builds viewer route for related image collection`() {
        assertEquals(
            "image-collections/collection-1",
            shortDetailImageCollectionViewerRoute(
                VideoImageCollectionDto(
                    id = "collection-1",
                    name = "剧照合集",
                ),
            ),
        )
        assertNull(shortDetailImageCollectionViewerRoute(null))
    }
}
