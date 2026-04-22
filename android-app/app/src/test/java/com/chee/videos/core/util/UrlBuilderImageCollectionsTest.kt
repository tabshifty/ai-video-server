package com.chee.videos.core.util

import org.junit.Assert.assertEquals
import org.junit.Test

class UrlBuilderImageCollectionsTest {
    @Test
    fun `exposes image collection endpoints`() {
        val instance = UrlBuilder
        val baseUrl = "https://media.example.com/"

        val listMethod = instance::class.java.getDeclaredMethod("imageCollections", String::class.java)
        val detailMethod = instance::class.java.getDeclaredMethod("imageCollectionDetail", String::class.java, String::class.java)
        val imageViewMethod = instance::class.java.getDeclaredMethod("appImageView", String::class.java, String::class.java)

        assertEquals(
            "https://media.example.com/api/v1/image-collections",
            listMethod.invoke(instance, baseUrl),
        )
        assertEquals(
            "https://media.example.com/api/v1/image-collections/collection-1",
            detailMethod.invoke(instance, baseUrl, "collection-1"),
        )
        assertEquals(
            "https://media.example.com/api/v1/images/image-1/view",
            imageViewMethod.invoke(instance, baseUrl, "image-1"),
        )
    }
}
