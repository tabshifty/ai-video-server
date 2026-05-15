package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Test

class AppNavigationTransitionSpecTest {
    @Test
    fun rootTabTransitionDirection_followsBottomTabOrder() {
        assertEquals(
            AppNavigationTransitionDirection.Forward,
            appNavigationTransitionDirection("home", "search", isPop = false),
        )
        assertEquals(
            AppNavigationTransitionDirection.Forward,
            appNavigationTransitionDirection("search", "image-collections", isPop = false),
        )
        assertEquals(
            AppNavigationTransitionDirection.Backward,
            appNavigationTransitionDirection("mine", "search", isPop = false),
        )
    }

    @Test
    fun pushTransitionDirection_entersFromRightForNestedPages() {
        assertEquals(
            AppNavigationTransitionDirection.Forward,
            appNavigationTransitionDirection("home", "detail/{videoId}?type={videoType}", isPop = false),
        )
        assertEquals(
            AppNavigationTransitionDirection.Forward,
            appNavigationTransitionDirection("mine", "player/{source}/{videoId}", isPop = false),
        )
        assertEquals(
            AppNavigationTransitionDirection.Forward,
            appNavigationTransitionDirection("image-collections", "image-collections/{collectionId}", isPop = false),
        )
    }

    @Test
    fun popTransitionDirection_returnsTowardLeftSourcePage() {
        assertEquals(
            AppNavigationTransitionDirection.Backward,
            appNavigationTransitionDirection("detail/{videoId}?type={videoType}", "home", isPop = true),
        )
        assertEquals(
            AppNavigationTransitionDirection.Backward,
            appNavigationTransitionDirection("image-collections/{collectionId}", "image-collections", isPop = true),
        )
    }

    @Test
    fun unknownRoutes_defaultToForwardPushDirection() {
        assertEquals(
            AppNavigationTransitionDirection.Forward,
            appNavigationTransitionDirection("home", "unclassified/{id}", isPop = false),
        )
    }

    @Test
    fun transitionSpec_usesShortIosLikeMotion() {
        val spec = appNavigationTransitionSpec()

        assertEquals(260, spec.durationMillis)
        assertEquals(0.18f, spec.fadeStartAlpha, 0.0001f)
    }
}
