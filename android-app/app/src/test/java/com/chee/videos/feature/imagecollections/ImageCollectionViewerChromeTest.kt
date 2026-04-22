package com.chee.videos.feature.imagecollections

import org.junit.Assert.assertEquals
import org.junit.Test

class ImageCollectionViewerChromeTest {
    @Test
    fun `double tap toggles immersive chrome visibility`() {
        val klass = Class.forName("com.chee.videos.feature.imagecollections.ImageCollectionViewerChromeKt")
        val toggleMethod = klass.getDeclaredMethod("toggleImageCollectionViewerChrome", Boolean::class.javaPrimitiveType)

        assertEquals(false, toggleMethod.invoke(null, true))
        assertEquals(true, toggleMethod.invoke(null, false))
    }
}
